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

package cluster

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
	"github.com/oceanbase/obshell/param"
)

type ClusterShowFlags struct {
	detail  bool
	verbose bool
}

func newShowCmd() *cobra.Command {
	opts := &ClusterShowFlags{}
	showCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_SHOW,
		Short:   "Show OceanBase cluster info.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			if err := clusterShow(opts); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: showCmdExample(),
	})

	showCmd.Flags().SortFlags = false
	showCmd.VarsPs(&opts.detail, []string{clientconst.FLAG_DETAIL, clientconst.FLAG_DETAIL_SH}, false, "Display detailed information.", false)
	showCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output.", false)
	return showCmd.Command
}

func clusterShow(flags *ClusterShowFlags) error {
	obInfo, err := api.GetObInfo()
	if err != nil {
		return err
	}

	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return err
	}

	buildVersionConsistent := checkBuildVersion(obInfo)
	agentVersionConsistent := checkAgentVersion(obInfo)

	overviewData := makeOverviewData(obInfo, agentStatus, buildVersionConsistent)
	agent2row := makeRowData(obInfo, flags.detail)

	printer.PrintShowTable(overviewData, agent2row, flags.detail)

	if !buildVersionConsistent {
		printer.PrintWarnForBuildVersion(agent2row)
	}

	if !agentVersionConsistent {
		printer.PrintWarnForAgentVersion(obInfo.Agents)
	}

	return nil
}

func makeOverviewData(obInfo *param.ObInfoResp, agentStatus *http.AgentStatus, buildVersionConsistent bool) (overviewData printer.ShowOverviewData) {
	stdio.Verbosef("My observer's status is %s", oceanbase.OBStateShortMap[agentStatus.OBState])
	if agentStatus.OBState == oceanbase.STATE_CONNECTION_AVAILABLE {
		overviewData.Connected = true
		overviewData.ID = obInfo.Config.ClusterID
		overviewData.Name = obInfo.Config.ClusterName
		if buildVersionConsistent {
			for _, servers := range obInfo.Config.ZoneConfig {
				overviewData.Version = strings.Replace(servers[0].BuildVersion, "_", "-", 1)
				break
			}
		} else {
			overviewData.Version = getLocalBuildVersion(obInfo, &agentStatus.Agent)
		}
		overviewData.UnderMaintenance = checkOceanbaseMaintenance()
	}

	overviewData.AgentVersion = constant.VERSION_RELEASE
	return overviewData
}

func makeRowData(obInfo *param.ObInfoResp, detail bool) (agent2row map[meta.AgentInfo]printer.ShowRowData) {
	agent2row = make(map[meta.AgentInfo]printer.ShowRowData)
	for _, agent := range obInfo.Agents {
		row := printer.ShowRowData{
			AgentInstance: agent,
			OBState:       "N/A",
		}
		agent2row[agent.AgentInfo] = row
	}

	if detail {
		agentsStatus, err := api.GetAllAgentsStatus()
		if err != nil {
			stdio.Verbosef("get all agents status failed %s", err)
		}
		for key, status := range agentsStatus {
			aKey := *meta.NewAgentInfoByString(key)
			row, ok := agent2row[aKey]
			if !ok {
				stdio.Verbosef("unmatched api response, miss agent: %v", key)
			} else {
				row.OBState = oceanbase.OBStateShortMap[status.OBState]
				row.UnderMaintenance = status.UnderMaintenance
			}
			agent2row[aKey] = row
		}
	}

	for _, servers := range obInfo.Config.ZoneConfig {
		for _, server := range servers {
			if server == nil {
				continue
			}
			akey := meta.AgentInfo{Ip: server.SvrIP, Port: server.AgentPort}
			row, ok := agent2row[akey]
			if !ok {
				stdio.Verbosef("unmatched api response, miss agent: %v", akey)
			} else {
				row.ServerConfig = param.ServerConfig{
					SvrPort:      server.SvrPort,
					SqlPort:      server.SqlPort,
					WithRootSvr:  server.WithRootSvr,
					Status:       server.Status,
					BuildVersion: server.BuildVersion,
				}
			}
			agent2row[akey] = row
		}
	}
	return agent2row
}

func checkBuildVersion(obInfo *param.ObInfoResp) (res bool) {
	stdio.Verbose("Checking my obcluster's build version")
	m := make(map[string]struct{})
	for _, servers := range obInfo.Config.ZoneConfig {
		for _, server := range servers {
			if server == nil {
				continue
			}
			m[server.BuildVersion] = struct{}{}
		}
	}
	res = len(m) <= 1
	stdio.Verbosef("Build version consistent: %v", res)
	return
}

func checkAgentVersion(obInfo *param.ObInfoResp) (res bool) {
	stdio.Verbose("Checking my agent's build version")
	m := make(map[string]struct{})
	for _, a := range obInfo.Agents {
		m[a.Version] = struct{}{}
	}
	res = len(m) <= 1
	stdio.Verbosef("Agent version consistent: %v", res)
	return
}

func checkOceanbaseMaintenance() bool {
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

func getLocalBuildVersion(obInfo *param.ObInfoResp, agent meta.AgentInfoInterface) string {
	for _, servers := range obInfo.Config.ZoneConfig {
		for _, server := range servers {
			if server.SvrIP == agent.GetIp() && server.AgentPort == agent.GetPort() {
				stdio.Verbosef("My observer version is %s", server.BuildVersion)
				return server.BuildVersion
			}
		}
	}
	stdio.Verbose("My observer version is not found in the response")
	return ""
}

func showCmdExample() string {
	return `  obshell cluster show`
}
