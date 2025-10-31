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

package pool

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	"github.com/oceanbase/obshell/ob/client/global"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
)

func newDropCmd() *cobra.Command {
	opts := &global.DropFlags{}
	dropCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_DROP,
		Short: "Drop a resource pool.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				return errors.Occur(errors.ErrCliUsageError, "resource pool name is required")
			}
			stdio.SetSkipConfirmMode(opts.SkipConfirm)
			stdio.SetVerboseMode(opts.Verbose)
			return rpDrop(args[0], opts)
		}),
		Example: `  obshell rp drop p1`,
	})

	dropCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<resource-pool-name>"}
	dropCmd.Flags().SortFlags = false
	dropCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	dropCmd.VarsPs(&opts.SkipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation of drop resource pool operation", false)
	return dropCmd.Command
}

func rpDrop(name string, opts *global.DropFlags) error {
	pass, err := stdio.Confirmf("Please confirm if you need to drop resource pool %s", name)
	if err != nil {
		return errors.Wrap(err, "ask for confirmation failed")
	}
	if !pass {
		return nil
	}
	// Drop rp
	stdio.StartLoadingf("drop resource pool %s", name)
	if err := api.CallApiWithMethod(http.DELETE, constant.URI_POOL_API_PREFIX+"/"+name, nil, nil); err != nil {
		return err
	}
	stdio.LoadSuccessf("drop resource pool %s", name)
	return nil
}
