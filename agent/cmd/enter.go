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

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/meta"
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
}

func SetFlags(cmd *cobra.Command, flag *CommonFlag) {
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&flag.AgentInfo.Ip, constant.FLAG_IP, "", "The IP address for the agent to bind to")
	cmd.Flags().IntVarP(&flag.AgentInfo.Port, constant.FLAG_PORT, constant.FLAG_PORT_SH, 0, "The operations port number")
	cmd.Flags().Int32Var(&flag.OldServerPid, constant.FLAG_PID, 0, "Old obshell pid, only used for upgrade")
	cmd.Flags().MarkHidden(constant.FLAG_PID)
	cmd.Flags().IntVar(&flag.IsTakeover, constant.FLAG_TAKE_OVER, 1, "If the agent is started for a takeover")
	cmd.Flags().BoolVar(&flag.NeedStartOB, constant.FLAG_START_OB, false, "If need to start observer")
	cmd.Flags().BoolVar(&flag.NeedBeCluster, constant.FLAG_NEED_BE_CLUSTER, false, "If need to be a cluster agent")
	cmd.Flags().MarkHidden(constant.FLAG_START_OB)
}

func (flag *CommonFlag) GetArgs() (args []string) {
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
	if flag.NeedStartOB {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_START_OB))
	}
	if flag.NeedBeCluster {
		args = append(args, fmt.Sprintf("--%s", constant.FLAG_NEED_BE_CLUSTER))
	}
	return args
}
