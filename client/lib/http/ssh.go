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

package http

import (
	"io"
	"os"
	osuser "os/user"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/utils"
)

const DEFALUT_SSH_PORT = 22
const DEFALUT_SSH_PATH = ".ssh"

// excludeFile is a list of files that not related to private keys
var excludeFile = []string{"authorized_keys", "config", "id_rsa.pub", "known_hosts"}

type SSHClient struct {
	Host           string
	User           string
	Port           int
	Password       string
	PrivateKeyFile string
	Passphrase     string
	client         *ssh.Client
}

func (s *SSHClient) SetPassword(password string) {
	s.Password = password
}

func (s *SSHClient) SetPrivateKeyFile(keyPath string, passphrase string) {
	s.PrivateKeyFile = keyPath
	s.Passphrase = passphrase
}

func (s *SSHClient) Exec(cmd string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "failed to create ssh session")
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute command: %s", cmd)
	}
	return string(output), nil
}

func (s *SSHClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *SSHClient) Connect() (*ssh.Client, error) {
	var err error
	if s.Password != "" {
		s.client, err = newSSHClientByPwd(s, s.Password)
	} else {
		s.client, err = newSSHClientByPK(s, s.PrivateKeyFile, s.Passphrase)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ssh client for %s", s.Host)
	}
	return s.client, nil
}

func NewSSHClient(host, user, sshPort string) (*SSHClient, error) {
	client := &SSHClient{
		Host: host,
		User: user,
		Port: DEFALUT_SSH_PORT,
	}
	if user == "" {
		userName, err := osuser.Current()
		if err != nil {
			return nil, err
		}
		client.User = userName.Username
	}
	if sshPort != "" {
		port, err := strconv.Atoi(sshPort)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert ssh port")
		}
		client.Port = port
	}
	return client, nil
}

func loadDefaultPrivateKeys() ([]ssh.Signer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	defaultDir := filepath.Join(home, DEFALUT_SSH_PATH)

	var signers []ssh.Signer
	files, err := os.ReadDir(defaultDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if utils.ContainsString(excludeFile, file.Name()) {
			continue
		}
		keyPath := filepath.Join(defaultDir, file.Name())
		// If an error occurs while loading the private key, ignore it, because it may have no permission
		signer, _ := loadPrivateKey(keyPath)
		if signer != nil {
			signers = append(signers, signer)
		}
	}
	return signers, nil
}

func loadPrivateKey(keyPath string) (ssh.Signer, error) {
	file, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keyData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadPrivateKeyWithPassphrase(keyPath string, passphrase string) (ssh.Signer, error) {
	file, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keyData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(passphrase))
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// ssh client by private key
// Only when a keypath is specified does the passphrase make sense.
// when user does not specify keyPath, while a passphrase has been configured, the client should notify the user with an appropriate message
func newSSHClientByPK(config *SSHClient, keyPath string, passphrase string) (*ssh.Client, error) {
	var signers []ssh.Signer
	var err error
	if keyPath == "" {
		signers, err = loadDefaultPrivateKeys()
		if err != nil {
			return nil, errors.Wrap(err, "failed to load default private key")
		}
	} else if passphrase != "" {
		signer, err := loadPrivateKeyWithPassphrase(keyPath, passphrase)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load private key with passphrase")
		}
		signers = append(signers, signer)
	} else {
		signer, err := loadPrivateKey(keyPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load private key")
		}
		signers = append(signers, signer)
	}
	return newClient(config, ssh.PublicKeys(signers...))
}

// ssh client by password
func newSSHClientByPwd(config *SSHClient, password string) (client *ssh.Client, err error) {
	return newClient(config, ssh.Password(password))
}

func newClient(config *SSHClient, auth ...ssh.AuthMethod) (*ssh.Client, error) {
	conf := &ssh.ClientConfig{
		User:            config.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	server := meta.NewAgentInfo(config.Host, config.Port)
	return ssh.Dial("tcp", server.String(), conf)
}
