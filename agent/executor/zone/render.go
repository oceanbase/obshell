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

package zone

import (
	"strings"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/param"
)

func RenderZoneParams(zoneList []param.ZoneParam) {
	for i := range zoneList {
		if zoneList[i].ReplicaType == "" {
			zoneList[i].ReplicaType = constant.REPLICA_TYPE_FULL
		} else {
			zoneList[i].ReplicaType = strings.ToUpper(zoneList[i].ReplicaType)
		}
	}
}
