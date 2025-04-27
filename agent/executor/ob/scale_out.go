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

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/coordinator"
	"github.com/oceanbase/obshell/agent/engine/executor"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/binary"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/lib/pkg"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/repository/db/oceanbase"
	oceanbaseModel "github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
)

var (
	paramExpectDeployNextStage   = 4 // Wait remoteDeployTask finish
	paramExpectStartNextStage    = 5 // Wait remoteStartTask finish
	paramWaitDeployRetryStage    = 2 // Wait deployRetryTask
	paramWaitStartRetryStage     = 4 // Wait startRetryTask
	paramExpectRollbackNextStage = 6 // PrevCheckTask
)

var WAIT_REMOTE_TASK_FINISH_TIMES = 30
var WAIT_REMOTE_TASK_FINISH_INTERVAL = 3 * time.Second
var SYNC_FROM_OB_RETRY_TIME = 30

type LocalScaleOutResp struct {
	task.DagDetailDTO
	param.JoinMasterParam
	ParamWaitDeployRetryStage int `json:"paramWaitDeployRetryStage" binding:"required"`
	ParamWaitStartRetryStage  int `json:"paramWaitStartRetryStage" binding:"required"`
}

func HandleClusterScaleOut(param param.ClusterScaleOutParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent", meta.OCS_AGENT.String())
	}
	// Check scaling agent status.
	var agent meta.AgentStatus
	if err := http.SendGetRequest(&param.AgentInfo, constant.URI_API_V1+constant.URI_INFO, nil, &agent); err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "get %s status failed", param.AgentInfo.String())
	}
	if !agent.IsSingleAgent() {
		return nil, errors.Occurf(errors.ErrKnown, "%s is not single agent", param.AgentInfo.String())
	}

	// check ob version is consistent
	if obVersion, _, err := binary.GetMyOBVersion(); err != nil {
		return nil, errors.Occurf(errors.ErrUnexpected, "get ob version failed: %s", err.Error())
	} else if obVersion != agent.OBVersion {
		return nil, errors.Occurf(errors.ErrBadRequest, "ob version is not consistent between %s(%s) and %s(%s)",
			param.AgentInfo.String(), agent.OBVersion, meta.OCS_AGENT.String(), obVersion)
	}

	var targetVersion string
	agentVersion := strings.Split(agent.Version, "-")[0]
	log.Infof("scale out agent %s(%s) into cluster agent %s(%s)", agent.String(), agentVersion, meta.OCS_AGENT.String(), constant.VERSION)
	if cmp := pkg.CompareVersion(constant.VERSION, agentVersion); cmp < 0 {
		return nil, errors.Occurf(errors.ErrBadRequest, "scale out a higer version agent(%s) into cluster agent(%s) is not allowed", agentVersion, constant.VERSION)
	} else if cmp > 0 {
		if pkg.CompareVersion(agentVersion, constant.AGENT_V4241) < 0 {
			return nil, errors.Occurf(errors.ErrBadRequest, "scale out a lower version agent(%s before %s) into cluster agent(%s) is not allowed", agentVersion, constant.AGENT_V4241, constant.VERSION)
		}
		if exist, err := agentService.TargetVersionAgentExists(constant.VERSION); err != nil {
			return nil, errors.Occur(errors.ErrUnexpected, "check target version agent exists failed")
		} else if !exist {
			return nil, errors.Occurf(errors.ErrBadRequest, "There is no aviailable agent(version: %s, architecture: %s) in OB", constant.VERSION, global.Architecture)
		}
		targetVersion = constant.VERSION
	}

	param.ObConfigs[constant.CONFIG_HOME_PATH] = agent.HomePath
	if err := paramToConfig(param.ObConfigs); err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}

	var rpcPort int
	rpcPortStr, ok := param.ObConfigs[constant.CONFIG_RPC_PORT]
	if ok {
		var err error
		if rpcPort, err = strconv.Atoi(rpcPortStr); err != nil {
			return nil, errors.Occur(errors.ErrIllegalArgument, "rpc_port is not a number")
		}
	} else {
		rpcPort = constant.DEFAULT_RPC_PORT
	}
	srvInfo := meta.NewAgentInfo(param.AgentInfo.Ip, rpcPort)

	// Check the server is not already in the cluster.
	if exist, err := obclusterService.IsServerExist(*srvInfo); err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	} else if exist {
		return nil, errors.Occurf(errors.ErrBadRequest, "server %s already exists in the cluster", srvInfo.String())
	}

	// Create Cluster Scale Out Dag
	dag, err := CreateClusterScaleOutDag(param, targetVersion)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}
	return dag, nil
}

func HandleLocalScaleOut(params param.LocalScaleOutParam) (*LocalScaleOutResp, *errors.OcsAgentError) {
	// Create Local Scale Out Dag.
	if !meta.OCS_AGENT.IsSingleAgent() {
		return nil, errors.Occurf(errors.ErrKnown, "%s is not single agent", meta.OCS_AGENT.String())
	}
	if params.Dirs[constant.CONFIG_HOME_PATH] != global.HomePath {
		return nil, errors.Occur(errors.ErrUnexpected, "home path is not right")
	}
	dag, err := CreateLocalScaleOutDag(params)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err.Error())
	}

	encryptedToken, err := secure.NewToken(&params.AgentInfo)
	if err != nil {
		return nil, errors.Occur(errors.ErrKnown, "create token failed")
	}
	return &LocalScaleOutResp{
		DagDetailDTO: *task.NewDagDetailDTO(dag),
		JoinMasterParam: param.JoinMasterParam{
			HomePath:     global.HomePath,
			Version:      meta.OCS_AGENT.GetVersion(),
			Os:           global.Os,
			Architecture: global.Architecture,
			PublicKey:    secure.Public(),
			Token:        encryptedToken,
		},
		ParamWaitDeployRetryStage: paramWaitDeployRetryStage,
		ParamWaitStartRetryStage:  paramWaitStartRetryStage,
	}, nil
}

func CreateClusterScaleOutDag(param param.ClusterScaleOutParam, targetVersion string) (*task.DagDetailDTO, error) {
	zone := param.Zone
	isZoneExist, err := obclusterService.IsZoneExistInOB(zone)
	if err != nil {
		return nil, errors.Wrap(err, "check zone exist failed")
	}
	// get target agent pk
	encryptAgentPassword, err := secure.EncryptForAgent(param.TargetAgentPassword, &param.AgentInfo)
	if err != nil {
		return nil, errors.Wrap(err, "encrypt agent password failed")
	}

	template := buildClusterScaleOutTaskTemplate(param, !isZoneExist)
	context := buildClusterScaleOutDagContext(param, !isZoneExist, targetVersion, encryptAgentPassword)
	dag, err := clusterTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, errors.Wrap(err, "create dag instance failed")
	}
	return task.NewDagDetailDTO(dag), nil
}

func CreateLocalScaleOutDag(param param.LocalScaleOutParam) (*task.Dag, error) {
	isRunning, err := localTaskService.IsRunning()
	if err != nil {
		return nil, errors.Wrap(err, "check local task service is running failed")
	}
	if !isRunning {
		dag, err := localTaskService.GetLastMaintenanceDag()
		if err != nil {
			return nil, errors.Wrap(err, "get last maintenance dag failed")
		}
		if dag.GetName() == DAG_NAME_LOCAL_SCALE_OUT {
			var uuid string
			if dag.GetContext().GetParamWithValue(PARAM_SCALE_OUT_UUID, &uuid) != nil {
				return nil, errors.Wrap(err, "get coordinate dag id failed")
			}
			if uuid == param.Uuid {
				return dag, nil
			}
		}
		return nil, errors.New("agent is under maintenance")
	}
	if err := secure.UpdateObPassword(param.RootPwd); err != nil {
		return nil, errors.Wrap(err, "update ob password failed")
	}
	dag, err := localTaskService.CreateDagInstanceByTemplate(buildLocalScaleOutTaskTemplate(param), buildLocalScaleOutDagContext(param))
	if err != nil {
		return nil, errors.Wrap(err, "create dag instance failed")
	}

	return dag, err
}

func buildClusterScaleOutTaskTemplate(param param.ClusterScaleOutParam, isNewZone bool) *task.Template {
	templateBuild := task.NewTemplateBuilder(DAG_NAME_CLUSTER_SCALE_OUT).
		AddTask(newIntegrateSingleObConfigTask(), false).
		AddTask(newCreateLocalScaleOutDagTask(), false).
		AddTask(newWaitScalingReadyTask(), false).
		AddTask(newWaitRemoteDeployTaskFinish(), false).
		AddTask(newWaitRemoteStartTaskFinish(), false).
		AddTask(newPrevCheckTask(), false)
	if isNewZone {
		templateBuild.AddTask(newAddNewZoneTask(), false).
			AddTask(newStartNewZoneTask(), false)
	}
	templateBuild.AddTask(newAddServerTask(), false).
		AddTask(newAddAgentTask(), false).
		AddTask(newFinishTask(), false)

	return templateBuild.SetMaintenance(task.GlobalMaintenance()).Build()
}

func buildLocalScaleOutTaskTemplate(param param.LocalScaleOutParam) *task.Template {
	builder := task.NewTemplateBuilder(DAG_NAME_LOCAL_SCALE_OUT).
		AddTask(newAgentBeScalingOutTask(), false).
		AddNode(task.NewNodeWithContext(newWaitDeployRetryTask(), false, task.NewTaskContext().SetParam(PARAM_EXPECT_MAIN_NEXT_STAGE, param.ParamExpectDeployNextStage))).
		AddTask(newDeployTask(), false).
		AddNode(task.NewNodeWithContext(newWaitStartRetryTask(), false, task.NewTaskContext().SetParam(PARAM_EXPECT_MAIN_NEXT_STAGE, param.ParamExpectStartNextStage))).
		AddTask(newStartObServerTask(), false).
		AddTask(newWatchDagTask(), false)
	if param.TargetVersion != "" {
		builder.AddTask(newScalingAgentUpdateBinaryTask(), false)
	}
	return builder.AddTask(newSyncFromOB(), false).
		SetMaintenance(task.GlobalMaintenance()).
		Build()
}

func buildClusterScaleOutDagContext(param param.ClusterScaleOutParam, isNewZone bool, targetVersion string, targetAgentPassword string) *task.TaskContext {
	context := task.NewTaskContext().
		SetParam(PARAM_ZONE, param.Zone).
		SetParam(PARAM_IS_NEW_ZONE, isNewZone).
		SetParam(PARAM_AGENT_INFO, param.AgentInfo).
		SetParam(PARAM_CONFIG, param.ObConfigs).
		SetParam(PARAM_TARGET_AGENT_VERSION, targetVersion).
		SetParam(PARAM_TARGET_AGENT_PASSWORD, targetAgentPassword)
	return context
}

func buildLocalScaleOutDagContext(param param.LocalScaleOutParam) *task.TaskContext {
	context := task.NewTaskContext().
		SetParam(PARAM_ZONE, param.Zone).
		SetParam(PARAM_ALL_AGENTS, param.AllAgents).
		SetParam(PARAM_COORDINATE_AGENT, param.AgentInfo).
		SetParam(PARAM_COORDINATE_DAG_ID, param.CoordinateDagId).
		SetParam(PARAM_MAIN_AGENT, param.AgentInfo).
		SetParam(PARAM_MAIN_DAG_ID, param.CoordinateDagId).
		SetParam(PARAM_EXPECT_ROLLBACK_NEXT_STAGE, param.ParamExpectRollbackNextStage).
		SetParam(PARAM_ROOT_PWD, param.RootPwd).
		SetParam(PARAM_SCALE_OUT_UUID, param.Uuid).
		SetParam(PARAM_HEALTH_CHECK, true).
		SetAgentData(meta.OCS_AGENT, PARAM_DIRS, param.Dirs).
		SetAgentData(meta.OCS_AGENT, PARAM_CONFIG, param.ObConfigs)
	if param.TargetVersion != "" {
		context.SetParam(PARAM_TARGET_AGENT_VERSION, param.TargetVersion)
	}
	return context
}

type scaleCoordinateTask struct {
	task.Task
	coordinateDagId string
	coordinateAgent meta.AgentInfo
	allAgent        []meta.AgentInfo
}

func (t *scaleCoordinateTask) getNode() (*task.Node, error) {
	if t.IsLocalTask() {
		return localTaskService.GetNodeBySubTask(t.GetID())
	} else {
		id, err := clusterTaskService.GetRemoteTaskIdByLocalTaskId(t.GetID())
		if err != nil {
			return nil, errors.Wrap(err, "get remote task id failed")
		}
		return clusterTaskService.GetNodeBySubTask(id)
	}
}

func newScaleCoordinateTask(name string) *scaleCoordinateTask {
	newTask := &scaleCoordinateTask{
		Task: *task.NewSubTask(name),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *scaleCoordinateTask) init() error {
	ctx := t.GetContext()
	if err := ctx.GetParamWithValue(PARAM_COORDINATE_AGENT, &t.coordinateAgent); err != nil {
		return errors.Wrap(err, "get coordinate agent failed")
	}
	if err := ctx.GetParamWithValue(PARAM_COORDINATE_DAG_ID, &t.coordinateDagId); err != nil {
		return errors.Wrap(err, "get coordinate dag id failed")
	}
	if err := ctx.GetParamWithValue(PARAM_ALL_AGENTS, &t.allAgent); err != nil {
		return errors.Wrap(err, "get all coordinate agent id failed")
	}
	return nil
}

func (t *scaleCoordinateTask) initFromData() error {
	ctx := t.GetContext()
	if ctx.GetData(PARAM_COORDINATE_DAG_ID) != nil {
		if err := ctx.GetDataWithValue(PARAM_COORDINATE_DAG_ID, &t.coordinateDagId); err != nil {
			return errors.Wrap(err, "get coordinate dag id failed")
		}
	} else {
		t.coordinateDagId = ""
		return nil
	}
	if err := ctx.GetDataWithValue(PARAM_COORDINATE_AGENT, &t.coordinateAgent); err != nil {
		return errors.Wrap(err, "get coordinate agent failed")
	}
	if err := ctx.GetDataWithValue(PARAM_ALL_AGENTS, &t.allAgent); err != nil {
		return errors.Wrap(err, "get coordinate all agent failed")
	}
	return nil
}

// tryOperateDag try to operate the dag, if success return true, else return false.
// Parameter stage is the index of the task in dag, and is used for retry.
func (t *scaleCoordinateTask) tryOperateDag(operator string, stage ...int) bool {
	uri := fmt.Sprintf("%s%s/%s", constant.URI_TASK_API_PREFIX, constant.URI_DAG, t.coordinateDagId)
	operatorParams := &task.DagOperator{Operator: operator}
	requestParams := &param.TaskQueryParams{ShowDetails: constant.PTR_TRUE}
	var coordinateDag task.DagDetailDTO
	for j := 0; j <= len(t.allAgent); j++ {
		for i := 0; i < DEFAULT_REMOTE_REQUEST_RETRY_TIMES; i++ {
			if resp, err := secure.SendGetRequestAndReturnResponse(&t.coordinateAgent, uri, requestParams, &coordinateDag); resp != nil && resp.IsError() {
				time.Sleep(1 * time.Second)
				continue
			} else if err != nil {
				break // Try aonther agent.
			}

			switch operator {
			case task.ROLLBACK_STR:
				if coordinateDag.Operator == task.ROLLBACK_STR && coordinateDag.State != task.FAILED_STR {
					t.ExecuteInfoLog("try rollback coordinate dag successfully")
					return true
				}
			case task.RETRY_STR:
				if len(stage) != 0 {
					if coordinateDag.Nodes[stage[0]].State != task.FAILED_STR {
						t.ExecuteInfoLogf("retry stage: %d successfully", stage[0]+1) // When print, should +1.
						return true
					}
				} else {
					if coordinateDag.Operator == task.RUN_STR && coordinateDag.State != task.FAILED_STR {
						return true
					}
				}
			}
			// Send rollback/retry request to cluster agent and ignore the error.
			secure.SendPostRequestAndReturnResponse(&t.coordinateAgent, uri, operatorParams, nil)
			time.Sleep(1 * time.Second)
		}
		if j == len(t.allAgent) {
			return false
		} else {
			t.coordinateAgent = t.allAgent[j]
		}
	}
	return false
}

func (t *scaleCoordinateTask) syncCoordinateDag() (bool, error) {
	if t.coordinateDagId == "" {
		return true, nil
	}
	node, err := t.getNode()
	if err != nil {
		return false, errors.Wrap(err, "get node failed")
	}
	if node.GetOperator() == task.ROLLBACK {
		// Rollback does not ask for success.
		return t.tryOperateDag(task.ROLLBACK_STR), nil
	} else if node.GetOperator() == task.RETRY {
		if t.tryOperateDag(task.RETRY_STR) {
			return true, nil
		} else {
			// Retry asks for success
			return false, errors.New("retry coordinate dag failed")
		}
	} else {
		return false, errors.Errorf("not support operator %s", task.OPERATOR_MAP[node.GetOperator()])
	}
}

func (t *scaleCoordinateTask) getCoordinateDag() (*task.DagDetailDTO, error) {
	var coordinateDag task.DagDetailDTO
	for i := 0; i <= len(t.allAgent); i++ {
		uri := fmt.Sprintf("%s%s/%s", constant.URI_TASK_API_PREFIX, constant.URI_DAG, t.coordinateDagId)
		requestParams := &param.TaskQueryParams{ShowDetails: constant.PTR_TRUE}
		if resp, err := secure.SendGetRequestAndReturnResponse(&t.coordinateAgent, uri, requestParams, &coordinateDag); resp != nil && resp.IsError() {
			return nil, errors.Errorf("get remote dag failed %v", resp.Error())
		} else if err != nil {
			if i == len(t.allAgent) {
				return nil, errors.Wrap(err, "get remote dag failed")
			} else {
				t.coordinateAgent = t.allAgent[i]
			}
		}
	}
	return &coordinateDag, nil
}

type AgentBeScalingOutTask struct {
	scaleCoordinateTask
}

func newAgentBeScalingOutTask() *AgentBeScalingOutTask {
	newTask := &AgentBeScalingOutTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_BE_SCALING),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *AgentBeScalingOutTask) Execute() error {
	if t.IsContinue() && meta.OCS_AGENT.IsScalingOutAgent() {
		t.ExecuteLog("agent is follower agent")
		return nil
	}
	if !meta.OCS_AGENT.IsSingleAgent() {
		return errors.New("agent is not single")
	}

	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("zone is not set")
	}
	if err := agentService.BeScalingOutAgent(zone); err != nil {
		return err
	}
	t.ExecuteLog("set agent identity to follower")
	return nil
}

func (t *AgentBeScalingOutTask) Rollback() error {
	t.init()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	return agentService.BeSingleAgent()
}

func (t *AgentBeScalingOutTask) GetAdditionalData() map[string]interface{} {
	uuid := t.GetContext().GetParam(PARAM_SCALE_OUT_UUID).(string)
	return map[string]interface{}{
		PARAM_SCALE_OUT_UUID: uuid,
	}
}

type WaitDeployRetryTask struct {
	CheckDagStageTask
}

func newWaitDeployRetryTask() *WaitDeployRetryTask {
	newTask := &WaitDeployRetryTask{
		CheckDagStageTask: *newCheckDagStageTask(),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *WaitDeployRetryTask) Execute() error {
	return t.CheckDagStageTask.Execute()
}

func (t *WaitDeployRetryTask) Rollback() error {
	return nil
}

type WaitStartRetryTask struct {
	CheckDagStageTask
}

func newWaitStartRetryTask() *WaitStartRetryTask {
	newTask := &WaitStartRetryTask{
		CheckDagStageTask: *newCheckDagStageTask(),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *WaitStartRetryTask) Execute() error {
	return t.CheckDagStageTask.Execute()
}

func (t *WaitStartRetryTask) Rollback() error {
	return nil
}

type WatchDagTask struct {
	scaleCoordinateTask
}

func newWatchDagTask() *WatchDagTask {
	newTask := &WatchDagTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_WATCH_DAG),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *WatchDagTask) Execute() error {
	t.init()
	t.ExecuteLog("Watch cluster scale out dag")
	for {
		t.TimeoutCheck()
		coordinateDag, err := t.getCoordinateDag()
		if err != nil {
			t.ExecuteWarnLogf("get remote dag failed, %s", err.Error())
			continue
		}
		if coordinateDag.Stage == coordinateDag.MaxStage {
			break
		}
		// Failed only when remote dag failed.
		if coordinateDag.Operator == task.RUN_STR && coordinateDag.State == task.FAILED_STR || coordinateDag.Operator == task.ROLLBACK_STR {
			return errors.Errorf("remote task %s %s failed", coordinateDag.GenericID, coordinateDag.Name)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (t *WatchDagTask) Rollback() error {
	t.init()
	if synced, err := t.syncCoordinateDag(); err != nil || !synced {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	var expectStage int
	if err := t.GetContext().GetParamWithValue(PARAM_EXPECT_ROLLBACK_NEXT_STAGE, &expectStage); err != nil {
		return errors.Wrap(err, "get expected stage failed")
	}
	expectStage -= 1
	for {
		t.TimeoutCheck()
		coordinateDag, err := t.getCoordinateDag()
		if err != nil {
			t.ExecuteWarnLogf("get remote dag failed, %s", err.Error())
			continue
		}
		node := coordinateDag.Nodes[expectStage]
		if node.IsSucceed() {
			// Expect that the stage roll back successfully.
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

type ScalingAgentUpdateBinaryTask struct {
	task.Task

	scaleCoordinateTask
	UpgradeToClusterAgentVersionTask
}

func newScalingAgentUpdateBinaryTask() *ScalingAgentUpdateBinaryTask {
	newTask := &ScalingAgentUpdateBinaryTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_INSTALL_CLUSTER_AGENT_VERSION),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *ScalingAgentUpdateBinaryTask) Execute() error {
	// Try to connect to the ocenabase.
	var cipherPassword, password string
	val := t.GetContext().GetParam(PARAM_ROOT_PWD)
	if val != nil {
		cipherPassword = val.(string)
		var err error
		password, err = secure.Decrypt(cipherPassword)
		if err != nil {
			return err
		}
	}
	for i := 0; i < SYNC_FROM_OB_RETRY_TIME; i++ {
		if _, err := oceanbase.GetOcsInstance(); err != nil {
			t.ExecuteLogf("get ocs instance failed: %s, try to connect", err.Error())
			if err := oceanbase.LoadOceanbaseInstance(config.NewObDataSourceConfig().SetPassword(password).SetTryTimes(10).SetSkipPwdCheck(true)); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
		}
		break
	}

	t.UpgradeToClusterAgentVersionTask.Task = t.Task
	return t.UpgradeToClusterAgentVersionTask.Execute()
}

func (t *ScalingAgentUpdateBinaryTask) Rollback() error {
	t.scaleCoordinateTask.Task = t.Task
	t.init()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	return nil
}

type SyncFromOB struct {
	scaleCoordinateTask
}

func newSyncFromOB() *SyncFromOB {
	newTask := &SyncFromOB{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_SYNC_FROM_OB),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *SyncFromOB) Execute() error {
	// Try to connect to the ocenabase.
	var cipherPassword, password string
	val := t.GetContext().GetParam(PARAM_ROOT_PWD)
	if val != nil {
		cipherPassword = val.(string)
		var err error
		password, err = secure.Decrypt(cipherPassword)
		if err != nil {
			return err
		}
	}
	for i := 0; i < SYNC_FROM_OB_RETRY_TIME; i++ {
		if _, err := oceanbase.GetOcsInstance(); err != nil {
			t.ExecuteLogf("get ocs instance failed: %s, try to connect", err.Error())
			if err := oceanbase.LoadOceanbaseInstance(config.NewObDataSourceConfig().SetPassword(password).SetTryTimes(10).SetSkipPwdCheck(true)); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			t.ExecuteLog("dump password")
			if err := secure.UpdateObPassword(cipherPassword); err != nil {
				return err
			}
			break
		} else {
			break
		}
	}

	if err := agentService.SyncAgentData(); err != nil {
		return err
	}

	if t.GetContext().GetParam(PARAM_TARGET_AGENT_VERSION) == nil {
		// If PARAM_TARGET_AGENT_VERSION is not set, it means that the agent version is the same as the cluster agent version.
		// But there may not be a binary for the current architecture in the cluster agent version.
		// Therefore, an attempt is made to upload the binary.
		go syncAgentBinary()
	}
	return nil
}

func (t *SyncFromOB) Rollback() error {
	t.init()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	return nil
}

type IntegrateSingleObConfigTask struct {
	task.Task
}

func newIntegrateSingleObConfigTask() *IntegrateSingleObConfigTask {
	newTask := &IntegrateSingleObConfigTask{
		Task: *task.NewSubTask(TASK_NAME_INTEGRATE_SERVER_CONFIG),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *IntegrateSingleObConfigTask) Execute() error {
	var configs map[string]string
	if err := t.GetContext().GetParamWithValue(PARAM_CONFIG, &configs); err != nil {
		return errors.Wrap(err, "get server configs failed")
	}
	var agentInfo meta.AgentInfo
	if t.GetContext().GetParamWithValue(PARAM_AGENT_INFO, &agentInfo) != nil {
		return errors.New("agent info is not set")
	}
	zone := t.GetContext().GetParam(PARAM_ZONE).(string)
	configs[constant.CONFIG_ZONE] = zone

	// Get cluster name and cluster id.
	var clusterName, clusterId string
	if err := observerService.GetOBParatemerByName(PARAM_CLUSTER_NAME, &clusterName); err != nil {
		return errors.Wrap(err, "get cluster name failed")
	}
	if err := observerService.GetOBParatemerByName(PARAM_CLUSTER_ID, &clusterId); err != nil {
		return errors.Wrap(err, "get cluster id failed")
	}
	configs[constant.CONFIG_CLUSTER_NAME] = clusterName
	configs[constant.CONFIG_CLUSTER_ID] = clusterId

	if err := fillPort(configs); err != nil {
		return errors.Wrap(err, "fill port failed")
	}
	fillDir(configs)

	dirs := make(map[string]string)
	for _, key := range allDirOrder {
		dirs[key] = configs[key]
		delete(configs, key)
	}
	t.GetContext().SetData(PARAM_DIRS, dirs)

	// Set observer configs
	t.GetContext().SetData(PARAM_CONFIG, configs)
	return nil
}

type CreateLocalScaleOutDagTask struct {
	scaleCoordinateTask
	targetAgentPassword string
}

func newCreateLocalScaleOutDagTask() *CreateLocalScaleOutDagTask {
	newTask := &CreateLocalScaleOutDagTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_CREATE_LOCAL_SCALE_OUT_DAG),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *CreateLocalScaleOutDagTask) Execute() error {
	var agentInfo meta.AgentInfo
	if err := t.GetContext().GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
		return errors.Wrap(err, "get agent info failed")
	}
	if err := t.GetContext().GetParamWithValue(PARAM_TARGET_AGENT_PASSWORD, &t.targetAgentPassword); err != nil {
		return errors.Wrap(err, "get target agent password failed")
	}
	// Send rpc to target agent.
	param, err := t.buildLocalScaleOutParam()
	if err != nil {
		return errors.Wrap(err, "build local scale out param failed")
	}

	if param.TargetVersion != "" {
		t.ExecuteLogf("create local scale out dag, target version: %s", param.TargetVersion)
	}

	var resp LocalScaleOutResp
	if err := secure.SendRequestWithPassword(&agentInfo, constant.URI_OB_RPC_PREFIX+constant.URI_SCALE_OUT, http.POST, t.targetAgentPassword, param, &resp); err != nil {
		return errors.Wrap(err, "send scale out rpc to target agent failed")
	}
	t.ExecuteLogf("create local scale out dag success, genericID:%s", resp.GenericID)
	t.GetContext().SetData(PARAM_COORDINATE_AGENT, agentInfo)
	t.GetContext().SetData(PARAM_ALL_AGENTS, []meta.AgentInfo{agentInfo})
	t.GetContext().SetData(PARAM_JOIN_MASTER_INFO, resp.JoinMasterParam)
	t.GetContext().SetData(PARAM_WAIT_DEPLOY_RETRY_STAGE, resp.ParamWaitDeployRetryStage)
	t.GetContext().SetData(PARAM_WAIT_START_RETRY_STAGE, resp.ParamWaitStartRetryStage)
	t.GetContext().SetData(PARAM_COORDINATE_DAG_ID, resp.GenericID)
	return nil
}

func (t *CreateLocalScaleOutDagTask) buildLocalScaleOutParam() (*param.LocalScaleOutParam, error) {
	var agentInfo meta.AgentInfo
	ctx := t.GetContext()
	if err := ctx.GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
		return nil, errors.Wrap(err, "get agent info failed")
	}
	var dirs map[string]string
	if err := ctx.GetDataWithValue(PARAM_DIRS, &dirs); err != nil {
		return nil, errors.Wrap(err, "get dirs failed")
	}
	var configs map[string]string
	if err := ctx.GetDataWithValue(PARAM_CONFIG, &configs); err != nil {
		return nil, errors.Wrap(err, "get configs failed")
	}
	zone := ctx.GetParam(PARAM_ZONE).(string)
	remoteTaskId, err := clusterTaskService.GetRemoteTaskIdByLocalTaskId(t.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "get remote dag id failed")
	}
	dagId, err := clusterTaskService.GetDagGenericIDBySubTaskId(remoteTaskId)
	if err != nil {
		return nil, errors.Wrap(err, "get dag generic id failed")
	}
	cipherPassword, err := secure.EncryptForAgent(meta.OCEANBASE_PWD, &agentInfo)
	if err != nil {
		return nil, err
	}
	allAgents, err := agentService.GetAllAgentsInfoFromOB()
	if err != nil {
		return nil, err
	}
	uuid := uuid.New().String()
	t.GetContext().SetAgentData(meta.OCS_AGENT, PARAM_SCALE_OUT_UUID, uuid)

	var targetVersion string
	if err := ctx.GetParamWithValue(PARAM_TARGET_AGENT_VERSION, &targetVersion); err != nil {
		return nil, errors.Wrap(err, "get target version failed")
	}
	param := param.LocalScaleOutParam{
		ScaleOutParam: param.ScaleOutParam{
			AgentInfo: meta.OCS_AGENT.GetAgentInfo(),
			ObConfigs: configs,
			Zone:      zone,
		},
		Dirs:                         dirs,
		CoordinateDagId:              dagId,
		AllAgents:                    allAgents,
		RootPwd:                      cipherPassword,
		Uuid:                         uuid,
		ParamExpectDeployNextStage:   paramExpectDeployNextStage,
		ParamExpectStartNextStage:    paramExpectStartNextStage,
		ParamExpectRollbackNextStage: paramExpectRollbackNextStage,
		TargetVersion:                targetVersion,
	}
	return &param, nil
}

func (t *CreateLocalScaleOutDagTask) Rollback() error {
	t.initFromData()
	if t.coordinateDagId != "" {
		// Local scale out dag has been created.
		if _, err := t.syncCoordinateDag(); err != nil {
			return errors.Wrap(err, "sync coordinate dag failed")
		}
	} else {
		var agentInfo meta.AgentInfo
		if err := t.GetContext().GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
			return errors.Wrap(err, "get agent info failed")
		}
		t.coordinateAgent = agentInfo
		if t.GetContext().GetData(PARAM_SCALE_OUT_UUID) == nil {
			return nil
		}
		dag, err := t.getCoordinaterLastMaintainDag()
		if err != nil {
			return errors.Wrap(err, "get coordinater last maintain dag failed")
		}
		if dag.AdditionalData == nil {
			return errors.New("get coordinater last maintain dag failed, additional data is nil")
		}
		additionalData := *(dag.AdditionalData)
		if dag == nil || additionalData[PARAM_SCALE_OUT_UUID].(string) != t.GetContext().GetData(PARAM_SCALE_OUT_UUID).(string) {
			t.ExecuteInfoLog("no need to rollback")
			return nil
		}
		t.coordinateDagId = dag.GenericID
		// Sync coordinate dag.
		if _, err := t.syncCoordinateDag(); err != nil {
			return errors.Wrap(err, "sync coordinate dag failed")
		} else {
			t.ExecuteInfoLog("sync coordinate dag successfully")
		}
	}
	return nil
}

func (t *CreateLocalScaleOutDagTask) getCoordinaterLastMaintainDag() (*task.DagDetailDTO, error) {
	uri := constant.URI_TASK_API_PREFIX + constant.URI_DAG + constant.URI_MAINTAIN + constant.URI_AGENT_GROUP
	var dag task.DagDetailDTO
	for count := 0; count < DEFAULT_REMOTE_REQUEST_RETRY_TIMES; count++ {
		if resp, err := secure.SendGetRequestAndReturnResponse(&t.coordinateAgent, uri, nil, &dag); resp != nil && resp.IsError() {
			t.ExecuteWarnLogf("get current maintain dag failed, err: %v", resp.Error())
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			return nil, errors.Wrap(err, "get current maintain dag failed")
		}
		return &dag, nil
	}
	return nil, nil
}

type WaitScalingReadyTask struct {
	scaleCoordinateTask
}

func newWaitScalingReadyTask() *WaitScalingReadyTask {
	newTask := &WaitScalingReadyTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_WAIT_SCALING_READY),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

// Execute will wait for the ‘Be Scaling Agent Task’ of local scale out dag to complete.
func (t *WaitScalingReadyTask) Execute() error {
	t.initFromData()
	var expectStage int
	if err := t.GetContext().GetDataWithValue(PARAM_WAIT_DEPLOY_RETRY_STAGE, &expectStage); err != nil {
		return errors.Wrap(err, "get expected stage failed")
	}
	expectStage -= 2
	for i := 0; i < WAIT_REMOTE_TASK_FINISH_TIMES; i++ {
		dag, err := t.getCoordinateDag()
		if err != nil {
			t.ExecuteErrorLogf("get remote dag failed, %s", err.Error())
		}
		if dag.Nodes[expectStage].IsSucceed() {
			t.ExecuteInfoLogf("local scale out dag task %s is succeed", dag.Nodes[expectStage].Name)
			return nil
		}
		if dag.Nodes[expectStage].IsFailed() {
			return errors.New("wait scaling out ready failed")
		}
		time.Sleep(WAIT_REMOTE_TASK_FINISH_INTERVAL)
	}
	return errors.New("wait scaling out ready failed, timeout...")
}

func (t *WaitScalingReadyTask) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	return nil
}

type WaitRemoteDeployTaskFinish struct {
	scaleCoordinateTask
}

func newWaitRemoteDeployTaskFinish() *WaitRemoteDeployTaskFinish {
	newTask := &WaitRemoteDeployTaskFinish{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_WAIT_REMOTE_DEPLOY_FINISH),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *WaitRemoteDeployTaskFinish) Execute() error {
	t.initFromData()
	var expectStage int
	if err := t.GetContext().GetDataWithValue(PARAM_WAIT_DEPLOY_RETRY_STAGE, &expectStage); err != nil {
		return errors.Wrap(err, "get expected stage failed")
	}
	expectStage -= 1
	// Send retry.
	for i := 0; i < DEFAULT_REMOTE_REQUEST_RETRY_TIMES; i++ {
		dag, err := t.getCoordinateDag()
		if err != nil {
			t.ExecuteErrorLogf("get remote dag failed, %s", err.Error())
		}
		if dag.Nodes[expectStage].IsFailed() {
			t.tryOperateDag(task.RETRY_STR, expectStage)
			time.Sleep(1 * time.Second)
			continue
		} else if dag.Nodes[expectStage].IsSucceed() {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Wait for succeed.
	for i := 0; i < WAIT_REMOTE_TASK_FINISH_TIMES; i++ {
		dag, err := t.getCoordinateDag()
		if err != nil {
			return errors.Wrap(err, "get remote dag failed")
		}
		node := dag.Nodes[expectStage+1]
		if node.IsFinished() {
			return t.getMessage(dag, node)
		}
		time.Sleep(WAIT_REMOTE_TASK_FINISH_INTERVAL)
	}
	return errors.New("wait remote deploy task finish failed")
}

func (t *scaleCoordinateTask) getMessage(dag *task.DagDetailDTO, node *task.NodeDetailDTO) error {
	var log string
	for _, task := range node.SubTasks {
		for _, log = range task.TaskLogs {
			t.ExecuteLogf("task %s: %s", task.Name, log)
		}
	}
	if node.State == task.FAILED_STR {
		return fmt.Errorf("dag %s %s failed: [%s] %s", dag.GenericID, dag.Name, node.Name, log)
	}
	return nil
}

func (t *WaitRemoteDeployTaskFinish) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	return nil
}

type WaitRemoteStartTaskFinish struct {
	scaleCoordinateTask
}

func newWaitRemoteStartTaskFinish() *WaitRemoteStartTaskFinish {
	newTask := &WaitRemoteStartTaskFinish{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_WAIT_REMOTE_START_FINISH),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *WaitRemoteStartTaskFinish) Execute() error {
	t.initFromData()
	var expectStage int
	if err := t.GetContext().GetDataWithValue(PARAM_WAIT_START_RETRY_STAGE, &expectStage); err != nil {
		return errors.Wrap(err, "get expected stage failed")
	}
	expectStage -= 1
	for i := 0; i < DEFAULT_REMOTE_REQUEST_RETRY_TIMES; i++ {
		dag, err := t.getCoordinateDag()
		if err != nil {
			t.ExecuteErrorLogf("get remote dag failed, %s", err.Error())
		}
		if dag.Nodes[expectStage].IsFailed() {
			t.tryOperateDag(task.RETRY_STR, expectStage)
			time.Sleep(1 * time.Second)
			continue
		} else if dag.Nodes[expectStage].IsSucceed() {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Wait for succeed.
	for i := 0; i < WAIT_REMOTE_TASK_FINISH_TIMES; i++ {
		dag, err := t.getCoordinateDag()
		if err != nil {
			return errors.Wrap(err, "get remote dag failed")
		}
		node := dag.Nodes[expectStage+1]
		if node.IsFinished() {
			return t.getMessage(dag, node)
		}
		time.Sleep(WAIT_REMOTE_TASK_FINISH_INTERVAL)
	}
	return errors.New("wait remote start observer task finish failed")
}

func (t *WaitRemoteStartTaskFinish) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	return nil
}

type PrevCheckTask struct {
	scaleCoordinateTask
}

func newPrevCheckTask() *PrevCheckTask {
	newTask := &PrevCheckTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_PREV_CHECK),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *PrevCheckTask) Execute() error {
	t.ExecuteLog("PrevCheckTask execute")
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("get zone failed")
	}
	isNewZone, ok := t.GetContext().GetParam(PARAM_IS_NEW_ZONE).(bool)
	if !ok {
		return errors.New("get isNewZone failed")
	}
	/* check if the zone is exist */
	isZoneExist, err := obclusterService.IsZoneExistInOB(zone)
	if err != nil {
		return errors.New("check if the zone is exist failed")
	}
	if isZoneExist == isNewZone {
		return errors.New("zone has been changed")
	}
	return nil
}

func (t *PrevCheckTask) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	return nil
}

type AddNewZoneTask struct {
	scaleCoordinateTask
}

func newAddNewZoneTask() *AddNewZoneTask {
	newTask := &AddNewZoneTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_ADD_NEW_ZONE),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *AddNewZoneTask) Execute() error {
	t.ExecuteLog("AddZoneTask execute")
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("get zone failed")
	}
	/* add a new zone */
	if err := obclusterService.AddZone(zone); err != nil {
		return errors.Errorf("add zone %s failed", zone)
	}
	return nil
}

func (t *AddNewZoneTask) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	t.ExecuteLog("AddZoneTask rollback")
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("get zone failed")
	}
	/* delete a zone */
	exist, err := obclusterService.IsZoneExistInOB(zone)
	if err != nil {
		return errors.Errorf("check zone %s exist failed", zone)
	}
	if !exist {
		return nil
	}
	return obclusterService.DeleteZone(zone)
}

type StartNewZoneTask struct {
	scaleCoordinateTask
}

func newStartNewZoneTask() *StartNewZoneTask {
	newTask := &StartNewZoneTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_START_NEW_ZONE),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *StartNewZoneTask) Execute() error {
	t.ExecuteLog("StartZoneTask execute")
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("get zone failed")
	}
	/* start zone */
	if err := obclusterService.StartZone(zone); err != nil {
		return errors.Errorf("start zone %s failed", zone)
	}
	return nil
}

func (t *StartNewZoneTask) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	t.ExecuteLog("StartZoneTask rollback")
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("get zone failed")
	}
	/* stop zone */
	if err := obclusterService.StopZone(zone); err != nil {
		return errors.Errorf("stop zone %s failed", zone)
	}
	return nil
}

/* add server task: add observer to ob cluster */
type AddServerTask struct {
	scaleCoordinateTask
}

func newAddServerTask() *AddServerTask {
	newTask := &AddServerTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_ADD_SERVER),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}
func (t *AddServerTask) Execute() error {
	t.ExecuteLog("AddServerTask execute")
	var agentInfo meta.AgentInfo
	ctx := t.GetContext()
	if err := ctx.GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
		return errors.Wrap(err, "get agent info failed")
	}
	var configs map[string]string
	if err := ctx.GetDataWithValue(PARAM_CONFIG, &configs); err != nil {
		return errors.Wrap(err, "get configs failed")
	}
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("get zone failed")
	}

	port, err := strconv.Atoi(configs[constant.CONFIG_RPC_PORT])
	if err != nil {
		return errors.Wrap(err, "convert port to integer failed")
	}

	serverInfo := meta.NewAgentInfo(agentInfo.Ip, port)
	err = obclusterService.AddServer(*serverInfo, zone)
	if err != nil {
		return errors.Errorf("add server %s failed", serverInfo.String())
	}

	t.GetContext().SetParam(PARAM_ADD_SERVER_SUCCEED, true)

	return nil
}

func (t *AddServerTask) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	var agentInfo meta.AgentInfo
	ctx := t.GetContext()
	if err := ctx.GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
		return errors.Wrap(err, "get agent info failed")
	}
	if agentInfo.Equal(meta.OCS_AGENT) {
		agentService.BeScalingOutAgent(t.GetContext().GetParam(PARAM_ZONE).(string))
		if err := scalingSelfRollback(); err != nil {
			return err
		}
	}

	if t.GetContext().GetParam(PARAM_ADD_SERVER_SUCCEED) == nil {
		// If add server falied, should not delete from obcluster.
		return nil
	}

	var configs map[string]string
	if err := ctx.GetDataWithValue(PARAM_CONFIG, &configs); err != nil {
		return errors.Wrap(err, "get configs failed")
	}
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("get zone failed")
	}

	port, err := strconv.Atoi(configs[constant.CONFIG_RPC_PORT])
	if err != nil {
		return errors.Wrap(err, "convert rpc port to integer failed")
	}
	serverInfo := meta.NewAgentInfo(agentInfo.Ip, port)

	// Check whether addserver task execute successfully.
	exist, err := obclusterService.IsServerExistWithZone(*serverInfo, zone)
	if err != nil {
		return errors.Errorf("check server %s exist failed", agentInfo.String())
	}
	if !exist {
		return nil
	}

	if err = obclusterService.DeleteServerInZone(*serverInfo, zone); err != nil {
		return errors.Errorf("delete server %s failed", serverInfo.String())
	}
	return nil
}

type AddAgentTask struct {
	scaleCoordinateTask
}

func newAddAgentTask() *AddAgentTask {
	newTask := &AddAgentTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_ADD_AGENT),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *AddAgentTask) Execute() error {
	var agentInfo meta.AgentInfo
	ctx := t.GetContext()
	if err := ctx.GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
		return errors.Wrap(err, "get agent info failed")
	}
	var configs map[string]string
	if err := ctx.GetDataWithValue(PARAM_CONFIG, &configs); err != nil {
		return errors.Wrap(err, "get configs failed")
	}
	zone, ok := t.GetContext().GetParam(PARAM_ZONE).(string)
	if !ok {
		return errors.New("get zone failed")
	}
	var scalingAgent param.JoinMasterParam
	if err := ctx.GetDataWithValue(PARAM_JOIN_MASTER_INFO, &scalingAgent); err != nil {
		return errors.Wrap(err, "get scaling agent info failed")
	}
	mysqlPort, err := strconv.Atoi(configs[constant.CONFIG_MYSQL_PORT])
	if err != nil {
		return errors.Wrap(err, "get mysql port failed")
	}
	rpcPort, err := strconv.Atoi(configs[constant.CONFIG_RPC_PORT])
	if err != nil {
		return errors.Wrap(err, "get rpc port failed")
	}
	agentInstance := oceanbaseModel.AllAgent{
		Ip:           agentInfo.Ip,
		Port:         agentInfo.Port,
		Identity:     string(meta.CLUSTER_AGENT),
		Os:           scalingAgent.Os,
		Architecture: scalingAgent.Architecture,
		Version:      scalingAgent.Version,
		Zone:         zone,
		HomePath:     scalingAgent.HomePath,
		PublicKey:    scalingAgent.PublicKey,
		MysqlPort:    mysqlPort,
		RpcPort:      rpcPort,
	}
	if err := agentService.AddAgentInOB(agentInstance); err != nil {
		return errors.New("add agent failed")
	}
	return nil
}
func (t *AddAgentTask) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	var agentInfo meta.AgentInfo
	if err := t.GetContext().GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
		return errors.Wrap(err, "get agent info failed")
	}
	if agentInfo.Equal(meta.OCS_AGENT) {
		agentService.BeScalingOutAgent(t.GetContext().GetParam(PARAM_ZONE).(string))
		if err := scalingSelfRollback(); err != nil {
			return err
		}
	}
	if err := agentService.DeleteAgentInOB(&agentInfo); err != nil {
		t.ExecuteErrorLogf("delete agent %s failed", agentInfo.String())
	}
	return nil
}

type FinishTask struct {
	scaleCoordinateTask
}

func newFinishTask() *FinishTask {
	newTask := &FinishTask{
		scaleCoordinateTask: *newScaleCoordinateTask(TASK_NAME_FINISH),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *FinishTask) Execute() error {
	t.initFromData()
	t.ExecuteLog("FinishTask execute")
	var agentInfo meta.AgentInfo
	if err := t.GetContext().GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
		return errors.Wrap(err, "get agent info failed")
	}
	var agent meta.AgentStatus
	for i := 0; i < DEFAULT_REMOTE_REQUEST_RETRY_TIMES; i++ {
		// Get the identity of the agent.
		if err := http.SendGetRequest(&agentInfo, constant.URI_API_V1+constant.URI_INFO, nil, &agent); err != nil {
			t.ExecuteWarnLogf("send info api to %s failed", agentInfo.String())
			time.Sleep(1 * time.Second)
			continue
		}
		if agent.IsClusterAgent() {
			break
		}
		time.Sleep(1 * time.Second)
		continue
	}
	if !agent.IsClusterAgent() {
		return errors.New("agent is not cluster agent")
	}
	return nil
}

func (t *FinishTask) Rollback() error {
	t.initFromData()
	if _, err := t.syncCoordinateDag(); err != nil {
		return errors.Wrap(err, "sync coordinate dag failed")
	}
	var agentInfo meta.AgentInfo
	if err := t.GetContext().GetParamWithValue(PARAM_AGENT_INFO, &agentInfo); err != nil {
		return errors.Wrap(err, "get agent info failed")
	}
	if agentInfo.Equal(meta.OCS_AGENT) {
		agentService.BeScalingOutAgent(t.GetContext().GetParam(PARAM_ZONE).(string))
		if err := scalingSelfRollback(); err != nil {
			return err
		}
	}

	return nil
}

func scalingSelfRollback() error {
	coordinator.OCS_COORDINATOR.Suspend()
	coordinator.OCS_COORDINATOR.Resume()
	executor.OCS_SYNCHRONIZER.Restart()
	if err := localTaskService.DeleteRemoteTask(); err != nil {
		return err
	}
	return nil
}
