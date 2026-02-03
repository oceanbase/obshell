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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/config"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/process"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/param"
)

type StartObserverTask struct {
	RemoteExecutableTask
	config          map[string]string
	mysqlPort       int
	rpcPort         int
	needHealthCheck bool
}

func newStartObServerTask() *StartObserverTask {
	newTask := &StartObserverTask{
		RemoteExecutableTask: *newRemoteExecutableTask(TASK_NAME_START),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanRollback().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func CreateStartSelfDag(config map[string]string, healthCheck bool) (*task.DagDetailDTO, error) {
	subTask := newStartObServerTask()
	builder := task.NewTemplateBuilder(subTask.GetName())
	builder.AddTask(subTask, false).SetMaintenance(task.GlobalMaintenance())
	dag, err := localTaskService.CreateDagInstanceByTemplate(builder.Build(), task.NewTaskContext().SetAgentData(meta.OCS_AGENT, PARAM_CONFIG, config).SetParam(PARAM_HEALTH_CHECK, healthCheck))
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func (t *StartObserverTask) Execute() error {
	if _, ok := t.GetContext().GetData(DATA_SKIP_START_TASK).(bool); ok {
		t.ExecuteLog("observer started.")
		return nil
	}

	// t.needHealthCheck defaults to false, so we ignore the error here.
	t.GetContext().GetParamWithValue(PARAM_HEALTH_CHECK, &t.needHealthCheck)

	agent := t.GetExecuteAgent()
	t.ExecuteLog("start observer")
	if err := t.GetContext().GetAgentDataWithValue(&agent, PARAM_CONFIG, &t.config); err != nil {
		if t.GetLocalData(PARAM_CONFIG) != nil ||
			(!meta.OCS_AGENT.IsClusterAgent() && t.GetContext().GetParam(PARAM_START_OWN_OBSVR) == nil) {
			return err
		}
		t.config = make(map[string]string)
	}

	if err := t.setPort(); err != nil {
		return err
	}

	if !agent.Equal(meta.OCS_AGENT) {
		return t.remoteStart()
	}

	if err := startObserver(t, t.config); err != nil {
		return err
	}

	if t.needHealthCheck {
		if err := t.observerHealthCheck(t.mysqlPort); err != nil {
			return errors.Wrap(err, "observer health check failed")
		}
	}
	return t.updatePort()
}

func (t *StartObserverTask) observerHealthCheck(mysqlPort int) error {
	dsConfig := config.NewObMysqlDataSourceConfig().
		SetTryTimes(1).
		SetDBName("").
		SetTimeout(10).
		SetPort(mysqlPort)

	const (
		maxRetries    = 300
		retryInterval = 2 * time.Second
	)

	for retryCount := 1; retryCount <= maxRetries; retryCount++ {
		time.Sleep(retryInterval)
		if retryCount%10 == 0 {
			t.TimeoutCheck()
		} else {
			t.ExecuteLogf("observer health check, retry [%d/%d]", retryCount, maxRetries)
		}

		// Check if the observer process exists
		if exist, err := process.CheckObserverProcess(); err != nil {
			return errors.Occur(errors.ErrObServerProcessCheckFailed, err.Error())
		} else if !exist {
			return errors.Occur(errors.ErrObServerProcessNotExist)
		}

		// Check if the SSTable file exists
		_, err := os.Stat(path.ObBlockFilePath())
		if os.IsNotExist(err) {
			continue
		}

		// Attempt to connect to the OceanBase instance for testing
		if err := oceanbase.LoadOceanbaseInstanceForTest(dsConfig); err != nil {
			continue // Connection failed, retry
		}

		// All checks passed, exit the loop
		return nil
	}

	// If retries run out, return a timeout error
	return errors.Occur(errors.ErrTaskDagExecuteTimeout, "observer health check")
}

func (t *StartObserverTask) setPort() (err error) {
	if val, ok := t.config[constant.CONFIG_MYSQL_PORT]; ok {
		t.mysqlPort, err = strconv.Atoi(val)
		if err != nil {
			return errors.Occur(errors.ErrCommonInvalidPort, val)
		}
	}
	if val, ok := t.config[constant.CONFIG_RPC_PORT]; ok {
		t.rpcPort, err = strconv.Atoi(val)
		if err != nil {
			return errors.Occur(errors.ErrCommonInvalidPort, val)
		}
	}
	return nil
}

func (t *StartObserverTask) updatePort() error {
	t.ExecuteLog("update self OB port")
	if err := agentService.UpdatePort(t.mysqlPort, t.rpcPort); err != nil {
		return err
	}
	t.ExecuteLog("update OB port in all_agent")
	agent := t.GetExecuteAgent()
	return agentService.UpdateAgentOBPort(&agent, t.mysqlPort, t.rpcPort)
}

func (t *StartObserverTask) Rollback() error {
	if _, ok := t.GetContext().GetData(DATA_SKIP_START_TASK).(bool); ok {
		return nil
	}

	agent := t.GetExecuteAgent()
	if !agent.Equal(meta.OCS_AGENT) {
		return t.remoteStop()
	}

	return stopObserver(t)
}

func (t *StartObserverTask) remoteStart() error {
	t.initial(constant.URI_OB_RPC_PREFIX+constant.URI_START, http.POST, param.StartTaskParams{Config: t.config, HealthCheck: t.needHealthCheck})
	if err := t.retmoteExecute(); err != nil {
		return err
	}
	agent := t.GetExecuteAgent()
	return agentService.UpdateAgentOBPort(&agent, t.mysqlPort, t.rpcPort)
}

func (t *StartObserverTask) remoteStop() error {
	t.initial(constant.URI_OB_RPC_PREFIX+constant.URI_STOP, http.POST, param.StopTaskParams{Force: true})
	t.rollbackTaskName = TASK_NAME_STOP
	return t.remoteRollback()
}

func startObserver(t task.ExecutableTask, config map[string]string) error {
	t.ExecuteLog("check if first start")
	if err := requireCheck(t, config); err != nil {
		return err
	}
	t.ExecuteLog("generate start cmd")
	cmd := generateStartCmd(config)
	t.ExecuteLogf("start cmd: %s", cmd)
	return execStartCmd(cmd)
}

// SafeStartObserver is a safe method to start the observer, ensuring that it has been successfully started at least once before using.
// This method allows an empty config and does not check whether the config contains the necessary startup configuration items.
func SafeStartObserver(config map[string]string) error {
	if isFirst, err := isFirstStart(); err != nil {
		return err
	} else if isFirst {
		return errors.Occur(errors.ErrObServerHasNotBeenStarted)
	}

	if config == nil {
		config = make(map[string]string)
	}
	cmd := generateStartCmd(config)

	log.Info("safty start observer, cmd: ", cmd)
	return execStartCmd(cmd)
}

func generateStartCmd(config map[string]string) string {
	cmd := fmt.Sprintf("export LD_LIBRARY_PATH='%s/lib'; %s/bin/observer ", global.HomePath, global.HomePath)
	startOptionsCmd := generateStartOpitonCmd(config)
	additionalCmd := generateAdditionalStartCmd(config)
	return fmt.Sprintf("%s %s %s", cmd, startOptionsCmd, additionalCmd)
}

func requireCheck(t task.ExecutableTask, config map[string]string) error {
	if isFirst, err := isFirstStart(); err != nil {
		return err
	} else if !isFirst {
		t.ExecuteLog("not first start, skip require check")
		return nil
	}

	fillStartConfig(config)
	for _, key := range requiredConfigItems {
		if _, ok := config[key]; !ok {
			return errors.Occurf(errors.ErrCommonUnexpected, "config %s is required", key)
		}
	}
	return nil
}

func fillStartConfig(config map[string]string) {
	if _, ok := config[constant.CONFIG_LOCAL_IP]; !ok {
		config[constant.CONFIG_LOCAL_IP] = meta.OCS_AGENT.GetIp()
	}
	if _, ok := config[constant.CONFIG_DATA_DIR]; !ok {
		config[constant.CONFIG_DATA_DIR] = filepath.Join(global.HomePath, constant.OB_DIR_STORE)
	}
}

func generateStartOpitonCmd(config map[string]string) string {
	cmd := ""
	agentIp := config[constant.CONFIG_LOCAL_IP]
	for name, value := range startOptionsMap {
		if val, ok := config[name]; ok {
			if name == constant.CONFIG_RS_LIST {
				cmd += fmt.Sprintf("%s '%s' ", value, val)
			} else {
				cmd += fmt.Sprintf("%s %s ", value, val)
			}
			delete(config, name)
		}
	}

	if meta.NewAgentInfo(agentIp, 0).IsIPv6() {
		cmd += "--ipv6 "
	}
	return cmd
}

func generateAdditionalStartCmd(config map[string]string) string {
	additionalOpts := make([]string, 0)
	// Delete non start items.
	for _, name := range nonStartItems {
		delete(config, name)
	}
	for k, v := range config {
		additionalOpts = append(additionalOpts, fmt.Sprintf("%s=%s", k, v))
	}
	if len(additionalOpts) == 0 {
		return ""
	}
	return fmt.Sprintf("-o '%s'", strings.Join(additionalOpts, ","))
}

func execStartCmd(bash string) error {
	if err := os.Chdir(global.HomePath); err != nil {
		return err
	}
	cmd := exec.Command("/bin/bash", "-c", bash)
	if stderr, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrap(err, string(stderr))
	}
	return nil
}

func CreateStartDag(params CreateSubDagParam) (*CreateSubDagResp, error) {
	lastMainDagId, err := checkMaintenanceAndPassDag(params.ID)
	if err != nil {
		return &CreateSubDagResp{ForcePassDagParam: param.ForcePassDagParam{ID: []string{lastMainDagId}}}, err
	}

	taskCtx := task.NewTaskContext().
		SetParam(PARAM_MAIN_DAG_ID, params.GenericID).
		SetParam(PARAM_EXPECT_MAIN_NEXT_STAGE, params.ExpectedStage).
		SetParam(PARAM_MAIN_AGENT, params.Agent).
		SetParam(PARAM_SCOPE, params.Scope).
		SetParam(PARAM_MAIN_DAG_NAME, params.MainDagName)

	resp := &CreateSubDagResp{
		SubDagInfo: SubDagInfo{
			ExpectedStage: MAIN_START_DAG_EXPECTED_SUB_NEXT_STAGE,
		},
	}

	template := task.NewTemplateBuilder(DAG_START_OBSERVER).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newCheckDagStageTask(), false)
	if params.NeedExecCmd {
		// If not need start, then skip start observer.
		template.AddTask(newCheckObserverForStartTask(), false).
			AddTask(newStartObServerTask(), false).
			AddTask(newAlterStartServerTask(), false)
		resp.ExpectedStage += 3
	}
	template.AddTask(newWaitPassOperatorTask(), false)

	dag, err := localTaskService.CreateDagInstanceByTemplate(template.Build(), taskCtx)
	if err != nil {
		return resp, err
	}
	resp.GenericID = task.NewDagDetailDTO(dag).GenericID
	return resp, nil
}

type CheckDagStageTask struct {
	task.Task
	expectedStage int
	mainDagID     string
	mainAgent     meta.AgentInfo
}

type CheckObserverForStartTask struct {
	task.Task
}

type AlterStartServerTask struct {
	task.Task
}

type WaitPassOperatorTask struct {
	task.Task
}

func newCheckDagStageTask() *CheckDagStageTask {
	newTask := &CheckDagStageTask{
		Task: *task.NewSubTask(TASK_START_PREPARATIONS),
	}
	newTask.
		SetCanContinue().
		SetCanPass().
		SetCanCancel().
		SetCanRetry()
	return newTask
}

func newCheckObserverForStartTask() *CheckObserverForStartTask {
	newTask := &CheckObserverForStartTask{
		Task: *task.NewSubTask(TASK_CHECK_OB_PROC_AND_CONIFG),
	}
	newTask.
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func newAlterStartServerTask() *AlterStartServerTask {
	newTask := &AlterStartServerTask{
		Task: *task.NewSubTask(TASK_EXEC_START_OBSERVER_SQL),
	}
	newTask.
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func newWaitPassOperatorTask() *WaitPassOperatorTask {
	newTask := &WaitPassOperatorTask{
		Task: *task.NewSubTask(TASK_WAIT_FOR_TASK_TO_END),
	}
	newTask.
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *CheckDagStageTask) Execute() error {
	if err := t.getParams(); err != nil {
		return err
	}
	log.Infof("wait for dag %s to reach stage %d", t.mainDagID, t.expectedStage)
	if t.isDagReachExpectedStage() {
		t.ExecuteLog("start to execute the task")
		return nil
	}
	t.ExecuteLog("wait for notification to start task")
	return task.ERR_WAIT_OPERATOR
}

func (t *CheckDagStageTask) getParams() error {
	taskCtx := t.GetContext()
	if err := taskCtx.GetParamWithValue(PARAM_EXPECT_MAIN_NEXT_STAGE, &t.expectedStage); err != nil {
		return err
	}
	if err := taskCtx.GetParamWithValue(PARAM_MAIN_DAG_ID, &t.mainDagID); err != nil {
		return err
	}
	if err := taskCtx.GetParamWithValue(PARAM_MAIN_AGENT, &t.mainAgent); err != nil {
		return err
	}
	return nil
}

func (t *CheckDagStageTask) GetAdditionalData() map[string]interface{} {
	t.getParams()
	return map[string]interface{}{
		ADDL_KEY_MAIN_DAG_ID: t.mainDagID,
	}
}

func (t *CheckDagStageTask) isDagReachExpectedStage() bool {
	var dagDetailDTO *task.DagDetailDTO
	err := secure.SendGetRequest(&t.mainAgent, constant.URI_TASK_API_PREFIX+constant.URI_DAG+"/"+t.mainDagID, nil, &dagDetailDTO)
	if err != nil {
		t.ExecuteErrorLog(err)
		return false
	}
	log.Infof("dag %s current stage %d", t.mainDagID, dagDetailDTO.Stage)
	return dagDetailDTO.Stage >= t.expectedStage
}

func (t *CheckObserverForStartTask) Execute() error {
	t.ExecuteLog("check if first start")
	isFirst, err := isFirstStart()
	if err != nil {
		return err
	}
	if isFirst {
		if t.GetContext().GetParam(PARAM_START_OWN_OBSVR) != nil {
			return errors.Occur(errors.ErrCommonUnexpected, "observer has not started yet")
		}
		t.ExecuteLog("first start, skip check observer process")
		t.GetContext().SetData(DATA_SKIP_START_TASK, true)
		return nil
	}
	t.ExecuteLog("check observer process config")
	exist, err := process.CheckObserverProcess()
	if err != nil {
		return errors.Wrap(err, "check observer process failed")
	}
	if exist {
		t.ExecuteLog("observer process exist")
		if err := t.checkObsvrProcConfig(); err != nil {
			return errors.Wrap(err, "The observer process and configs are inconsistent")
		}
	}
	return nil
}

func (t *CheckObserverForStartTask) checkObsvrProcConfig() error {
	t.GetContext().SetData(DATA_SKIP_START_TASK, true)
	return nil
}

func (t *AlterStartServerTask) Execute() error {
	t.ExecuteInfoLog("exec start server sql")
	var rpcPort int
	if err := observerService.GetObConfigValueByName(constant.CONFIG_RPC_PORT, &rpcPort); err != nil {
		return errors.Wrap(err, "get rpc port failed")
	}
	if err := getOceanbaseInstance(); err != nil {
		return err
	}

	sql := fmt.Sprintf("alter system start server '%s'", meta.NewAgentInfo(meta.OCS_AGENT.GetIp(), rpcPort).String())
	log.Info(sql)
	if err := obclusterService.ExecuteSql(sql); err != nil {
		return errors.Wrap(err, "alter start server failed")
	}
	return nil
}

func getOceanbaseInstance() (err error) {
	for i := 1; i <= constant.MAX_GET_INSTANCE_RETRIES; i++ {
		if _, err = oceanbase.GetInstance(); err == nil {
			return nil
		}
		log.Infof("get db instance failed: %v , retry [%d/%d]", err, i, constant.MAX_GET_INSTANCE_RETRIES)
		time.Sleep(time.Second * constant.GET_INSTANCE_RETRY_INTERVAL)
	}
	return errors.Wrap(err, "get db instance timeout")
}

func (t *WaitPassOperatorTask) Execute() error {
	return task.ERR_WAIT_OPERATOR
}
