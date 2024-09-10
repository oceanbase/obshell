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
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
)

var header = []string{"Name", "OriginalName", "Can UnDrop", "Can Purge"}

func newShowCmd() *cobra.Command {
	verbose := false
	showCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SHOW,
		Short: "Show tenant in recyclebin.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(verbose)
			if err := tenantShow(args...); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: `  obshell recyclebin tenant show t1
  obshell recyclebin tenant show '__recycle_$_1_1720679549921648'
  obshell recyclebin tenant show`,
	})
	showCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "[tenant-name|object-name]"}
	showCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return showCmd.Command
}

func tenantShow(name ...string) error {
	tenants := make([]oceanbase.DbaRecyclebin, 0)
	data := make([][]string, 0)
	// Get recyclebin tenant
	if err := api.CallApiWithMethod(http.GET, constant.URI_API_V1+constant.URI_RECYCLEBIN_GROUP+constant.URI_TENANTS_GROUP, nil, &tenants); err != nil {
		return err
	}
	if len(name) > 0 {
		for _, n := range tenants {
			if n.Name == name[0] || n.OriginalName == name[0] {
				data = append(data, []string{n.Name, n.OriginalName, n.CanUndrop, n.CanPurge})
			}
		}
		if len(data) == 0 {
			return fmt.Errorf("Tenant '%s' is not exsit in recyclebin", name[0])
		}
	} else {
		for _, n := range tenants {
			data = append(data, []string{n.Name, n.OriginalName, n.CanUndrop, n.CanPurge})
		}
	}
	stdio.PrintTable(header, data)

	return nil
}
