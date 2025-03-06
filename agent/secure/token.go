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

package secure

import (
	"errors"

	"github.com/google/uuid"

	"github.com/oceanbase/obshell/agent/meta"
)

// NewToken generates a token for the agent to join/scale-out an existing cluster
func NewToken(targetAgent meta.AgentInfoInterface) (string, error) {
	token, err := getTokenByAgentInfo(meta.OCS_AGENT)
	if err != nil {
		return "", err
	}
	if token == "" {
		token = uuid.New().String()
	}
	if err := updateToken(meta.OCS_AGENT, token); err != nil {
		return "", err
	}
	encryptedToken, err := EncryptToOther([]byte(token), targetAgent)
	if err != nil {
		return "", err
	}
	return encryptedToken, nil
}

func VerifyToken(token string) error {
	agentToken, err := getTokenByAgentInfo(meta.OCS_AGENT)
	if err != nil {
		return err
	}
	if agentToken != token || token == "" {
		return errors.New("wrong token")
	}
	return nil
}

func VerifyTokenByAgentInfo(token string, agentInfo meta.AgentInfo) error {
	agentToken, err := getTokenByAgentInfo(&agentInfo)
	if err != nil {
		return err
	}
	if agentToken != token {
		return errors.New("wrong token")
	}
	return nil
}
