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

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/process"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/param"
)

func HandleObStop(param param.ObStopParam) (*task.DagDetailDTO, error) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}
	if err := CheckStopObParam(&param); err != nil {
		return nil, err
	}

	template := buildStopTemplate(param.Force, param.Terminate)
	taskCtx, err := buildStopTaskContext(param)
	if err != nil {
		return nil, err
	}
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, taskCtx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func buildStopTaskContext(param param.ObStopParam) (*task.TaskContext, error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}
	needStopAgents, err := GenerateTargetAgentList(param.Scope)
	if err != nil {
		return nil, err
	}
	log.Infof("need stop agents are %v", needStopAgents)
	ctx := task.NewTaskContext().
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_ALL_AGENTS, agents).
		SetParam(PARAM_SCOPE, param.Scope).
		SetParam(PARAM_FORCE_PASS_DAG, param.ForcePassDagParam).
		SetParam(PARAM_URI, constant.URI_OB_RPC_PREFIX+constant.URI_STOP).
		SetParam(PARAM_EXPECT_MAIN_NEXT_STAGE, SUB_STOP_DAG_EXPECT_MAIN_NEXT_STAGE).
		SetParam(PARAM_MAIN_DAG_NAME, DAG_STOP_OB).
		SetParam(PARAM_STOP_OBSERVER_PROCESS, param.Force || param.Terminate)
	if param.Force || param.Terminate {
		for _, agent := range needStopAgents {
			ctx.SetAgentData(&agent, DATA_SUB_DAG_NEED_EXEC_CMD, true)
		}
	}
	return ctx, nil
}

func buildStopTemplate(force bool, terminate bool) *task.Template {
	task := task.NewTemplateBuilder(DAG_STOP_OB).
		SetMaintenance(task.UnMaintenance()).
		AddTask(newCreateSubStopDagTask(), true).
		AddTask(newCheckSubStopDagReadyTask(), false)
	if terminate {
		task.AddTask(newMinorFreezeTask(), false)
	}
	task.AddTask(newRetrySubStopDagTask(), false).
		AddTask(newWaitSubStopDagFinishTask(), true)
	if !force && !terminate {
		task.AddTask(newExecStopSqlTask(), false)
	}
	task.AddTask(newPassSubStopDagTask(), false)
	return task.Build()
}

type CreateSubStopDagTask struct {
	CreateSubDagTask
}

type CheckSubStopDagReadyTask struct {
	CheckSubDagReadyTask
}

type RetrySubStopDagTask struct {
	RetrySubDagTask
}

type WaitSubStopDagFinishTask struct {
	WaitSubDagFinishTask
}

type ExecStopSqlTask struct {
	PassSubDagTask
	scope param.Scope
}

type PassSubStopDagTask struct {
	PassSubDagTask
}

func newCreateSubStopDagTask() *CreateSubStopDagTask {
	return &CreateSubStopDagTask{
		CreateSubDagTask: *NewCreateSubDagTask("Inform all agents to prepare to stop observer"),
	}
}

func newCheckSubStopDagReadyTask() *CheckSubStopDagReadyTask {
	return &CheckSubStopDagReadyTask{
		*NewCheckSubDagReadyTask(),
	}
}

func newRetrySubStopDagTask() *RetrySubStopDagTask {
	return &RetrySubStopDagTask{
		*NewRetrySubDagTask(),
	}
}

func newWaitSubStopDagFinishTask() *WaitSubStopDagFinishTask {
	return &WaitSubStopDagFinishTask{
		*NewWaitSubDagFinishTask(),
	}
}

func newExecStopSqlTask() *ExecStopSqlTask {
	newTask := &ExecStopSqlTask{
		PassSubDagTask: *NewPassSubDagTask("Execute stop sql"),
	}
	newTask.SetCanContinue().SetCanCancel()
	return newTask
}

func newPassSubStopDagTask() *PassSubStopDagTask {
	newTask := &PassSubStopDagTask{
		*NewPassSubDagTask("Inform all agents to end the task"),
	}
	newTask.SetCanCancel()
	return newTask
}

const (
	SUB_STOP_DAG_EXPECT_MAIN_NEXT_STAGE      = 3
	MAIN_STOP_DAG_EXPECTED_SUB_NEXT_STAGE    = 2
	SUB_STOP_ZONE_DAG_EXPECT_MAIN_NEXT_STAGE = 3
)

func (t *CreateSubStopDagTask) Execute() error {
	return t.execute()
}

func (t *CheckSubStopDagReadyTask) Execute() (err error) {
	return t.execute()
}

func (t *RetrySubStopDagTask) Execute() (err error) {
	return t.execute()
}

func (t *WaitSubStopDagFinishTask) Execute() (err error) {
	return t.WaitSubDagFinishTask.Execute()
}

func (t *ExecStopSqlTask) Execute() (err error) {
	ctx := t.GetContext()
	if err := ctx.GetDataWithValue(DATA_ALL_AGENT_DAG_MAP, &t.allAgentDagMap); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			t.pass()
		}
	}()

	if exist, err := process.CheckObserverProcess(); err != nil || !exist {
		return err
	}

	if err = ctx.GetParamWithValue(PARAM_SCOPE, &t.scope); err != nil {
		return err
	}

	if err := getOceanbaseInstance(); err != nil {
		return err
	}

	switch t.scope.Type {
	case SCOPE_ZONE:
		return t.stopZone()
	case SCOPE_SERVER:
		return t.stopServer()
	}
	return errors.Occur(errors.ErrObClusterScopeInvalid, t.scope.Type)
}

func (t *ExecStopSqlTask) stopZone() (err error) {
	t.ExecuteLog("Stop Zone")
	for _, zone := range t.scope.Target {
		if err = obclusterService.StopZone(zone); err != nil {
			return err
		}

		active, err := obclusterService.IsZoneActive(zone)
		if !active {
			t.ExecuteLogf("%s stopped", zone)
			continue
		}
		if err != nil {
			t.ExecuteErrorLog(err)
		}
	}
	return nil
}

func (t *ExecStopSqlTask) stopServer() (err error) {
	t.ExecuteLog("Stop observer")
	agents, err := agentService.GetAllAgentsDOFromOB()
	if err != nil {
		return err
	}
	for _, server := range t.scope.Target {
		t.ExecuteLogf("Stop %s", server)
		agentInfo, err := meta.ConvertAddressToAgentInfo(server)
		if err != nil {
			return errors.Wrapf(err, "convert server '%s' to agent info failed", server)
		}
		for _, agent := range agents {
			if agentInfo.Ip == agent.Ip && agentInfo.Port == agent.Port {
				serverInfo := meta.NewAgentInfo(agent.Ip, agent.RpcPort)
				sql := fmt.Sprintf("alter system stop server '%s'", serverInfo.String())
				log.Info(sql)
				if err = obclusterService.ExecuteSql(sql); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func (t *PassSubStopDagTask) Execute() (err error) {
	return t.execute()
}

// validateStopZoneMajorityCondition checks if stopping the zone will satisfy majority condition for all log streams
func validateStopZoneMajorityCondition(zoneName string) error {
	// Get all user tenants for error message
	tenants, err := tenantService.GetAllUserTenants()
	if err != nil {
		return errors.Wrap(err, "get all user tenants failed")
	}

	// Find which tenants are affected
	var invalidTenants []string
	tenantChecked := make(map[int]bool)

	// Get all log streams in the zone
	logStreams, err := obclusterService.GetLogInfosInZone(zoneName)
	if err != nil {
		return errors.Wrap(err, "get log streams in zone failed")
	}

	// Check each log stream to find invalid tenants
	for _, ls := range logStreams {
		alive, err := obclusterService.IsLsMultiPaxosAliveAfterStopZone(ls.LsId, ls.TenantId, zoneName)
		if err != nil {
			log.Warnf("validateStopZoneMajorityCondition: failed to check tenant %d LS %d: %v", ls.TenantId, ls.LsId, err)
			continue
		}
		if !alive {
			// Find tenant name
			for _, tenant := range tenants {
				if tenant.TenantID == ls.TenantId && !tenantChecked[tenant.TenantID] {
					invalidTenants = append(invalidTenants, tenant.TenantName)
					tenantChecked[tenant.TenantID] = true

					break
				}
			}
		}
	}

	if len(invalidTenants) > 0 {
		return errors.Occur(errors.ErrObClusterTenantReplicaInvalid, zoneName, invalidTenants)
	}

	return nil
}

func CheckStopObParam(param *param.ObStopParam) error {
	if param.Scope.Type == SCOPE_GLOBAL && (!param.Force && !param.Terminate) {
		return errors.Occur(errors.ErrObClusterForceStopOrTerminateRequired)
	}
	if param.Force && param.Terminate {
		return errors.Occur(errors.ErrObClusterStopModeConflict)
	}
	if (!param.Force || param.Scope.Type == SCOPE_ZONE) && oceanbase.GetState() != oceanbase.STATE_CONNECTION_AVAILABLE {
		return errors.Occur(errors.ErrObClusterForceStopRequired)
	}

	// If need to execute stop sql, check if has other stop task

	if param.Scope.Type == SCOPE_SERVER {
		if !param.Terminate && !param.Force {
			// Check all servers is in one zone
			servers, err := agentService.GetAllAgentsDOFromOB()
			if err != nil {
				return errors.Wrap(err, "get all agents do from ob failed")
			}
			var targetToZoneMap = make(map[string]string)
			for _, server := range servers {
				targetToZoneMap[fmt.Sprintf("%s:%d", server.Ip, server.Port)] = server.Zone
			}
			var zone string
			for _, target := range param.Scope.Target {
				if zone == "" {
					zone = targetToZoneMap[target]
				} else if targetToZoneMap[target] != zone {
					return errors.Occur(errors.ErrObServerStoppedInMultiZone)
				}
			}
			// Check whether has other stopped server in the same zone.
			if exist, err := obclusterService.HasOtherStopTask(zone); err != nil {
				return errors.Wrap(err, "check if has other stop task failed")
			} else if exist {
				return errors.Occur(errors.ErrObServerStoppedInMultiZone)
			}
		}
	}

	if param.Scope.Type == SCOPE_ZONE {
		if !param.Terminate && !param.Force {
			if len(param.Scope.Target) > 1 {
				return errors.Occur(errors.ErrObServerStoppedInMultiZone)
			} else if len(param.Scope.Target) != 0 {
				// Check whether has other stopped zone or server in other zone.
				zone := param.Scope.Target[0]
				if exist, err := obclusterService.HasOtherStopTask(zone); err != nil {
					return errors.Wrap(err, "check if has other stop task failed")
				} else if exist {
					return errors.Occur(errors.ErrObServerStoppedInMultiZone)
				}

				// Validate majority condition before stopping zone
				if err := validateStopZoneMajorityCondition(zone); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
