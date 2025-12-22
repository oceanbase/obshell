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

package param

// Target represents a target host with IP and port
type Target struct {
	IP   string `json:"ip" binding:"required"` // IP address
	Port int    `json:"port"`                  // Port number, default is 22 if not provided
}

type SshCredentialProperty struct {
	Targets    []Target `json:"targets" binding:"required"`    // Target host list, each Target contains ip and port. If port is not provided, default port is 22. List cannot be empty, must contain at least one target.
	Username   string   `json:"username" binding:"required"`   // Username for SSH connection
	Type       string   `json:"type" binding:"required"`       // Authentication type, currently only supports "PASSWORD"
	Passphrase *string  `json:"passphrase" binding:"required"` // Password for SSH connection (plain text, will be encrypted before storage), can be empty.
}

type CreateCredentialParam struct {
	TargetType            string                `json:"target_type" binding:"required"` // Target type, currently only supports "HOST"
	Name                  string                `json:"name" binding:"required"`        // Credential name
	Description           string                `json:"description"`                    // Credential description
	SshCredentialProperty SshCredentialProperty `json:"ssh_credential_property" binding:"required"`
}

type UpdateCredentialParam struct {
	Name                  string                `json:"name" binding:"required"`
	Description           string                `json:"description"`
	SshCredentialProperty SshCredentialProperty `json:"ssh_credential_property" binding:"required"`
}

type ValidateCredentialParam struct {
	TargetType            string                `json:"target_type" binding:"required"` // Target type, currently only supports "HOST"
	SshCredentialProperty SshCredentialProperty `json:"ssh_credential_property" binding:"required"`
}

type BatchValidateCredentialParam struct {
	CredentialIdList []int `json:"credential_id_list" binding:"required"`
}

type BatchDeleteCredentialParam struct {
	CredentialIdList []int `json:"credential_id_list" binding:"required"`
}

type ListCredentialQueryParam struct {
	CredentialId int    `form:"credential_id"` // Credential ID
	TargetType   string `form:"target_type"`   // Target type, currently only supports "HOST"
	KeyWord      string `form:"key_word"`      // Keyword for searching (matches name or username)
	Page         int    `form:"page"`          // Page number, default is 1
	PageSize     int    `form:"page_size"`     // Page size, default is 10
	Sort         string `form:"sort"`          // Sort field
	SortOrder    string `form:"sort_order"`    // Sort order (asc/desc)
}

func (p *ListCredentialQueryParam) Format() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
}

// UpdateCredentialEncryptSecretKeyParam carries new AES key (Base64 string of raw key bytes)
type UpdateCredentialEncryptSecretKeyParam struct {
	AesKey string `json:"aes_key" binding:"required"`
}
