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

package coordinator

import (
	"testing"
	"time"
)

func TestWatcherWaitUsesRemainingLease(t *testing.T) {
	c := NewCoordinator()
	c.identity = WATCHER
	c.Maintainer.setLifeTime(3.8)
	// The five-second snapshot has only 1.2 seconds left. Sleeping five
	// more seconds leaves a healthy maintainer falsely inactive locally.
	if duration := c.waitDuration(); duration != 1200*time.Millisecond {
		t.Fatalf("watcher waits %v with only 1.2s remaining in the lease", duration)
	}
}

func TestMaintainerWaitUsesRemainingRenewalInterval(t *testing.T) {
	c := NewCoordinator()
	c.identity = MAINTAINER
	c.Maintainer.setLifeTime(3.8)
	if duration := c.waitDuration(); duration != time.Second {
		t.Fatalf("maintainer waits %v instead of the minimum reconcile interval", duration)
	}
}

func TestRefreshLifeTimeUsesCurrentSnapshot(t *testing.T) {
	c := NewCoordinator()
	c.Maintainer.setLifeTime(3.8)
	// Another RPC refreshed the snapshot during the coordinator's wait.
	c.Maintainer = &Maintainer{}
	c.Maintainer.setLifeTime(0.1)
	c.refreshLifeTime()
	if life := c.Maintainer.GetLifeTime(); life < 0.1 || life > 1 {
		t.Fatalf("refreshed snapshot has incorrect age %f", life)
	}
}
