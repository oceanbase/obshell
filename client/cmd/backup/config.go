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
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/client/cmd/cluster"
	"github.com/oceanbase/obshell/client/cmd/tenant"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
)

type SetConfigFlags struct {
	tenant.TenantBackupFlags
}

func NewSetConfigCmd() *cobra.Command {
	opts := &SetConfigFlags{}
	setConfigCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_SET_CONFIG,
		Short:   "Set the backup configuration for the OceanBase cluster.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.Verbose)
			stdio.SetSilenceMode(false)
			stdio.SetSkipConfirmMode(opts.SkipConfirm)
			return SetbackupConfig(opts)
		}),
		Example: setConfigCmdExample(),
	})

	setConfigCmd.Flags().SortFlags = false

	setConfigCmd.VarsPs(&opts.TenantName, []string{tenant.FLAG_TENANT_NAME, tenant.FLAG_TENANT_NAME_SH}, "", "The name of the tenant to set the backup configuration for.", false)
	setConfigCmd.VarsPs(&opts.DataBackupUri, []string{tenant.FLAG_DATA_BACKUP_URI, tenant.FLAG_DATA_BACKUP_URI_SH}, "", "The URI for the data backup. Only used when the tenant is set to backup.", false)
	setConfigCmd.VarsPs(&opts.ArchiveLogUri, []string{tenant.FLAG_ARCHIVE_LOG_URI, tenant.FLAG_ARCHIVE_LOG_URI_SH}, "", "The URI for the archive log. Only used when the tenant is set to backup.", false)

	setConfigCmd.VarsPs(&opts.BackupBaseUri, []string{cluster.FLAG_PATH, cluster.FLAG_PATH_SH}, "", "The base URI for the backup data. Only used when the tenant is not set.", false)

	setConfigCmd.VarsPs(&opts.LogArchiveConcurrency, []string{cluster.FLAG_LOG_ARCHIVE_CONCURRENCY, cluster.FLAG_LOG_ARCHIVE_CONCURRENCY_SH}, "", "Configure the total number of working threads for log archiving.", false)
	setConfigCmd.VarsPs(&opts.Binding, []string{cluster.FLAG_BINDING, cluster.FLAG_BINDING_SH}, "", fmt.Sprintf("Set the archiving and business priority mode. Supports '%s' and '%s' modes. Defaults: '%s'.", constant.BINDING_MODE_OPTIONAL, constant.BINDING_MODE_MANDATORY, constant.BINDING_MODE_OPTIONAL), false)
	setConfigCmd.VarsPs(&opts.HaLowThreadScore, []string{cluster.FLAG_HA_LOW_THREAD_SCORE, cluster.FLAG_HA_LOW_THREAD_SCORE_SH}, "", "Specifies the number of current working threads for low-priority threads.", false)
	setConfigCmd.VarsPs(&opts.PieceSwitchInterval, []string{cluster.FLAG_PIECE_SWITCH_INTERVAL, cluster.FLAG_PIECE_SWITCH_INTERVAL_SH}, "", "Configure the piece switch interval. Range: [1d, 7d].", false)
	setConfigCmd.VarsPs(&opts.ArchiveLagTarget, []string{cluster.FLAG_ARCHIVE_LAG_TARGET, cluster.FLAG_ARCHIVE_LAG_TARGET_SH}, "", "Sets the target lag time for log archiving processes", false)
	setConfigCmd.VarsPs(&opts.Policy, []string{cluster.FLAG_DELETE_POLICY, cluster.FLAG_DELETE_POLICY_SH}, "", fmt.Sprintf("Policy for deleting backup data. Only supports '%s'.", constant.DELETE_POLICY_DEFAULT), false)
	setConfigCmd.VarsPs(&opts.RecoveryWindow, []string{cluster.FLAG_RECOVERY_WINDOW, cluster.FLAG_RECOVERY_WINDOW_SH}, "", "Defines the recovery window for which data delete policies apply.", false)

	setConfigCmd.VarsPs(&opts.SkipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt", false)
	setConfigCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return setConfigCmd.Command
}

func SetbackupConfig(opts *SetConfigFlags) error {
	// Confirm the operation.
	res, err := stdio.Confirm("Please confirm if you need to set the specified configuration for backup.")
	if err != nil {
		return errors.Wrap(err, "ask for confirmation failed")
	}
	if !res {
		return errors.Occur(errors.ErrCliOperationCancelled)
	}

	// Set the backup configuration.
	if opts.TenantName != "" {
		stdio.Verbosef("Setting the backup configuration for tenant %s", opts.TenantName)

		if opts.BackupBaseUri != "" {

			return errors.Occur(errors.ErrCliUsageError, "'backup_base_uri' is not required when setting the backup configuration for a specific tenant")
		}

		// Set the backup configuration for the tenant.
		tenantBackupConfigParam, err := opts.ToTenantBackupConfigParam()
		if err != nil {
			return err
		}

		uri := fmt.Sprintf("%s/%s%s%s", constant.URI_TENANT_API_PREFIX, opts.TenantName, constant.URI_BACKUP, constant.URI_CONFIG)
		dag, err := api.CallPatchApiAndPrintStage(uri, tenantBackupConfigParam)
		if err != nil {
			return err
		}
		log.Infof("updated backup config %s", dag.GenericID)

	} else {
		stdio.Verbose("Setting the backup configuration for the entire cluster")
		if opts.DataBackupUri != "" || opts.ArchiveLogUri != "" {
			return errors.Occur(errors.ErrCliUsageError, "'data_backup_uri' and 'archive_log_uri' are not required when setting the backup configuration for the entire cluster")
		}

		// Set the backup configuration for the cluster.
		backupConfigParam, err := opts.ToClusterBackupConfigParam()
		if err != nil {
			return err
		}

		uri := constant.URI_OBCLUSTER_API_PREFIX + constant.URI_BACKUP + constant.URI_CONFIG
		dag, err := api.CallPatchApiAndPrintStage(uri, backupConfigParam)
		if err != nil {
			return err
		}
		log.Infof("updated backup config %s", dag.GenericID)

	}
	return nil
}

func setConfigCmdExample() string {
	return `  Specifying the backup configuration for the entire cluster:
    obshell backup set-config --path /path/to/backup --logArchiveConcurrency 4 --binding optional

  Specifying the backup configuration for a specific tenant:
    obshell backup set-config -t tenant1 -d /path/to/backup/data -a /path/to/backup/archive`
}
