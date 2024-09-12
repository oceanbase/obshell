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
	"time"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
)

var shortHeader = []string{"Name", "Id", "Mode", "Locality", "Primary Zone", "Status", "Locked"}
var longHeader = []string{"Name", "Id", "Mode", "Locality", "Primary Zone", "Status", "Unit Num(Each Zone)", "Unit Config", "Locked", "Whitelist", "Create Time"}

type tenantShowFlags struct {
	showDetail bool
	verbose    bool
}

func newShowCmd() *cobra.Command {
	opts := tenantShowFlags{}
	showCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SHOW,
		Short: "Show tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			if err := tenantShow(opts.showDetail, args...); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
	})
	showCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "[tenant-name]"}
	showCmd.VarsPs(&opts.showDetail, []string{clientconst.FLAG_DETAIL, clientconst.FLAG_DETAIL_SH}, false, "Show tenant detail.", false)
	showCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Show verbose output.", false)
	return showCmd.Command
}

func tenantShow(showDetails bool, name ...string) error {
	tenants := make([]oceanbase.DbaObTenant, 0)
	tenants = append(tenants, oceanbase.DbaObTenant{})
	if len(name) == 0 {
		// show all
		err := api.CallApiWithMethod(http.GET, constant.URI_API_V1+constant.URI_TENANTS_GROUP+constant.URI_OVERVIEW, nil, &tenants)
		if err != nil {
			return err
		}
	} else {
		err := api.CallApiWithMethod(http.GET, constant.URI_TENANT_API_PREFIX+"/"+name[0], nil, &tenants[0])
		if err != nil {
			return err
		}
	}
	data := make([][]string, 0)
	for _, tenant := range tenants {
		if !showDetails {
			data = append(data, []string{tenant.TenantName, fmt.Sprint(tenant.TenantID), tenant.Mode, tenant.Locality, tenant.PrimaryZone, tenant.Status, tenant.Locked})
		} else {
			info := bo.TenantInfo{}
			err := api.CallApiWithMethod(http.GET, constant.URI_TENANT_API_PREFIX+"/"+tenant.TenantName, nil, &info)
			if err != nil {
				return err
			}
			if len(info.Pools) == 0 {
				data = append(data, []string{tenant.TenantName, fmt.Sprint(tenant.TenantID), tenant.Mode, tenant.Locality, tenant.PrimaryZone, tenant.Status, tenant.Locked, info.Whitelist, "", ""})
			} else {
				unitConfigs := ""
				for _, pool := range info.Pools {
					unitConfigs += fmt.Sprintf("%s(%s);", pool.ZoneList, pool.Unit.Name)
				}
				data = append(data, []string{tenant.TenantName, fmt.Sprint(tenant.TenantID), tenant.Mode, tenant.Locality, tenant.PrimaryZone, tenant.Status, fmt.Sprint(info.Pools[0].UnitNum), unitConfigs, tenant.Locked, info.Whitelist, tenant.CreatedTime.Format(time.DateTime)})
			}
		}
	}
	if !showDetails {
		stdio.PrintTable(shortHeader, data)
	} else {
		stdio.PrintTable(longHeader, data)
	}
	return nil
}
