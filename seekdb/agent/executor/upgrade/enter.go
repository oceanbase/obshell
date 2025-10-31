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

package upgrade

import (
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/service/agent"
	"github.com/oceanbase/obshell/seekdb/agent/service/obcluster"
	taskservice "github.com/oceanbase/obshell/seekdb/agent/service/task"
)

const (
	// for upgrade
	PARAM_VERSION                  = "version"
	PARAM_BUILD_NUMBER             = "buildNumber"
	PARAM_DISTRIBUTION             = "distribution"
	PARAM_RELEASE_DISTRIBUTION     = "releaseDistribution"
	PARAM_UPGRADE_DIR              = "upgradeDir"
	PARAM_UPGRADE_PKG_INSTALL_INFO = "upgradePkgInstallInfo"
	PARAM_TASK_TIME                = "taskTime"
	PARAM_UPGRADE_CHECK_TASK_DIR   = "upgradeCheckTaskDir"

	PARAM_TARGET_AGENT_BUILD_VERSION = "targetAgentBuildVersion"

	DATA_SKIP_START_TASK = "skipStartTask"

	// for upgrade
	DATA_BACKUP_DIR = "backupDir"

	// task name
	TASK_CHECK_PYTHON_ENV              = "Check the python environment"
	TASK_BACKUP_FOR_UPGRADE            = "Backup for upgrade"
	TASK_INSTALL_NEW_OBSHELL           = "Install new obshell"
	TASK_RESTART_OBSHELL               = "Restart obshell"
	TASK_CREATE_UPGRADE_DIR            = "Create upgrade dir"
	TASK_REMOVE_UPGRADE_CHECK_TASK_DIR = "Remove upgrade check task dir"
	TASK_CHECK_ALL_REQUIRED_PKGS       = "Check all required packages"
	TASK_GET_ALL_REQUIRED_PKGS         = "Download all required packages"
	TASK_INSTALL_ALL_REQUIRED_PKGS     = "Unpack all required packages"
	TASK_UPGRADE_POST_TABLE_MAINTAIN   = "Upgrade post table maintain"

	// dag name
	DAG_CHECK_AND_UPGRADE_OBSHELL = "Check and upgrade obshell"
	DAG_UPGRADE_OBSHELL           = "Upgrade obshell"
	DAG_UPGRADE_CHECK_OBSHELL     = "Upgrade check obshell"
)

var (
	// start

	agentService     = agent.AgentService{}
	obclusterService = obcluster.ObclusterService{}
	taskService      = taskservice.NewClusterTaskService()
)

func RegisterUpgradeTask() {
	// upgrade check
	task.RegisterTaskType(CreateUpgradeDirTask{})
	task.RegisterTaskType(GetAllRequiredPkgsTask{})
	task.RegisterTaskType(CheckAllRequiredPkgsTask{})
	task.RegisterTaskType(InstallAllRequiredPkgsTask{})
	task.RegisterTaskType(RemoveUpgradeCheckDirTask{})

	// agent upgrade
	task.RegisterTaskType(BackupAgentForUpgradeTask{})
	task.RegisterTaskType(InstallNewAgentTask{})
	task.RegisterTaskType(RestartAgentTask{})
	task.RegisterTaskType(UpgradePostTableMaintainTask{})
}
