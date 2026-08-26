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
	"sync"
	"time"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	metricexecutor "github.com/oceanbase/obshell/seekdb/agent/executor/metric"
	agentservice "github.com/oceanbase/obshell/seekdb/agent/service/agent"
)

const metricVersionCacheTTL = 60 * time.Second

type metricVersionResolver struct {
	mu            sync.Mutex
	cachedVersion string
	validUntil    time.Time
	ttl           time.Duration
	now           func() time.Time
	liveVersion   func() (string, error)
	storedVersion func() (string, error)
}

func newMetricVersionResolver(liveVersion, storedVersion func() (string, error)) *metricVersionResolver {
	return &metricVersionResolver{
		ttl:           metricVersionCacheTTL,
		now:           time.Now,
		liveVersion:   liveVersion,
		storedVersion: storedVersion,
	}
}

func (r *metricVersionResolver) resolve() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	if r.cachedVersion != "" && now.Before(r.validUntil) {
		return r.cachedVersion, nil
	}

	liveVersion, liveErr := r.liveVersion()
	if liveErr == nil && metricexecutor.IsValidProductVersion(liveVersion) {
		r.cache(liveVersion, now)
		return liveVersion, nil
	}
	if liveErr == nil {
		liveErr = fmt.Errorf("invalid live seekdb product version %q", liveVersion)
	}

	storedVersion, storedErr := r.storedVersion()
	if storedErr == nil && metricexecutor.IsValidProductVersion(storedVersion) {
		r.cache(storedVersion, now)
		return storedVersion, nil
	}
	if storedErr == nil {
		storedErr = fmt.Errorf("invalid stored seekdb product version %q", storedVersion)
	}
	return "", fmt.Errorf("resolve seekdb product version failed: live: %v; sqlite: %v", liveErr, storedErr)
}

func (r *metricVersionResolver) cache(version string, now time.Time) {
	r.cachedVersion = version
	r.validUntil = now.Add(r.ttl)
}

var defaultMetricVersionResolver = newMetricVersionResolver(
	func() (string, error) {
		return clusterService.GetObVersion()
	},
	func() (string, error) {
		var version string
		err := (&agentservice.AgentService{}).GetObConfig(constant.CONFIG_OB_VERSION, &version)
		return version, err
	},
)
