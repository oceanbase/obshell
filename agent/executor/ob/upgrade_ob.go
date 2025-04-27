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

package ob

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/engine/coordinator"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/pkg"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
)

const (
	UPGRADE_CHECKER_FILE        = "etc/upgrade_checker.py"
	UPGRADE_HEALTH_CHECKER_FILE = "etc/upgrade_health_checker.py"
	UPGRADE_PRE_SCRIPT_FILE     = "etc/upgrade_pre.py"
	UPGRADE_POST_SCRIPT_FILE    = "etc/upgrade_post.py"
)

type obUpgradeParams struct {
	currentVersion  string
	targetVersion   string
	rollingUpgrade  bool
	upgradeRoute    []RouteNode
	dbaObZones      []oceanbase.DbaObZones
	allAgents       []meta.AgentInfo
	agentsInZoneMap map[string][]meta.AgentInfo
	agents          []meta.AgentInfo
	RequestParam    *param.ObUpgradeParam
}

func CheckAndUpgradeOb(param param.ObUpgradeParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	log.Info("check and upgrade ob")
	p, err := preCheckForObUpgrade(param)
	if err != nil {
		log.WithError(err).Error("pre check for ob upgrade failed")
		return nil, err
	}
	e := p.initParamsForObUpgrade()
	if e != nil {
		log.WithError(e).Error("init params for ob upgrade failed")
		return nil, errors.Occur(errors.ErrUnexpected, e)
	}

	checkAndUpgradeObTemplate := buildCheckAndUpgradeObTemplate(p)
	checkAndUpgradeObTaskContext := buildCheckAndUpgradeObTaskContext(p)
	checkAndUpgradeObDag, e := taskService.CreateDagInstanceByTemplate(checkAndUpgradeObTemplate, checkAndUpgradeObTaskContext)
	if e != nil {
		return nil, errors.Occur(errors.ErrUnexpected, e)
	}
	return task.NewDagDetailDTO(checkAndUpgradeObDag), nil
}

func buildCheckAndUpgradeObTaskContext(p *obUpgradeParams) *task.TaskContext {
	ctx := task.NewTaskContext()
	buildNumber, distribution, _ := pkg.SplitRelease(p.RequestParam.Release)
	taskTime := strconv.Itoa(int(time.Now().UnixMilli()))
	ctx.SetParam(PARAM_ALL_AGENTS, p.agents).
		SetParam(PARAM_UPGRADE_DIR, p.RequestParam.UpgradeDir).
		SetParam(PARAM_TASK_TIME, taskTime).
		SetParam(task.EXECUTE_AGENTS, p.agents).
		SetParam(PARAM_VERSION, p.RequestParam.Version).
		SetParam(PARAM_BUILD_NUMBER, buildNumber).
		SetParam(PARAM_DISTRIBUTION, distribution).
		SetParam(PARAM_RELEASE_DISTRIBUTION, p.RequestParam.Release).
		SetParam(PARAM_UPGRADE_ROUTE, p.upgradeRoute).
		SetParam(PARAM_ROLLING_UPGRADE, p.rollingUpgrade)
	return ctx
}

func buildCheckAndUpgradeObTemplate(p *obUpgradeParams) *task.Template {
	if p.rollingUpgrade {
		return buildObClusterCrossVeriosnRollingUpgradeTemplate(p)
	} else {
		return buildObClusterCrossVeriosnStopServiceUpgradeTemplate(p)
	}
}

func buildObClusterCrossVeriosnStopServiceUpgradeTemplate(p *obUpgradeParams) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_OB_STOP_SVC_UPGRADE, p.RequestParam.Version, p.RequestParam.Release)
	builder := task.NewTemplateBuilder(name).SetMaintenance(task.GlobalMaintenance()).
		AddTemplate(newCheckAndUpgradeAgentTemplate())
	for i := 0; i < len(p.upgradeRoute); i++ {
		builder.AddTemplate(newStopServiceUpgradeProcessTemplate(i, p))
	}
	builder.AddTask(newRemoveUpgradeCheckDirTask(), true)
	return builder.Build()
}

func newStopServiceUpgradeProcessTemplate(idx int, p *obUpgradeParams) *task.Template {
	name := fmt.Sprintf("Upgrade process %d", idx)
	return task.NewTemplateBuilder(name).
		AddTask(newBackupParametersTask(), false).
		AddNode(newExecUpgradeCheckerNode("", idx)).
		AddNode(newExecPreScriptNode("", idx)).
		AddNode(newExecHealthCheckerNode("", idx)).
		AddNode(newReinstallAndRestartObNode("", p.allAgents, idx)).
		AddNode(newExecHealthCheckerNode("", idx)).
		AddNode(newExecPostScriptNode("", idx)).
		AddTask(newRestoreParametersTask(), false).
		Build()
}

func newCheckAndUpgradeAgentTemplate() *task.Template {
	return task.NewTemplateBuilder(DAG_CHECK_AND_UPGRADE_OBSHELL).
		AddTask(newCheckEnvTask(), true).
		AddTemplate(buildAgentCheckAndUpgradeTemplate()).
		Build()
}

func buildObClusterCrossVeriosnRollingUpgradeTemplate(p *obUpgradeParams) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_OB_ROLLING_UPGRADE, p.RequestParam.Version, p.RequestParam.Release)
	builder := task.NewTemplateBuilder(name).SetMaintenance(task.GlobalMaintenance()).
		AddTemplate(newCheckAndUpgradeAgentTemplate())
	for i := 0; i < len(p.upgradeRoute); i++ {
		builder.AddTemplate(newRollingUpgradeProcessTemplate(i, p))
	}
	builder.AddTask(newRemoveUpgradeCheckDirTask(), true)
	return builder.Build()
}

func newRollingUpgradeProcessTemplate(idx int, p *obUpgradeParams) *task.Template {
	name := fmt.Sprintf("Upgrade process %d", idx)
	builder := task.NewTemplateBuilder(name)
	builder.
		AddTask(newBackupParametersTask(), false).
		AddNode(newExecUpgradeCheckerNode("", idx)).
		AddNode(newExecPreScriptNode("", idx))
	for _, dbaObZone := range p.dbaObZones {
		agents := p.agentsInZoneMap[dbaObZone.Zone]
		builder.
			AddNode(newExecHealthCheckerNode("", idx)).
			AddNode(newStopZoneNode(dbaObZone.Zone)).
			AddNode(newReinstallAndRestartObNode(dbaObZone.Zone, agents, idx)).
			AddNode(newExecZoneHealthCheckerNode(dbaObZone.Zone, idx)).
			AddNode(newStartZoneNode(dbaObZone.Zone))
	}
	builder.
		AddNode(newExecPostScriptNode("", idx)).
		AddTask(newRestoreParametersTask(), false)
	return builder.Build()
}

func newExecUpgradeCheckerNode(zone string, idx int) *task.Node {
	ctx := task.NewTaskContext()
	ctx.SetParam(PARAM_SCRIPT_FILE, UPGRADE_CHECKER_FILE).
		SetParam(PARAM_UPGRADE_ROUTE_INDEX, idx).
		SetParam(PARAM_ZONE, zone)
	return task.NewNodeWithContext(&ExecScriptTask{
		Task: *task.NewSubTask(TASK_EXEC_UPGRADE_CHECKER_SCRIPT).
			SetCanRetry().
			SetCanContinue()},
		false, ctx)
}

func newExecPreScriptNode(zone string, idx int) *task.Node {
	ctx := task.NewTaskContext()
	ctx.SetParam(PARAM_SCRIPT_FILE, UPGRADE_PRE_SCRIPT_FILE).
		SetParam(PARAM_UPGRADE_ROUTE_INDEX, idx).
		SetParam(PARAM_ZONE, zone)
	return task.NewNodeWithContext(&ExecScriptTask{
		Task: *task.NewSubTask(TASK_EXEC_UPGRADE_PRE_SCRIPT).
			SetCanRetry()},
		false, ctx)

}

func newExecPostScriptNode(zone string, idx int) *task.Node {
	ctx := task.NewTaskContext()
	ctx.SetParam(PARAM_SCRIPT_FILE, UPGRADE_POST_SCRIPT_FILE).
		SetParam(PARAM_UPGRADE_ROUTE_INDEX, idx).
		SetParam(PARAM_ZONE, zone)
	return task.NewNodeWithContext(&ExecScriptTask{
		Task: *task.NewSubTask(TASK_EXEC_UPGRADE_POST_SCRIPT).
			SetCanRetry().
			SetCanContinue()},
		false, ctx)
}

func newExecHealthCheckerNode(zone string, idx int) *task.Node {
	ctx := task.NewTaskContext()
	ctx.SetParam(PARAM_SCRIPT_FILE, UPGRADE_HEALTH_CHECKER_FILE).
		SetParam(PARAM_UPGRADE_ROUTE_INDEX, idx).
		SetParam(PARAM_ZONE, zone)
	return task.NewNodeWithContext(&ExecScriptTask{
		Task: *task.NewSubTask(TASK_EXEC_UPGRADE_HEALTH_CHECKER_SCRIPT).
			SetCanRetry().
			SetCanContinue()},
		false, ctx)
}

func newExecZoneHealthCheckerNode(zone string, idx int) *task.Node {
	ctx := task.NewTaskContext()
	ctx.SetParam(PARAM_SCRIPT_FILE, UPGRADE_HEALTH_CHECKER_FILE).
		SetParam(PARAM_UPGRADE_ROUTE_INDEX, idx).
		SetParam(PARAM_ZONE, zone)
	return task.NewNodeWithContext(&ExecScriptTask{
		Task: *task.NewSubTask(TASK_EXEC_UPGRADE_ZONE_HEALTH_CHECKER_SCRIPT).
			SetCanRetry().
			SetCanContinue()},
		false, ctx)
}

func (p *obUpgradeParams) initParamsForObUpgrade() (err error) {
	log.Info("init params for ob upgrade")
	p.targetVersion = p.RequestParam.Version
	p.rollingUpgrade = p.RequestParam.Mode == PARAM_ROLLING_UPGRADE
	p.agentsInZoneMap = make(map[string][]meta.AgentInfo, 0)
	p.dbaObZones, err = obclusterService.GetAllZone()
	if err != nil {
		return
	}
	p.allAgents, err = agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return
	}
	for _, dbaObZone := range p.dbaObZones {
		agents, e := agentService.GetAgentInfoByZoneFromOB(dbaObZone.Zone)
		if e != nil {
			return
		}
		p.agentsInZoneMap[dbaObZone.Zone] = agents
	}
	return nil
}

func preCheckForObUpgrade(param param.ObUpgradeParam) (p *obUpgradeParams, err *errors.OcsAgentError) {
	log.Info("start precheck for ob upgrade")
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrObclusterNotFound, "Cannot be upgraded. Please execute `init` first.")
	}
	allAgents, agentErr := agentService.GetAllAgentsInfoFromOB()
	if agentErr != nil {
		return nil, errors.Occur(errors.ErrUnexpected, "Failed to query all agents from ob")
	}
	agentInfo := coordinator.OCS_COORDINATOR.Maintainer
	agentsStatus := make(map[string]http.AgentStatus)
	resErr := secure.SendGetRequest(agentInfo, "/api/v1/agents/status", nil, &agentsStatus)
	if resErr != nil {
		return nil, errors.Occur(errors.ErrUnexpected, "Failed to query all agents status")
	}
	unavailableAgents := make([]string, 0)
	unavailableObservers := make([]string, 0)
	for agent, agentStatus := range agentsStatus {
		if agentStatus.State != 2 {
			unavailableAgents = append(unavailableAgents, agent)
		}
		if agentStatus.OBState != 3 {
			unavailableObservers = append(unavailableObservers, fmt.Sprintf("%s:%d", agentStatus.Agent.GetIp(), agentStatus.SqlPort))
		}
	}
	for _, agent := range allAgents {
		if _, ok := agentsStatus[agent.String()]; !ok {
			unavailableAgents = append(unavailableAgents, agent.String())
		}
	}
	if len(unavailableAgents) > 0 {
		return nil, errors.Occur(errors.ErrUnexpected, fmt.Sprintf("Found agent %s not running", strings.Join(unavailableAgents, ",")))
	}
	if len(unavailableObservers) > 0 {
		return nil, errors.Occur(errors.ErrUnexpected, fmt.Sprintf("Found observer %s not running", strings.Join(unavailableObservers, ",")))
	}
	currentBuildVersion, e := obclusterService.GetObBuildVersion()
	if e != nil {
		return nil, errors.Occur(errors.ErrUnexpected, e)
	}
	p = &obUpgradeParams{
		RequestParam:   &param,
		currentVersion: strings.Split(currentBuildVersion, "_")[0],
	}

	if err = checkUpgradeMode(&param); err != nil {
		return
	}

	if e = checkUpgradeDir(&param.UpgradeDir); e != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, e)
	}

	p.upgradeRoute, e = checkForAllRequiredPkgs(param.Version, param.Release)
	if e != nil {
		return nil, errors.Occur(errors.ErrUnexpected, e)
	}
	p.agents, e = agentService.GetAllAgentsInfoFromOB()
	if e != nil {
		return nil, errors.Occur(errors.ErrUnexpected, e)
	}
	return p, nil
}

const (
	PARAM_ROLLING_UPGRADE      = "ROLLING"
	PARAM_STOP_SERVICE_UPGRADE = "STOPSERVICE"
)

func checkUpgradeMode(param *param.ObUpgradeParam) *errors.OcsAgentError {
	param.Mode = strings.ToUpper(param.Mode)
	log.Info("check upgrade mode")
	if param.Mode != PARAM_ROLLING_UPGRADE && param.Mode != PARAM_STOP_SERVICE_UPGRADE {
		return errors.Occurf(errors.ErrKnown, "upgrade mode '%s' is not supported", param.Mode)
	}
	if param.Mode == PARAM_STOP_SERVICE_UPGRADE {
		return nil
	}

	zoneCount, err := obclusterService.GetZoneCount()
	if err != nil {
		return errors.Occur(errors.ErrUnexpected, err)
	}
	if zoneCount < 3 {
		return errors.Occur(errors.ErrKnown, "not support rolling upgrade when zone num is lower than 3")
	}
	return nil
}
