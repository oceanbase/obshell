/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package obcluster

import (
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/cavaliergopher/rpm"
	"gorm.io/gorm"

	obdriver "github.com/oceanbase/go-oceanbase-driver"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/ob/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/sqlite"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

func (obclusterService *ObclusterService) ExecuteSql(sql string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Exec(sql).Error
}

func (obclusterService *ObclusterService) ExecuteSqlWithoutIdentityCheck(sql string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Exec(sql).Error
}

func (obclusterService *ObclusterService) Bootstrap(sql string) (err error) {
	db, err := oceanbasedb.GetRestrictedInstance()
	if err != nil {
		return err
	}
	if err = obclusterService.setSessionObQueryTimeout(db, 1000000000); err != nil {
		return err
	}
	return db.Exec(sql).Error
}

// setSessionObQueryTimeout sets the session-level query timeout for an OceanBase database session.
// The `time` parameter specifies the timeout duration in microseconds(us).
func (obclusterService *ObclusterService) setSessionObQueryTimeout(db *gorm.DB, time int) (err error) {
	return db.Exec(fmt.Sprintf("SET SESSION ob_query_timeout=%d", time)).Error
}

func (obclusterService *ObclusterService) StartZone(zone string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM START ZONE '%s'", zone)
	err = db.Exec(sql).Error
	return
}

// StopZone attempts to stop a given zone. It is safe to call this function
// even if the zone is already stopped (INACTIVE).
func (obclusterService *ObclusterService) StopZone(zone string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER SYSTEM STOP ZONE '%s'", zone)
	err = db.Exec(sql).Error
	return
}

func (obclusterService *ObclusterService) GetUTCTime() (t time.Time, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return t, err
	}
	err = db.Raw("SELECT UTC_TIMESTAMP(6)").Scan(&t).Error
	return
}

func (*ObclusterService) GetCurrentTimestamp() (t time.Time, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return t, err
	}
	err = db.Raw("SELECT CURRENT_TIMESTAMP(6)").Scan(&t).Error
	return
}

func (ObclusterService *ObclusterService) GetServerCheckpointScn(servers []oceanbase.OBServer) (map[oceanbase.OBServer]uint64, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var result = make(map[oceanbase.OBServer]uint64)
	for _, server := range servers {
		var checkpointScn uint64
		err = db.Raw("SELECT checkpoint_scn FROM oceanbase.__all_virtual_ls_info WHERE svr_ip = ? AND svr_port = ? ORDER BY checkpoint_scn ASC LIMIT 1", server.SvrIp, server.SvrPort).Scan(&checkpointScn).Error
		if err != nil {
			return nil, err
		}
		result[server] = checkpointScn
	}
	return result, nil
}

// MinorFreeze is only for sys tenant, and only support one server.
func (obclusterService *ObclusterService) MinorFreeze(servers []oceanbase.OBServer) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	var targetCmd []string
	for _, server := range servers {
		targetCmd = append(targetCmd, meta.NewAgentInfo(server.SvrIp, server.SvrPort).String())
	}
	serverList := strings.Join(targetCmd, "','")
	sql := fmt.Sprintf("alter system minor freeze server = ('%[1]s');", serverList)

	return db.Exec(sql).Error
}

// IsLsCheckpointAfterTs will check the smallest checkpoint of the log stream of on the target server.
func (obclusterService *ObclusterService) IsLsCheckpointAfterTs(server oceanbase.OBServer) (checkpintScn uint64, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return checkpintScn, err
	}

	var systemTimeZone string
	if err = db.Raw("SELECT @@system_time_zone").Scan(&systemTimeZone).Error; err != nil {
		return
	}

	// Get checkpoint of target server.
	sql := fmt.Sprintf("select checkpoint_scn from oceanbase.__all_virtual_ls_info where svr_ip = '%s' and svr_port = %d order by checkpoint_scn asc limit 1", server.SvrIp, server.SvrPort)
	err = db.Raw(sql).Scan(&checkpintScn).Error
	return

}

func (obclusterService *ObclusterService) AddZoneInRegion(zone string, region string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("alter system add zone '%s' region '%s'", zone, region)
	err = db.Exec(sql).Error
	return
}

func (obclusterService *ObclusterService) AddZone(zone string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("alter system add zone '%s'", zone)
	err = db.Exec(sql).Error
	return
}

func (obclusterService *ObclusterService) DeleteZone(zoneName string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	deleteZoneSql := "ALTER SYSTEM DELETE ZONE " + zoneName
	return db.Exec(deleteZoneSql).Error
}

func (obclusterService *ObclusterService) AddServers(sql []string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	for _, sql := range sql {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return
}

func (obclusterService *ObclusterService) GetUpgradePkgInfoByVersionAndReleaseDist(name, version, releaseDist, arch string) (pkgInfo oceanbase.UpgradePkgInfo, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Model(&oceanbase.UpgradePkgInfo{}).Where("name = ? and version = ? and release_distribution = ? and architecture = ? ", name, version, releaseDist, arch).Last(&pkgInfo).Error
	return
}

func (ObclusterService *ObclusterService) GetAllArchs() (archs []string, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Model(&oceanbase.AllAgent{}).Distinct("architecture").Pluck("architecture", &archs).Error
	return
}

func (ObclusterService *ObclusterService) GetAllUpgradePkgInfos() (pkgInfos []oceanbase.UpgradePkgInfo, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Model(&oceanbase.UpgradePkgInfo{}).Find(&pkgInfos).Error
	return
}

func (obclusterService *ObclusterService) GetUpgradePkgInfoByVersion(name, version, arch, distribution string, deprecatedInfo []string) (pkgInfo oceanbase.UpgradePkgInfo, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	if len(deprecatedInfo) == 0 {
		err = oceanbaseDb.Model(&oceanbase.UpgradePkgInfo{}).Where("name = ? and version = ? and distribution = ? and architecture = ? ", name, version, arch, distribution).Last(&pkgInfo).Error
	} else {
		err = oceanbaseDb.Model(&oceanbase.UpgradePkgInfo{}).Where("name = ? and version = ? and distribution = ? and architecture = ? and `release` not in ?", name, version, distribution, arch, deprecatedInfo).Last(&pkgInfo).Error
	}
	return
}

func (obclusterService *ObclusterService) GetUpgradePkgInfoByVersionAndRelease(name, version, release, distribution, arch string) (pkgInfo oceanbase.UpgradePkgInfo, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Model(&oceanbase.UpgradePkgInfo{}).Where("name = ? and version = ? and distribution = ? and architecture = ? and `release` = ?", name, version, distribution, arch, release).Last(&pkgInfo).Error
	return
}

func (obclusterService *ObclusterService) AddServer(svrInfo meta.ObserverSvrInfo, zoneName string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	alterSql := fmt.Sprintf("ALTER SYSTEM ADD SERVER '%s' ZONE '%s'", svrInfo.String(), zoneName)
	return db.Exec(alterSql).Error
}

func (obclusterService *ObclusterService) DeleteServerInZone(svrInfo meta.ObserverSvrInfo, zoneName string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	alterSql := fmt.Sprintf("ALTER SYSTEM DELETE SERVER '%s' ZONE '%s'", svrInfo.String(), zoneName)
	return db.Exec(alterSql).Error
}

func (obclusterService *ObclusterService) DeleteServer(svrInfo meta.ObserverSvrInfo) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	alterSql := fmt.Sprintf("ALTER SYSTEM DELETE SERVER '%s'", svrInfo.String())
	return db.Exec(alterSql).Error
}

func (ObclusterService *ObclusterService) CancelDeleteServer(svrInfo meta.ObserverSvrInfo) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	alterSql := fmt.Sprintf("ALTER SYSTEM CANCEL DELETE SERVER '%s'", svrInfo.String())
	return db.Exec(alterSql).Error
}

func (obclusterService *ObclusterService) IsServerExist(svrInfo meta.ObserverSvrInfo) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int
	err = db.Raw("select count(*) from oceanbase.dba_ob_servers where svr_ip = ? and svr_port = ?", svrInfo.GetIp(), svrInfo.GetPort()).First(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (obclusterService *ObclusterService) IsServerExistWithZone(svrInfo meta.ObserverSvrInfo, zone string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Table(DBA_OB_SERVERS).Where("svr_ip = ? and svr_port = ? and zone = ?", svrInfo.GetIp(), svrInfo.GetPort(), zone).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (obclusterService *ObclusterService) GetObZonesName() (zones []string, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = db.Raw("select zone from oceanbase.dba_ob_zones").Find(&zones).Error
	return
}

func (obclusterService *ObclusterService) IsZoneActive(zone string) (bool, error) {
	var count int
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	err = oceanbaseDb.Raw("SELECT count(*) FROM oceanbase.DBA_OB_ZONES WHERE STATUS = 'ACTIVE' AND ZONE = ?", zone).Find(&count).Error
	return count == 1, err
}

// HasOtherStopTask returns true if there has other zone which is stopped or has stopped server.
func (obclusterService *ObclusterService) HasOtherStopTask(excludeZone string) (bool, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = oceanbaseDb.Raw("SELECT COUNT(*) FROM (SELECT zone FROM oceanbase.DBA_OB_SERVERS WHERE stop_time > 0 AND zone != ? UNION SELECT zone FROM oceanbase.DBA_OB_ZONES WHERE status = 'INACTIVE' AND zone != ?)", excludeZone, excludeZone).Scan(&count).Error
	return count > 0, err
}

func (obclusterService *ObclusterService) IsZoneExist(zone string) (res bool, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	var count int64
	err = db.Model(&sqlite.AllAgent{}).Where("zone=?", zone).Count(&count).Error
	return count > 0, err
}

func (obclusterService *ObclusterService) IsZoneExistInOB(zone string) (res bool, err error) {
	var count int
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	err = oceanbaseDb.Raw("select count(*) from oceanbase.dba_ob_zones where zone = ?", zone).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (obclusterService *ObclusterService) MigrateAllAgentToOb() (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	var allAgent []sqlite.AllAgent
	if err = sqliteDb.Model(&sqlite.AllAgent{}).Find(&allAgent).Error; err != nil {
		return err
	}
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Transaction(func(tx *gorm.DB) error {
		if err = tx.Delete(&oceanbase.AllAgent{}, "1=1").Error; err != nil {
			return err
		}

		for _, agent := range allAgent {
			if err = tx.Model(&oceanbase.AllAgent{}).Create(&agent).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (obclusterService *ObclusterService) ModifyUserPwd(user, pwd string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	pwd = strings.ReplaceAll(pwd, "\"", "\\\"")
	sql := fmt.Sprintf("ALTER USER %s IDENTIFIED BY \"%s\"", user, pwd)
	return oceanbaseDb.Exec(sql).Error
}

func (obclusterService *ObclusterService) UpdateAllAgentIdentity(identity string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	err = oceanbaseDb.Model(&oceanbase.AllAgent{}).Where("identity != ''").Updates(&oceanbase.AllAgent{Identity: identity}).Error
	return
}

func (obclusterService *ObclusterService) GetAllOBServers() (server []oceanbase.OBServer, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Raw("SELECT * FROM oceanbase.DBA_OB_SERVERS").Scan(&server).Error
	return
}

func (obclusterService *ObclusterService) GetObVersion() (version string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return version, err
	}
	err = oceanbaseDb.Raw("select ob_version()").Scan(&version).Error
	return
}

func (obclusterService *ObclusterService) MigrateObSysParameter() (err error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	return sqliteDb.Transaction(func(sqliteTx *gorm.DB) error {
		err := sqliteTx.Exec("delete from " + constant.TABLE_OB_SYS_PARAMETER).Error
		if err != nil {
			return err
		}
		var obSysParameter []sqlite.ObSysParameter
		err = oceanbaseDb.Raw("select * from oceanbase.__all_virtual_sys_parameter_stat").Find(&obSysParameter).Error
		if err != nil {
			return err
		}
		for _, parameter := range obSysParameter {
			err = sqliteTx.Model(&sqlite.ObSysParameter{}).Create(&parameter).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (obclusterService *ObclusterService) GetObBuildVersion() (build_version string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("select BUILD_VERSION from oceanbase.dba_ob_servers where oceanbase.dba_ob_servers.SVR_IP = ? and oceanbase.dba_ob_servers.SQL_PORT = ?", meta.OCS_AGENT.GetIp(), meta.MYSQL_PORT).First(&build_version).Error
	if err != nil {
		return
	}
	build_version = strings.Split(build_version, "-")[0]
	return
}

func (obclusterService *ObclusterService) GetInactiveServerCount() (count int, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("select count(*) from oceanbase.DBA_OB_SERVERS where STATUS != 'ACTIVE' or STOP_TIME is not NULL or START_SERVICE_TIME is NULL").Find(&count).Error
	return
}

func (obclusterService *ObclusterService) GetNotInSyncServerCount() (count int, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("select count(*) from oceanbase.GV$OB_LOG_STAT where in_sync = 'NO'").Find(&count).Error
	return
}

func (obclusterService *ObclusterService) GetZoneCount() (count int, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("select count(*) from oceanbase.DBA_OB_ZONES").Find(&count).Error
	return
}

func (obclusterService *ObclusterService) GetAllZone() (zones []oceanbase.DbaObZones, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("SELECT * FROM oceanbase.DBA_OB_ZONES").Scan(&zones).Error
	return
}

func (obclusterService *ObclusterService) GetZone(zoneName string) (zone *oceanbase.DbaObZones, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(DBA_OB_ZONES).Where("ZONE = ?", zoneName).Scan(&zone).Error
	return
}

func (obclusterService *ObclusterService) IsZoneInactive(zone string) (bool, error) {
	var count int
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	err = oceanbaseDb.Raw("SELECT count(*) FROM oceanbase.DBA_OB_ZONES WHERE STATUS = 'INACTIVE' AND ZONE = ?", zone).Find(&count).Error
	return count == 1, err
}

func (obclusterService *ObclusterService) GetOBServersByZone(zone string) (observers []oceanbase.OBServer, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("SELECT * FROM oceanbase.DBA_OB_SERVERS WHERE ZONE = ?", zone).Scan(&observers).Error
	return
}

func (obclusterService *ObclusterService) GetOBServer(svrInfo meta.ObserverSvrInfo) (observer *oceanbase.OBServer, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(DBA_OB_SERVERS).Where("SVR_IP = ? AND SVR_PORT = ?", svrInfo.GetIp(), svrInfo.GetPort()).Scan(&observer).Error
	return
}

func (obclusterService *ObclusterService) GetOBServerByAgentInfo(agent meta.AgentInfo) (server *oceanbase.OBServer, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("SELECT ip AS SvrIp, rpc_port AS SvrPort FROM ocs.all_agent WHERE Ip = ? AND Port = ?", agent.GetIp(), agent.GetPort()).Scan(&server).Error
	return
}

func (obclusterService *ObclusterService) DownloadUpgradePkgChunkInBatch(filepath string, pkgId, chunkCount int) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < chunkCount; i++ {
		chunk, err := obclusterService.GetUpgradePkgChunkByPkgIdAndChunkId(pkgId, i)
		if err != nil {
			return err
		}
		_, err = file.Write(chunk.Chunk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obclusterService *ObclusterService) GetUpgradePkgChunkByPkgIdAndChunkId(pkgId, chunkId int) (chunk oceanbase.UpgradePkgChunk, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return chunk, err
	}
	oceanbaseDb.Exec("SET SESSION ob_query_timeout=1000000000") // Ignore the error because it is not non-essential.
	oceanbaseDb.Exec("SET SESSION ob_trx_timeout=1000000000")
	err = oceanbaseDb.Model(&oceanbase.UpgradePkgChunk{}).Where("pkg_id = ? and chunk_id = ?", pkgId, chunkId).First(&chunk).Error
	return
}

func (obclusterService *ObclusterService) GetUpgradePkgChunkCountByPkgId(pkgId int) (count int64, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return 0, err
	}
	err = oceanbaseDb.Model(&oceanbase.UpgradePkgChunk{}).Where("pkg_id = ?", pkgId).Count(&count).Error
	return
}

func (obclusterService *ObclusterService) DumpUpgradePkgInfoAndChunkTx(rpmPkg *rpm.Package, file multipart.File, upgradeDepYml string) (pkgInfo *oceanbase.UpgradePkgInfo, err error) {
	payloadSize := uint64(rpmPkg.Signature.GetTag(1000).Int64())
	chunkCount := payloadSize / constant.CHUNK_SIZE
	if payloadSize%constant.CHUNK_SIZE != 0 {
		chunkCount++
	}
	pkgInfo = &oceanbase.UpgradePkgInfo{
		Name:                rpmPkg.Name(),
		Version:             rpmPkg.Version(),
		ReleaseDistribution: rpmPkg.Release(),
		Distribution:        strings.Split(rpmPkg.Release(), ".")[1],
		Release:             strings.Split(rpmPkg.Release(), ".")[0],
		Architecture:        rpmPkg.Architecture(),
		Size:                rpmPkg.Size(),
		ChunkCount:          int(chunkCount),
		UpgradeDepYaml:      upgradeDepYml,
		PayloadSize:         payloadSize,
		Md5:                 hex.EncodeToString(rpmPkg.Signature.Tags[1004].Bytes()),
	}

	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Transaction(func(tx *gorm.DB) error {
		tx.Exec("SET SESSION ob_query_timeout=1000000000") // Ignore the error because it is not non-essential.
		tx.Exec("SET SESSION ob_trx_timeout=1000000000")
		if err := tx.Model(&oceanbase.UpgradePkgInfo{}).Create(&pkgInfo).Error; err != nil {
			return err
		}
		chunkBuffer := make([]byte, constant.CHUNK_SIZE)
		_, err = file.Seek(0, 0)
		if err != nil {
			return errors.Wrap(err, "Seek failed")
		}
		for i := 0; i < pkgInfo.ChunkCount; i++ {
			n, err := file.Read(chunkBuffer)
			if err != nil {
				return err
			}
			record := &oceanbase.UpgradePkgChunk{
				PkgId:      pkgInfo.PkgId,
				ChunkId:    i,
				ChunkCount: pkgInfo.ChunkCount,
				Chunk:      chunkBuffer[:n]}
			if err = tx.Model(&oceanbase.UpgradePkgChunk{}).Create(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (obclusterService *ObclusterService) RestoreParamsForUpgrade(params []oceanbase.ObParameters) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	err = oceanbaseDb.Transaction(func(tx *gorm.DB) error {
		for _, param := range params {
			var sql string
			switch param.Scope {
			case "TENANT":
				var tenantName string
				if err = tx.Raw("SELECT TENANT_NAME FROM oceanbase.DBA_OB_TENANTS WHERE TENANT_ID = ?", param.TenantId).Scan(&tenantName).Error; err != nil {
					return err
				}
				sql = fmt.Sprintf("ALTER SYSTEM SET %s = '%s' TENANT = %s", param.Name, param.Value, tenantName)
			case "CLUSTER":
				sql = fmt.Sprintf("ALTER SYSTEM SET %s = '%s' SERVER = '%s'", param.Name, param.Value, meta.NewAgentInfo(param.SvrIp, param.SvrPort).String())
			default:
				return errors.Occur(errors.ErrObParameterScopeInvalid, param.Scope)
			}
			err = tx.Exec(sql).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (obclusterService *ObclusterService) GetObParametersForUpgrade(params []string) (res []oceanbase.ObParameters, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Raw("SELECT SVR_IP, SVR_PORT, ZONE, SCOPE, TENANT_ID, NAME, VALUE FROM oceanbase.GV$OB_PARAMETERS WHERE NAME IN ?", params).Find(&res).Error
	if err != nil {
		return
	}
	return
}

func (*ObclusterService) GetAllUnhiddenParameters() ([]oceanbase.ObParameters, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}

	var unhiddenParams []oceanbase.ObParameters
	err = oceanbaseDb.Table(GV_OB_PARAMETERS).Where("NAME NOT LIKE ?", `\_%`).Find(&unhiddenParams).Error

	return unhiddenParams, err
}

func (obclusterService *ObclusterService) GetParameterByName(name string) (param *oceanbase.ObParameters, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(GV_OB_PARAMETERS).Where("NAME = ?", name).Scan(&param).Error
	return
}

func (obclusterService *ObclusterService) SetParameter(parameter param.SetParameterParam) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("ALTER SYSTEM SET `%s` = \"%v\"", parameter.Name, parameter.Value)
	if parameter.Zone != "" {
		sql += fmt.Sprintf(" ZONE = `%s`", parameter.Zone)
	} else if parameter.Server != "" {
		sql += fmt.Sprintf(" SERVER = '%s'", parameter.Server)
	} else if parameter.Tenant != "" {
		sql += fmt.Sprintf(" TENANT = `%s`", parameter.Tenant) // when tenant is not empty, zone and server won't influence the sql.
	}
	return oceanbaseDb.Exec(sql).Error
}

func (ObclusterService *ObclusterService) GetAllZonesWithRegion() (zones []oceanbase.DbaObZones, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_ZONES).Scan(&zones).Error
	return
}

func (obclusterService *ObclusterService) GetServerByZone(name string) (servers []oceanbase.OBServer, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_SERVERS).Where("ZONE = ?", name).Scan(&servers).Error
	return
}

func (obclusterService *ObclusterService) GetCharsetAndCollation(charset string, collation string) (*oceanbase.ObCollation, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}

	var charsetInfo *oceanbase.ObCollation
	if collation == "" {
		err = oceanbaseDb.Table(COLLATIONS).Select("CHARACTER_SET_NAME, COLLATION_NAME").Where("CHARACTER_SET_NAME = ?", charset).Scan(&charsetInfo).Error
		return charsetInfo, err
	} else if charset == "" {
		err = oceanbaseDb.Table(COLLATIONS).Select("CHARACTER_SET_NAME, COLLATION_NAME").Where("COLLATION_NAME = ?", collation).Scan(&charsetInfo).Error
		return charsetInfo, err
	} else {
		err = oceanbaseDb.Table(COLLATIONS).Select("CHARACTER_SET_NAME, COLLATION_NAME").Where("CHARACTER_SET_NAME = ? AND COLLATION_NAME = ?", charset, collation).Scan(&charsetInfo).Error
		return charsetInfo, err
	}
}

func (*ObclusterService) GetAllCharsets() (charsets []oceanbase.ObCharset, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Raw("SHOW CHARACTER SET").Scan(&charsets).Error
	return
}

func (*ObclusterService) GetAllCollations() (collations []oceanbase.ObCollation, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(COLLATIONS).Scan(&collations).Error
	return
}

func (*ObclusterService) GetCollationMap() (collationMap map[int]oceanbase.ObCollation, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var collations []oceanbase.ObCollation
	err = oceanbaseDb.Table(COLLATIONS).Scan(&collations).Error
	if err != nil {
		return nil, err
	}
	collationMap = make(map[int]oceanbase.ObCollation)
	for _, collation := range collations {
		collationMap[collation.Id] = collation
	}
	return
}

func (ObclusterService *ObclusterService) GetObUnitsOnServer(svrIp string, svrPort int) (units []oceanbase.DbaObUnit, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(DBA_OB_UNITS).Where("SVR_IP = ? AND SVR_PORT = ?", svrIp, svrPort).Scan(&units).Error
	return
}

func (ObclusterService *ObclusterService) IsLsMultiPaxosAlive(lsId int, tenantId int, svrInfo meta.ObserverSvrInfo) (bool, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	if err = oceanbaseDb.Raw("select count(*) from oceanbase.GV$OB_LOG_STAT as a inner join oceanbase.GV$OB_LOG_STAT as b on a.tenant_id = b.tenant_id and a.ls_id = b.ls_id and b.role = 'LEADER' and b.paxos_member_list like concat('%',a.svr_ip,':',a.svr_port,'%') and a.in_sync = 'YES' and a.ls_id = ? AND a.tenant_id = ? and (a.svr_ip != ? OR a.svr_port != ?)", lsId, tenantId, svrInfo.GetIp(), svrInfo.GetPort()).Count(&count).Error; err != nil {
		return false, err
	}
	var paxosMember int64
	if err = oceanbaseDb.Table(GV_OB_LOG_STAT).Select("paxos_replica_num").Where("ls_id = ? AND tenant_id = ? AND ROLE = 'LEADER'", lsId, tenantId).Scan(&paxosMember).Error; err != nil {
		return false, err
	}
	if count > paxosMember/2 {
		return true, nil
	} else {
		return false, nil
	}
}

// GetLogInfosInServer returns the log stat in target server
// only contains tenant_id and ls_id.
func (*ObclusterService) GetLogInfosInServer(svrInfo meta.ObserverSvrInfo) (logStats []oceanbase.ObLogStat, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Table(GV_OB_LOG_STAT).Distinct("TENANT_ID", "LS_ID").Where("SVR_IP = ? AND SVR_PORT = ?", svrInfo.GetIp(), svrInfo.GetPort()).Find(&logStats).Error
	return
}

func (*ObclusterService) HasUnitInZone(zone string) (exist bool, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = oceanbaseDb.Table(DBA_OB_UNITS).Where("ZONE = ?", zone).Count(&count).Error
	return count > 0, err
}

func (obclusterService *ObclusterService) CreateProxyroUser(password string) error {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}

	sqlText := fmt.Sprintf("CREATE USER IF NOT EXISTS `%s`@`%s`", constant.SYS_USER_PROXYRO, "%")
	if password != "" {
		sqlText += fmt.Sprintf(" IDENTIFIED BY '%s'", strings.ReplaceAll(password, "'", "'\"'\"'"))
	}
	if err = oceanbaseDb.Exec(sqlText).Error; err != nil {
		return err
	}
	if err := oceanbaseDb.Exec(fmt.Sprintf("GRANT SELECT ON oceanbase.* TO %s", constant.SYS_USER_PROXYRO)).Error; err != nil {
		return err
	}
	return nil
}

func (obclusterService *ObclusterService) GetRsListStr() (rsListStr string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return "", err
	}
	err = oceanbaseDb.Table(GV_OB_PARAMETERS).
		Select("VALUE").
		Where("NAME = ?", "rootservice_list").
		Scan(&rsListStr).Error
	return
}

func (obclusterService *ObclusterService) IsCommunityEdition() (bool, error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int64
	err = oceanbaseDb.Raw("select version() REGEXP 'OceanBase[\\s_]CE'").Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (obclusterService *ObclusterService) GetOBType() (obType modelob.OBType, err error) {
	obType = modelob.OBTypeUnknown
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}

	var count int64
	err = oceanbaseDb.Raw("select version() REGEXP 'OceanBase[\\s_]CE'").Scan(&count).Error
	if err != nil {
		return
	}
	if count > 0 {
		return modelob.OBTypeCommunity, nil
	}
	err = oceanbaseDb.Raw("SELECT COUNT(*) FROM oceanbase.DBA_OB_LICENSE").Scan(&count).Error
	if err != nil {
		if obErr, ok := err.(*obdriver.MySQLError); ok {
			if obErr.Number == 1146 { // table not found
				return modelob.OBTypeBusiness, nil
			}
		}
		return
	}
	if count > 0 {
		obType = modelob.OBTypeStandalone
	} else {
		obType = modelob.OBTypeBusiness
	}

	return
}

func (*ObclusterService) GetAllZoneRootServers() (rootServersMap map[string]oceanbase.RootServer, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var rootServers []oceanbase.RootServer
	err = oceanbaseDb.Table(CDB_OB_LS_LOCATIONS).Select("SVR_IP, SVR_PORT, ZONE, ROLE").Where("LS_ID = 1 AND TENANT_ID = 1").Scan(&rootServers).Error
	if err != nil {
		return nil, err
	}
	rootServersMap = make(map[string]oceanbase.RootServer)
	for _, server := range rootServers {
		rootServersMap[server.Zone] = server
	}
	return
}

func (ObclusterService) GetObserverCapacityByZone(zone string) (servers []oceanbase.ObServerCapacity, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(GV_OB_SERVERS).Where("ZONE = ?", zone).Scan(&servers).Error
	return
}

func (ObclusterService) GetAllObserverResourceMap() (observerResourceMap map[meta.ObserverSvrInfo]oceanbase.ObServerCapacity, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var serverResources []oceanbase.ObServerCapacity
	err = db.Table(GV_OB_SERVERS).Scan(&serverResources).Error
	if err != nil {
		return nil, err
	}
	observerResourceMap = make(map[meta.ObserverSvrInfo]oceanbase.ObServerCapacity)
	for _, serverResource := range serverResources {
		observerResourceMap[meta.ObserverSvrInfo{
			Ip:   serverResource.SvrIp,
			Port: serverResource.SvrPort,
		}] = serverResource
	}
	return
}

func (*ObclusterService) GetTenantSysStat(tenantId int, StatId int) (sysStat oceanbase.SysStat, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return sysStat, err
	}
	err = oceanbaseDb.Model(oceanbase.SysStat{}).Where("CON_ID = ? AND STAT_ID = ?", tenantId, StatId).Scan(&sysStat).Error
	return
}

func (obclusterService *ObclusterService) GetTenantMutilSysStat(tenantId int, StatIds []int) (sysStatMap map[int64]oceanbase.SysStat, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	var sysStats []oceanbase.SysStat
	err = oceanbaseDb.Model(oceanbase.SysStat{}).Select("CON_ID, SVR_IP, SVR_PORT, NAME, CLASS, VALUE_TYPE, STAT_ID, sum(VALUE) AS VALUE").Where("CON_ID = ? AND STAT_ID IN ?", tenantId, StatIds).Group("STAT_ID").Scan(&sysStats).Error
	if err != nil {
		return nil, err
	}
	sysStatMap = make(map[int64]oceanbase.SysStat)
	for _, sysStat := range sysStats {
		sysStatMap[int64(sysStat.StatId)] = sysStat
	}
	return sysStatMap, nil
}

func (*ObclusterService) GetTabletInMergingCount() (count int, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, err
	}
	// There is no view in oceanbase, so we need to use the table directly.
	sql := "select count(1) from oceanbase.__all_virtual_tablet_compaction_info where max_received_scn > finished_scn and max_received_scn > 0"
	err = oceanbaseDb.Raw(sql).Scan(&count).Error
	return count, err
}

func (*ObclusterService) GetRunningBackupTaskCount() (count int, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, err
	}
	sql := "select count(1) from oceanbase.CDB_OB_BACKUP_JOBS"
	err = oceanbaseDb.Raw(sql).Scan(&count).Error
	return count, err
}

func (obclusterService *ObclusterService) GetObLicense() (license *oceanbase.ObLicense, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = oceanbaseDb.Model(&oceanbase.ObLicense{}).First(&license).Error
	if err != nil {
		return nil, err
	}
	return
}
