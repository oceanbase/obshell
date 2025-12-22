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
package sshutil

import (
	"fmt"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"golang.org/x/crypto/ssh"
)

// ValidateSSHConnection validates SSH connection using provided credentials.
func ValidateSSHConnection(host string, port int, username, passphrase string) error {
	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(passphrase)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	server := meta.NewAgentInfo(host, port)
	client, err := ssh.Dial("tcp", server.String(), config)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("SSH connection failed to %s", server.String()))
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return errors.Wrap(err, "create SSH session failed")
	}
	session.Close()
	return nil
}
