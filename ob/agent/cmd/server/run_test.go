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

func TestSyncAgentMetadataBacksOffVersionRetries(t *testing.T) {
	const unavailableAttempts = 7
	versionAttempts := 0
	waits := make([]time.Duration, 0, unavailableAttempts)

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
		func(duration time.Duration) { waits = append(waits, duration) },
	)

	expected := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second, time.Minute, time.Minute}
	if len(waits) != len(expected) {
		t.Fatalf("wait count = %d, expected %d", len(waits), len(expected))
	}
	for index := range expected {
		if waits[index] != expected[index] {
			t.Fatalf("wait[%d] = %s, expected %s", index, waits[index], expected[index])
		}
	}
}

func TestSyncAgentMetadataPreservesPublicKeyRetryCadence(t *testing.T) {
	versionAttempts := 0
	waits := make([]time.Duration, 0, publicKeySyncMaxAttempts+1)

	syncAgentMetadataToOBWhenReady(
		true,
		func() error { return errors.New("public key is not ready") },
		func() error {
			versionAttempts++
			if versionAttempts <= publicKeySyncMaxAttempts {
				return errors.New("database is not ready")
			}
			return nil
		},
		func(duration time.Duration) { waits = append(waits, duration) },
	)

	if len(waits) != publicKeySyncMaxAttempts {
		t.Fatalf("wait count = %d, expected %d", len(waits), publicKeySyncMaxAttempts)
	}
	for index, duration := range waits {
		if duration != metadataSyncInterval {
			t.Fatalf("wait[%d] = %s, expected %s", index, duration, metadataSyncInterval)
		}
	}
}

func TestShouldSyncAgentVersionOnStartup(t *testing.T) {
	testCases := []struct {
		name     string
		identity meta.AgentIdentity
		expected bool
	}{
		{name: "cluster agent", identity: meta.CLUSTER_AGENT, expected: true},
		{name: "unidentified agent", identity: meta.UNIDENTIFIED, expected: false},
		{name: "takeover master", identity: meta.TAKE_OVER_MASTER, expected: false},
		{name: "scaling out agent", identity: meta.SCALING_OUT, expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := shouldSyncAgentVersionOnStartup(testCase.identity)
			if actual != testCase.expected {
				t.Fatalf("shouldSyncAgentVersionOnStartup(%q) = %t, expected %t", testCase.identity, actual, testCase.expected)
			}
		})
	}
}
