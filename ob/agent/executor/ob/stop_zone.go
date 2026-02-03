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

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
	"github.com/oceanbase/obshell/ob/param"
	log "github.com/sirupsen/logrus"
)

func HandleObZoneStop(p param.ObZoneStopParam) (*task.DagDetailDTO, error) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		return nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT)
	}

	agent := getAgentWithAvailableOB()
	if meta.OCS_AGENT.Equal(agent) {
		if err := CheckZoneStopValidate(p.ZoneName); err != nil {
			return nil, err
		}
	} else { // forward to other obshell to do this check
		if err := secure.SendPostRequest(agent, fmt.Sprintf("%s%s/%s%s%s", constant.URI_OB_API_PREFIX, constant.URI_ZONE_GROUP, p.ZoneName, constant.URI_STOP, constant.URI_CHECK), &p, nil); err != nil {
			return nil, err
		}
	}

	template := buildStopZoneTemplate(&p)
	taskCtx, err := buildStopZoneTaskContext(&p)
	if err != nil {
		return nil, err
	}
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, taskCtx)
	if err != nil {
		return nil, err
	}
	return task.NewDagDetailDTO(dag), nil
}

func CheckZoneStopValidate(zoneName string) error {
	// check if the zone is exist
	if exist, err := obclusterService.IsZoneExistInOB(zoneName); err != nil {
		return errors.Wrap(err, "check if zone exist failed")
	} else if !exist {
		return errors.Occur(errors.ErrObZoneNotExist, zoneName)
	}

	if exist, err := obclusterService.HasOtherStopTask(zoneName); err != nil {
		return errors.Wrap(err, "check if has other stop task failed")
	} else if exist {
		return errors.Occur(errors.ErrObServerStoppedInMultiZone)
	}

	// Validate majority condition before stopping zone
	if err := validateStopZoneMajorityCondition(zoneName); err != nil {
		return err
	}
	return nil
}

func buildStopZoneTaskContext(p *param.ObZoneStopParam) (*task.TaskContext, error) {
	agents, err := agentService.GetAllAgentsInfo()
	if err != nil {
		return nil, err
	}
	var paramScope param.Scope = param.Scope{
		Type:   SCOPE_ZONE,
		Target: []string{p.ZoneName},
	}
	needStopAgents, err := GenerateTargetAgentList(paramScope)
	if err != nil {
		return nil, err
	}
	log.Infof("need stop agents are %v", needStopAgents)
	ctx := task.NewTaskContext().
		SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_ALL_AGENTS, agents).
		SetParam(PARAM_SCOPE, paramScope).
		SetParam(PARAM_FORCE_PASS_DAG, p.ForcePassDagParam).
		SetParam(PARAM_URI, constant.URI_OB_RPC_PREFIX+constant.URI_STOP).
		SetParam(PARAM_EXPECT_MAIN_NEXT_STAGE, SUB_STOP_DAG_EXPECT_MAIN_NEXT_STAGE).
		SetParam(PARAM_MAIN_DAG_NAME, DAG_STOP_OB).
		SetParam(PARAM_STOP_OBSERVER_PROCESS, true)
	for _, agent := range needStopAgents {
		ctx.SetAgentData(&agent, DATA_SUB_DAG_NEED_EXEC_CMD, true)
	}
	return ctx, nil
}

func buildStopZoneTemplate(p *param.ObZoneStopParam) *task.Template {
	task := task.NewTemplateBuilder(DAG_STOP_OB).
		SetMaintenance(task.UnMaintenance()).
		AddTask(newCreateSubStopDagTask(), true).
		AddTask(newCheckSubStopDagReadyTask(), false)
	task.AddTask(newExecStopSqlTask(), false)
	if p.FreezeServer {
		task.AddTask(newMinorFreezeTask(), false)
	}
	task.AddTask(newRetrySubStopDagTask(), false).
		AddTask(newWaitSubStopDagFinishTask(), true)
	task.AddTask(newPassSubStopDagTask(), false)
	return task.Build()
}
