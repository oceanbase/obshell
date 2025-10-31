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

package cluster

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	cmdlib "github.com/oceanbase/obshell/ob/client/lib/cmd"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/param"
)

type ClusterBackupFlags struct {
	BackupBaseUri string
	BaseBackupConfigFlags

	Verbose     bool
	SkipConfirm bool
}

type BaseBackupConfigFlags struct {
	Location            string
	Binding             string
	PieceSwitchInterval string

	LogArchiveConcurrency string
	ArchiveLagTarget      string
	Encryption            string
	HaLowThreadScore      string
	Policy                string
	RecoveryWindow        string

	Mode        string
	PlusArchive bool
}

func newBackupCmd() *cobra.Command {
	opts := &ClusterBackupFlags{}
	backupCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_BACKUP,
		Short:   "Initiate a backup operation with options to specify the backup mode and various other configurations.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.Verbose)
			stdio.SetSkipConfirmMode(opts.SkipConfirm)
			stdio.SetSilenceMode(false)
			return clusterBackup(opts)
		}),
		Example: backupCmdExample(),
	})

	backupCmd.Flags().SortFlags = false

	backupCmd.VarsPs(&opts.BackupBaseUri, []string{FLAG_PATH, FLAG_PATH_SH}, "", "The directory path where the backup and archive logs will be stored.", false)

	backupCmd.VarsPs(&opts.Mode, []string{FLAG_BACKUP_MODE, FLAG_BACKUP_MODE_SH}, "", fmt.Sprintf("The backup mode: '%s' for incremental backup or '%s' for a full backup. Defaults: '%s'.", constant.BACKUP_MODE_INCREMENTAL, constant.BACKUP_MODE_FULL, constant.BACKUP_MODE_FULL), false)
	backupCmd.VarsPs(&opts.LogArchiveConcurrency, []string{FLAG_LOG_ARCHIVE_CONCURRENCY, FLAG_LOG_ARCHIVE_CONCURRENCY_SH}, "", "Configure the total number of working threads for log archiving.", false)
	backupCmd.VarsPs(&opts.Binding, []string{FLAG_BINDING, FLAG_BINDING_SH}, "", fmt.Sprintf("Set the archiving and business priority mode. Supports '%s' and '%s' modes. Defaults: '%s'.", constant.BINDING_MODE_OPTIONAL, constant.BINDING_MODE_MANDATORY, constant.BINDING_MODE_OPTIONAL), false)
	backupCmd.VarsPs(&opts.Encryption, []string{FLAG_ENCRYPTION, FLAG_ENCRYPTION_SH}, "", "The password for encrypting the backup set.", false)
	backupCmd.VarsPs(&opts.HaLowThreadScore, []string{FLAG_HA_LOW_THREAD_SCORE, FLAG_HA_LOW_THREAD_SCORE_SH}, "", "Specifies the number of current working threads for low-priority threads.", false)
	backupCmd.VarsPs(&opts.PieceSwitchInterval, []string{FLAG_PIECE_SWITCH_INTERVAL, FLAG_PIECE_SWITCH_INTERVAL_SH}, "", "Configure the piece switch interval. Range: [1d, 7d].", false)
	backupCmd.VarsPs(&opts.ArchiveLagTarget, []string{FLAG_ARCHIVE_LAG_TARGET, FLAG_ARCHIVE_LAG_TARGET_SH}, "", "Sets the target lag time for log archiving processes", false)
	backupCmd.VarsPs(&opts.PlusArchive, []string{FLAG_PLUS_ARCHIVE, FLAG_PLUS_ARCHIVE_SH}, false, "Bool. Whether to include archive logs within the backup process for a combined data and log backup.", false)
	backupCmd.VarsPs(&opts.Policy, []string{FLAG_DELETE_POLICY, FLAG_DELETE_POLICY_SH}, "", fmt.Sprintf("Policy for deleting backup data. Only supports '%s'.", constant.DELETE_POLICY_DEFAULT), false)
	backupCmd.VarsPs(&opts.RecoveryWindow, []string{FLAG_RECOVERY_WINDOW, FLAG_RECOVERY_WINDOW_SH}, "", "Defines the recovery window for which data delete policies apply.", false)

	backupCmd.VarsPs(&opts.SkipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt", false)
	backupCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return backupCmd.Command
}

func clusterBackup(flags *ClusterBackupFlags) (err error) {
	clusterBackupConfigParam, err := flags.ToClusterBackupConfigParam()
	if err != nil {
		return err
	}
	clusterBackupParam := flags.NewBackupParam()

	if err = ConfirmBackup(); err != nil {
		return err
	}

	uri := constant.URI_OBCLUSTER_API_PREFIX + constant.URI_BACKUP + constant.URI_CONFIG
	dag, err := api.CallPatchApiAndPrintStage(uri, clusterBackupConfigParam)
	if err != nil {
		return err
	}
	log.Infof("updated backup config %s", dag.GenericID)

	uri = constant.URI_OBCLUSTER_API_PREFIX + constant.URI_BACKUP
	dag, err = api.CallApiAndPrintStage(uri, clusterBackupParam)
	if err != nil {
		return err
	}
	log.Infof("backup operation %s", dag.GenericID)
	return nil
}

func ConfirmBackup() error {
	msg := "Please confirm if you need to back up with the specified configuration"
	res, err := stdio.Confirm(msg)
	if err != nil {
		return errors.Wrap(err, "ask for backup confirmation failed")
	}
	if !res {
		return errors.Occur(errors.ErrCliOperationCancelled)
	}
	return nil
}

func (f *BaseBackupConfigFlags) NewBackupParam() *param.BackupParam {
	res := &param.BackupParam{}
	if f.Mode != "" {
		res.Mode = &f.Mode
	}
	if f.PlusArchive {
		res.PlusArchive = &f.PlusArchive
	}
	if f.Encryption != "" {
		res.Encryption = &f.Encryption
	}
	return res
}

func (f *BaseBackupConfigFlags) NewBackupConfigParam() (*param.BackupConfigParam, error) {
	res := &param.BackupConfigParam{}
	if f.Location != "" {
		res.Location = &f.Location
	}
	if f.Binding != "" {
		res.Binding = &f.Binding
	}
	if f.PieceSwitchInterval != "" {
		res.PieceSwitchInterval = &f.PieceSwitchInterval
	}
	if f.LogArchiveConcurrency != "" {
		val, err := strconv.Atoi(f.LogArchiveConcurrency)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid value for %s", FLAG_LOG_ARCHIVE_CONCURRENCY)
		}
		res.LogArchiveConcurrency = &val
	}
	if f.ArchiveLagTarget != "" {
		res.ArchiveLagTarget = &f.ArchiveLagTarget
	}
	if f.HaLowThreadScore != "" {
		val, err := strconv.Atoi(f.HaLowThreadScore)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid value for %s", FLAG_HA_LOW_THREAD_SCORE)
		}
		res.HaLowThreadScore = &val
	}
	if f.Policy != "" || f.RecoveryWindow != "" {
		res.DeletePolicy = &param.BackupDeletePolicy{
			Policy:         f.Policy,
			RecoveryWindow: f.RecoveryWindow,
		}
	}
	return res, nil
}

func (f *ClusterBackupFlags) ToClusterBackupConfigParam() (*param.ClusterBackupConfigParam, error) {
	backupConfigParam, err := f.BaseBackupConfigFlags.NewBackupConfigParam()
	if err != nil {
		return nil, err
	}
	res := &param.ClusterBackupConfigParam{
		LogArchiveDestConf:    backupConfigParam.LogArchiveDestConf,
		BaseBackupConfigParam: backupConfigParam.BaseBackupConfigParam,
	}
	if f.BackupBaseUri != "" {
		res.BackupBaseUri = &f.BackupBaseUri
	}
	return res, nil
}

func backupCmdExample() string {
	return `  Triggering a full backup without specifying the path (assuming the path is already configured):
    obshell cluster backup --backup_mode full

  Triggering an incremental backup, specifying the path and using encryption:
    obshell cluster backup -u /path/to/backup --backup_mode incremental --encryption MySecretPassword`
}
