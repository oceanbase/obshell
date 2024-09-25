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

package tenant

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type TenantBackupFlags struct {
	TenantName    string
	DataBackupUri string
	ArchiveLogUri string

	cluster.ClusterBackupFlags
}

func newBackupCmd() *cobra.Command {
	opts := &TenantBackupFlags{}
	backupCmd := command.NewCommand(&cobra.Command{
		Use:     cluster.CMD_BACKUP,
		Short:   "Backup the specified tenant.",
		PreRunE: cmdlib.ValidateArgTenantName,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.Verbose)
			stdio.SetSkipConfirmMode(opts.SkipConfirm)
			stdio.SetSilenceMode(false)

			opts.TenantName = args[0]
			if err := tenantBackup(opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: tenantBackupExample(),
	})

	backupCmd.Flags().SortFlags = false
	backupCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}

	backupCmd.VarsPs(&opts.DataBackupUri, []string{FLAG_DATA_BACKUP_URI, FLAG_DATA_BACKUP_URI_SH}, "", "The directory path where the backup will be stored.", false)
	backupCmd.VarsPs(&opts.ArchiveLogUri, []string{FLAG_ARCHIVE_LOG_URI, FLAG_ARCHIVE_LOG_URI_SH}, "", "The directory path where the archive logs will be stored.", false)

	backupCmd.VarsPs(&opts.Mode, []string{cluster.FLAG_BACKUP_MODE, cluster.FLAG_BACKUP_MODE_SH}, "", fmt.Sprintf("The backup mode: '%s' for incremental backup or '%s' for a full backup. Defaults: '%s'.", constant.BACKUP_MODE_INCREMENTAL, constant.BACKUP_MODE_FULL, constant.BACKUP_MODE_FULL), false)
	backupCmd.VarsPs(&opts.LogArchiveConcurrency, []string{cluster.FLAG_LOG_ARCHIVE_CONCURRENCY, cluster.FLAG_LOG_ARCHIVE_CONCURRENCY_SH}, "", "Configure the total number of working threads for log archiving.", false)
	backupCmd.VarsPs(&opts.Binding, []string{cluster.FLAG_BINDING, cluster.FLAG_BINDING_SH}, "", fmt.Sprintf("Set the archiving and business priority mode. Supports '%s' and '%s' modes. Defaults: '%s'.", constant.BINDING_MODE_OPTIONAL, constant.BINDING_MODE_MANDATORY, constant.BINDING_MODE_OPTIONAL), false)
	backupCmd.VarsPs(&opts.Encryption, []string{cluster.FLAG_ENCRYPTION, cluster.FLAG_ENCRYPTION_SH}, "", "The password for encrypting the backup set.", false)
	backupCmd.VarsPs(&opts.HaLowThreadScore, []string{cluster.FLAG_HA_LOW_THREAD_SCORE, cluster.FLAG_HA_LOW_THREAD_SCORE_SH}, "", "Specifies the number of current working threads for low-priority threads.", false)
	backupCmd.VarsPs(&opts.PieceSwitchInterval, []string{cluster.FLAG_PIECE_SWITCH_INTERVAL, cluster.FLAG_PIECE_SWITCH_INTERVAL_SH}, "", "Configure the piece switch interval. Range: [1d, 7d].", false)
	backupCmd.VarsPs(&opts.ArchiveLagTarget, []string{cluster.FLAG_ARCHIVE_LAG_TARGET, cluster.FLAG_ARCHIVE_LAG_TARGET_SH}, "", "Sets the target lag time for log archiving processes", false)
	backupCmd.VarsPs(&opts.PlusArchive, []string{cluster.FLAG_PLUS_ARCHIVE, cluster.FLAG_PLUS_ARCHIVE_SH}, false, "Whether to include archive logs within the backup process for a combined data and log backup.", false)
	backupCmd.VarsPs(&opts.Policy, []string{cluster.FLAG_DELETE_POLICY, cluster.FLAG_DELETE_POLICY_SH}, "", fmt.Sprintf("Policy for deleting backup data. Only supports '%s'.", constant.DELETE_POLICY_DEFAULT), false)
	backupCmd.VarsPs(&opts.RecoveryWindow, []string{cluster.FLAG_RECOVERY_WINDOW, cluster.FLAG_RECOVERY_WINDOW_SH}, "", "Defines the recovery window for which data delete policies apply.", false)

	backupCmd.VarsPs(&opts.SkipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt", false)
	backupCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return backupCmd.Command
}

func tenantBackup(opts *TenantBackupFlags) error {
	tenantBackupConfigParam, err := opts.ToTenantBackupConfigParam()
	if err != nil {
		return err
	}

	tenantBackupParam := opts.NewBackupParam()

	if err = cluster.ConfirmBackup(); err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/%s%s%s", constant.URI_TENANT_API_PREFIX, opts.TenantName, constant.URI_BACKUP, constant.URI_CONFIG)
	dag, err := api.CallPatchApiAndPrintStage(uri, tenantBackupConfigParam)
	if err != nil {
		return err
	}
	log.Infof("updated backup config %s", dag.GenericID)

	uri = fmt.Sprintf("%s/%s%s", constant.URI_TENANT_API_PREFIX, opts.TenantName, constant.URI_BACKUP)
	dag, err = api.CallApiAndPrintStage(uri, tenantBackupParam)
	if err != nil {
		return err
	}
	log.Info(dag)

	return nil
}

func (f *TenantBackupFlags) ToTenantBackupConfigParam() (*param.TenantBackupConfigParam, error) {
	backupConfigParam, err := (&f.BaseBackupConfigFlags).NewBackupConfigParam()
	if err != nil {
		return nil, err
	}
	res := &param.TenantBackupConfigParam{
		LogArchiveDestConf:    backupConfigParam.LogArchiveDestConf,
		BaseBackupConfigParam: backupConfigParam.BaseBackupConfigParam,
	}
	if f.ArchiveLogUri != "" {
		res.ArchiveBaseUri = &f.ArchiveLogUri
	}
	if f.DataBackupUri != "" {
		res.DataBaseUri = &f.DataBackupUri
	}
	return res, nil
}

func tenantBackupExample() string {
	return `  Triggering a full backup without specifying the path (assuming the path is already configured):
    obshell tenant backup t1 --backup_mode full

  Triggering an incremental backup, specifying the path and using encryption:
    obshell tenant backup t1 -d /path/to/backup/data --backup_mode incremental --encryption MySecretPassword`

}
