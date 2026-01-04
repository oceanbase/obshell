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
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/coordinator"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/secure"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
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
	freezeServer    bool
	upgradeRoute    []RouteNode
	dbaObZones      []oceanbase.DbaObZones
	allAgents       []meta.AgentInfo
	agentsInZoneMap map[string][]meta.AgentInfo
	agents          []meta.AgentInfo
	RequestParam    *param.ObUpgradeParam
}

func CheckAndUpgradeOb(param param.ObUpgradeParam) (*task.DagDetailDTO, error) {
	log.Info("check and upgrade ob")
	obType, err := obclusterService.GetOBType()
	if err != nil {
		return nil, err
	}
	p, err := preCheckForObUpgrade(param, obType)
	if err != nil {
		log.WithError(err).Error("pre check for ob upgrade failed")
		return nil, err
	}
	e := p.initParamsForObUpgrade()
	if e != nil {
		log.WithError(e).Error("init params for ob upgrade failed")
		return nil, e
	}

	checkAndUpgradeObTemplate := buildCheckAndUpgradeObTemplate(p, obType)
	checkAndUpgradeObTaskContext := buildCheckAndUpgradeObTaskContext(p, obType)
	checkAndUpgradeObDag, e := taskService.CreateDagInstanceByTemplate(checkAndUpgradeObTemplate, checkAndUpgradeObTaskContext)
	if e != nil {
		return nil, e
	}
	return task.NewDagDetailDTO(checkAndUpgradeObDag), nil
}

func buildCheckAndUpgradeObTaskContext(p *obUpgradeParams, obType modelob.OBType) *task.TaskContext {
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
		SetParam(PARAM_ROLLING_UPGRADE, p.rollingUpgrade).
		SetParam(PARAM_OB_TYPE, obType)
	return ctx
}

func buildCheckAndUpgradeObTemplate(p *obUpgradeParams, obType modelob.OBType) *task.Template {
	if p.rollingUpgrade {
		return buildObClusterCrossVeriosnRollingUpgradeTemplate(p, obType)
	} else {
		return buildObClusterCrossVeriosnStopServiceUpgradeTemplate(p, obType)
	}
}

func buildObClusterCrossVeriosnStopServiceUpgradeTemplate(p *obUpgradeParams, obType modelob.OBType) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_OB_STOP_SVC_UPGRADE, p.RequestParam.Version, p.RequestParam.Release)
	builder := task.NewTemplateBuilder(name).SetMaintenance(task.GlobalMaintenance())
	if obType == modelob.OBTypeBusiness {
		builder.AddTask(newCheckEnvTask(), true).
			AddTask(newCreateUpgradeDirTask(), true).
			AddTask(newGetAllRequiredPkgsTask(), true).
			AddTask(newCheckAllRequiredPkgsTask(), true).
			AddTask(newInstallAllRequiredPkgsTask(), true)
	} else {
		builder.AddTemplate(newCheckAndUpgradeAgentTemplate())
	}
	for i := 0; i < len(p.upgradeRoute); i++ {
		builder.AddTemplate(newStopServiceUpgradeProcessTemplate(i, p))
	}
	builder.AddTask(newRemoveUpgradeCheckDirTask(), true)
	return builder.Build()
}

func newStopServiceUpgradeProcessTemplate(idx int, p *obUpgradeParams) *task.Template {
	name := fmt.Sprintf("Upgrade process %d", idx)
	templateBuilder := task.NewTemplateBuilder(name).
		AddTask(newBackupParametersTask(), false).
		AddNode(newExecUpgradeCheckerNode("", idx))
	if p.freezeServer {
		templateBuilder.AddNode(task.NewNodeWithContext(newMinorFreezeTask(), false, task.NewTaskContext().SetParam(PARAM_SCOPE, param.Scope{Type: SCOPE_GLOBAL})))
	}
	templateBuilder.AddNode(newExecPreScriptNode("", idx)).
		AddNode(newExecHealthCheckerNode("", idx)).
		AddNode(newReinstallAndRestartObNode("", p.allAgents, idx)).
		AddNode(newExecHealthCheckerNode("", idx)).
		AddNode(newExecPostScriptNode("", idx)).
		AddTask(newRestoreParametersTask(), false)
	return templateBuilder.Build()
}

func newCheckAndUpgradeAgentTemplate() *task.Template {
	return task.NewTemplateBuilder(DAG_CHECK_AND_UPGRADE_OBSHELL).
		AddTask(newCheckEnvTask(), true).
		AddTemplate(buildAgentCheckAndUpgradeTemplate()).
		Build()
}

func buildObClusterCrossVeriosnRollingUpgradeTemplate(p *obUpgradeParams, obType modelob.OBType) *task.Template {
	name := fmt.Sprintf("%s %s-%s", DAG_OB_ROLLING_UPGRADE, p.RequestParam.Version, p.RequestParam.Release)
	builder := task.NewTemplateBuilder(name).SetMaintenance(task.GlobalMaintenance())
	if obType == modelob.OBTypeBusiness {
		builder.AddTask(newCheckEnvTask(), true).
			AddTask(newCreateUpgradeDirTask(), true).
			AddTask(newGetAllRequiredPkgsTask(), true).
			AddTask(newCheckAllRequiredPkgsTask(), true).
			AddTask(newInstallAllRequiredPkgsTask(), true)
	} else {
		builder.AddTemplate(newCheckAndUpgradeAgentTemplate())
	}
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
			AddNode(newStopZoneNode(dbaObZone.Zone))
		if p.freezeServer {
			builder.AddNode(task.NewNodeWithContext(newMinorFreezeTask(),
				false,
				task.NewTaskContext().
					SetParam(PARAM_SCOPE, param.Scope{Type: SCOPE_ZONE, Target: []string{dbaObZone.Zone}})))
		}
		builder.AddNode(newReinstallAndRestartObNode(dbaObZone.Zone, agents, idx)).
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
	p.freezeServer = p.RequestParam.FreezeServer
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

func preCheckForObUpgrade(param param.ObUpgradeParam, obType modelob.OBType) (p *obUpgradeParams, err error) {
	log.Info("start precheck for ob upgrade")
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}
	allAgents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return nil, err
	}
	agentInfo := coordinator.OCS_COORDINATOR.Maintainer
	agentsStatus := make(map[string]http.AgentStatus)
	resErr := secure.SendGetRequest(agentInfo, "/api/v1/agents/status", nil, &agentsStatus)
	if resErr != nil {
		return nil, errors.Wrap(resErr, "failed to query all agents status")
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
		return nil, errors.Occur(errors.ErrAgentUnavailable, strings.Join(unavailableAgents, ","))
	}
	if len(unavailableObservers) > 0 {
		return nil, errors.Occur(errors.ErrObServerUnavailable, strings.Join(unavailableObservers, ","))
	}
	currentBuildVersion, e := obclusterService.GetObBuildVersion()
	if e != nil {
		return nil, err
	}
	p = &obUpgradeParams{
		RequestParam:   &param,
		currentVersion: strings.Split(currentBuildVersion, "_")[0],
	}

	if err = checkUpgradeMode(&param); err != nil {
		return
	}

	// Check python and module dependencies on real execute agents
	if err = checkPythonEnvOnRealExecuteAgents(allAgents); err != nil {
		return nil, err
	}

	if err = checkUpgradeDir(&param.UpgradeDir); err != nil {
		return nil, err
	}

	// Check if there are any tenants in the cluster with unsynchronized schema
	if err = checkAllTenantsSchemaInSync(); err != nil {
		return nil, err
	}

	// Check if there are any tablet is merging
	if err = checkTabletNotInMerging(); err != nil {
		return nil, err
	}

	// Check if there are any tenants in major compaction
	if err = checkTenantNotInMajorCompaction(); err != nil {
		return nil, err
	}

	// Check if data version is sync
	if err = checkDataVersionSync(); err != nil {
		return nil, err
	}

	// Check if there are any running backup tasks
	if err = checkNoRunningBackupTask(); err != nil {
		return nil, err
	}

	p.upgradeRoute, err = checkForAllRequiredPkgs(param.Version, param.Release, obType)
	if err != nil {
		return nil, err
	}
	log.Infof("upgrade route: %v", p.upgradeRoute)
	p.agents, err = agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return nil, err
	}

	return p, nil
}

const (
	PARAM_ROLLING_UPGRADE      = "ROLLING"
	PARAM_STOP_SERVICE_UPGRADE = "STOPSERVICE"
)

func checkUpgradeMode(param *param.ObUpgradeParam) error {
	param.Mode = strings.ToUpper(param.Mode)
	log.Info("check upgrade mode")
	if param.Mode != PARAM_ROLLING_UPGRADE && param.Mode != PARAM_STOP_SERVICE_UPGRADE {
		return errors.Occur(errors.ErrObUpgradeModeNotSupported, param.Mode)
	}
	if param.Mode == PARAM_STOP_SERVICE_UPGRADE {
		return nil
	}

	zoneCount, err := obclusterService.GetZoneCount()
	if err != nil {
		return err
	}
	if zoneCount < 3 {
		return errors.Occur(errors.ErrObUpgradeUnableToRollingUpgrade)
	}
	return nil
}

func checkAllTenantsSchemaInSync() error {
	if tenants, err := tenantService.GetUnfreshedTenants(); err != nil {
		return err
	} else if len(tenants) > 0 {
		return errors.Occur(errors.ErrObUpgradeTenantsSchemaNotRefreshed, strings.Join(tenants, ","))
	}
	return nil
}

func checkTabletNotInMerging() error {
	if count, err := obclusterService.GetTabletInMergingCount(); err != nil {
		return err
	} else if count > 0 {
		return errors.Occur(errors.ErrObUpgradeTabletInMerging, count)
	}
	return nil
}

func checkTenantNotInMajorCompaction() error {
	if count, err := tenantService.GetMajorCompactionTenantCount(); err != nil {
		return err
	} else if count > 0 {
		return errors.Occur(errors.ErrObUpgradeTenantInMajorCompaction, count)
	}
	return nil
}

func checkDataVersionSync() error {
	minObserverVersions, err := tenantService.GetDinstinctParameterValue("min_observer_version")
	if err != nil {
		return err
	}
	if len(minObserverVersions) > 1 {
		return errors.Occur(errors.ErrObUpgradeMinObserverVersionNotSync)
	}
	if strings.Compare(minObserverVersions[0], "4.3.3.0") < 0 {
		tenantCompatibles, err := tenantService.GetDinstinctParameterValue("compatible")
		if err != nil {
			return err
		}
		if len(tenantCompatibles) > 1 {
			return errors.Occur(errors.ErrObUpgradeTenantCompatibleNotSync)
		}
	}
	return nil
}

func checkNoRunningBackupTask() error {
	count, err := obclusterService.GetRunningBackupTaskCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.Occur(errors.ErrObUpgradeHasRunningBackupTask, count)
	}
	return nil
}

// getRealExecuteAgentForMachine returns the real execute agent for a given machine IP.
// For machines with multiple obshell instances, it returns the first agent in the list
// with the same IP (consistent with isRealExecuteAgent logic).
func getRealExecuteAgentForMachine(agents []meta.AgentInfo, machineIP string) *meta.AgentInfo {
	for _, agent := range agents {
		if agent.Ip == machineIP {
			return &agent
		}
	}
	return nil
}

// checkPythonEnvOnRealExecuteAgents checks python and module dependencies on real execute agents.
// For machines with multiple obshell instances, only the real execute agent is checked.
func checkPythonEnvOnRealExecuteAgents(agents []meta.AgentInfo) error {
	// Group agents by IP to handle machines with multiple obshell instances
	ipToAgents := make(map[string][]meta.AgentInfo)
	for _, agent := range agents {
		ipToAgents[agent.Ip] = append(ipToAgents[agent.Ip], agent)
	}

	// Check python environment on each machine's real execute agent
	for ip := range ipToAgents {
		realExecuteAgent := getRealExecuteAgentForMachine(agents, ip)
		if realExecuteAgent == nil {
			return errors.Occurf(errors.ErrCommonUnexpected, "failed to find real execute agent for machine %s", ip)
		}

		log.Infof("Checking python environment on real execute agent %s (machine %s)", realExecuteAgent.String(), ip)
		if err := checkPythonEnvOnAgent(*realExecuteAgent); err != nil {
			return errors.Wrapf(err, "python environment check failed on agent %s", realExecuteAgent.String())
		}
		log.Infof("Python environment check passed on agent %s", realExecuteAgent.String())
	}

	return nil
}

// checkPythonEnvOnAgent checks python and module dependencies on a specific agent.
func checkPythonEnvOnAgent(agent meta.AgentInfo) error {
	// If checking local agent, check directly
	if agent.Equal(meta.OCS_AGENT) {
		return checkPythonEnvironmentInternal(nil)
	}

	// For remote agents, send RPC request to check python environment
	return checkPythonEnvRemote(agent)
}

// checkPythonEnvironmentInternal checks python and module dependencies.
// If taskLogger is provided, it uses ExecuteLog/ExecuteLogf for logging;
// otherwise, it uses log.Infof for logging.
func checkPythonEnvironmentInternal(taskLogger task.TaskLogInterface) error {
	var logFunc func(string)
	var logFuncf func(string, ...interface{})

	if taskLogger != nil {
		logFunc = taskLogger.ExecuteLog
		logFuncf = taskLogger.ExecuteLogf
	} else {
		logFunc = func(msg string) { log.Info(msg) }
		logFuncf = log.Infof
	}

	logFunc("Checking if python is installed.")
	cmd := exec.Command("python", "-c", "import sys; print(sys.version_info.major)")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return errors.Occur(errors.ErrEnvironmentWithoutPython)
	}
	pythonVersion := strings.TrimSpace(out.String())
	logFuncf("Python major version %s", pythonVersion)

	for _, module := range modules {
		logFuncf("Checking if python module '%s' is installed.", module)
		cmd = exec.Command("python", "-c", "import "+module)
		if err := cmd.Run(); err != nil {
			logFuncf("Check python module '%s' failed: %v", module, err)
			return errors.Occur(errors.ErrEnvironmentWithoutPythonModule, module)
		}
	}

	return nil
}

// checkPythonEnvRemote checks python and module dependencies on a remote agent via RPC.
func checkPythonEnvRemote(agent meta.AgentInfo) error {
	// No request body needed, just send POST request
	err := secure.SendPostRequest(&agent, constant.URI_API_V1+constant.URI_UPGRADE+constant.URI_ENV+constant.URI_CHECK, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to check python environment on agent %s", agent.String())
	}

	log.Infof("Python environment check passed on agent %s", agent.String())
	return nil
}
