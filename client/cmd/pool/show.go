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
	"errors"
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

var header = []string{"Name", "Id", "Zone List", "Replica Type", "Unit Num", "Unit Config Id", "Tenant Id"}

func newShowCmd() *cobra.Command {
	var verbose bool
	showCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SHOW,
		Short: "Show resource pool.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(verbose)
			if err := rpShow(args...); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: ` obshell rp show`,
	})
	showCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "[resource-pool-name]"}
	showCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Show verbose output.", false)
	return showCmd.Command
}

func rpShow(name ...string) error {
	data := make([][]string, 0)
	rps := make([]oceanbase.DbaOBResourcePool, 0)
	// show all
	err := api.CallApiWithMethod(http.GET, constant.URI_API_V1+constant.URI_POOLS_GROUP, nil, &rps)
	if err != nil {
		return err
	}

	if len(name) != 0 {
		for _, rp := range rps {
			if rp.Name == name[0] {
				data = append(data, []string{rp.Name, fmt.Sprint(rp.Id), rp.ZoneList, rp.ReplicaType, fmt.Sprint(rp.UnitNum), fmt.Sprint(rp.UnitConfigId), fmt.Sprint(rp.TenantId)})
			}
		}
	} else {
		for _, rp := range rps {
			data = append(data, []string{rp.Name, fmt.Sprint(rp.Id), rp.ZoneList, rp.ReplicaType, fmt.Sprint(rp.UnitNum), fmt.Sprint(rp.UnitConfigId), fmt.Sprint(rp.TenantId)})
		}
	}
	if len(data) == 0 {
		return errors.New("no resource pool found")
	}
	stdio.PrintTable(header, data)
	return nil
}
