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
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
)

type CancelFlags struct {
	TenantName string

	verbose     bool
	skipConfirm bool
}

func newCancelCmd() *cobra.Command {
	opts := &CancelFlags{}
	cancelCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_CANCEL,
		Short:   "Cancel the restore task for the specific tenant.",
		PreRunE: cmdlib.ValidateArgTenantName,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetSilenceMode(false)

			opts.TenantName = args[0]
			if err := cancel(opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: cancelCmdExample(),
	})

	cancelCmd.Flags().SortFlags = false
	cancelCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	cancelCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output.", false)
	cancelCmd.VarsPs(&opts.skipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt.", false)

	return cancelCmd.Command
}

func cancel(opts *CancelFlags) error {
	// confirm the operation
	skip, err := stdio.Confirmf("Please confirm if you need to cancel the restore task for tenant '%s'", opts.TenantName)
	if err != nil {
		return err
	}
	if !skip {
		return errors.New("Operation canceled")
	}

	var dag task.DagDetailDTO
	url := fmt.Sprintf("%s/%s%s", constant.URI_TENANT_API_PREFIX, opts.TenantName, constant.URI_RESTORE)
	if err = api.CallApiWithMethod(http.DELETE, url, nil, &dag); err != nil {
		return err
	}
	if dag.GenericDTO == nil {
		stdio.Infof("There is no restore task for tenant '%s'.", opts.TenantName)
		return nil
	}
	return api.NewDagHandler(&dag).PrintDagStage()
}

func cancelCmdExample() string {
	return `  # Cancel the restore task for the specific tenant.
	obshell restore cancel tenant1`
}
