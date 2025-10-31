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

	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	cmdlib "github.com/oceanbase/obshell/ob/client/lib/cmd"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/client/utils/printer"
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
		PreRunE: cmdlib.ValidateArgTenantName,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)

			opts.TenantName = args[0]
			return show(opts)
		}),
		Example: showCmdExample(),
	})

	showCmd.Flags().SortFlags = false
	showCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
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
	return `  obshell restore show tenant1
`
}
