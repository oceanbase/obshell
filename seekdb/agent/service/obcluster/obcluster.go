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

	"github.com/cavaliergopher/rpm"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	oceanbasedb "github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/agent/repository/model/oceanbase"
)

func (obclusterService *ObclusterService) ExecuteSql(sql string) (err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Exec(sql).Error
}

// setSessionObQueryTimeout sets the session-level query timeout for an OceanBase database session.
// The `time` parameter specifies the timeout duration in microseconds(us).
func (obclusterService *ObclusterService) setSessionObQueryTimeout(db *gorm.DB, time int) (err error) {
	return db.Exec(fmt.Sprintf("SET SESSION ob_query_timeout=%d", time)).Error
}

func (ObclusterService *ObclusterService) GetServerCheckpointScn() (uint64, error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return 0, err
	}
	var checkpointScn uint64
	err = db.Raw("SELECT checkpoint_scn FROM oceanbase.__all_virtual_ls_info ORDER BY checkpoint_scn ASC LIMIT 1").Scan(&checkpointScn).Error
	if err != nil {
		return 0, err
	}
	return checkpointScn, nil
}

// MinorFreeze is only for sys tenant, and only support one server.
func (obclusterService *ObclusterService) MinorFreeze() (err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return err
	}
	return db.Exec("alter system minor freeze").Error
}

func (obclusterService *ObclusterService) GetUpgradePkgInfoByVersionAndReleaseDist(name, version, releaseDist, arch string) (pkgInfo oceanbase.UpgradePkgInfo, err error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Model(&oceanbase.UpgradePkgInfo{}).Where("name = ? and version = ? and release_distribution = ? and architecture = ? ", name, version, releaseDist, arch).Last(&pkgInfo).Error
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

func (obclusterService *ObclusterService) GetObVersion() (version string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return version, err
	}
	err = oceanbaseDb.Raw("select ob_version()").Scan(&version).Error
	return
}

func (obclusterService *ObclusterService) GetObBuildVersion() (revision string, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return revision, err
	}
	err = oceanbaseDb.Raw("select ob_build_version()").Scan(&revision).Error
	return
}

func (obclusterService *ObclusterService) GetOBServer() (observer *oceanbase.OBServer, err error) {
	oceanbaseDb, err := oceanbasedb.GetInstance()
	if err != nil {
		return
	}
	err = oceanbaseDb.Table(DBA_OB_SERVERS).Scan(&observer).Error
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

func (obclusterService *ObclusterService) DumpUpgradePkgInfoAndChunkTx(rpmPkg *rpm.Package, file multipart.File) (pkgInfo *oceanbase.UpgradePkgInfo, err error) {
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

func (obclusterService *ObclusterService) DeleteUpgradePkgInfoAndChunkTx(name, version, releaseDist, arch string) error {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}
	return oceanbaseDb.Transaction(func(tx *gorm.DB) error {
		// get pkg id
		var pkgId int
		err = oceanbaseDb.Model(&oceanbase.UpgradePkgInfo{}).Select("pkg_id").Where("name = ? and version = ? and release_distribution = ? and architecture = ?", name, version, releaseDist, arch).Scan(&pkgId).Error
		if err != nil {
			return err
		}
		if pkgId == 0 {
			return nil
		}
		err = oceanbaseDb.Model(&oceanbase.UpgradePkgInfo{}).Where("name = ? and version = ? and release_distribution = ? and architecture = ?", name, version, releaseDist, arch).Delete(&oceanbase.UpgradePkgInfo{}).Error
		if err != nil {
			return err
		}
		err = oceanbaseDb.Model(&oceanbase.UpgradePkgChunk{}).Where("pkg_id = ?", pkgId).Delete(&oceanbase.UpgradePkgChunk{}).Error
		if err != nil {
			return err
		}
		return nil
	})
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

func (ObclusterService) GetObserverResource() (resource *oceanbase.ObServerResource, err error) {
	db, err := oceanbasedb.GetInstance()
	if err != nil {
		return nil, err
	}
	err = db.Table(GV_OB_SERVERS).Scan(&resource).Error
	return
}
