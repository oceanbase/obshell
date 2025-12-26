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
	"fmt"
	"strconv"
	"strings"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/executor/inspection/constant"
	"github.com/oceanbase/obshell/ob/agent/lib/sshutil"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/secure"
	modelob "github.com/oceanbase/obshell/ob/model/oceanbase"
)

func extractNodeIndex(key string) int {
	// key pattern obcluster.servers.nodes[%d].ip or similar
	start := strings.Index(key, "[")
	end := strings.Index(key, "]")
	if start == -1 || end == -1 || end <= start+1 {
		return 0
	}
	idxStr := key[start+1 : end]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return 0
	}
	return idx
}

// findTargetForIP finds the target in Targets that matches the given IP and returns the IP and port
// Returns: matchedIP, port, error
func findTargetForIP(targets []modelob.Target, targetIP string) (string, int, error) {
	const DEFAULT_SSH_PORT = 22
	for _, target := range targets {
		targetIPFromTarget := target.IP
		port := target.Port
		if port == 0 {
			// Port not provided, use default port
			port = DEFAULT_SSH_PORT
		}

		if targetIPFromTarget == targetIP {
			return targetIPFromTarget, port, nil
		}
	}
	return "", 0, errors.Occur(errors.ErrObClusterInspectionHostCredentialNotFound, targetIP)
}

// fillSSHCredentialConfig fetches host credentials and injects into config before executing obdiag.
func (t *InspectionTask) fillSSHCredentialConfig() error {
	// collect node ips
	type nodeInfo struct {
		index int
		ip    string
	}
	nodes := make([]nodeInfo, 0)
	for k, v := range t.configs {
		if strings.HasPrefix(k, "obcluster.servers.nodes[") && strings.HasSuffix(k, "].ip") {
			nodes = append(nodes, nodeInfo{index: extractNodeIndex(k), ip: v})
		}
	}

	// usePasswordlessSSH is false means at least one node needs credential
	// checkSSHAccess() has already verified that each node either has passwordless SSH or credential
	// So we only fill credentials for nodes that have credentials (passwordless nodes are skipped)
	for _, node := range nodes {
		if node.ip == meta.OCS_AGENT.GetIp() {
			continue
		}

		cred, err := credentialService.GetByHost(node.ip)
		if err != nil {
			return errors.Wrap(err, "get host credential failed")
		}
		if cred == nil {
			// Node has passwordless SSH, skip credential filling
			continue
		}

		var secret modelob.CredentialSecretData
		if err := secret.ParseFrom(cred.Secret); err != nil {
			return errors.WrapRetain(errors.ErrCredentialSecretFormatInvalid, err)
		}
		passphrase, err := secure.DecryptCredentialPassphrase(secret.Passphrase)
		if err != nil {
			return errors.WrapRetain(errors.ErrCredentialDecryptFailed, err)
		}
		// Find matching target in Targets array
		matchedIP, port, err := findTargetForIP(secret.Targets, node.ip)
		if err != nil {
			return errors.WrapRetain(errors.ErrObClusterInspectionHostCredentialNotFound, err)
		}
		if matchedIP != node.ip {
			// credential stored ip mismatch
			return errors.Occur(errors.ErrObClusterInspectionHostCredentialNotFound, node.ip)
		}
		if err := sshutil.ValidateSSHConnection(node.ip, port, secret.Username, passphrase); err != nil {
			return errors.WrapRetain(errors.ErrCredentialSSHValidationFailed, err)
		}

		t.configs[fmt.Sprintf(constant.CONFIG_OBCLUSTER_SERVERS_NODES_SSH_USERNAME, node.index)] = secret.Username
		t.configs[fmt.Sprintf(constant.CONFIG_OBCLUSTER_SERVERS_NODES_SSH_PASSWORD, node.index)] = passphrase
	}
	return nil
}
