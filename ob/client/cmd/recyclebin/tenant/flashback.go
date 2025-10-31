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
	"github.com/oceanbase/obshell/ob/param"
)

func newFlashbackCmd() *cobra.Command {
	verbose := false
	var newName string
	flashbackCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_FLASHBACK,
		Short: "Flashback a tenant in recyclebin.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				return errors.Occur(errors.ErrCliUsageError, "tenant name is required")
			}
			stdio.SetVerboseMode(verbose)
			return tenantFlashback(cmd, args[0], newName)
		}),
		Example: `  obshell recyclebin tenant flashback t1
  obshell recyclebin tenant flashback t1 -n t2`,
	})

	flashbackCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name|object-name>"}
	flashbackCmd.Flags().SortFlags = false
	flashbackCmd.VarsPs(&newName, []string{FLAG_NEW_NAME_SH, FLAG_NEW_NAME}, "", "New tenant name", false)
	flashbackCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return flashbackCmd.Command
}

func tenantFlashback(cmd *cobra.Command, name string, newName string) error {
	params := param.FlashBackTenantParam{}
	if cmd.Flags().Changed(FLAG_NEW_NAME) || cmd.Flags().Changed(FLAG_NEW_NAME_SH) {
		params.NewName = &newName
		stdio.Verbosef("Flashback tenant %s to %s", name, newName)
	} else {
		stdio.Verbosef("Flashback tenant %s to %s", name, name)
	}
	// flashback tenant
	stdio.StartLoadingf("flashback tenant %s", name)
	if err := api.CallApiWithMethod(http.POST, constant.URI_API_V1+constant.URI_RECYCLEBIN_GROUP+constant.URI_TENANT_GROUP+"/"+name, params, nil); err != nil {
		return err
	}
	stdio.LoadSuccessf("flashback tenant %s", name)
	return nil
}
