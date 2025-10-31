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

package replica

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	"github.com/oceanbase/obshell/ob/client/global"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/param"
)

type replicaDeleteFlags struct {
	zones string
	global.DropFlags
}

func newDeleteCmd() *cobra.Command {
	opts := &replicaDeleteFlags{}
	deleteCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_DELETE,
		Short: "Delete tenant replicas.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				return errors.Occur(errors.ErrCliUsageError, "tenant name is required")
			}
			stdio.SetVerboseMode(opts.Verbose)
			return replicaDelete(args[0], opts)
		}),
		Example: `  obshell tenant replica delete t1 -z zone3`,
	})

	deleteCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	deleteCmd.Flags().SortFlags = false
	deleteCmd.VarsPs(&opts.zones, []string{FLAG_ZONE, FLAG_ZONE_SH}, "", "The zones of the tenant.", true)
	deleteCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	deleteCmd.VarsPs(&opts.SkipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation of operation", false)

	return deleteCmd.Command
}

func replicaDelete(tenantName string, opts *replicaDeleteFlags) (err error) {
	zones := strings.Split(opts.zones, ",")
	params := param.ScaleInTenantReplicasParam{
		Zones: zones,
	}

	dag := task.DagDetailDTO{}
	if err := api.CallApiWithMethod(http.DELETE, constant.URI_TENANT_API_PREFIX+"/"+tenantName+constant.URI_REPLICAS, params, &dag); err != nil {
		return err
	}
	if dag.GenericDTO == nil {
		stdio.Infof("Tenant %s has no replicas in %s", tenantName, opts.zones)
		return nil
	}
	return api.NewDagHandler(&dag).PrintDagStage()
}
