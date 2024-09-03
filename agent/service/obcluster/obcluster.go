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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

type ObclusterService struct{}

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

// MinorFreeze is only for sys tenant, and only support one server.
func (obclusterService *ObclusterService) MinorFreeze(servers []oceanbase.OBServer) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	var targetCmd []string
	for _, server := range servers {
		targetCmd = append(targetCmd, fmt.Sprintf("'%s:%d'", server.SvrIp, server.SvrPort))
	}
	serverList := strings.Join(targetCmd, ",")
	sql := fmt.Sprintf("alter system minor freeze server = (%[1]s);", serverList)

	return db.Exec(sql).Error
}

// IsLsCheckpointAfterTs will check the smallest checkpoint of the log stream of on the target server.
func (obclusterService *ObclusterService) IsLsCheckpointAfterTs(server oceanbase.OBServer) (t time.Time, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return t, err
	}

	var systemTimeZone string
	if err = db.Raw("SELECT @@system_time_zone").Scan(&systemTimeZone).Error; err != nil {
		return
	}

	// Get checkpoint of target server.
	sql := fmt.Sprintf("select CONVERT_TZ(scn_to_timestamp(checkpoint_scn), '%s', '+00:00') from oceanbase.__all_virtual_ls_info where svr_ip = '%s' and svr_port = %d order by checkpoint_scn asc limit 1", systemTimeZone, server.SvrIp, server.SvrPort)
	err = db.Raw(sql).Scan(&t).Error
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

func (obclusterService *ObclusterService) AddServer(ip, port, zoneName string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	alterSql := fmt.Sprintf("ALTER SYSTEM ADD SERVER '%s:%s' ZONE '%s'", ip, port, zoneName)
	return db.Exec(alterSql).Error
}

func (obclusterService *ObclusterService) DeleteServer(ip, port, zoneName string) (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	alterSql := fmt.Sprintf("ALTER SYSTEM DELETE SERVER '%s:%s' ZONE '%s'", ip, port, zoneName)
	return db.Exec(alterSql).Error
}

func (obclusterService *ObclusterService) IsServerExist(ip string, port string) (bool, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return false, err
	}
	var count int
	err = db.Raw("select count(*) from oceanbase.dba_ob_servers where svr_ip = ? and svr_port = ?", ip, port).First(&count).Error
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
	err = oceanbaseDb.Raw("select count(*) from oceanbase.dba_ob_zones where zone = ?", zone).First(&count).Error
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

func (obclusterService *ObclusterService) GetOBServerByAgentInfo(ip string, port int) (server oceanbase.OBServer, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Raw("SELECT ip AS SvrIp, rpc_port AS SvrPort FROM ocs.all_agent WHERE Ip = ? AND Port = ?", ip, port).Scan(&server).Error
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
				sql = fmt.Sprintf("ALTER SYSTEM SET %s = '%s' SERVER = '%s:%d'", param.Name, param.Value, param.SvrIp, param.SvrPort)
			default:
				return errors.New("unknown scope")
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
