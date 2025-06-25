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

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
)

func Rebuild(agentInstance *meta.AgentInstance) error {
	// Get the entry, check if the agent port is the same.
	if agentInstance.GetPort() != meta.OCS_AGENT.GetPort() {
		log.Errorf("agent port is not the same, agent port in all_agents: %d, agent port now: %d", agentInstance.GetPort(), meta.OCS_AGENT.GetPort())
		return errors.Occur(errors.ErrAgentRebuildPortNotSame, agentInstance.GetPort(), meta.OCS_AGENT.GetPort())
	}

	// Check version consistent.
	if agentInstance.GetVersion() != meta.OCS_AGENT.GetVersion() {
		log.Errorf("agent version is not the same, agent version in all_agents: %s, agent version now: %s", agentInstance.GetVersion(), meta.OCS_AGENT.GetVersion())
		return errors.Occur(errors.ErrAgentRebuildVersionNotSame, agentInstance.GetVersion(), meta.OCS_AGENT.GetVersion())
	}

	if err := agentService.UpdateAgentPublicKey(secure.Public()); err != nil {
		return err
	}

	// Rebuild sqlite.
	return createRebuildDag(agentInstance)
}

func createRebuildDag(agent *meta.AgentInstance) error {
	isRunning, err := localTaskService.IsRunning()
	if err != nil {
		return err
	} else if !isRunning {
		log.Infof("The agent is already under maintenance.")
		return nil
	}
	log.Infof("create rebuild dag, identity: %s", agent.GetIdentity())
	templateName := fmt.Sprintf("Rebuild %s", agent.GetIdentity())
	templateBuilder := task.NewTemplateBuilder(templateName).
		SetMaintenance(task.GlobalMaintenance()).
		AddTask(newAgentSyncTask(), false)

	if agent.IsMasterAgent() || agent.IsTakeOverMasterAgent() {
		templateBuilder.AddTemplate(newConvertClusterTemplate())
	}
	ctx, err := newConvertClusterContext()
	if err != nil {
		return err
	}
	dag, err := localTaskService.CreateDagInstanceByTemplate(templateBuilder.Build(), ctx)
	if err != nil {
		return err
	}
	log.Infof("create rebuild dag '%s' success", task.NewDagDetailDTO(dag).GenericID)
	return nil
}
