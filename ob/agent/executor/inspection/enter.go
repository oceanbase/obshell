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
package inspection

import (
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/service/agent"
	"github.com/oceanbase/obshell/ob/agent/service/credential"
	"github.com/oceanbase/obshell/ob/agent/service/inspection"
	"github.com/oceanbase/obshell/ob/agent/service/obcluster"
	taskservice "github.com/oceanbase/obshell/ob/agent/service/task"
)

var (
	DAG_TRIGGER_INSPECTION = "Cluster inspection"

	TASK_NAME_INSTALL_OBDIAG  = "Install Obdiag"
	TASK_NAME_GENERATE_CONFIG = "Generate Inspection Config"
	TASK_NAME_INSPECTION      = "Inspection Task"
	TASK_NAME_GENERATE_REPORT = "Generate Inspection Report"

	PARAM_SCENARIO     = "scenario"
	PARAM_VERSION      = "version"
	PARAM_RELEASE      = "release"
	PARAM_DISTRIBUTION = "distribution"
	PARAM_ARCHITECTURE = "architecture"

	PARAM_USE_PASSWORDLESS_SSH = "use_passwordless_ssh"
	PARAM_USE_WORK_PATH        = "use_work_path"
	PARAM_OBDIAG_BIN_PATH      = "obdiag_bin_path"

	DATA_INSPECTION_CONFIG      = "inspection_config"
	DATA_INSPECTION_RESULT      = "inspection_result"
	DATA_INSPECTION_START_TIME  = "inspection_start_time"
	DATA_INSPECTION_FINISH_TIME = "inspection_finish_time"

	localTaskService  = taskservice.NewLocalTaskService()
	obclusterService  = obcluster.ObclusterService{}
	agentService      = agent.AgentService{}
	inspectionService = inspection.InspectionService{}
	credentialService = credential.CredentialService{}
)

func RegisterInspectionTask() {
	task.RegisterTaskType(InstallObdiagTask{})
	task.RegisterTaskType(GenerateConfigTask{})
	task.RegisterTaskType(InspectionTask{})
	task.RegisterTaskType(GenerateReportTask{})
}

type ObdiagResult struct {
	Data struct {
		Observer struct {
			Fail     map[string][]string `json:"fail"`
			Critical map[string][]string `json:"critical"`
			Warning  map[string][]string `json:"warning"`
			All      map[string][]string `json:"all"`
		} `json:"observer"`
	} `json:"data"`
	ErrorData string `json:"error_data"`
}
