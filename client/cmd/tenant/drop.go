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
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/global"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type tenantDropFlags struct {
	needRecycle bool
	global.DropFlags
}

func newDropCmd() *cobra.Command {
	opts := &tenantDropFlags{}
	dropCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_DROP,
		Short: "Drop a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			if len(args) <= 0 {
				stdio.Error("tenant name is required")
				cmd.SilenceUsage = false
				return errors.New("tenant name is required")
			}
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetSkipConfirmMode(opts.SkipConfirm)
			stdio.SetVerboseMode(opts.Verbose)
			if err := tenantDrop(args[0], opts); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: `  obshell tenant drop t1 --recycle`,
	})
	dropCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	dropCmd.Flags().SortFlags = false
	dropCmd.VarsPs(&opts.needRecycle, []string{FLAG_RECYCLE}, false, "Drop tenant bu reserver resource pool.", false)
	dropCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	dropCmd.VarsPs(&opts.SkipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation of drop unit config operation", false)
	return dropCmd.Command
}

func tenantDrop(name string, opts *tenantDropFlags) error {
	pass, err := stdio.Confirmf("Please confirm if you need to drop tenant %s", name)
	if err != nil {
		return errors.New("ask for confirmation failed")
	}
	if !pass {
		return nil
	}
	param := buildDropTenantParams(name, opts)
	dag := task.DagDetailDTO{}
	if opts.IfExist {
		tenants := make([]oceanbase.DbaObUnitConfig, 0)
		// show all
		err := api.CallApiWithMethod(http.GET, constant.URI_API_V1+constant.URI_TENANTS_GROUP+constant.URI_OVERVIEW, nil, &tenants)
		if err != nil {
			return err
		}
		doDrop := false
		for _, tenant := range tenants {
			if tenant.Name == name {
				doDrop = true
			}
		}
		if !doDrop {
			return nil
		}
	}
	// Drop tenant
	if err := api.CallApiWithMethod(http.DELETE, constant.URI_TENANT_API_PREFIX+"/"+name, param, &dag); err != nil {
		return err
	}
	if dag.GenericDTO == nil {
		return nil
	}
	if err = api.NewDagHandler(&dag).PrintDagStage(); err != nil {
		return err
	}
	return nil
}

func buildDropTenantParams(name string, opts *tenantDropFlags) *param.DropTenantParam {
	params := param.DropTenantParam{}
	params.NeedRecycle = &opts.needRecycle
	return &params
}