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

package observer

import (
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"
)

const (
	ETC_KEY_MYSQL_PORT      = "mysql_port"
	ETC_KEY_IP              = "local_ip"
	ETC_KEY_USE_IPV6        = "use_ipv6"
	ETC_KEY_ALL_SERVER_LIST = "all_server_list"
)

// isFirstStart returns true when seekdb has not been started (no meta.db yet).
// Seekdb stores config in ./etc/meta.db table __all_sys_parameter, not in seekdb.config.bin.
func isFirstStart() (bool, error) {
	metaPath := path.MetaDbPath()
	if _, err := os.Stat(metaPath); err == nil {
		return false, nil
	} else if os.IsNotExist(err) {
		return true, nil
	} else {
		return false, err
	}
}

func HasStarted() (bool, error) {
	isFirstStart, err := isFirstStart()
	return !isFirstStart, err
}

// getParamFromMetaDb reads a parameter value by name from meta.db __all_sys_parameter.
// Returns empty string and false if meta.db does not exist or the name is not found.
func getParamFromMetaDb(name string) (string, bool) {
	metaPath := path.MetaDbPath()
	db, err := gorm.Open(sqlite.Open(metaPath), &gorm.Config{})
	if err != nil {
		log.WithError(err).Errorf("open meta.db failed: %s", metaPath)
		return "", false
	}
	sqlDb, err := db.DB()
	if err != nil {
		return "", false
	}
	defer sqlDb.Close()

	var value string
	err = db.Table(constant.OB_SYS_PARAMETER_TABLE).Where("name = ?", name).Select("value").Limit(1).Scan(&value).Error
	if err != nil || value == "" {
		return "", false
	}
	return value, true
}

func LoadOBConfigFromConfigFile() (err error) {
	// Load ob config from ./etc/meta.db table __all_sys_parameter (seekdb does not use seekdb.config.bin).
	log.Info("load ob config from meta.db __all_sys_parameter")
	ip, mysqlPort := GetConfFromObMeta()
	if mysqlPort == 0 {
		log.Info("load ob config without mysql port, use default mysql port")
		mysqlPort = constant.DEFAULT_MYSQL_PORT
	}
	if err = agentService.UpdateAgentIP(ip); err != nil {
		return err
	}
	return agentService.UpdatePort(mysqlPort)
}

// GetConfFromObMeta reads mysql_port and local_ip from ./etc/meta.db __all_sys_parameter.
// If meta.db or a parameter is missing, returns constant.LOCAL_IP and 0 for port.
func GetConfFromObMeta() (ip string, mysqlPort int) {
	metaPath := path.MetaDbPath()
	log.Infof("get conf from meta.db %s table %s", metaPath, constant.OB_SYS_PARAMETER_TABLE)

	if v, ok := getParamFromMetaDb(ETC_KEY_MYSQL_PORT); ok {
		mysqlPort, _ = strconv.Atoi(v)
	}
	ip = constant.LOCAL_IP
	if useIPv6, ok := getParamFromMetaDb(ETC_KEY_USE_IPV6); ok && (useIPv6 == "1" || strings.ToLower(useIPv6) == "true") {
		ip = constant.LOCAL_IP_V6
	}
	log.Infof("get conf from meta.db, ip: %s, mysqlPort: %d", ip, mysqlPort)
	return
}
