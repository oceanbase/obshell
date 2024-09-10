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
	"errors"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
)

func newRenameCmd() *cobra.Command {
	verbose := false
	renameCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_RENAME,
		Short: "Rename a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			if len(args) < 1 {
				stdio.Error("tenant name is required")
				cmd.SilenceUsage = false
				return errors.New("tenant name is required")
			}
			if len(args) < 2 {
				stdio.Error("new name is required")
				cmd.SilenceUsage = false
				return errors.New("new name is required")
			}
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(verbose)
			if err := tenantRename(args[0], args[1]); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: `  obshell tenant rename t1 new_name`,
	})
	renameCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	renameCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return renameCmd.Command
}

func tenantRename(name string, newName string) error {
	// Rename tenant
	params := map[string]string{
		"new_name": newName,
	}
	stdio.StartLoadingf("Renaming tenant %s to %s", name, newName)
	if err := api.CallApiWithMethod(http.PUT, constant.URI_TENANT_API_PREFIX+"/"+name+constant.URI_NAME, params, nil); err != nil {
		stdio.LoadFailedWithoutMsg()
		return err
	}
	stdio.LoadSuccessf("Renaming tenant %s to %s", name, newName)
	return nil
}
