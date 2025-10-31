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

package ob

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
)

const (
	ETC_KEY_MYSQL_PORT      = "mysql_port"
	ETC_KEY_RPC_PORT        = "rpc_port"
	ETC_KEY_IP              = "local_ip"
	ETC_KEY_ZONE            = "zone"
	ETC_KEY_ALL_SERVER_LIST = "all_server_list"
)

func isFirstStart() (bool, error) {
	filePath := path.ObConfigPath()
	if _, err := os.Stat(filePath); err == nil {
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

func IsOceanBaseSeekdb() bool {
	// execute the observer binary, if it is success, return true
	cmd := exec.Command(filepath.Join(path.RunDir(), constant.PROC_OBSERVER), "-V")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	res := string(out)
	return strings.Contains(res, "OceanBase SeekDB")
}

func LoadOBConfigFromConfigFile() (err error) {
	// Load ob port from $homepath/etc/observer.config.bin.
	log.Info("load ob config from config file")
	filePath := path.ObConfigPath()

	file, err := os.Open(filePath)
	if err != nil {
		return errors.Wrapf(err, "read file %s", filePath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "read file failed")
	}

	ip, zone, mysqlPort, rpcPort := GetConfFromObConfFile()
	if mysqlPort == 0 || rpcPort == 0 || zone == "" {
		return errors.Occur(errors.ErrCommonUnexpected, "load observer configs from config file failed")
	}
	if err = agentService.UpdateAgentIP(ip); err != nil {
		return err
	}
	return agentService.UpdatePortAndZone(mysqlPort, rpcPort, zone)
}

func GetConfFromObConfFile() (ip, zone string, mysqlPort, rpcPort int) {
	f := path.ObConfigPath()
	log.Info("get conf from ob conf file ", f)
	file, err := os.Open(f)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if err = scanner.Err(); err != nil {
		return
	}
	re := regexp.MustCompile("\x00*([_a-zA-Z]+)=(.*)")

	for scanner.Scan() {
		if ip != "" && zone != "" && mysqlPort != 0 && rpcPort != 0 {
			break
		}
		line := scanner.Text()
		match := re.FindStringSubmatch(line)

		if len(match) != 3 {
			continue
		}

		switch match[1] {
		case ETC_KEY_IP:
			ip = match[2]
		case ETC_KEY_ZONE:
			zone = match[2]
		case ETC_KEY_MYSQL_PORT:
			mysqlPort, _ = strconv.Atoi(match[2])
		case ETC_KEY_RPC_PORT:
			rpcPort, _ = strconv.Atoi(match[2])
		}
	}
	log.Infof("get conf from ob conf file, ip: %s, zone: %s, mysqlPort: %d, rpcPort: %d", ip, zone, mysqlPort, rpcPort)
	return
}
