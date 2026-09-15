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

package trustedkeys

import "embed"

// Files contains only the OceanBase release public keys distributed by the
// official OceanBase RPM repository. Keeping the keys in the OBShell binary
// makes verification independent of the host RPM database and rpm command.
//
//go:embed RPM-GPG-KEY-OceanBase RPM-GPG-KEY-OceanBase-el7 RPM-GPG-KEY-OceanBase-old
var Files embed.FS

const (
	currentKeyName = "RPM-GPG-KEY-OceanBase"
	el7KeyName     = "RPM-GPG-KEY-OceanBase-el7"
	legacyKeyName  = "RPM-GPG-KEY-OceanBase-old"
)

// Names returns a new slice so callers cannot mutate the process-wide trust
// configuration.
func Names() []string {
	return []string{
		currentKeyName,
		el7KeyName,
		legacyKeyName,
	}
}
