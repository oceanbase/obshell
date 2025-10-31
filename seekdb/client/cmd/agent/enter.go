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
	"github.com/spf13/cobra"

	agentcmd "github.com/oceanbase/obshell/seekdb/agent/cmd"
	"github.com/oceanbase/obshell/seekdb/agent/cmd/admin"
	"github.com/oceanbase/obshell/seekdb/client/constant"
)

// Upgrade command configuration constants.
const (
	CMD_UPGRADE         = "upgrade"
	FLAG_PKG_DIR        = "pkg_directory"
	FLAG_PKG_DIR_SH     = "d"
	FLAG_VERSION        = "target_version"
	FLAG_VERSION_SH     = "V"
	FLAG_UPGRADE_DIR    = "tmp_directory"
	FLAG_UPGRADE_DIR_SH = "t"
)

func NewAgentCmd() *cobra.Command {
	adminCmd := admin.NewAdminCmd()
	agentCmd := adminCmd
	agentCmd.Use = constant.CMD_AGENT
	agentCmd.Short = "Manage obshell agent."
	agentCmd.Hidden = false
	for _, c := range agentCmd.Commands() {
		switch c.Use {
		case agentcmd.CMD_START:
			initStartCmd(c)
		case agentcmd.CMD_STOP:
			initStopCmd(c)
		case agentcmd.CMD_RESTART:
			initRestartCmd(c)
		}
	}

	agentCmd.AddCommand(newUpgradeCmd())
	return agentCmd
}
