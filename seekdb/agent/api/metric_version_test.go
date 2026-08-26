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

package api

import (
	"fmt"
	"testing"
	"time"
)

func TestMetricVersionResolverCachesSuccessfulVersion(t *testing.T) {
	now := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	liveCalls := 0
	storedCalls := 0
	resolver := newMetricVersionResolver(
		func() (string, error) {
			liveCalls++
			return "1.4.0.0", nil
		},
		func() (string, error) {
			storedCalls++
			return "1.3.0.0", nil
		},
	)
	resolver.now = func() time.Time { return now }

	for _, advance := range []time.Duration{0, 59 * time.Second} {
		now = now.Add(advance)
		version, err := resolver.resolve()
		if err != nil || version != "1.4.0.0" {
			t.Fatalf("resolve version = %q, %v", version, err)
		}
	}
	if liveCalls != 1 || storedCalls != 0 {
		t.Fatalf("unexpected calls during cache lifetime: live=%d stored=%d", liveCalls, storedCalls)
	}

	now = now.Add(time.Second)
	if _, err := resolver.resolve(); err != nil {
		t.Fatalf("refresh cached version: %v", err)
	}
	if liveCalls != 2 || storedCalls != 0 {
		t.Fatalf("cache was not refreshed after 60 seconds: live=%d stored=%d", liveCalls, storedCalls)
	}
}

func TestMetricVersionResolverFallsBackToSQLite(t *testing.T) {
	storedCalls := 0
	resolver := newMetricVersionResolver(
		func() (string, error) {
			return "", fmt.Errorf("database unavailable")
		},
		func() (string, error) {
			storedCalls++
			return "1.3.2.0", nil
		},
	)
	version, err := resolver.resolve()
	if err != nil || version != "1.3.2.0" || storedCalls != 1 {
		t.Fatalf("fallback version = %q, err=%v, storedCalls=%d", version, err, storedCalls)
	}
}

func TestMetricVersionResolverRejectsInvalidVersions(t *testing.T) {
	resolver := newMetricVersionResolver(
		func() (string, error) {
			return "seekdb-v1.4", nil
		},
		func() (string, error) {
			return "", nil
		},
	)
	version, err := resolver.resolve()
	if err == nil || version != "" {
		t.Fatalf("invalid versions must not be cached or returned: version=%q err=%v", version, err)
	}
}
