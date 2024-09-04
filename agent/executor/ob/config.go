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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/model/sqlite"
	"github.com/oceanbase/obshell/param"
)

type IntegrateObConfigTask struct {
	task.Task
	agents       map[meta.AgentInfo]sqlite.AllAgent
	zoneConfig   map[string]map[string]sqlite.ObConfig // zone -> configName -> config
	globalConfig map[string]sqlite.ObConfig            // configName -> config

	portMap map[sqlite.AllAgent]map[string]int // agent -> configName -> port

	zoneOrder   []string
	zoneRS      map[string]meta.AgentInfo         // AgentInfo.Port is rpc port
	unRS        map[string][]meta.AgentInfo       // AgentInfo.Port is rpc port
	observerMap map[meta.AgentInfo]meta.AgentInfo // the key AgentInfo.Port is rpc port
}

type UpdateOBClusterConfigTask struct {
	task.Task
}

type UpdateOBServerConfigTask struct {
	task.Task
	config    param.ObServerConfigParams
	deleteAll bool
}

func newUpdateOBServerConfigTask() *UpdateOBServerConfigTask {
	newTask := &UpdateOBServerConfigTask{
		Task: *task.NewSubTask(TASK_NAME_UPDATE_OB_CONFIG),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry()
	return newTask
}

func CheckOBServerConfigParams(params param.ObServerConfigParams) error {
	for _, key := range DeniedConfig {
		if _, ok := params.ObServerConfig[key]; ok {
			return fmt.Errorf("config %s is not allowed to set", key)
		}
	}
	return nil
}

func CreateUpdateOBServerConfigDag(params param.ObServerConfigParams, deleteAll bool) (*task.DagDetailDTO, error) {
	subTask := newUpdateOBServerConfigTask()
	template := task.NewTemplateBuilder(subTask.GetName()).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(subTask, false).
		Build()

	if err := paramToConfig(params.ObServerConfig); err != nil {
		return nil, err
	}

	ctx := task.NewTaskContext().SetParam(PARAM_CONFIG, params).SetParam(PARAM_DELETE_ALL, deleteAll)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func paramToConfig(obServerConfig map[string]string) error {
	// Check if both mysql_porth and mysqlPort are set.
	for k, v := range constant.OB_CONFIG_COMPATIBLE_MAP {
		if val, ok := obServerConfig[k]; ok {
			if val2, ok2 := obServerConfig[v]; ok2 && val != val2 {
				return errors.Errorf("You cannot set both %s and %s, use %s instead.", k, v, k)
			}
		} else if val, ok := obServerConfig[v]; ok {
			obServerConfig[k] = val
		}
		delete(obServerConfig, v)
	}
	return nil
}

func (t *UpdateOBServerConfigTask) Execute() error {
	ctx := t.GetContext()
	if err := ctx.GetParamWithValue(PARAM_CONFIG, &t.config); err != nil {
		return errors.Wrap(err, "get config error")
	}
	t.deleteAll = ctx.GetParam(PARAM_DELETE_ALL).(bool)

	switch strings.ToUpper(t.config.Scope.Type) {
	case SCOPE_GLOBAL:
		return t.updateGlobalConfig()
	case SCOPE_ZONE:
		return t.updateZoneConfig()
	case SCOPE_SERVER:
		return t.updateServerConfig()
	}
	return fmt.Errorf("invalid scope type: %s", t.config.Scope.Type)
}

func (t *UpdateOBServerConfigTask) updateGlobalConfig() error {
	return observerService.UpdateGlobalConfig(t.config.ObServerConfig, t.deleteAll)
}

func (t *UpdateOBServerConfigTask) updateZoneConfig() error {
	return observerService.UpdateZoneConfig(t.config.ObServerConfig, t.config.Scope.Target, t.deleteAll)
}

func (t *UpdateOBServerConfigTask) updateServerConfig() error {
	agents := make([]meta.AgentInfoInterface, 0)
	for _, server := range t.config.Scope.Target {
		agent := ConvertAgentInfo(server)
		if agent == nil {
			return fmt.Errorf("invalid server: %s", server)
		}
		agents = append(agents, agent)
	}
	return observerService.UpdateServerConfig(t.config.ObServerConfig, agents, t.deleteAll)
}

func ConvertAgentInfo(str string) meta.AgentInfoInterface {
	re := regexp.MustCompile(`(\[?[^\[\]:]+\]?):(\d+)`)
	matches := re.FindAllStringSubmatch(str, -1)
	if len(matches) != 1 || len(matches[0]) != 3 {
		return nil
	}
	port, err := strconv.Atoi(matches[0][2])
	if err != nil {
		return nil
	}
	return meta.NewAgentInfo(matches[0][1], port)
}

func newUpdateOBClusterConfigTask() *UpdateOBClusterConfigTask {
	newTask := &UpdateOBClusterConfigTask{
		Task: *task.NewSubTask(TASK_NAME_UPDATE_CONFIG),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry()
	return newTask
}

func CreateUpdateOBClusterConfigDag(params param.ObClusterConfigParams, deleteAll bool) (*task.DagDetailDTO, error) {
	subTask := newUpdateOBClusterConfigTask()
	template := task.NewTemplateBuilder(subTask.GetName()).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(subTask, false).
		Build()
	ctx := task.NewTaskContext().SetParam(PARAM_CONFIG, params).SetParam(PARAM_DELETE_ALL, deleteAll)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, ctx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func (t *UpdateOBClusterConfigTask) Execute() error {
	var params param.ObClusterConfigParams
	ctx := t.GetContext()
	if err := ctx.GetParamWithValue(PARAM_CONFIG, &params); err != nil {
		return err
	}
	deleteAll := ctx.GetParam(PARAM_DELETE_ALL).(bool)

	config := make(map[string]string)
	if params.ClusterId != nil {
		config[constant.CONFIG_CLUSTER_ID] = fmt.Sprint(*params.ClusterId)
	}
	if params.ClusterName != nil {
		config[constant.CONFIG_CLUSTER_NAME] = *params.ClusterName
	}
	if params.RootPwd != nil {
		config[constant.CONFIG_ROOT_PASSWORD] = *params.RootPwd
	}
	if params.RsList != nil {
		config[constant.CONFIG_RS_LIST] = *params.RsList
	}
	return obclusterService.UpdateClusterConfig(config, deleteAll)
}

func newIntegrateObConfigTask() *IntegrateObConfigTask {
	newTask := &IntegrateObConfigTask{
		Task: *task.NewSubTask(TASK_NAME_INTEGRATE_CONFIG),
	}
	newTask.SetCanCancel().SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *IntegrateObConfigTask) Execute() error {
	if err := t.init(); err != nil {
		return err
	}

	if err := t.setObserverConfig(); err != nil {
		return err
	}

	// SetBootsrapConfig will use t.agents, t.zoneConfig, t.globalConfig, t.portMap and t.observerMap.
	return t.setBootsrapConfig()
}

func (t *IntegrateObConfigTask) setObserverConfig() error {
	ctx := t.GetContext()
	t.observerMap = make(map[meta.AgentInfo]meta.AgentInfo)
	t.portMap = make(map[sqlite.AllAgent]map[string]int)
	for agent, agentDO := range t.agents {
		observerConfig, err := t.getObserverConfig(agent)
		if err != nil {
			return err
		}
		t.setDirs(agent, observerConfig)
		ctx.SetAgentData(&agent, PARAM_CONFIG, observerConfig)

		portMap := t.portMap[agentDO]
		observerInfo := *meta.NewAgentInfo(agent.Ip, portMap[constant.CONFIG_RPC_PORT])
		t.observerMap[observerInfo] = agent
	}
	return nil
}

func (t *IntegrateObConfigTask) init() (err error) {
	if err = t.getAgents(); err != nil {
		return err
	}
	t.ExecuteLog("get global config")
	if t.globalConfig, err = observerService.GetObGlobalConfigMap(); err != nil {
		return err
	}
	t.setRootPWD()
	if err = t.initZoneConfig(); err != nil {
		return err
	}
	return nil
}

func (t *IntegrateObConfigTask) getAgents() error {
	t.ExecuteLog("get all agents")
	t.agents = make(map[meta.AgentInfo]sqlite.AllAgent)
	agentsDO, err := agentService.GetAllAgentsDO()
	if err != nil {
		return err
	}
	for _, agentDO := range agentsDO {
		agent := *meta.NewAgentInfo(agentDO.Ip, agentDO.Port)
		t.agents[agent] = agentDO
	}
	return nil
}

func (t *IntegrateObConfigTask) setRootPWD() {
	if rootPWD, ok := t.globalConfig[constant.CONFIG_ROOT_PASSWORD]; ok {
		delete(t.globalConfig, constant.CONFIG_ROOT_PASSWORD)
		if rootPWD.Value != "" {
			t.ExecuteLog("set root password")
			t.GetContext().SetData(PARAM_ROOT_PWD, rootPWD.Value)
		}
	}
}

func (t *IntegrateObConfigTask) initZoneConfig() (err error) {
	t.ExecuteLog("init all zone config")
	t.zoneConfig = make(map[string]map[string]sqlite.ObConfig)
	for _, agentDO := range t.agents {
		if _, err = t.buildZoneConfig(agentDO.Zone); err != nil {
			return
		}
	}
	return
}

func (t *IntegrateObConfigTask) buildZoneConfig(zone string) (zoneConfig map[string]sqlite.ObConfig, err error) {
	if zoneConfig, ok := t.zoneConfig[zone]; ok {
		return zoneConfig, nil
	}

	t.ExecuteLogf("build %s zone config", zone)
	if zoneConfig, err = observerService.GetObZoneConfigMap(zone); err != nil {
		return
	}
	t.ExecuteLogf("merge %s zone config", zone)
	zoneConfig = mergeConfig(zoneConfig, t.globalConfig)
	t.zoneConfig[zone] = zoneConfig
	return
}

func (t *IntegrateObConfigTask) setBootsrapConfig() error {
	t.ExecuteLog("setting bootstrap config")
	t.zoneRS = make(map[string]meta.AgentInfo)
	t.zoneOrder = make([]string, 0)
	t.unRS = make(map[string][]meta.AgentInfo)

	if val, ok := t.globalConfig[constant.CONFIG_RS_LIST]; ok && val.Value != "" {
		t.ExecuteLog("rsList has set, check it")
		if err := t.hasRSList(); err != nil {
			return err
		}
		if err := t.setUNRootServer(); err != nil {
			return err
		}
	} else {
		t.generateRS()
	}

	t.GetContext().SetData(PARAM_ZONE_RS, t.zoneRS).SetData(PARAM_ZONE_ORDER, t.zoneOrder).SetData(PARAM_UNRS, t.unRS)
	return nil
}

func (t *IntegrateObConfigTask) generateRS() {
	// Add self first.
	agentInfo := *meta.NewAgentInfoByInterface(meta.OCS_AGENT)
	zone := t.agents[agentInfo].Zone
	t.zoneRS[zone] = agentInfo // Hold, will update to observerInfo.
	t.zoneOrder = append(t.zoneOrder, zone)

	for observerInfo, agentInfo := range t.observerMap {
		zone := t.agents[agentInfo].Zone
		if agentInfo.Equal(meta.OCS_AGENT) {
			t.zoneRS[zone] = observerInfo // Update to observerInfo.
			continue
		}

		if _, ok := t.zoneRS[zone]; ok {
			// If this zone already has a rootserver, add to unRS.
			t.unRS[zone] = append(t.unRS[zone], observerInfo)
		} else {
			t.zoneRS[zone] = observerInfo
			t.zoneOrder = append(t.zoneOrder, zone)
			t.unRS[zone] = make([]meta.AgentInfo, 0)
		}
	}
}

func (t *IntegrateObConfigTask) hasRSList() error {
	val := t.globalConfig[constant.CONFIG_RS_LIST]
	delete(t.globalConfig, constant.CONFIG_RS_LIST)

	rsServers := strings.Split(val.Value, ";")
	t.ExecuteLogf("check rsList config: %s", rsServers)
	for _, rsServer := range rsServers {
		info := strings.Split(rsServer, ":")
		if len(info) != 3 {
			return fmt.Errorf("invalid rsList config: %s", rsServer)
		}
		rpcPort, err := strconv.Atoi(info[1])
		if err != nil {
			return fmt.Errorf("invalid rsList config: %s", rsServer)
		}
		observerInfo := *meta.NewAgentInfo(info[0], rpcPort)
		agentInfo, ok := t.observerMap[observerInfo]
		if !ok {
			return fmt.Errorf("invalid rsList config: %s not in cluster", rsServer)
		}

		mysqlPort, err := strconv.Atoi(info[2])
		if err != nil {
			return fmt.Errorf("invalid rsList config: %s", rsServer)
		}
		portMap := t.portMap[t.agents[agentInfo]]
		if portMap[constant.CONFIG_MYSQL_PORT] != mysqlPort {
			return fmt.Errorf("invalid rsList config: %s mysql port not match", rsServer)
		}

		zone := t.agents[agentInfo].Zone
		if _, ok := t.zoneRS[zone]; !ok {
			t.zoneRS[zone] = observerInfo
			t.zoneOrder = append(t.zoneOrder, zone)
		} else {
			return fmt.Errorf("invalid rsList config: %s has more than one rootserver", zone)
		}
	}

	for zone, _ := range t.zoneConfig {
		if _, ok := t.zoneRS[zone]; !ok {
			return fmt.Errorf("invalid rsList config: %s has no rootserver", zone)
		}
	}
	return nil
}

func (t *IntegrateObConfigTask) setUNRootServer() error {
	for observerInfo, agentInfo := range t.observerMap {
		zone := t.agents[agentInfo].Zone
		if rs, ok := t.zoneRS[zone]; !ok {
			return fmt.Errorf("invalid rsList config: %s has no rootserver", zone)
		} else if rs.Equal(&observerInfo) {
			continue
		}
		if _, ok := t.unRS[zone]; !ok {
			t.unRS[zone] = make([]meta.AgentInfo, 0)
		}
		t.unRS[zone] = append(t.unRS[zone], observerInfo)
	}
	return nil
}

func (t *IntegrateObConfigTask) getObserverConfig(agent meta.AgentInfo) (map[string]string, error) {
	agentStr := fmt.Sprintf("%s:%d", agent.Ip, agent.Port)
	t.ExecuteLogf("Integrating %s agent config", agentStr)

	t.ExecuteLogf("get %s agent config", agentStr)
	serverConfig, err := observerService.GetObServerConfigMap(&agent)
	if err != nil {
		return nil, err
	}
	agentDO := t.agents[agent]
	zoneConfig, err := t.buildZoneConfig(agentDO.Zone)
	if err != nil {
		return nil, err
	}

	config := make(map[string]string)
	t.ExecuteLogf("merge %s agent config", agentStr)
	serverConfig = mergeConfig(serverConfig, zoneConfig)
	for key, item := range serverConfig {
		config[key] = item.Value
	}

	t.ExecuteLogf("fill %s agent config", agentStr)
	err = t.fillConfig(agentDO, config)

	delete(config, constant.CONFIG_RS_LIST)
	return config, err
}

func (t *IntegrateObConfigTask) setDirs(agent meta.AgentInfo, config map[string]string) {
	dirs := make(map[string]string)
	for _, key := range allDirOrder {
		dirs[key] = config[key]
		delete(config, key)
	}
	t.GetContext().SetAgentData(&agent, PARAM_DIRS, dirs)
}

func (t *IntegrateObConfigTask) fillConfig(agentDO sqlite.AllAgent, config map[string]string) error {
	config[constant.CONFIG_HOME_PATH] = agentDO.HomePath
	config[constant.CONFIG_ZONE] = agentDO.Zone

	t.portMap[agentDO] = make(map[string]int)
	if err := fillPort(config); err != nil {
		return err
	}
	for key, _ := range defaultPortMap {
		p, _ := strconv.Atoi(config[key])
		t.portMap[agentDO][key] = p
	}

	return fillDir(config)
}

func fillPort(config map[string]string) error {
	for key, port := range defaultPortMap {
		if isValidConfig(config, key) {
			p, err := strconv.Atoi(config[key])
			if err != nil {
				return fmt.Errorf("invalid %s: %s not a int", key, config[key])
			}
			if p < 1025 || p > 65535 {
				return fmt.Errorf("invalid %s: %d. port must between 1025 and 65535", key, p)
			}
		} else {
			config[key] = fmt.Sprint(port)
		}
	}
	return nil
}

func fillDir(config map[string]string) error {
	for _, dir := range storeDirOrder {
		if !isValidConfig(config, dir) {
			if parentDirKey, ok := parentDirKeys[dir]; ok {
				if parentDir, ok := config[parentDirKey]; ok {
					config[dir] = filepath.Join(parentDir, dirMap[dir])
				} else {
					return fmt.Errorf("fill config error, key %s not in config", parentDirKey)
				}
			} else {
				return fmt.Errorf("fill config error, key %s not in parentDirKeys", dir)
			}
		}
	}
	return nil
}

func isValidConfig(config map[string]string, name string) bool {
	val, ok := config[name]
	return ok && val != ""
}

func mergeConfig(config1, config2 map[string]sqlite.ObConfig) map[string]sqlite.ObConfig {
	for key, config := range config2 {
		if item, ok := config1[key]; ok && item.GmtModify.After(config.GmtModify) {
			// There is a corresponding configuration on config1,
			// and the modification time is later than cofnig2,
			// so the config1 configuration will not be updated.
			continue
		}
		config1[key] = config
	}
	return config1
}
