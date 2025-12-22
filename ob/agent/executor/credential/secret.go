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

package credential

import (
	"encoding/json"
	"fmt"

	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/model/oceanbase"
)

// SerializeSecret serializes credential information to JSON string format
func SerializeSecret(username string, targets []oceanbase.Target, SshType string, encryptedPassphrase string) string {
	secretData := oceanbase.CredentialSecretData{
		Username:   username,
		Targets:    targets,
		SshType:    SshType,
		Passphrase: encryptedPassphrase,
	}
	jsonBytes, err := json.Marshal(secretData)
	if err != nil {
		// Fallback to simple format if JSON marshal fails (should not happen)
		targetsJSON, _ := json.Marshal(targets)
		return fmt.Sprintf(`{"username":"%s","targets":%s,"ssh_type":"%s","passphrase":"%s"}`, username, string(targetsJSON), SshType, encryptedPassphrase)
	}
	return string(jsonBytes)
}

// DeserializeSecret deserializes JSON secret string to CredentialSecretData
// Returns: *CredentialSecretData, error
func DeserializeSecret(secret string) (*oceanbase.CredentialSecretData, error) {
	var secretData oceanbase.CredentialSecretData
	if err := json.Unmarshal([]byte(secret), &secretData); err != nil {
		return nil, errors.WrapRetain(errors.ErrCredentialSecretFormatInvalid, err)
	}

	// Validate required fields
	if secretData.Username == "" || len(secretData.Targets) == 0 || secretData.Passphrase == "" {
		return nil, errors.Occur(errors.ErrCredentialSecretFormatInvalid, "invalid secret format: missing required fields")
	}

	// Set default type if not present (backward compatibility)
	if secretData.SshType == "" {
		secretData.SshType = AUTH_TYPE_PASSWORD
	}

	return &secretData, nil
}
