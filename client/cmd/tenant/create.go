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

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/client/cmd/tenant/parameter"
	"github.com/oceanbase/obshell/client/cmd/tenant/replica"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type TenantCreateFlags struct {
	mode         string
	charset      string
	collate      string
	primary_zone string
	info         string
	read_only    bool
	parameters   string
	variables    string
	whitelist    string
	scenario     string
	verbose      bool
	pwd          string
	importScript bool
	replica.ZoneParamsFlags
}

func newCreateCmd() *cobra.Command {
	opts := &TenantCreateFlags{}
	createCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_CREATE,
		Short: "Create a tenant.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				return errors.Occur(errors.ErrCliUsageError, "tenant name is required")
			}
			stdio.SetVerboseMode(opts.verbose)
			return tenantCreate(cmd, args[0], opts)
		}),
		Example: `  obshell tenant create t1 -u s1
  obshell tenant create t1 -z zone1,zone2,zone3
    --zone1.unit=s1 --zone2.unit=s2  --zone3.unit=s3
    --zone1.replica_type=FULL --zone2.replica_type=FULL --zone3.replica_type=READONLY
    --root_password 111`,
	})

	createCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	createCmd.Flags().SortFlags = false
	createCmd.VarsPs(&opts.Zones, []string{FLAG_ZONE, FLAG_ZONE_SH}, "", "The zones of the tenant.", false)
	createCmd.VarsPs(&opts.UnitNum, []string{FLAG_UNIT_NUM}, 1, "The number of units in each zone", false)
	createCmd.VarsPs(&opts.UnitConfigName, []string{FLAG_UNIT, FLAG_UNIT_SH}, "", "The unit config name.", false)
	createCmd.VarsPs(&opts.ReplicaType, []string{FLAG_REPLICA_TYPE}, "", "The replica type of the tenant.", false)
	createCmd.VarsPs(&opts.charset, []string{FLAG_CHARSET}, "", "Tenant charset. Default to utf8mb4.", false)
	createCmd.VarsPs(&opts.collate, []string{FLAG_COLLATE}, "", "Tenant collote. Default to utf8mb4_general_ci.", false)
	createCmd.VarsPs(&opts.primary_zone, []string{FLAG_PRIMARY_ZONE}, "", "Tenant primary zone. Default to 'RANDOM'.", false)
	createCmd.VarsPs(&opts.info, []string{FLAG_INFO}, "", "Tenant information.", false)
	createCmd.VarsPs(&opts.read_only, []string{FLAG_READ_ONLY}, false, "Whether the tenant is read-only.", false)
	createCmd.VarsPs(&opts.parameters, []string{FLAG_PARAMETERS}, "", "Tenant parameters.", false)
	createCmd.VarsPs(&opts.variables, []string{FLAG_VARIABLES}, "", "Tenant variables.", false)
	createCmd.VarsPs(&opts.whitelist, []string{FLAG_WHITELIST}, "", "Tenant whitelist.", false)
	createCmd.VarsPs(&opts.scenario, []string{FLAG_SCENARIO}, "", "Tenant scenario.", false)
	createCmd.VarsPs(&opts.pwd, []string{FLAG_ROOT_PASSWORD}, "", "Tenant password.", false)
	createCmd.VarsPs(&opts.importScript, []string{FLAG_IMPORT_SCRIPT}, false, "Import the observer's scripts for tenant.", false)
	createCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	createCmd.VarsPs(&opts.mode, []string{FLAG_MODE}, "", "Tenant mode. Default to 'MYSQL'.", false)

	createCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return replica.FlagErrorFunc(cmd, err, &opts.ZoneParamsFlags)
	})

	return createCmd.Command
}

func buildCreateTenantParams(cmd *cobra.Command, tenantName string, opts *TenantCreateFlags) (*param.CreateTenantParam, error) {
	params := &param.CreateTenantParam{
		Name:         &tenantName,
		Mode:         opts.mode,
		Charset:      opts.charset,
		Collation:    opts.collate,
		PrimaryZone:  opts.primary_zone,
		ReadOnly:     opts.read_only,
		Comment:      opts.info,
		Scenario:     opts.scenario,
		ImportScript: opts.importScript,
		RootPassword: opts.pwd,
	}

	if cmd.Flags().Changed(FLAG_WHITELIST) {
		params.Whitelist = &opts.whitelist
	}
	zoneList, err := replica.BuildZoneParams(cmd, &opts.ZoneParamsFlags)
	if err != nil {
		return nil, err
	}
	params.ZoneList = zoneList
	// get variables
	if varialbes, err := parameter.BuildVariableOrParameterMap(opts.variables); err != nil {
		return nil, err
	} else {
		params.Variables = varialbes
	}

	if variables, err := parameter.BuildVariableOrParameterMap(opts.parameters); err != nil {
		return nil, err
	} else {
		params.Parameters = variables
	}

	return params, err

}

func tenantCreate(cmd *cobra.Command, tenantName string, opts *TenantCreateFlags) (err error) {
	params, err := buildCreateTenantParams(cmd, tenantName, opts)
	if err != nil {
		return err
	}
	dag := task.DagDetailDTO{}
	if err := api.CallApiWithMethod(http.POST, constant.URI_TENANT_API_PREFIX, params, &dag); err != nil {
		return err
	}
	if err = api.NewDagHandler(&dag).PrintDagStage(); err != nil {
		return err
	}
	return nil
}
