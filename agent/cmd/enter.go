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
	"os"
	"unsafe"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/client/command"
	"github.com/oceanbase/obshell/utils"
)

const (
	CMD_ADMIN   = "admin"
	CMD_DAEMON  = "daemon"
	CMD_SERVER  = "server"
	CMD_VERSION = "version"
	CMD_V       = "V"
	CMD_START   = "start"
	CMD_STOP    = "stop"
	CMD_RESTART = "restart"
	CMD_INFO_IP = "info-ip"
)

type CommonFlag struct {
	AgentInfo     meta.AgentInfo
	OldServerPid  int32
	IsTakeover    int
	NeedStartOB   bool
	NeedBeCluster bool
	RootPassword  *string // Call HiddenPassword() before use RootPassword, or use GetRootPassword() instead.
	rootPassword  string  // Just use for input. Don't use it in the code.
	hiden         bool
}

func SetCommandFlags(cmd *command.Command, flag *CommonFlag) {
	cmd.Flags().SortFlags = false
	cmd.VarsPs(&flag.AgentInfo.Ip, []string{constant.FLAG_IP}, "", "The IP address for the agent to bind to", false)
	cmd.VarsPs(&flag.AgentInfo.Port, []string{constant.FLAG_PORT, constant.FLAG_PORT_SH}, 0, "The operations port number", false)
	cmd.VarsPs(&flag.OldServerPid, []string{constant.FLAG_PID}, int32(0), "Old obshell pid, only used for upgrade", false)
	cmd.Flags().MarkHidden(constant.FLAG_PID)
	cmd.VarsPs(&flag.IsTakeover, []string{constant.FLAG_TAKE_OVER}, 1, "If the agent is started for a takeover", false)
	cmd.VarsPs(&flag.rootPassword, []string{constant.FLAG_ROOT_PWD_SH, constant.FLAG_ROOT_PWD}, "", "The password for OceanBase root@sys user, only used for takeover", false)
	cmd.VarsPs(&flag.NeedStartOB, []string{constant.FLAG_START_OB}, false, "If need to start observer", false)
	cmd.VarsPs(&flag.NeedBeCluster, []string{constant.FLAG_NEED_BE_CLUSTER}, false, "If need to be a cluster agent", false)
	cmd.Flags().MarkHidden(constant.FLAG_START_OB)
}

func (flag *CommonFlag) GetArgs() (args []string) {
	flag.HiddenPassword()
	if flag.AgentInfo.GetIp() != "" {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_IP), flag.AgentInfo.GetIp())
	}
	if flag.AgentInfo.GetPort() != 0 {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_PORT), fmt.Sprint(flag.AgentInfo.GetPort()))
	}
	if flag.OldServerPid != 0 {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_PID), fmt.Sprint(flag.OldServerPid))
	}
	if flag.IsTakeover == 0 {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_TAKE_OVER), fmt.Sprint(flag.IsTakeover))
	}
	if flag.RootPassword != nil {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_ROOT_PWD), *flag.RootPassword)
	}
	if flag.NeedStartOB {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_START_OB))
	}
	if flag.NeedBeCluster {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_NEED_BE_CLUSTER))
	}
	return args
}

func (flag *CommonFlag) GetRootPassword() *string {
	flag.HiddenPassword()
	return flag.RootPassword
}

func (flag *CommonFlag) HiddenPassword() {
	if flag.hiden {
		return
	}

	password := string([]byte(flag.rootPassword)) // Deep copy the password to avoid being modified by hiddenPassword.
	if hiddenPassword(fmt.Sprintf("--%s", constant.FLAG_ROOT_PWD), fmt.Sprintf("--%s", constant.FLAG_ROOT_PWD_SH)) {
		flag.RootPassword = &password
	}
	flag.hiden = true
}

func hiddenPassword(flags ...string) bool {
	hiden := false
	for idx, arg := range os.Args {
		if hiden {
			maskArgs(idx)
			return true
		} else if utils.ContainsString(flags, arg) {
			hiden = true
			maskArgs(idx)
		} else if utils.ContainsPrefix(flags, arg) {
			maskArgs(idx)
			return true
		}
	}
	return hiden
}

func maskArgs(idx int) {
	baseAddr := uintptr(unsafe.Pointer(&os.Args[idx]))
	argPtr := (*[]byte)(unsafe.Pointer(baseAddr))
	oldLen := len(os.Args[idx])
	for i := 0; i < oldLen; i++ {
		(*argPtr)[i] = 0
	}
}
