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

package meta

import (
	"testing"

	"github.com/oceanbase/obshell/ob/agent/constant"
)

func TestStandaloneModeFromEnv(t *testing.T) {
	t.Setenv(constant.ENV_OBSHELL_STANDALONE_MODE, "true")
	standalone, err := StandaloneModeFromEnv()
	if err != nil {
		t.Fatalf("StandaloneModeFromEnv returned error: %v", err)
	}
	if !standalone {
		t.Fatal("StandaloneModeFromEnv returned false, expected true")
	}
}

func TestStandaloneModeFromEnvDefaultsToFalse(t *testing.T) {
	t.Setenv(constant.ENV_OBSHELL_STANDALONE_MODE, "")
	standalone, err := StandaloneModeFromEnv()
	if err != nil {
		t.Fatalf("StandaloneModeFromEnv returned error: %v", err)
	}
	if standalone {
		t.Fatal("StandaloneModeFromEnv returned true without an explicit marker")
	}
}

func TestStandaloneModeFromEnvRejectsInvalidValue(t *testing.T) {
	t.Setenv(constant.ENV_OBSHELL_STANDALONE_MODE, "not-a-boolean")
	if _, err := StandaloneModeFromEnv(); err == nil {
		t.Fatal("StandaloneModeFromEnv accepted an invalid value")
	}
}

func TestIsStandaloneLoopback(t *testing.T) {
	tests := []struct {
		name       string
		standalone bool
		ip         string
		expected   bool
	}{
		{name: "standalone loopback", standalone: true, ip: "127.0.0.1", expected: true},
		{name: "standalone physical ip", standalone: true, ip: "10.0.0.8", expected: false},
		{name: "community loopback", standalone: false, ip: "127.0.0.1", expected: false},
		{name: "community physical ip", standalone: false, ip: "10.0.0.8", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := IsStandaloneLoopback(test.standalone, test.ip); actual != test.expected {
				t.Fatalf("IsStandaloneLoopback(%v, %q) = %v, expected %v", test.standalone, test.ip, actual, test.expected)
			}
		})
	}
}
