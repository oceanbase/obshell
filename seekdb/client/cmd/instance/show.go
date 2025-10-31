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

package instance

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/http"
	"github.com/oceanbase/obshell/seekdb/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/seekdb/client/command"
	clientconst "github.com/oceanbase/obshell/seekdb/client/constant"
	cmdlib "github.com/oceanbase/obshell/seekdb/client/lib/cmd"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
	"github.com/oceanbase/obshell/seekdb/client/utils/api"
	"github.com/oceanbase/obshell/seekdb/client/utils/printer"
	obmodel "github.com/oceanbase/obshell/seekdb/model/observer"
)

type ClusterShowFlags struct {
	detail  bool
	verbose bool
}

func newShowCmd() *cobra.Command {
	opts := &ClusterShowFlags{}
	showCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_SHOW,
		Short:   "Show seekdb info.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			return clusterShow(opts)
		}),
		Example: showCmdExample(),
	})

	showCmd.Flags().SortFlags = false
	showCmd.VarsPs(&opts.detail, []string{clientconst.FLAG_DETAIL, clientconst.FLAG_DETAIL_SH}, false, "Display detailed information.", false)
	showCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output.", false)
	return showCmd.Command
}

func clusterShow(flags *ClusterShowFlags) error {
	stdio.StartLoading("Get ob info")
	obInfo, err := api.GetObserverInfo()
	if err != nil {
		return err
	}
	stdio.StopLoading()

	stdio.StartLoading("Get agent status")
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return err
	}
	stdio.StopLoading()

	overviewData := makeOverviewData(obInfo, agentStatus)
	agent2row := makeRowData(obInfo, agentStatus, flags.detail)

	printer.PrintShowTable(overviewData, agent2row, flags.detail)

	return nil
}

func makeOverviewData(obInfo *obmodel.ObserverInfo, agentStatus *http.AgentStatus) (overviewData printer.ShowOverviewData) {
	stdio.Verbosef("My observer's status is %s", oceanbase.OBStateShortMap[agentStatus.OBState])
	if agentStatus.OBState == oceanbase.STATE_CONNECTION_AVAILABLE {
		overviewData.Connected = true
		overviewData.Version = obInfo.Version
		overviewData.UnderMaintenance = checkOceanbaseMaintenance()
	}

	overviewData.AgentVersion = constant.VERSION_RELEASE
	overviewData.Name = obInfo.ClusterName
	return overviewData
}

func makeRowData(obInfo *obmodel.ObserverInfo, agentStatus *http.AgentStatus, detail bool) (agentRowData printer.ShowRowData) {
	agentRowData = printer.ShowRowData{
		AgentPort: agentStatus.Port,
		OBState:   "N/A",
	}
	if agentStatus != nil {
		agentRowData.OBState = oceanbase.OBStateShortMap[agentStatus.OBState]
		if detail {
			agentRowData.UnderMaintenance = agentStatus.UnderMaintenance
		}
	}
	agentRowData.ObserverInfo = *obInfo
	return agentRowData
}

func checkOceanbaseMaintenance() bool {
	defer stdio.StopLoading()
	stdio.StartLoading("Checking oceanbase maintenance")
	dag, err := api.GetObLastMaintenanceDag()
	if err != nil {
		if errors.IsTaskNotFoundErr(err) {
			stdio.Verbose("No oceanbase maintenance dag found")
			return false
		}
		stdio.Verbosef("check oceanbase maintenance failed, err: %s", err)
		return false
	}

	return !dag.IsSucceed()
}

func showCmdExample() string {
	return `  obshell seekdb show
  obshell seekdb show --port 2886`
}
