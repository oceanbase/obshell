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

package observer

import (
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/agent/secure"
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
	return agentService.Rebuild(agentInstance)
}
