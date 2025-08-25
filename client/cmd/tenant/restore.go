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
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/client/cmd/tenant/replica"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type TenantRestoreFlags struct {
	TenantName string

	DataBackupUri string
	ArchiveLogUri string

	Timestamp           string `json:"timestamp" time_format:"2006-01-02T15:04:05.000Z07:00"`
	SCN                 int64
	PrimaryZone         string
	Concurrency         string
	HaHighThreadScore   string
	Decryption          string
	IncBackupDecryption string
	KmsEncryptInfo      string

	verbose     bool
	skipConfirm bool

	replica.ZoneParamsFlags
}

func newRestoreCmd() *cobra.Command {
	opts := &TenantRestoreFlags{}
	restoreCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_RESTORE,
		Short:   "Restore tenant from backup",
		PreRunE: cmdlib.ValidateArgTenantName,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)

			opts.TenantName = args[0]
			return tenantRestore(cmd, opts)
		}),
		Example: RestoreCmdExample(),
	})

	restoreCmd.Flags().SortFlags = false
	restoreCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	restoreCmd.VarsPs(&opts.DataBackupUri, []string{FLAG_DATA_BACKUP_URI, FLAG_DATA_BACKUP_URI_SH}, "", "The directory path where the backups are stored.", true)

	restoreCmd.VarsPs(&opts.Zones, []string{FLAG_ZONE, FLAG_ZONE_SH}, "", "The zones of the tenant.", false)
	restoreCmd.VarsPs(&opts.UnitNum, []string{FLAG_UNIT_NUM}, 1, "The number of units in each zone", false)
	restoreCmd.VarsPs(&opts.UnitConfigName, []string{FLAG_UNIT, FLAG_UNIT_SH}, "", "The unit config name.", false)
	restoreCmd.VarsPs(&opts.ReplicaType, []string{FLAG_REPLICA_TYPE}, "", "The replica type of the tenant.", false)
	restoreCmd.VarsPs(&opts.PrimaryZone, []string{FLAG_PRIMARY_ZONE, FLAG_PRIMARY_ZONE_SH}, "", "The primary zone of the tenant to be restored.", false)

	restoreCmd.VarsPs(&opts.Timestamp, []string{FLAG_TIMESTAMP, FLAG_TIMESTAMP_SH}, "", "The timestamp to restore to.", false)
	restoreCmd.VarsPs(&opts.SCN, []string{FLAG_SCN, FLAG_SCN_SH}, int64(0), "The SCN to restore to.", false)
	restoreCmd.VarsPs(&opts.ArchiveLogUri, []string{FLAG_ARCHIVE_LOG_URI, FLAG_ARCHIVE_LOG_URI_SH}, "", "The directory path where the archive logs are stored.", false)
	restoreCmd.VarsPs(&opts.HaHighThreadScore, []string{FLAG_HA_HIGH_THREAD_SCORE, FLAG_HA_HIGH_THREAD_SCORE_SH}, "", "The high thread score for HA. Range: [0, 100]", false)
	restoreCmd.VarsPs(&opts.Concurrency, []string{FLAG_CONCURRENCY, FLAG_CONCURRENCY_SH}, "", "The number of threads to use for the restore operation.", false)
	restoreCmd.VarsPs(&opts.Decryption, []string{FLAG_DECRYPTION, FLAG_DECRYPTION_SH}, "", "The decryption password for all backups.", false)
	restoreCmd.VarsPs(&opts.KmsEncryptInfo, []string{FLAG_KMS_ENCRYPT_INFO, FLAG_KMS_ENCRYPT_INFO_SH}, "", "The KMS encryption information.", false)

	restoreCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt", false)
	restoreCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return restoreCmd.Command
}

func tenantRestore(cmd *cobra.Command, opts *TenantRestoreFlags) error {
	param, err := opts.toRestoreParam(cmd)
	if err != nil {
		return err
	}

	if err = ConfirmRestore(); err != nil {
		return err
	}

	uri := constant.URI_TENANT_API_PREFIX + constant.URI_RESTORE
	dag, err := api.CallApiAndPrintStage(uri, param)
	if err != nil {
		return err
	}
	log.Info("Restore tenant successfully, DAG ID: ", dag.DagID)
	return nil
}

func ConfirmRestore() error {
	msg := "Please confirm if you need to restore the tenant"
	res, err := stdio.Confirm(msg)
	if err != nil {
		return errors.Wrap(err, "ask for restore confirmation failed")
	}
	if !res {
		return errors.Occur(errors.ErrCliOperationCancelled)
	}
	return nil
}

func (f *TenantRestoreFlags) toRestoreParam(cmd *cobra.Command) (*param.RestoreParam, error) {
	zoneList, err := replica.BuildZoneParams(cmd, &f.ZoneParamsFlags)
	if err != nil {
		return nil, err
	}

	restoreParam := &param.RestoreParam{
		TenantName: f.TenantName,
		RestoreStorageParam: param.RestoreStorageParam{
			DataBackupUri: f.DataBackupUri,
		},
		ZoneList: zoneList,
	}
	stdio.Verbosef("Zone list is %v", restoreParam.ZoneList)

	if f.Timestamp != "" {
		timestamp, err := time.Parse(time.RFC3339, f.Timestamp)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid timestamp")
		}
		restoreParam.Timestamp = &timestamp
	}

	if f.SCN != 0 {
		restoreParam.SCN = &f.SCN
	}

	if f.ArchiveLogUri != "" {
		restoreParam.ArchiveLogUri = &f.ArchiveLogUri
	}

	if f.HaHighThreadScore != "" {
		haHighThreadScore, err := strconv.Atoi(f.HaHighThreadScore)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid haHighThreadScore")
		}
		restoreParam.HaHighThreadScore = &haHighThreadScore
	}

	if f.Concurrency != "" {
		concurrency, err := strconv.Atoi(f.Concurrency)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid concurrency")
		}
		restoreParam.Concurrency = &concurrency
	}

	if f.Decryption != "" {
		pwds := (strings.Split(strings.TrimSpace(f.Decryption), ","))
		restoreParam.Decryption = &pwds
	}

	if f.KmsEncryptInfo != "" {
		restoreParam.KmsEncryptInfo = &f.KmsEncryptInfo
	}

	return restoreParam, nil
}

func RestoreCmdExample() string {
	return `  Initiating a restore operation to a specific time, using a previously configured backup path:
    obshell tenant restore mytenant --timestamp "2021-01-01T00:00:00.000+08:00" -z "zone1,zone2,zone3" -d '/path/to/backup/data' -a '/path/to/backup/clog' -u unit1
	`
}
