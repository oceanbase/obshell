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
package inspection

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	obconstant "github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/inspection/constant"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/path"
	"github.com/oceanbase/obshell/ob/agent/lib/pkg"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

func TriggerInspection(p *param.InspectionParam) (*task.DagDetailDTO, error) {
	if strings.ToUpper(p.Scenario) != constant.SCENARIO_BASIC && strings.ToUpper(p.Scenario) != constant.SCENARIO_PERFORMANCE {
		return nil, errors.Occur(errors.ErrObClusterInspectionScenarioNotSupported, p.Scenario, constant.SCENARIO_BASIC, constant.SCENARIO_PERFORMANCE)
	}

	obdiagBinPath, needInstall, useWorkPath, err := checkObdiagAvailability()
	if err != nil {
		return nil, err
	}

	usePasswordlessSSH, err := checkSSHAccess()
	if err != nil {
		return nil, err
	}

	taskCtx := task.NewTaskContext().
		SetParam(PARAM_SCENARIO, p.Scenario).
		SetParam(PARAM_USE_PASSWORDLESS_SSH, usePasswordlessSSH).
		SetParam(PARAM_USE_WORK_PATH, useWorkPath)

	if !needInstall && obdiagBinPath != "" {
		taskCtx.SetParam(PARAM_OBDIAG_BIN_PATH, obdiagBinPath)
	}

	taskTemplate := task.NewTemplateBuilder(DAG_TRIGGER_INSPECTION)

	if needInstall {
		taskTemplate.AddTask(newInstallObdiagTask(), false)
	}

	taskTemplate.
		AddTask(newGenerateConfigTask(), false).
		AddTask(newInspectionTask(), false).
		AddTask(newGenerateReportTask(), false)

	dag, err := localTaskService.CreateDagInstanceByTemplate(taskTemplate.Build(), taskCtx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

type InstallObdiagTask struct {
	task.Task
	version      string
	release      string
	distribution string
	architecture string
}

func newInstallObdiagTask() *InstallObdiagTask {
	newTask := &InstallObdiagTask{
		Task: *task.NewSubTask(TASK_NAME_INSTALL_OBDIAG),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *InstallObdiagTask) Execute() error {
	var err error
	defer setupInspectionReportStatusUpdate(t)(&err)

	pkgInfo, err := obclusterService.GetLatestUpgradePkgInfo(obconstant.PKG_OCEANBASE_DIAGNOSTIC_TOOL, global.Architecture, obconstant.DIST)
	if err != nil {
		return errors.Wrap(err, "failed to get latest obdiag package from OCS")
	}

	t.version = pkgInfo.Version
	t.release = pkgInfo.Release
	t.distribution = pkgInfo.Distribution
	t.architecture = pkgInfo.Architecture

	if err = os.RemoveAll(path.ObdiagPath()); err != nil {
		return errors.Wrap(err, "failed to remove obdiag directory")
	}
	if err = os.MkdirAll(path.ObdiagPath(), 0755); err != nil {
		return errors.Wrap(err, "failed to create obdiag directory")
	}
	rpmPath := filepath.Join(path.ObdiagPath(), fmt.Sprintf("obdiag-%s-%s-%s-%s.rpm", t.version, t.release, t.distribution, t.architecture))

	if err = obclusterService.DownloadUpgradePkgChunkInBatch(rpmPath, pkgInfo.PkgId, pkgInfo.ChunkCount); err != nil {
		return errors.Wrap(err, "failed to download obdiag package")
	}
	if err := pkg.InstallRpmPkgToTargetDir(rpmPath, path.ObdiagPath()); err != nil {
		return errors.Wrap(err, "failed to install obdiag package")
	}
	if err = os.Chmod(path.ObdiagBinPath(), 0755); err != nil {
		return errors.Wrap(err, "failed to chmod obdiag binary")
	}

	if err := os.Remove(rpmPath); err != nil {
		return err
	}

	obdiagBinPath := path.ObdiagBinPath()
	t.GetContext().SetParam(PARAM_OBDIAG_BIN_PATH, obdiagBinPath)
	t.GetContext().SetParam(PARAM_USE_WORK_PATH, true)
	t.ExecuteLogf("Obdiag installed successfully, path: %s", obdiagBinPath)

	return nil
}

func (t *InstallObdiagTask) Rollback() error {
	obdiagDirPath := filepath.Join(path.AgentDir(), fmt.Sprintf("oceanbase-diagnostic-tool/%s-%s/%s/%s", t.version, t.release, t.distribution, t.architecture))
	if err := os.RemoveAll(obdiagDirPath); err != nil {
		t.ExecuteWarnLogf("failed to rollback install obdiag, err: %s", err)
	}
	return nil
}

type GenerateConfigTask struct {
	task.Task
	scenario string
}

func newGenerateConfigTask() *GenerateConfigTask {
	newTask := &GenerateConfigTask{
		Task: *task.NewSubTask(TASK_NAME_GENERATE_CONFIG),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *GenerateConfigTask) Execute() error {
	var err error
	defer setupInspectionReportStatusUpdate(t)(&err)

	if err := t.GetContext().GetParamWithValue(PARAM_SCENARIO, &t.scenario); err != nil {
		return err
	}
	t.scenario = strings.ToUpper(t.scenario)

	var usePasswordlessSSH bool
	if err := t.GetContext().GetParamWithValue(PARAM_USE_PASSWORDLESS_SSH, &usePasswordlessSSH); err != nil {
		return err
	}

	t.ExecuteLogf("Generating inspection config for scenario: %s, use_passwordless_ssh: %v", t.scenario, usePasswordlessSSH)

	config := make(map[string]string)
	serversWithDataDir, err := obclusterService.GetParametersByName(obconstant.CONFIG_DATA_DIR)
	if err != nil {
		return err
	}

	observers, err := obclusterService.GetAllOBServers()
	if err != nil {
		return err
	}

	agents, err := agentService.GetAllAgentsDOFromOB()
	if err != nil {
		return err
	}

	agentHomePathMap := make(map[string]string)
	for _, agent := range agents {
		key := fmt.Sprintf("%s:%d", agent.Ip, agent.RpcPort)
		agentHomePathMap[key] = agent.HomePath
	}

	dataDirMap := make(map[string]string)
	for _, server := range serversWithDataDir {
		key := fmt.Sprintf("%s:%d", server.SvrIp, server.SvrPort)
		dataDirMap[key] = server.Value
	}

	for i, observer := range observers {
		config[fmt.Sprintf(constant.CONFIG_OBCLUSTER_SERVERS_NODES_IP, i)] = observer.SvrIp
		key := fmt.Sprintf("%s:%d", observer.SvrIp, observer.SvrPort)

		var dataDir string
		if dir, exists := dataDirMap[key]; exists {
			dataDir = dir
		} else if homePath, exists := agentHomePathMap[key]; exists && homePath != "" {
			dataDir = fmt.Sprintf("%s/store", homePath)
		}

		if dataDir != "" {
			config[fmt.Sprintf(constant.CONFIG_OBCLUSTER_SERVERS_NODES_DATA_DIR, i)] = dataDir
			config[fmt.Sprintf(constant.CONFIG_OBCLUSTER_SERVERS_NODES_REDO_DIR, i)] = dataDir
		}

		if homePath, exists := agentHomePathMap[key]; exists && homePath != "" {
			config[fmt.Sprintf(constant.CONFIG_OBCLUSTER_SERVERS_NODES_HOME_PATH, i)] = homePath
		}
	}

	config[constant.CONFIG_TENANT_SYS_USER] = "root@sys"

	value, err := obclusterService.GetParameterByName(obconstant.OB_PARAM_CLUSTER_NAME)
	if err != nil {
		return err
	}
	config[constant.CONFIG_CLUSTER_NAME] = value.Value

	t.GetContext().SetData(DATA_INSPECTION_CONFIG, config)
	t.ExecuteLog("Inspection config generated successfully")
	return nil
}

type InspectionTask struct {
	task.Task
	scenario           string
	configs            map[string]string
	usePasswordlessSSH bool
	useWorkPath        bool
}

func newInspectionTask() *InspectionTask {
	newTask := &InspectionTask{
		Task: *task.NewSubTask(TASK_NAME_INSPECTION),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *InspectionTask) Execute() error {
	var err error
	defer setupInspectionReportStatusUpdate(t)(&err)

	if err := t.GetContext().GetParamWithValue(PARAM_SCENARIO, &t.scenario); err != nil {
		return err
	}
	t.scenario = strings.ToUpper(t.scenario)

	if err := t.GetContext().GetDataWithValue(DATA_INSPECTION_CONFIG, &t.configs); err != nil {
		return err
	}

	if err := t.GetContext().GetParamWithValue(PARAM_USE_PASSWORDLESS_SSH, &t.usePasswordlessSSH); err != nil {
		return err
	}
	if err := t.GetContext().GetParamWithValue(PARAM_USE_WORK_PATH, &t.useWorkPath); err != nil {
		return err
	}

	t.GetContext().SetData(DATA_INSPECTION_START_TIME, time.Now())

	result, err := t.executeInspection()
	if err != nil {
		return err
	}
	t.GetContext().SetData(DATA_INSPECTION_FINISH_TIME, time.Now())
	t.GetContext().SetData(DATA_INSPECTION_RESULT, result)
	return nil
}

func (t *InspectionTask) executeInspection() (result string, err error) {
	t.ExecuteLogf("Executing %s inspection...", t.scenario)

	obdiagBinPath := constant.BINARY_OBDIAG
	if pathVal := t.GetContext().GetParam(PARAM_OBDIAG_BIN_PATH); pathVal != nil {
		if path, ok := pathVal.(string); ok && path != "" {
			obdiagBinPath = path
		}
	}

	if !t.usePasswordlessSSH {
		if err := t.fillSSHCredentialConfig(); err != nil {
			return "", err
		}
	}

	args := []string{
		"check", "run",
	}
	for k, v := range t.configs {
		args = append(args, "--config", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, "--config", fmt.Sprintf("%s=127.0.0.1", constant.CONFIG_DB_HOST))
	args = append(args, "--config", fmt.Sprintf("%s=%d", constant.CONFIG_DB_PORT, meta.MYSQL_PORT))
	args = append(args, "--config", fmt.Sprintf("%s=%s", constant.CONFIG_TENANT_SYS_PASSWORD, meta.GetOceanbasePwd()))

	var observerTasks string
	switch t.scenario {
	case constant.SCENARIO_BASIC:
		observerTasks = strings.Join(constant.INSPECTION_BASIC_OBSERVER_TASKS, ";")
	case constant.SCENARIO_PERFORMANCE:
		observerTasks = strings.Join(constant.INSPECTION_PERFORMANCE_OBSERVER_TASKS, ";")
	}
	args = append(args, fmt.Sprintf("--observer_tasks=\"%s\"", observerTasks), "--inner_config", "obdiag.logger.silent=True")
	if t.useWorkPath {
		checkWorkPath := path.ObdiagCheckPluginsPath()
		args = append(args, "--inner_config", fmt.Sprintf("check.work_path=%s", checkWorkPath))
	}

	var argsForPrint []string
	for _, arg := range args {
		if strings.Contains(arg, "password") {
			argsForPrint = append(argsForPrint, fmt.Sprintf("%s=********", strings.Split(arg, "=")[0]))
		} else {
			argsForPrint = append(argsForPrint, arg)
		}
	}
	t.ExecuteLogf("Executing command: %s", strings.Join(append([]string{obdiagBinPath}, argsForPrint...), " "))

	output, err := exec.Command(obdiagBinPath, args...).CombinedOutput()
	res := strings.TrimSpace(string(output))
	if err != nil {
		return "", errors.Wrap(err, "failed to execute obdiag") // ignore the output, because the output has sensitive information
	}
	return res, nil
}

type GenerateReportTask struct {
	task.Task
	result   string
	scenario string
}

func newGenerateReportTask() *GenerateReportTask {
	newTask := &GenerateReportTask{
		Task: *task.NewSubTask(TASK_NAME_GENERATE_REPORT),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *GenerateReportTask) Execute() error {
	var err error
	defer setupInspectionReportStatusUpdate(t)(&err)

	t.ExecuteLog("Generating inspection report...")

	err = t.GetContext().GetDataWithValue(DATA_INSPECTION_RESULT, &t.result)
	if err != nil {
		return err
	}

	err = t.GetContext().GetParamWithValue(PARAM_SCENARIO, &t.scenario)
	if err != nil {
		return err
	}
	t.scenario = strings.ToUpper(t.scenario)

	var startTime time.Time
	err = t.GetContext().GetDataWithValue(DATA_INSPECTION_START_TIME, &startTime)
	if err != nil {
		return err
	}
	var finishTime time.Time
	err = t.GetContext().GetDataWithValue(DATA_INSPECTION_FINISH_TIME, &finishTime)
	if err != nil {
		return err
	}

	report, err := parseResult(t.result)
	if err != nil {
		return err
	}

	reportBriefInfo, err := parseReportBriefInfo(report)
	if err != nil {
		return err
	}

	marshalReport, err := json.Marshal(report)
	if err != nil {
		return err
	}

	// Get DAG ID to find the inspection report
	dag, err := localTaskService.GetDagBySubTaskId(t.GetID())
	if err != nil {
		return errors.Wrap(err, "failed to get dag by task id")
	}

	// Get Generic ID for the dag
	dagGenericID := task.ConvertIDToGenericID(dag.GetID(), dag.IsLocalTask(), dag.GetDagType())

	// Get existing inspection report (must exist at this point), including SUCCEED status
	existingReport, err := inspectionService.GetInspectionReportByLocalTaskIdIncludeSucceed(dagGenericID)
	if err != nil {
		return errors.Wrap(err, "inspection report not found, should have been created earlier")
	}

	// If report is already SUCCEED, skip update
	if existingReport.Status == constant.INSPECTION_STATUS_SUCCEED {
		return nil
	}

	// Update existing report with complete information and SUCCEED status
	updateReport := &oceanbase.InspectionReport{
		Id:            existingReport.Id,
		Report:        string(marshalReport),
		StartTime:     startTime,
		FinishTime:    finishTime,
		CriticalCount: reportBriefInfo.CriticalCount,
		FailCount:     reportBriefInfo.FailedCount,
		WarningCount:  reportBriefInfo.WarningCount,
		PassCount:     reportBriefInfo.PassCount,
		Status:        constant.INSPECTION_STATUS_SUCCEED,
	}
	if err := inspectionService.UpdateInspectionReportComplete(updateReport); err != nil {
		return errors.Wrap(err, "failed to update inspection report")
	}

	// Success case - err is nil, defer won't update status
	return nil
}

// setupInspectionReportStatusUpdate sets up status update for inspection report
// It updates status to RUNNING at the start and returns a defer function to handle errors
func setupInspectionReportStatusUpdate(t task.ExecutableTask) func(*error) {
	// Update inspection report status to RUNNING as the first action
	if err := updateInspectionReportStatus(t, constant.INSPECTION_STATUS_RUNNING, ""); err != nil {
		t.ExecuteWarnLogf("failed to update inspection report status: %v", err)
	}

	// Return defer function to handle errors
	return func(err *error) {
		if *err != nil {
			if updateErr := updateInspectionReportStatus(t, constant.INSPECTION_STATUS_FAILED, (*err).Error()); updateErr != nil {
				t.ExecuteWarnLogf("failed to update inspection report status on error: %v", updateErr)
			}
		}
	}
}

// updateInspectionReportStatus updates the inspection report status by task ID
func updateInspectionReportStatus(t task.ExecutableTask, status string, errorMessage string) error {
	dag, err := localTaskService.GetDagBySubTaskId(t.GetID())
	if err != nil {
		return errors.Wrap(err, "failed to get dag by task id")
	}

	// Get Generic ID for the dag
	dagGenericID := task.ConvertIDToGenericID(dag.GetID(), dag.IsLocalTask(), dag.GetDagType())

	report, err := inspectionService.GetInspectionReportByLocalTaskId(dagGenericID)
	if err != nil {
		// Report not found, create a new one
		// Get scenario from task context
		var scenario string
		if err := t.GetContext().GetParamWithValue(PARAM_SCENARIO, &scenario); err != nil {
			// If scenario not found, try to get from context directly
			if scenarioVal := t.GetContext().GetParam(PARAM_SCENARIO); scenarioVal != nil {
				if s, ok := scenarioVal.(string); ok {
					scenario = s
				}
			}
			if scenario == "" {
				return errors.Wrap(err, "failed to get scenario from task context")
			}
		}
		scenario = strings.ToUpper(scenario)

		now := time.Now()
		newReport := &oceanbase.InspectionReport{
			Scenario:     scenario,
			LocalTaskId:  dagGenericID,
			Status:       status,
			StartTime:    now,
			FinishTime:   constant.ZERO_TIME,
			ErrorMessage: errorMessage,
			Report:       "{}",
		}
		// Only set FinishTime for completed statuses (FAILED or SUCCEED)
		// For RUNNING status, FinishTime should remain zero
		if status == constant.INSPECTION_STATUS_FAILED || status == constant.INSPECTION_STATUS_SUCCEED {
			newReport.FinishTime = now
		}
		err = inspectionService.SaveInspectionReport(newReport)
		if err != nil {
			return errors.Wrap(err, "failed to create inspection report")
		}
		// Get the newly created report
		_, err = inspectionService.GetInspectionReportByLocalTaskId(dagGenericID)
		if err != nil {
			return errors.Wrap(err, "failed to get newly created inspection report")
		}
		return nil
	}

	if errorMessage != "" {
		return inspectionService.UpdateInspectionReportStatusAndError(report.Id, status, errorMessage)
	}
	return inspectionService.UpdateInspectionReportStatus(report.Id, status)
}

// getTaskStatusForInspectionReport gets the task status for an inspection report
// localTaskId is the Generic ID of the task
func getTaskStatusForInspectionReport(localTaskId string) (string, error) {
	// Convert Generic ID to numeric ID
	dagID, agentInfo, err := task.ConvertGenericID(localTaskId)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert generic ID to dag ID")
	}
	if agentInfo != nil && !meta.OCS_AGENT.Equal(agentInfo) {
		return constant.INSPECTION_STATUS_RUNNING, nil
	}

	dag, err := localTaskService.GetDagInstance(dagID)
	if err != nil {
		return "", err
	}

	state := dag.GetState()
	switch state {
	case task.RUNNING:
		return constant.INSPECTION_STATUS_RUNNING, nil
	case task.FAILED:
		return constant.INSPECTION_STATUS_FAILED, nil
	case task.SUCCEED:
		return constant.INSPECTION_STATUS_SUCCEED, nil
	default:
		return constant.INSPECTION_STATUS_UNKNOWN, nil
	}
}
