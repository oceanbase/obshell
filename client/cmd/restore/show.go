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

package restore

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/cmd/tenant"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
)

type ShowFlags struct {
	TenantName string
	verbose    bool
	detail     bool
}

func newShowCmd() *cobra.Command {
	opts := &ShowFlags{}
	showCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_SHOW,
		Short:   "Displays the restore status for the specific tenant.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			if err := show(opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: showCmdExample(),
	})

	showCmd.Flags().SortFlags = false
	showCmd.VarsPs(&opts.TenantName, []string{tenant.FLAG_TENANT_NAME, tenant.FLAG_TENANT_NAME_SH}, "", "The name of the tenant to show backup jobs.", true)
	showCmd.VarsPs(&opts.detail, []string{clientconst.FLAG_DETAIL, clientconst.FLAG_DETAIL_SH}, false, "Display detailed information.", false)
	showCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output.", false)

	return showCmd.Command
}

func show(opts *ShowFlags) error {
	stdio.Verbosef("Get the restore status for tenant %s", opts.TenantName)
	overview, err := api.GetTenantRestoreOverview(opts.TenantName)
	if err != nil {
		return err
	}

	if opts.detail {
		printer.PrintDetailedTenantRestoreOverview(overview)
	} else {
		printer.PrintTenantRestoreOverview(overview)
	}

	return nil

}

func showCmdExample() string {
	return `  obshell restore show -t tenant1
`
}
