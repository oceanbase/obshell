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

package oceanbase

import (
	"encoding/json"
	"errors"
)

// Target represents a target host with IP and port
type Target struct {
	IP   string `json:"ip"`   // IP address
	Port int    `json:"port"` // Port number, default is 22 if not provided
}

// CredentialSecretData represents the structure of secret stored in database
// This is an internal model, not returned to users
type CredentialSecretData struct {
	Username            string   `json:"username"`
	Targets             []Target `json:"targets"`  // Target host list, each Target contains ip and port
	SshType             string   `json:"ssh_type"` // Authentication type, e.g., "PASSWORD"
	LocalPrivateKeyPath string   `json:"local_private_key_path,omitempty"`
	Passphrase          string   `json:"passphrase"` // encrypted passphrase
}

// ParseFrom parses secret json string into the receiver with basic validation.
// It ensures username, targets, and passphrase are not empty.
func (c *CredentialSecretData) ParseFrom(secret string) error {
	if err := json.Unmarshal([]byte(secret), c); err != nil {
		return err
	}
	if c.Username == "" || c.Passphrase == "" {
		return errors.New("invalid credential secret: missing username or passphrase")
	}
	if len(c.Targets) == 0 {
		return errors.New("invalid credential secret: targets list is empty")
	}
	return nil
}
