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

package alert

import (
	"github.com/oceanbase/obshell/ob/model/alarm"
	"github.com/oceanbase/obshell/ob/model/oceanbase"
)

type AlertFilter struct {
	Severity     alarm.Severity           `json:"severity,omitempty"`
	InstanceType oceanbase.OBInstanceType `json:"instance_type,omitempty"`
	Instance     *oceanbase.OBInstance    `json:"instance,omitempty"`
	StartTime    int64                    `json:"start_time,omitempty"`
	EndTime      int64                    `json:"end_time,omitempty"`
	Keyword      string                   `json:"keyword,omitempty"`
}
