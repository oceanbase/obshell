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

package agent

import (
	"os"
	"strconv"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (*AgentService) DeleteObproxy() error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM obproxy_info").Error; err != nil {
			return err
		}
		meta.OBPROXY_HOME_PATH = ""
		meta.OBPROXY_SQL_PORT = 0
		return nil
	})
}

func (*AgentService) AddObproxy(homePath string, sqlPort int, version, enObproxySysPwd, enObproxyProxyroPwd string) error {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}
	infos := make(map[string]string)
	infos[constant.OBPROXY_INFO_OBPROXY_SYS_PASSWORD] = enObproxySysPwd
	infos[constant.OBPROXY_INFO_PROXYRO_PASSWORD] = enObproxyProxyroPwd
	infos[constant.OBPROXY_INFO_HOME_PATH] = homePath
	infos[constant.OBPROXY_INFO_SQL_PORT] = strconv.Itoa(sqlPort)
	infos[constant.OBPROXY_INFO_VERSION] = version

	return db.Transaction(func(tx *gorm.DB) error {
		for k, v := range infos {
			// create or update
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(&sqlite.ObproxyInfo{
				Name:  k,
				Value: v,
			}).Error; err != nil {
				return err
			}
		}
		meta.OBPROXY_HOME_PATH = homePath
		return nil
	})
}

func (*AgentService) GetUpgradePkgInfoByVersion(name, version, arch, distribution string, deprecatedInfo []string) (pkgInfo sqlite.UpgradePkgInfo, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	if len(deprecatedInfo) == 0 {
		err = db.Model(&sqlite.UpgradePkgInfo{}).Where("name = ? and version = ? and distribution = ? and architecture = ? ", name, version, arch, distribution).Last(&pkgInfo).Error
	} else {
		err = db.Model(&sqlite.UpgradePkgInfo{}).Where("name = ? and version = ? and distribution = ? and architecture = ? and `release` not in ?", name, version, distribution, arch, deprecatedInfo).Last(&pkgInfo).Error
	}
	return
}

func (*AgentService) GetUpgradePkgInfoByVersionAndRelease(name, version, release, distribution, arch string) (pkgInfo sqlite.UpgradePkgInfo, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	err = db.Model(&sqlite.UpgradePkgInfo{}).Where("name = ? and version = ? and distribution = ? and architecture = ? and `release` = ?", name, version, distribution, arch, release).Last(&pkgInfo).Error
	return
}

func (agentService *AgentService) DownloadUpgradePkgChunkInBatch(filepath string, pkgId, chunkCount int) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < chunkCount; i++ {
		chunk, err := agentService.GetUpgradePkgChunkByPkgIdAndChunkId(pkgId, i)
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

func (agentService *AgentService) GetUpgradePkgChunkByPkgIdAndChunkId(pkgId, chunkId int) (chunk sqlite.UpgradePkgChunk, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return chunk, err
	}
	err = db.Model(&sqlite.UpgradePkgChunk{}).Where("pkg_id = ? and chunk_id = ?", pkgId, chunkId).First(&chunk).Error
	return
}

func (agentService *AgentService) GetUpgradePkgChunkCountByPkgId(pkgId int) (count int64, err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return 0, err
	}
	err = db.Model(&sqlite.UpgradePkgChunk{}).Where("pkg_id = ?", pkgId).Count(&count).Error
	return
}
