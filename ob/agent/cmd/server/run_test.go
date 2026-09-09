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

package server

import (
	"errors"
	"testing"
	"time"

	"github.com/oceanbase/obshell/ob/agent/meta"
)

func TestSyncAgentMetadataRetriesVersionBeyondPublicKeyWindow(t *testing.T) {
	const unavailableAttempts = publicKeySyncMaxAttempts + 1
	versionAttempts := 0

	syncAgentMetadataToOBWhenReady(
		true,
		func() error { return nil },
		func() error {
			versionAttempts++
			if versionAttempts <= unavailableAttempts {
				return errors.New("database is not ready")
			}
			return nil
		},
		func(time.Duration) {},
	)

	if versionAttempts != unavailableAttempts+1 {
		t.Fatalf("version attempts = %d, expected %d", versionAttempts, unavailableAttempts+1)
	}
}

func TestShouldSyncAgentVersionOnStartup(t *testing.T) {
	testCases := []struct {
		name        string
		upgradeMode bool
		identity    meta.AgentIdentity
		expected    bool
	}{
		{name: "external upgrade cluster agent", upgradeMode: true, identity: meta.CLUSTER_AGENT, expected: true},
		{name: "normal cluster agent", upgradeMode: false, identity: meta.CLUSTER_AGENT, expected: false},
		{name: "unidentified agent", upgradeMode: true, identity: meta.UNIDENTIFIED, expected: false},
		{name: "takeover master", upgradeMode: true, identity: meta.TAKE_OVER_MASTER, expected: false},
		{name: "scaling out agent", upgradeMode: true, identity: meta.SCALING_OUT, expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := shouldSyncAgentVersionOnStartup(testCase.upgradeMode, testCase.identity)
			if actual != testCase.expected {
				t.Fatalf("shouldSyncAgentVersionOnStartup(%t, %q) = %t, expected %t", testCase.upgradeMode, testCase.identity, actual, testCase.expected)
			}
		})
	}
}
