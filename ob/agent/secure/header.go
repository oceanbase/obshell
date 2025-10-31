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
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/crypto"
	"github.com/oceanbase/obshell/ob/agent/lib/json"
	"github.com/oceanbase/obshell/ob/agent/meta"
)

const (
	NotForward = iota
	AutoForward
	ManualForward
)

type HttpHeader struct {
	Auth         string
	Ts           string
	Token        string
	Uri          string
	Keys         []byte
	Sha256       string
	ForwardType  int
	ForwardAgent meta.AgentInfo
}

func BuildAgentHeader(agentInfo meta.AgentInfoInterface, password string, uri string, isForword bool, keys ...[]byte) map[string]string {
	auth := buildHeader(agentInfo, password, uri, isForword, keys...)
	header := map[string]string{
		constant.OCS_AGENT_HEADER: auth,
	}
	return header
}

func BuildHeader(agentInfo meta.AgentInfoInterface, uri string, isForword bool, keys ...[]byte) map[string]string {
	auth := buildHeader(agentInfo, meta.OCEANBASE_PWD, uri, isForword, keys...)
	header := map[string]string{
		constant.OCS_HEADER: auth,
	}
	return header
}

func buildHeader(agentInfo meta.AgentInfoInterface, password string, uri string, isForword bool, keys ...[]byte) string {
	pk := GetAgentPublicKey(agentInfo)
	if pk == "" {
		log.Warnf("no key for agent '%s'", agentInfo.String())
		return ""
	}

	var token string
	if isForword && !meta.OCS_AGENT.IsMasterAgent() {
		token, _ = getTokenByAgentInfo(meta.OCS_AGENT)
	} else {
		token, _ = getTokenByAgentInfo(agentInfo)
	}

	var aesKeys []byte
	if len(keys) != 2 {
		aesKeys = nil
	} else {
		aesKeys = append(keys[0], keys[1]...)
	}
	header := HttpHeader{
		Auth:  password,
		Ts:    fmt.Sprintf("%d", time.Now().Add(getAuthExpiredDuration()).Unix()),
		Token: token,
		Uri:   uri,
		Keys:  aesKeys,
	}

	if isForword {
		header.ForwardType = ManualForward
		header.ForwardAgent = meta.OCS_AGENT.GetAgentInfo()
	}

	mAuth, err := json.Marshal(header)
	if err != nil {
		log.WithError(err).Error("json marshal failed")
		return ""
	}
	auth, err := crypto.RSAEncrypt(mAuth, pk)
	if err != nil {
		log.WithError(err).Error("rsa encrypt failed")
		return ""
	}
	return auth
}

func DecryptHeader(ciphertext string) (HttpHeader, error) {
	var headers HttpHeader
	decHeader, err := Crypter.DecryptAndReturnBytes(ciphertext)
	if err == nil {
		err = json.Unmarshal(decHeader, &headers)
	}
	return headers, err
}

func RepackageHeaderForAutoForward(header *HttpHeader, agentInfo meta.AgentInfoInterface) (headers map[string]string, err error) {
	err = errors.Occur(errors.ErrSecurityAuthenticationUnauthorized)

	header.ForwardType = AutoForward
	header.ForwardAgent = meta.OCS_AGENT.GetAgentInfo()
	// encrypt for master
	pk := GetAgentPublicKey(agentInfo)
	if pk == "" {
		log.Warnf("no key for agent: %s", agentInfo.String())
		return
	}

	mHeader, err := json.Marshal(header)
	if err != nil {
		log.WithError(err).Error("json marshal failed")
		return
	}

	encryptedHeader, err := crypto.RSAEncrypt(mHeader, pk)
	if err != nil {
		log.WithError(err).Error("rsa encrypt failed")
		return
	}

	headers = map[string]string{
		constant.OCS_HEADER: string(encryptedHeader),
	}
	return headers, nil
}
