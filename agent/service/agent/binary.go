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

	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/path"
	oceanbasedb "github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

func (s *AgentService) IsBinarySynced() (bool, error) {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return false, err
	}

	var status string
	err = sqliteDb.Model(sqlite.OcsInfo{}).Select("Value").Where("name = ?", constant.OCS_INFO_BIN_SYNCED).Scan(&status).Error
	if err != nil {
		return false, err
	}
	return status == "1", nil
}

func (s *AgentService) SetBinarySynced(synced bool) error {
	sqliteDb, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return err
	}

	value := "0"
	if synced {
		value = "1"
	}
	info := &sqlite.OcsInfo{
		Name:  constant.OCS_INFO_BIN_SYNCED,
		Value: value,
	}
	return s.updateInfo(sqliteDb, info)
}

func (s *AgentService) UpgradeBinary() error {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}

	file, err := os.Open(path.ObshellBinPath())
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	return oceanbaseDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET SESSION ob_query_timeout=1000000000").Error; err != nil {
			return err
		}
		if err := tx.Exec("SET SESSION ob_trx_timeout=1000000000").Error; err != nil {
			return err
		}

		info := &oceanbase.AgentBinaryInfo{
			Version:      constant.VERSION,
			Architecture: global.Architecture,
			Distribution: constant.DIST,
		}
		err := tx.Model(info).Create(info).Error
		if err != nil {
			if dbErr, ok := err.(*mysql.MySQLError); ok {
				if dbErr.Number == 1062 || dbErr.Number == 1205 {
					// check if the binary has been upgraded
					if err = tx.Model(info).Where("version = ?", constant.VERSION).Where("architecture = ?", global.Architecture).Where("distribution = ?", constant.DIST).Scan(info).Error; err != nil {
						return errors.Wrap(err, "check binary compatibility failed")
					} else if info.BinId == 0 {
						// other agent is upgrading
						return ErrOtherAgentUpgrading
					} else if info.ChunkCount > 0 {
						// binary has been upgraded
						return nil
					} else {
						// Only the binary info is created, but the binary is not uploaded
					}
				}
			} else {
				return err
			}
		}

		info.ChunkCount = int(stat.Size() / int64(constant.CHUNK_SIZE))
		chunkBuffer := make([]byte, constant.CHUNK_SIZE)
		if stat.Size()%int64(constant.CHUNK_SIZE) != 0 {
			info.ChunkCount++
		}

		for i := 0; i < info.ChunkCount; i++ {
			log.Info("Upgrading binary, chunk: ", i)
			n, err := file.Read(chunkBuffer)
			if err != nil {
				return err
			}

			chunk := &oceanbase.AgentBinaryChunk{
				BinId:      info.BinId,
				ChunkId:    i,
				ChunkCount: info.ChunkCount,
				Chunk:      chunkBuffer[:n],
			}
			if err = tx.Model(chunk).Create(chunk).Error; err != nil {
				return err
			}
		}
		return tx.Model(info).Where("bin_id = ?", info.BinId).UpdateColumn("chunk_count", info.ChunkCount).Error
	})

}

func (s *AgentService) DownloadBinary(filePath, version string) error {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return err
	}

	info := &oceanbase.AgentBinaryInfo{
		Version: version,
	}

	return oceanbaseDb.Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(info).Where("version = ?", version).Where("architecture = ?", global.Architecture).Where("distribution <= ?", constant.DIST).Scan(info).Error; err != nil {
			return err
		} else if info.ChunkCount == 0 {
			return errors.Occur(errors.ErrAgentBinaryNotFound, version, global.Architecture, constant.DIST)
		}

		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		for i := 0; i < info.ChunkCount; i++ {
			chunk := &oceanbase.AgentBinaryChunk{
				BinId:   info.BinId,
				ChunkId: i,
			}
			if err = tx.Model(chunk).Where("bin_id = ?", info.BinId).Where("chunk_id = ?", i).First(chunk).Error; err != nil {
				return err
			}

			if _, err = file.Write(chunk.Chunk); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AgentService) TargetVersionAgentExists(version string) (bool, error) {
	oceanbaseDb, err := oceanbasedb.GetOcsInstance()
	if err != nil {
		return false, err
	}

	var count int64
	if err = oceanbaseDb.Model(&oceanbase.AgentBinaryInfo{}).Where("version = ?", version).Where("architecture = ?", global.Architecture).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
