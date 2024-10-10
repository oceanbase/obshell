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

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type ArchiveLogFlags struct {
	tenantName string
	status     string

	verbose     bool
	skipConfirm bool
}

func newArchiveLogCmd() *cobra.Command {
	opts := &ArchiveLogFlags{
		status: constant.ARCHIVELOG_STATUS_DOING,
	}
	archiveLogCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_ARCHIVE_LOG,
		Short:   "Open the archive log of the specified tenant.",
		PreRunE: cmdlib.ValidateArgTenantName,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)

			opts.tenantName = args[0]
			if err := tenantOperatorArchiveLog(opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: `  obshell tenant archivelog t1`,
	})

	archiveLogCmd.Flags().SortFlags = false
	archiveLogCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}

	archiveLogCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	archiveLogCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip confirmation", false)

	return archiveLogCmd.Command
}

func confirmArchiveLog() error {
	msg := "Are you sure you want to operator the archive log of the specified tenant?"
	res, err := stdio.Confirm(msg)
	if err != nil {
		return errors.Wrap(err, "ask for archivelog confirmation failed")
	}
	if !res {
		return errors.New("cancel backup")
	}
	return nil
}

func tenantOperatorArchiveLog(opts *ArchiveLogFlags) error {
	if err := confirmArchiveLog(); err != nil {
		return err
	}

	stdio.Infof("Operator the archive log of the specified tenant %s to %s", opts.tenantName, opts.status)
	param := param.ArchiveLogStatusParam{Status: &opts.status}
	uri := fmt.Sprintf("%s/%s%s%s", constant.URI_TENANT_API_PREFIX, opts.tenantName, constant.URI_BACKUP, constant.URI_ARCHIVE)
	if err := api.CallApiWithMethod(http.PATCH, uri, &param, nil); err != nil {
		return err
	}
	stdio.Successf("Operator the archive log of the specified tenant %s", opts.tenantName)
	return nil
}
