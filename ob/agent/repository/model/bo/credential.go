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

package bo

import "time"

// ConnectionResult represents the result of a connection attempt
type ConnectionResult string

// Connection result constants
const (
	ConnectionResultSuccess       ConnectionResult = "SUCCESS"
	ConnectionResultConnectFailed ConnectionResult = "CONNECT_FAILED"
)

// Target represents a target host with IP and port
type Target struct {
	IP   string `json:"ip"`   // IP address
	Port int    `json:"port"` // Port number
}

type SshSecret struct {
	Targets  []Target `json:"targets"` // Target host list, each Target contains ip and port
	Username string   `json:"username"`
	Type     string   `json:"type"`
}

type Credential struct {
	CredentialId int64     `json:"credential_id"`
	Name         string    `json:"name"`
	TargetType   string    `json:"target_type"`
	Description  string    `json:"description"`
	SshSecret    SshSecret `json:"ssh_secret"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// PaginatedCredentialResponse keeps list credentials response consistent with other paginated APIs.
type PaginatedCredentialResponse struct {
	Contents []Credential `json:"contents"`
	Page     CustomPage   `json:"page"`
}

type ValidationDetail struct {
	Target           Target           `json:"target"`            // Target information, contains ip and port
	ConnectionResult ConnectionResult `json:"connection_result"` // Connection result, e.g., ConnectionResultSuccess, ConnectionResultConnectFailed, etc. (for extensibility)
	Message          string           `json:"message"`           // Error message (empty string when validation succeeds)
}

type ValidationResult struct {
	CredentialId   int64              `json:"credential_id,omitempty"` // Credential ID, indicates which credential's validation result
	TargetType     string             `json:"target_type"`             // Target type, e.g., "HOST", "OB"
	SucceededCount int                `json:"succeeded_count"`         // Number of successfully validated targets for this credential
	FailedCount    int                `json:"failed_count"`            // Number of failed validated targets for this credential
	Details        []ValidationDetail `json:"details"`                 // Validation details, one detail per target
}
