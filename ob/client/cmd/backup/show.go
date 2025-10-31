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

package backup

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/client/cmd/tenant"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	cmdlib "github.com/oceanbase/obshell/ob/client/lib/cmd"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/client/utils/printer"
)

type BackupShowFlags struct {
	TenantName string
	verbose    bool
	detail     bool
}

func newShowCmd() *cobra.Command {
	opts := &BackupShowFlags{}
	showCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_SHOW,
		Short:   "Displays the backup status for the entire cluster or a specific tenant.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			return backupShow(opts)
		}),
		Example: showCmdExample(),
	})

	showCmd.Flags().SortFlags = false
	showCmd.VarsPs(&opts.TenantName, []string{tenant.FLAG_TENANT_NAME, tenant.FLAG_TENANT_NAME_SH}, "", "The name of the tenant to show backup jobs.", false)
	showCmd.VarsPs(&opts.detail, []string{clientconst.FLAG_DETAIL, clientconst.FLAG_DETAIL_SH}, false, "Display detailed information.", false)
	showCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output.", false)

	return showCmd.Command
}

func backupShow(opts *BackupShowFlags) error {
	if opts.TenantName != "" {

		overview, err := api.GetTenantBackupOverview(opts.TenantName)
		if err != nil {
			return err
		}

		log.Infof("Backup status for tenant %s: %#+v", opts.TenantName, *overview)
		if opts.detail {
			printer.PrintDetailedTenantBackupOverview(overview)
		} else {
			printer.PrintTenantBackupOverview(overview)
		}

	} else {

		overview, err := api.GetClusterBackupOverview()
		if err != nil {
			return err
		}
		log.Infof("Backup status for the entire cluster: %#+v", *overview)

		if opts.detail {
			printer.PrintDetailedClusterBackupOverview(overview)
		} else {
			printer.PrintClusterBackupOverview(overview)
		}

	}
	return nil

}

func showCmdExample() string {
	return `  Show the backup status for the entire cluster:
    obshell backup show -d

  Show the backup status for a specific tenant:
    obshell backup show -t tenant1 -d
`
}
