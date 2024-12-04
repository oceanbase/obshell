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

package obcluster

type ObserverService struct{}
type ObclusterService struct{}

const (
	ob_parameters_view = "oceanbase.V$OB_PARAMETERS"

	COLLATIONS       = "information_schema.collations"
	DBA_OB_SERVERS   = "oceanbase.DBA_OB_SERVERS"
	DBA_OB_ZONES     = "oceanbase.DBA_OB_ZONES"
	DBA_OB_UNITS     = "oceanbase.DBA_OB_UNITS"
	GV_OB_LOG_STAT   = "oceanbase.GV$OB_LOG_STAT"
	GV_OB_PARAMETERS = "oceanbase.GV$OB_PARAMETERS"
)
