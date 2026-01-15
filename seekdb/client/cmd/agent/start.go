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

	"github.com/oceanbase/obshell/seekdb/agent/config"
	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/global"
	ocsagentlog "github.com/oceanbase/obshell/seekdb/agent/log"
)

func initStartCmd(c *cobra.Command) {
	c.Hidden = false
	c.Short = "Start the background running obshell agent process for OceanBase seekdb."
	c.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
		global.InitGlobalVariable()
	}

	c.Example = startCmdExample()
	hideFlags(c, constant.FLAG_PID)
}

func hideFlags(c *cobra.Command, names ...string) {
	for _, n := range names {
		c.Flags().MarkHidden(n)
	}
}

func startCmdExample() string {
	return `  obshell agent start -P 2886 --base-dir /var/lib/oceanbase
  obshell agent start -base-dir /var/lib/oceanbase`
}
