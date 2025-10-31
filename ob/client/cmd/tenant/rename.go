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
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
)

func newRenameCmd() *cobra.Command {
	verbose := false
	renameCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_RENAME,
		Short: "Rename a tenant.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.Occur(errors.ErrCliUsageError, "tenant name is required")
			}
			if len(args) < 2 {
				return errors.Occur(errors.ErrCliUsageError, "new name is required")
			}
			stdio.SetVerboseMode(verbose)
			return tenantRename(args[0], args[1])
		}),
		Example: `  obshell tenant rename t1 t2`,
	})
	renameCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name> <new-name>"}
	renameCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return renameCmd.Command
}

func tenantRename(name string, newName string) error {
	// Rename tenant
	params := map[string]string{
		"new_name": newName,
	}
	stdio.StartLoadingf("rename tenant %s to %s", name, newName)
	if err := api.CallApiWithMethod(http.PUT, constant.URI_TENANT_API_PREFIX+"/"+name+constant.URI_NAME, params, nil); err != nil {
		return err
	}
	stdio.LoadSuccessf("rename tenant %s to %s", name, newName)
	return nil
}
