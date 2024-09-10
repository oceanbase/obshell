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

package unit

import (
	"fmt"
	"strconv"
	"time"

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

var header = []string{"Create Time", "Name", "Memory Size", "Max Cpu", "Min Cpu", "Log Disk Size", "Max Iops", "Min Iops"}

func newShowCmd() *cobra.Command {
	var verbose bool
	showCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SHOW,
		Short: "Show resource unit config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(verbose)
			if err := unitConfigShow(args...); err != nil {
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
	})
	showCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "[unit-config-name]"}
	showCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Show verbose output.", false)
	return showCmd.Command
}

func transferCapacity(capacity int) string {
	var cap = []string{"B", "KB", "MB", "GB", "TB", "PB"}
	i := 0
	for capacity >= 1024 {
		if capacity%1024 != 0 {
			break
		}
		capacity /= 1024
		i++
	}
	return fmt.Sprint(capacity) + cap[i]
}

func unitConfigShow(name ...string) error {
	data := make([][]string, 0)
	if len(name) == 0 {
		unitConfigs := make([]oceanbase.DbaObUnitConfig, 0)
		// show all
		stdio.StartLoading("Get all unit configs")
		if err := api.CallApiWithMethod(http.GET, constant.URI_API_V1+constant.URI_UNITS_GROUP, nil, &unitConfigs); err != nil {
			stdio.LoadFailedWithoutMsg()
			return err
		}
		stdio.LoadSuccess("Get all unit configs")
		for _, unitConfig := range unitConfigs {
			data = append(data, []string{unitConfig.GmtCreate.Format(time.DateTime), unitConfig.Name, transferCapacity(unitConfig.MemorySize), fmt.Sprint(unitConfig.MaxCpu), fmt.Sprint(unitConfig.MinCpu), transferCapacity(unitConfig.LogDiskSize), strconv.Itoa(unitConfig.MaxIops), strconv.Itoa(unitConfig.MinIops)})
		}
	} else {
		var unitConfig oceanbase.DbaObUnitConfig
		stdio.StartLoadingf("Get unit config %s", name[0])
		if err := api.CallApiWithMethod(http.GET, constant.URI_UNIT_GROUP_PREFIX+"/"+name[0], nil, &unitConfig); err != nil {
			stdio.LoadFailedWithoutMsg()
			return err
		}
		stdio.LoadSuccessf("Get unit config %s", name[0])
		data = append(data, []string{unitConfig.GmtCreate.Format(time.DateTime), unitConfig.Name, transferCapacity(unitConfig.MemorySize), fmt.Sprint(unitConfig.MaxCpu), fmt.Sprint(unitConfig.MinCpu), transferCapacity(unitConfig.LogDiskSize), strconv.Itoa(unitConfig.MaxIops), strconv.Itoa(unitConfig.MinIops)})
	}
	stdio.PrintTable(header, data)
	return nil
}
