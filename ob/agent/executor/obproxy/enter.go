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

package obproxy

import (
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	agentservice "github.com/oceanbase/obshell/ob/agent/service/agent"
	obclusterservice "github.com/oceanbase/obshell/ob/agent/service/obcluster"
	obproxyservice "github.com/oceanbase/obshell/ob/agent/service/obproxy"
	taskservice "github.com/oceanbase/obshell/ob/agent/service/task"
)

var obproxyService = obproxyservice.ObproxyService{}
var obclusterService = obclusterservice.ObclusterService{}
var localTaskService = taskservice.NewLocalTaskService()
var agentService = agentservice.AgentService{}

const (
	WORK_MODE_RS_LIST    = "rsList"
	WORK_MODE_CONFIG_URL = "configUrl"
)

var (
	// task name for obproxy
	DAG_ADD_OBPROXY     = "Add obproxy"
	DAG_START_OBPROXY   = "Start obproxy"
	DAG_STOP_OBPROXY    = "Stop obproxy"
	DAG_UPGRADE_OBPROXY = "Upgrade obproxy"
	DAG_DELETE_OBPROXY  = "Delete obproxy"

	TASK_START_OBPROXY             = "Start obproxy"
	TASK_START_OBPROXYD            = "Start obproxyd"
	TASK_SET_OBPROXY_SYS_PASSWORD  = "Set obproxy sys password"
	TASK_SET_OBPROXY_USER_PASSWORD = "Set proxyro password for connect"
	TASK_PERSIST_OBPROXY_INFP      = "Persist obproxy info"
	TASK_STOP_OBPROXY              = "Stop obproxy"
	TASK_CHECK_OBPROXY_STATUS      = "Check obproxy status"
	TASK_CHECK_PROXYRO_PASSWORD    = "Check proxyro password"
	TASK_DELETE_OBPROXY            = "Delete obproxy"
	TASK_CLEAN_OBPROXY_DIR         = "Clean obproxy dir"

	TASK_COPY_CONFIG_DB_FILE             = "Copy obproxy config db file"
	TASK_HOT_RESTART_OBPROXY             = "Hot restart obproxy"
	TASK_WAIT_HOT_RESTART_OBPROXY_FINISH = "Wait hot restart obproxy finish"
	TASK_RECORD_OBPROXY_INFO             = "Record obproxy info"
	TASK_REINSTALL_OBPROXY_BIN           = "Reinstall obproxy bin"
	TASK_DOWNLOAD_RPM_FROM_SQLITE        = "Download obproxy pkg from sqlite"
	TASK_CHECK_OBPROXY_PKG               = "Check obproxy pkg"
	TASK_INSTALL_ALL_REQUIRED_PKGS       = "Install all required pkgs"
	TASK_BACKUP_FOR_UPGRADE              = "Backup for upgrade"
	TASK_REMOVE_UPGRADE_DIR              = "Remove upgrade dir"
	TASK_CREATE_UPGRADE_DIR              = "Create upgrade dir"

	PARAM_ADD_OBPROXY_OPTION         = "addObproxyOption"
	PARAM_OBPROXY_HOME_PATH          = "homePath"
	PARAM_OBPROXY_SQL_PORT           = "sqlPort"
	PARAM_OBPROXY_EXPORTER_PORT      = "exporterPort"
	PARAM_OBPROXY_VERSION            = "version"
	PARAM_OBPROXY_START_PARAMS       = "startParams"
	PARAM_OBPROXY_START_WITH_OPTIONS = "startWithOptions"
	PARAM_OBPROXY_APP_NAME           = "appName"
	PARAM_OBPROXY_WORK_MODE          = "workMode"
	PARAM_OBPROXY_RS_LIST            = "rsList"
	PARAM_OBPROXY_CONFIG_URL         = "configUrl"
	PARAM_OBPROXY_SYS_PASSWORD       = "obproxySysPassword"
	PARAM_OBPROXY_PROXYRO_PASSWORD   = "proxyroPassword"
	PARAM_OBPROXY_CLUSTER_NAME       = "clusterName"
	PARAM_PERSIST_OBPROXY_INFO_PARAM = "persistObproxyInfoParam"
	PARAM_EXPECT_OBPROXY_AGENT       = "expectObproxyAgent"

	PARAM_HOT_UPGRADE_ROLLBACK_TIMEOUT = "hotUpgradeRollbackTimeout"
	PARAM_HOT_UPGRADE_EXIT_TIMEOUT     = "hotUpgradeExitTimeout"
	PARAM_OLD_OBPROXY_PID              = "oldObproxyPid"

	// for upgrade
	PARAM_VERSION                 = "version"
	PARAM_BUILD_NUMBER            = "buildNumber"
	PARAM_DISTRIBUTION            = "distribution"
	PARAM_RELEASE_DISTRIBUTION    = "releaseDistribution"
	PARAM_UPGRADE_DIR             = "upgradeDir"
	PARAM_TASK_TIME               = "taskTime"
	PARAM_ONLY_FOR_AGENT          = "onlyForAgent"
	PARAM_SCRIPT_FILE             = "scriptFile"
	PARAM_OBPROXY_RPM_PKG_PATH    = "obproxyRpmPkgPath"
	PARAM_CREATE_UPGRADE_DIR_FLAG = "createUpgradeDirFlag"

	// stop obproxy or obproxyd retry times
	STOP_PROCESS_MAX_RETRY_TIME = 15
	STOP_PROCESS_RETRY_INTERVAL = 5
)

func RegisterTaskType() {
	task.RegisterTaskType(StartObproxyTask{})
	task.RegisterTaskType(SetObproxyUserPasswordForObTask{})
	task.RegisterTaskType(PersistObproxyInfoTask{})
	task.RegisterTaskType(StopObproxyTask{})
	task.RegisterTaskType(PrepareForAddObproxyTask{})

	task.RegisterTaskType(StopObproxyTask{})

	task.RegisterTaskType(CopyConfigDbFileTask{})
	task.RegisterTaskType(HotRestartObproxyTask{})
	task.RegisterTaskType(WaitHotRestartObproxyFinishTask{})
	task.RegisterTaskType(RecordObproxyInfoTask{})
	task.RegisterTaskType(ReinstallObproxyBinTask{})
	task.RegisterTaskType(GetObproxyPkgTask{})
	task.RegisterTaskType(CheckObproxyPkgTask{})
	task.RegisterTaskType(BackupObproxyForUpgradeTask{})
	task.RegisterTaskType(RemoveUpgradeObproxyDirTask{})
	task.RegisterTaskType(CreateObproxyUpgradeDirTask{})

	task.RegisterTaskType(CleanObproxyDirTask{})
	task.RegisterTaskType(DeleteObproxyTask{})
}
