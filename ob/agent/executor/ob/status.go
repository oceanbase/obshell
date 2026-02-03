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

package ob

import (
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
)

const (
	OBSERVER_STATUS_RUNNING          = iota //           = "RUNNING"
	OBSERVER_STATUS_UNAVAILABLE      = 1    //  = "UNAVAILABLE"
	OBSERVER_STATUS_SERVICE_STOPPED  = 2    //  = "SERVICE_STOPPED"
	OBSERVER_STATUS_PROCESS_STOPPED  = 3    //  = "PROCESS_STOPPED"
	OBSERVER_STATUS_SERVICE_STOPPING = 6    //  = "SERVICE_STOPPING"
	OBSERVER_STATUS_PROCESS_STOPPING = 7    //  = "PROCESS_STOPPING"
	OBSERVER_STATUS_STARTING         = 8    //  = "STARTING"
	OBSERVER_STATUS_DELETING         = 11   //  = "DELETING"
)

var observerStatusMap = map[int]string{
	OBSERVER_STATUS_RUNNING:          "RUNNING",
	OBSERVER_STATUS_UNAVAILABLE:      "UNAVAILABLE",
	OBSERVER_STATUS_SERVICE_STOPPED:  "SERVICE_STOPPED",
	OBSERVER_STATUS_PROCESS_STOPPED:  "PROCESS_STOPPED",
	OBSERVER_STATUS_SERVICE_STOPPING: "SERVICE_STOPPING",
	OBSERVER_STATUS_PROCESS_STOPPING: "PROCESS_STOPPING",
	OBSERVER_STATUS_STARTING:         "STARTING",
	OBSERVER_STATUS_DELETING:         "DELETING",
}

const (
	ZONE_STATUS_RUNNING     = iota //
	ZONE_STATUS_UNAVAILABLE = 1    //
	ZONE_STATUS_STOPPED     = 2    //
	ZONE_STATUS_CREATING    = 3    // not used
	ZONE_STATUS_DELETED     = 4    // not used
	ZONE_STATUS_STOPPING    = 5    //
	ZONE_STATUS_STARTING    = 6    //
	ZONE_STATUS_RESTARTING  = 7    // not used
	ZONE_STATUS_DELETING    = 8    //
	ZONE_STATUS_UNKNOWN     = 9    //
)

var zoneStatusMap = map[int]string{
	ZONE_STATUS_RUNNING:     "RUNNING",
	ZONE_STATUS_UNAVAILABLE: "UNAVAILABLE",
	ZONE_STATUS_STOPPED:     "STOPPED",
	ZONE_STATUS_CREATING:    "CREATING",
	ZONE_STATUS_DELETED:     "DELETED",
	ZONE_STATUS_STOPPING:    "STOPPING",
	ZONE_STATUS_STARTING:    "STARTING",
	ZONE_STATUS_RESTARTING:  "RESTARTING",
	ZONE_STATUS_DELETING:    "DELETING",
	ZONE_STATUS_UNKNOWN:     "UNKNOWN",
}

// fixObserverStatus determines the observer status based on OB internal status, zone status, process state, and maintenance tasks
func fixObserverStatus(server *bo.Observer, zoneName string, obState int, taskInfo *MainDagTaskInfo, allAgents []bo.AllAgent) int {
	// 1. Check maintenance intermediate state (based on main dag)
	if taskInfo != nil && taskInfo.TaskId != "" {
		// Check if current observer or zone is in the maintenance scope
		if isInMaintenanceScope(server, zoneName, taskInfo, allAgents) {
			switch taskInfo.DagName {
			case DAG_START_OB:
				return OBSERVER_STATUS_STARTING
			case DAG_STOP_OB:
				if taskInfo.StopObserverProcess {
					return OBSERVER_STATUS_PROCESS_STOPPING
				}
				return OBSERVER_STATUS_SERVICE_STOPPING
				// Add other main dag types as needed
			}
		}
	}

	// 2. Check DELETING
	if server.InnerStatus == "DELETING" {
		return OBSERVER_STATUS_DELETING
	}

	// 3. Check UNAVAILABLE (INACTIVE and not PROCESS_STOPPED)
	if server.InnerStatus == "INACTIVE" {
		// If process is not running, it's PROCESS_STOPPED, not UNAVAILABLE
		if obState == oceanbase.STATE_PROCESS_NOT_RUNNING {
			return OBSERVER_STATUS_PROCESS_STOPPED
		}
		return OBSERVER_STATUS_UNAVAILABLE
	}

	// 4. Check SERVICE_STOPPED (based on stopTime)
	if !server.StopTime.IsZero() {
		// If process is not running, it's PROCESS_STOPPED
		if obState == oceanbase.STATE_PROCESS_NOT_RUNNING {
			return OBSERVER_STATUS_PROCESS_STOPPED
		}
		return OBSERVER_STATUS_SERVICE_STOPPED
	}

	// 6. Check RUNNING
	if server.InnerStatus == "ACTIVE" && obState == oceanbase.STATE_CONNECTION_AVAILABLE {
		return OBSERVER_STATUS_RUNNING
	}

	// 7. Check UNAVAILABLE (process running but connection unavailable)
	if obState == oceanbase.STATE_PROCESS_RUNNING || obState == oceanbase.STATE_CONNECTION_RESTRICTED {
		return OBSERVER_STATUS_UNAVAILABLE
	}

	// 8. Check PROCESS_STOPPED
	if obState == oceanbase.STATE_PROCESS_NOT_RUNNING {
		return OBSERVER_STATUS_PROCESS_STOPPED
	}

	// Default to UNAVAILABLE
	return OBSERVER_STATUS_UNAVAILABLE
}

// fixZoneStatus determines the zone status based on OB internal status and maintenance tasks
func fixZoneStatus(zone *bo.Zone, taskInfo *MainDagTaskInfo) int {
	// 1. Check maintenance intermediate state (based on main dag)
	if taskInfo != nil && taskInfo.TaskId != "" {
		// Check if current zone is in the maintenance scope
		if isZoneInMaintenanceScope(zone.Name, taskInfo) {
			switch taskInfo.DagName {
			case DAG_START_OB:
				if taskInfo.ScopeType == SCOPE_ZONE {
					return ZONE_STATUS_STARTING
				}
				// For GLOBAL scope, all zones are starting
				return ZONE_STATUS_STARTING
			case DAG_STOP_OB:
				return ZONE_STATUS_STOPPING
			}
		}
	}

	// 2. Check SERVICE_STOPPED (INACTIVE)
	if zone.InnerStatus == "INACTIVE" {
		return ZONE_STATUS_STOPPED
	}

	// 3. Check RUNNING (ACTIVE)
	if zone.InnerStatus == "ACTIVE" {
		return ZONE_STATUS_RUNNING
	}

	// Default to UNKNOWN
	return ZONE_STATUS_UNKNOWN
}

// isZoneInMaintenanceScope checks if the zone is in the maintenance scope of the main dag
func isZoneInMaintenanceScope(zoneName string, taskInfo *MainDagTaskInfo) bool {
	if taskInfo == nil || taskInfo.TaskId == "" {
		return false
	}

	switch taskInfo.ScopeType {
	case SCOPE_GLOBAL:
		// GLOBAL scope affects all zones
		return true
	case SCOPE_ZONE:
		// ZONE scope affects only specified zones
		for _, target := range taskInfo.Targets {
			if target == zoneName {
				return true
			}
		}
	case SCOPE_SERVER:
		// SERVER scope doesn't affect zone status directly
		// Zone status will be determined by its own OB internal status
		return false
	}
	return false
}
