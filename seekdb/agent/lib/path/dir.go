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

package path

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
)

var mypath string

var binDir string

const (
	ENV_INTERNAL_BASE_DIR = "OBSHELL_INTERNAL_BASE_DIR"
	ENV_OBSHELL_PORT      = "OBSHELL_PORT_FOR_SEEKDB"
)

func initialize() {
	if mypath == "" {
		var err error
		mypath, err = getMyPathByArgs() // if error, getMyPathByArgs will exit process
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
	}
}

func isAgentStartCommand() bool {
	for i, arg := range os.Args {
		if arg == "agent" ||
			arg == "admin" {
			if i+1 < len(os.Args) {
				if os.Args[i+1] == "start" {
					return true
				}
			}
		}
		if arg == "daemon" ||
			arg == "server" {
			return true
		}
	}
	return false
}

func getMyPathByArgs() (string, error) {
	isForSeekdb := false
	for _, arg := range os.Args {
		if arg == "seekdb" || arg == "--seekdb" || strings.HasPrefix(arg, "--base-dir") {
			isForSeekdb = true
			break
		}
	}
	if !isForSeekdb {
		// not change any args of os.Args
		return "", nil
	}

	// get the binary path of the current process
	// use os.Executable() for cross-platform compatibility (works on Linux, macOS, Windows)
	binaryPath, err := os.Executable()
	if err != nil {
		return "", errors.Occurf(errors.ErrCommonUnexpected, "failed to get executable path: %v", err)
	}
	// resolve symlinks to get the actual binary path
	binaryPath, err = filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return "", errors.Occurf(errors.ErrCommonUnexpected, "failed to resolve executable symlink: %v", err)
	}
	binDir = filepath.Dir(binaryPath)

	var baseDir string
	var isSetBaseDir bool
	var obshellPort int

	isAgentStart := isAgentStartCommand()
	if isAgentStart {
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			if strings.HasPrefix(arg, "--base-dir=") {
				baseDir = strings.TrimPrefix(arg, "--base-dir=")
				isSetBaseDir = true
				break
			} else if arg == "--base-dir" && i+1 < len(os.Args) {
				baseDir = os.Args[i+1]
				isSetBaseDir = true
				break
			}
		}
		if isSetBaseDir {
			if baseDir == "" {
				return "", errors.Occur(errors.ErrAgentBaseDirInvalid, baseDir, "base dir is empty")
			}
			ret, err := filepath.Abs(baseDir)
			if err != nil {
				return "", err
			}
			if _, err := os.Stat(baseDir); err != nil {
				return "", errors.Occur(errors.ErrAgentBaseDirInvalid, baseDir, err.Error())
			}
			return ret, nil
		} else if internalBaseDir := os.Getenv(ENV_INTERNAL_BASE_DIR); internalBaseDir != "" {
			if !strings.HasPrefix(internalBaseDir, "/") {
				return "", errors.Occur(errors.ErrAgentBaseDirInvalid, internalBaseDir, "base-dir is not absolute path")
			}
			if _, err := os.Stat(internalBaseDir); err != nil {
				return "", errors.Occur(errors.ErrAgentBaseDirInvalid, internalBaseDir, err.Error())
			}
			return internalBaseDir, nil
		} else {
			// retrun current directory
			return os.Getwd()
		}
	} else {
		var useIPv6 bool
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			if strings.HasPrefix(arg, "--port=") {
				portStr := strings.TrimPrefix(arg, "--port=")
				if p, err := strconv.Atoi(portStr); err == nil {
					obshellPort = p
				} else {
					return "", errors.Occur(errors.ErrCommonInvalidPort, portStr)
				}
				os.Args = append(os.Args[:i], os.Args[i+1:]...)
				break
			} else if arg == "--port" && i+1 < len(os.Args) {
				if p, err := strconv.Atoi(os.Args[i+1]); err == nil {
					obshellPort = p
				} else {
					return "", errors.Occur(errors.ErrCommonInvalidPort, os.Args[i+1])
				}
				os.Args = append(os.Args[:i], os.Args[i+2:]...)
				break
			} else if arg == "--use-ipv6" || arg == "-6" {
				useIPv6 = true
			}
		}
		if obshellPort == 0 { // read from env
			// read from env OBSHELL_INTERNAL_BASE_DIR
			obshellPortStr := os.Getenv(ENV_OBSHELL_PORT)
			if obshellPortStr != "" {
				obshellPort, err = strconv.Atoi(obshellPortStr)
				if err != nil {
					return "", errors.Occur(errors.ErrCommonInvalidPort, obshellPortStr)
				}
			} else {
				obshellPort = constant.DEFAULT_AGENT_PORT
			}
		}

		respStruct := struct {
			Data struct {
				HomePath string `json:"homePath"`
				Type     string `json:"type"`
			} `json:"data"`
		}{}

		ip := constant.LOCAL_IP
		if useIPv6 {
			ip = constant.LOCAL_IP_V6
		}
		var agentInfo meta.AgentInfo
		agentInfo.Ip = ip
		agentInfo.Port = obshellPort
		resp, err := resty.New().R().Get(fmt.Sprintf("http://%s/api/v1/info", agentInfo.String()))
		if err != nil {
			return "", err
		}
		if resp.IsError() {
			return "", errors.New("failed to get agent info")
		}
		err = json.Unmarshal(resp.Body(), &respStruct)
		if err != nil {
			return "", err
		}
		if respStruct.Data.Type != "seekdb" {
			return "", errors.Occur(errors.ErrCliUsageError, "the target obshell agent does not manage a seekdb instance")
		}
		return respStruct.Data.HomePath, nil
	}
}

func AgentDir() string {
	initialize()
	return mypath
}

func RunDir() string {
	initialize()
	return filepath.Join(AgentDir(), constant.DIR_RUN)
}

func ObserverBinPath() string {
	initialize()
	return filepath.Join(RunDir(), constant.PROC_SEEKDB)
}

func ObserverClusterIdFilePath() string {
	return filepath.Join(RunDir(), "telemetry.json")
}

func BinDir() string {
	initialize()
	return binDir
}

func LogDir() string {
	initialize()
	return filepath.Join(AgentDir(), constant.DIR_LOG_OBSHELL)
}

func EtcDir() string {
	initialize()
	return filepath.Join(AgentDir(), constant.OB_DIR_ETC)
}

func CertificateDir() string {
	initialize()
	return filepath.Join(AgentDir(), constant.DIR_CA)
}
