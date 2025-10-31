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

package cmd

import (
	"fmt"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/config"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/ob"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/lib/process"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
	ocsagentlog "github.com/oceanbase/obshell/ob/agent/log"
	"github.com/oceanbase/obshell/ob/agent/service/agent"
)

var (
	agentService = agent.AgentService{}
)

func PreHandler() {
	// Check whether is oceanbase seekdb
	if len(os.Args) >= 2 {
		arg := strings.ToLower(os.Args[1])
		if arg != "version" && arg != "info-ip" && arg != "-v" && arg != "--versoin" && ob.IsOceanBaseSeekdb() {
			process.ExitWithErrorWithoutLog(constant.EXIT_CODE_ERROR_AGENT_START_FAILED, errors.Occur(errors.ErrAgentStartForOceanbaseSeekdbIncorrectly))
		}
	}
	if len(os.Args) >= 2 {
		switch strings.ToLower(os.Args[1]) {
		case CMD_ADMIN:
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			handleBackupBin()
		case CMD_DAEMON:
			ocsagentlog.InitLogger(config.DefaultDaemonLoggerConifg())
			handleBackupBin()
		case CMD_SERVER:
			ocsagentlog.InitLogger(config.DefaultAgentLoggerConifg())
			handleBackupBin()
		default:
			return
		}
	}
}

func HandleVersionFlag() {
	fmt.Printf("OBShell %s (for OceanBase_CE)\n\n", constant.VERSION)
	fmt.Printf("REVISION: %s-%s\n", constant.RELEASE, config.GitCommitId)
	fmt.Printf("BUILD_BRANCH: %s\n", config.GitBranch)
	fmt.Printf("BUILD_TIME: %s\n", config.BuildTime)
	fmt.Printf("BUILD_FLAGS: %s\n", config.Mode)
	fmt.Print("BUILD_INFO: \n\nCopyright (c) 2011-present OceanBase Inc.\n\n")
	os.Exit(0)
}

func handleInfoIp() {
	ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
	ocsagentlog.SetDBLoggerLevel(ocsagentlog.Silent)
	var err error
	var ip string
	defer func() {
		fmt.Println(ip)
		os.Exit(0)
	}()

	if ip, err = agentService.GetIP(); ip != "" {
		return
	} else if err != nil {
		log.WithError(err).Error("get ip from sqlite failed")
	}

	if obHasStarted, err := ob.HasStarted(); obHasStarted {
		if ip, _, _, _ = ob.GetConfFromObConfFile(); ip != "" {
			return
		}
		if err != nil {
			log.WithError(err).Error("get ip from conf failed")
		}
	}
	ip, err = GetHostFirstIP()
	if err != nil {
		log.WithError(err).Error("get ip from host failed")
	}
}

// GetHostFirstIP retrieves the first IP address of the network interface
// available on the host machine.
func GetHostFirstIP() (ip string, err error) {
	hostname, err := os.Hostname()
	if err != nil {
		return
	}

	addrs := make([]string, 0)
	address, err := net.LookupHost(hostname)
	if err != nil {
		return
	}

	for _, addr := range address {
		ip := net.ParseIP(addr)
		if ip.To4() != nil && !ip.IsLoopback() {
			addrs = append(addrs, addr)
		}
	}

	return addrs[0], nil
}

func NewInfoIpCmd() *cobra.Command {
	return &cobra.Command{
		Use:    CMD_INFO_IP,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			handleInfoIp()
		},
	}
}

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:    CMD_VERSION,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(constant.VERSION_RELEASE)
			os.Exit(0)
		},
	}
}

func handleBackupBin() {
	log.Info("current obshell version is ", constant.VERSION_RELEASE)
	if needExecBackupBin() {
		if err := execBackupBin(); err != nil {
			log.WithError(err).Error("execute backup binary failed")
			os.Exit(constant.EXIT_CODE_ERROR_EXEC_BINARY_FAILED)
		}
		os.Exit(0)
	}

	if needBackupBin() {
		if err := BackupBin(); err != nil {
			log.WithError(err).Error("backup binary failed1")
			os.Exit(constant.EXIT_CODE_ERROR_BACKUP_BINARY_FAILED)
		}
		return
	}
}

func BackupBinExist() bool {
	return system.IsFileExist(path.ObshellBinBackupPath())
}

func needExecBackupBin() bool {
	log.Info("check backup binary ", path.ObshellBinBackupPath())
	if !BackupBinExist() {
		log.Info("backup binary not exist")
		return false
	}

	// Compare the backup binary version with the current binary versio.
	backupVersion, err := system.GetBinaryVersion(path.ObshellBinBackupPath())
	if err != nil {
		return false
	}
	log.Info("backup binary version is ", backupVersion)
	return pkg.CompareVersion(backupVersion, constant.VERSION_RELEASE) == 1
}

func execBackupBin() error {
	//Replace the current binary version with the backup binary.
	log.Info("replace current binary with backup binary")
	if err := replaceBin(); err != nil {
		return errors.Wrap(err, "replace current binary failed")
	}

	// Execute the backup binary.
	log.Info("execute backup binary")
	return process.ExecuteBinary(path.ObshellBinPath(), os.Args[1:])
}

func needBackupBin() bool {
	backupExist := system.IsFileExist(path.ObshellBinBackupPath())
	if !backupExist {
		return true
	}

	// Compare the backup binary version with the current binary version.
	backupVersion, err := system.GetBinaryVersion(path.ObshellBinBackupPath())
	if err != nil {
		return false
	}
	return pkg.CompareVersion(backupVersion, constant.VERSION_RELEASE) == -1
}

func BackupBin() error {
	log.Info("prepare to backup binary")
	if err := os.RemoveAll(path.ObshellBinBackupPath()); err != nil {
		return err
	}
	log.Info("copy current binary to backup binary")
	if err := system.CopyFile(path.ObshellBinPath(), path.ObshellBinBackupPath()); err != nil {
		return err
	}
	return nil
}

func replaceBin() error {
	log.Info("remove current binary")
	if err := os.RemoveAll(path.ObshellBinPath()); err != nil {
		return err
	}

	log.Info("copy backup binary to current binary")
	if err := system.CopyFile(path.ObshellBinBackupPath(), path.ObshellBinPath()); err != nil {
		return err
	}

	return nil
}
