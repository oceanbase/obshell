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

	obdriver "github.com/oceanbase/go-oceanbase-driver"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	oceanbasedb "github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	sqlitedb "github.com/oceanbase/obshell/ob/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/sqlite"
)

// agentBinaryChunkSize is the size of each chunk when uploading the agent binary to OceanBase.
// Kept smaller than the shared CHUNK_SIZE to reduce per-statement latency and OB MemStore pressure.
const agentBinaryChunkSize = 1024 * 1024 * 4 // 4 MB

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
	db, err := oceanbasedb.GetOcsInstance()
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

	chunkCount := int(stat.Size() / agentBinaryChunkSize)
	if stat.Size()%agentBinaryChunkSize != 0 {
		chunkCount++
	}

	// Step 1: create the binary info record (ChunkCount=0 acts as "upload in progress" marker).
	info := &oceanbase.AgentBinaryInfo{
		Version:      constant.VERSION,
		Architecture: global.Architecture,
		Distribution: constant.DIST,
	}
	if err = db.Model(info).Create(info).Error; err != nil {
		if dbErr, ok := err.(*obdriver.MySQLError); ok && (dbErr.Number == 1062 || dbErr.Number == 1205) {
			if err = db.Model(info).
				Where("version = ?", constant.VERSION).
				Where("architecture = ?", global.Architecture).
				Where("distribution = ?", constant.DIST).
				Scan(info).Error; err != nil {
				return errors.Wrap(err, "check binary compatibility failed")
			} else if info.BinId == 0 {
				return ErrOtherAgentUpgrading
			} else if info.ChunkCount > 0 {
				// binary already fully uploaded
				return nil
			}
			// ChunkCount == 0: info exists but upload was interrupted; clean up partial chunks.
			if err = db.Where("bin_id = ?", info.BinId).Delete(&oceanbase.AgentBinaryChunk{}).Error; err != nil {
				return errors.Wrap(err, "clean up partial binary chunks failed")
			}
		} else {
			return err
		}
	}

	if err = db.Exec("SET SESSION ob_query_timeout=1000000000").Error; err != nil {
		return err
	}
	if err = db.Exec("SET SESSION ob_trx_timeout=1000000000").Error; err != nil {
		return err
	}

	// Step 2: write each chunk as an independent auto-commit statement to avoid large MemStore pressure.
	chunkBuffer := make([]byte, agentBinaryChunkSize)
	for i := 0; i < chunkCount; i++ {
		log.Info("Upgrading binary, chunk: ", i)
		n, err := file.Read(chunkBuffer)
		if err != nil {
			return err
		}

		chunkData := make([]byte, n)
		copy(chunkData, chunkBuffer[:n])

		chunk := &oceanbase.AgentBinaryChunk{
			BinId:      info.BinId,
			ChunkId:    i,
			ChunkCount: chunkCount,
			Chunk:      chunkData,
		}
		if err = db.Model(chunk).Create(chunk).Error; err != nil {
			return err
		}
	}

	// Step 3: update ChunkCount to mark the binary as fully uploaded.
	return db.Model(info).Where("bin_id = ?", info.BinId).UpdateColumn("chunk_count", chunkCount).Error
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
