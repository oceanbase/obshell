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

package obproxy

import (
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"

	"github.com/cavaliergopher/rpm"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	obproxydb "github.com/oceanbase/obshell/agent/repository/db/obproxy"
	sqlitedb "github.com/oceanbase/obshell/agent/repository/db/sqlite"
	"github.com/oceanbase/obshell/agent/secure"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
)

func (obproxyService *ObproxyService) SetSysPassword(password string) (err error) {
	return obproxyService.SetGlobalConfig(constant.OBPROXY_CONFIG_OBPROXY_SYS_PASSWORD, password)
}

func (obproxyService *ObproxyService) SetProxyroPassword(password string) error {
	return obproxyService.SetGlobalConfig(constant.OBPROXY_CONFIG_PROXYRO_PASSWORD, password)
}

func (*ObproxyService) SetGlobalConfig(name string, value string) error {
	db, err := obproxydb.GetObproxyInstance()
	if err != nil {
		return err
	}

	if err := db.Exec(fmt.Sprintf("ALTER proxyconfig SET %s = %s ", name, value)).Error; err != nil {
		return err
	}
	return nil
}

func (*ObproxyService) GetObproxyVersion() (version string, err error) {
	db, err := obproxydb.GetObproxyInstance()
	if err != nil {
		return
	}
	var proxyInfo bo.ObproxyInfo
	if err = db.Raw("show proxyinfo binary").Scan(&proxyInfo).Error; err != nil {
		return "", err
	}
	// parse obproxy version
	re := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+-\d+`)
	version = re.FindString(proxyInfo.Info)
	return version, err
}

func (*ObproxyService) GetGlobalConfig(name string) (value string, err error) {
	db, err := obproxydb.GetObproxyInstance()
	if err != nil {
		return
	}
	var proxyConfig bo.ProxyConfig
	err = db.Raw(fmt.Sprintf("show proxyconfig like '%s'", name)).Scan(&proxyConfig).Error
	return proxyConfig.Value, err
}

func (obproxyService *ObproxyService) UpdateSqlPort(sqlPort int) (err error) {
	if err := obproxyService.UpdateObproxyInfo(constant.OBPROXY_INFO_SQL_PORT, strconv.Itoa(sqlPort)); err != nil {
		return err
	}
	meta.OBPROXY_SQL_PORT = sqlPort
	return nil
}

func (obproxyService *ObproxyService) UpdateObproxySysPassword(obproxySysPassword string) (err error) {
	encryptPwd, err := secure.Encrypt(obproxySysPassword)
	if err != nil {
		return err
	}
	if err := obproxyService.UpdateObproxyInfo(constant.OBPROXY_INFO_OBPROXY_SYS_PASSWORD, encryptPwd); err != nil {
		return err
	}
	meta.OBPROXY_SYS_PWD = obproxySysPassword
	return nil
}

func (obproxyService *ObproxyService) UpdateObproxyInfo(name string, value string) (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	obproxyInfo := &sqlite.ObproxyInfo{
		Name:  name,
		Value: value,
	}
	return db.Model(obproxyInfo).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(obproxyInfo).Error
}

func (*ObproxyService) GetObclusterName(db *gorm.DB) (name string, err error) {
	err = db.Table(GV_OB_PARAMETERS).Where("name = ?", "cluster").Select("value").Scan(&name).Error
	return
}

func (*ObproxyService) ClearObproxyInfo() (err error) {
	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return
	}
	return db.Delete(&sqlite.ObproxyInfo{}).Error
}

func (*ObproxyService) DumpUpgradePkgInfoAndChunkTx(rpmPkg *rpm.Package, file multipart.File) (pkgInfo *sqlite.UpgradePkgInfo, err error) {
	payloadSize := uint64(rpmPkg.Signature.GetTag(1000).Int64())
	chunkCount := payloadSize / constant.CHUNK_SIZE
	if payloadSize%constant.CHUNK_SIZE != 0 {
		chunkCount++
	}
	pkgInfo = &sqlite.UpgradePkgInfo{
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

	db, err := sqlitedb.GetSqliteInstance()
	if err != nil {
		return nil, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&sqlite.UpgradePkgInfo{}).Create(&pkgInfo).Error; err != nil {
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
			record := &sqlite.UpgradePkgChunk{
				PkgId:      pkgInfo.PkgId,
				ChunkId:    i,
				ChunkCount: pkgInfo.ChunkCount,
				Chunk:      chunkBuffer[:n]}
			if err = tx.Model(&sqlite.UpgradePkgChunk{}).Create(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return
}
