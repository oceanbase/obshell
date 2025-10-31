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

package admin

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/cmd"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/client/command"
)

const (
	WAIT_DAEMON_TIME_LIMIT = 100
)

// NewAdminCmd returns the admin command,
// admin command is used to start, stop and restart daemon.
// They are hidden commands, and used by obshell.
func NewAdminCmd() *cobra.Command {
	adminCmd := command.NewCommand(&cobra.Command{
		Use:    cmd.CMD_ADMIN,
		Hidden: true,
		Args:   cobra.NoArgs,
	})
	adminCmd.AddCommand(newStartCmd(), newStopCmd(), newRestartCmd())
	return adminCmd.Command
}

type Admin struct {
	daemonPid    int32
	oldServerPid int32
	isTakeover   int
	agent        meta.AgentInfoInterface
	upgradeMode  bool
	flags        *cmd.CommonFlag
}

func NewAdmin(flag *cmd.CommonFlag) *Admin {
	if flag == nil {
		return &Admin{
			isTakeover: 1,
		}
	}
	return &Admin{
		agent:        &flag.AgentInfo,
		oldServerPid: flag.OldServerPid,
		isTakeover:   flag.IsTakeover,
		flags:        flag,
	}
}
