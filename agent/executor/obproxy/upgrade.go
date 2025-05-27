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
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/global"
	"github.com/oceanbase/obshell/agent/lib/parse"
	"github.com/oceanbase/obshell/agent/lib/path"
	"github.com/oceanbase/obshell/agent/lib/pkg"
	"github.com/oceanbase/obshell/agent/lib/process"
	"github.com/oceanbase/obshell/agent/lib/system"
	"github.com/oceanbase/obshell/agent/meta"
	obproxydb "github.com/oceanbase/obshell/agent/repository/db/obproxy"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/param"
	"github.com/oceanbase/obshell/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const waitPeriod = 5 // seconds

func UpgradeObproxy(param param.UpgradeObproxyParam) (*task.DagDetailDTO, *errors.OcsAgentError) {
	if !meta.IsObproxyAgent() {
		return nil, errors.Occur(errors.ErrBadRequest, "not obproxy agent")
	}
	if alive, err := process.CheckObproxyProcess(); err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	} else if !alive {
		return nil, errors.Occur(errors.ErrBadRequest, "obproxy is not running")
	}

	if err := checkVersionSupport(param.Version, param.Release); err != nil {
		return nil, err
	}
	if err := checkUpgradeDir(&param.UpgradeDir); err != nil {
		return nil, errors.Occur(errors.ErrIllegalArgument, err)
	}
	if err := findTargetPkg(param.Version, param.Release); err != nil {
		return nil, err
	}

	template := buildUpgradeObproxyTemplate()
	context := buildUpgradeObproxyTaskContext(param)
	dag, err := localTaskService.CreateDagInstanceByTemplate(template, context)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return task.NewDagDetailDTO(dag), nil
}

func checkVersionSupport(version, release string) *errors.OcsAgentError {
	// Check obproxy version
	curObproxyVersion, err := obproxyService.GetObproxyVersion()
	if err != nil {
		return errors.Occur(errors.ErrUnexpected, err)
	}
	buildNumber, _, err := pkg.SplitRelease(release)
	if err != nil {
		return errors.Occur(errors.ErrUnexpected, err)
	}
	if pkg.CompareVersion(curObproxyVersion, fmt.Sprintf("%s-%s", version, buildNumber)) >= 0 {
		return errors.Occur(errors.ErrBadRequest, "current obproxy version is greater than or equal to the target version")
	}
	return nil
}

func findTargetPkg(version, release string) *errors.OcsAgentError {
	buildNumber, distribution, _ := pkg.SplitRelease(release)
	_, err := agentService.GetUpgradePkgInfoByVersionAndRelease(constant.PKG_OBPROXY_CE, version, buildNumber, distribution, global.Architecture)
	if err != nil {
		return errors.Occurf(errors.ErrBadRequest, "find target pkg '%s-%s-%s.%s.rpm' failed", constant.PKG_OBPROXY_CE, version, release, global.Architecture)
	}
	return nil
}

func checkUpgradeDir(path *string) (err error) {
	log.Infof("checking upgrade directory: '%s'", *path)
	str := *path

	*path = strings.TrimSpace(*path)
	if len(*path) == 0 {
		return nil
	}

	return utils.CheckPathValid(str)
}

func buildUpgradeObproxyTemplate() *task.Template {
	return task.NewTemplateBuilder(DAG_UPGRADE_OBPROXY).
		SetMaintenance(task.ObproxyMaintenance()).
		SetType(task.DAG_OBPROXY).
		AddTask(newCreateObproxyUpgradeDirTask(), false).
		AddTask(newGetObproxyPkgTask(), false).
		AddTask(newCheckObproxyPkgTask(), false).
		AddTask(newBackupObproxyForUpgradeTask(), false).
		AddTask(newReinstallObproxyBinTask(), false).
		AddTask(newCopyConfigDbFileTask(), false).
		AddTask(newRecordObproxyInfoTask(), false).
		AddTask(newHotRestartObproxyTask(), false).
		AddTask(newWaitHotRestartObproxyFinishTask(), false).
		AddTask(newRemoveUpgradeCheckDirTask(), false).
		Build()
}

func buildUpgradeObproxyTaskContext(param param.UpgradeObproxyParam) *task.TaskContext {
	if param.UpgradeDir == "" {
		param.UpgradeDir = meta.OBPROXY_HOME_PATH
	}
	buildNumber, distribution, _ := pkg.SplitRelease(param.Release)
	return task.NewTaskContext().
		SetParam(PARAM_UPGRADE_DIR, fmt.Sprintf("%s/%s-%d", param.UpgradeDir, "obproxy-upgrade-dir", time.Now().Unix())).
		SetParam(PARAM_VERSION, param.Version).
		SetParam(PARAM_BUILD_NUMBER, buildNumber).
		SetParam(PARAM_DISTRIBUTION, distribution).
		SetParam(PARAM_RELEASE_DISTRIBUTION, param.Release)
}

type CopyConfigDbFileTask struct {
	task.Task
	targetVersion string
}

func newCopyConfigDbFileTask() *CopyConfigDbFileTask {
	newTask := &CopyConfigDbFileTask{
		Task: *task.NewSubTask(TASK_COPY_CONFIG_DB_FILE),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *CopyConfigDbFileTask) Execute() error {
	if err := t.GetContext().GetParamWithValue(PARAM_VERSION, &t.targetVersion); err != nil {
		return err
	}
	if pkg.CompareVersion(t.targetVersion, "4.1.0.0") >= 0 {
		if _, err := os.Stat(path.ObproxyNewConfigDbFile()); err == nil {
			return nil
		} else {
			return system.CopyFile(path.ObproxyOldConfigDbFile(), path.ObproxyNewConfigDbFile())
		}
	}

	return nil
}

func (t *CopyConfigDbFileTask) Rollback() error {
	if pkg.CompareVersion(t.targetVersion, "4.1.0.0") >= 0 {
		if _, err := os.Stat(path.ObproxyNewConfigDbFile()); err == nil {
			return nil
		} else {
			return system.CopyFile(path.ObproxyOldConfigDbFile(), path.ObproxyNewConfigDbFile())
		}
	}
	return nil
}

type HotRestartObproxyTask struct {
	task.Task
}

func newHotRestartObproxyTask() *HotRestartObproxyTask {
	newTask := &HotRestartObproxyTask{
		Task: *task.NewSubTask(TASK_HOT_RESTART_OBPROXY),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel()
	return newTask
}

func (t *HotRestartObproxyTask) Execute() error {
	t.ExecuteLogf("set %s to %s", constant.OBPROXY_CONFIG_PROXY_LOCAL_CMD, constant.RESTART_FOR_PROXY_LOCAL_CMD)
	return obproxyService.SetGlobalConfig(constant.OBPROXY_CONFIG_PROXY_LOCAL_CMD, constant.RESTART_FOR_PROXY_LOCAL_CMD)
}

type RecordObproxyInfoTask struct {
	task.Task
}

func newRecordObproxyInfoTask() *RecordObproxyInfoTask {
	newTask := &RecordObproxyInfoTask{
		Task: *task.NewSubTask(TASK_RECORD_OBPROXY_INFO),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanRollback()
	return newTask
}

func (t *RecordObproxyInfoTask) Execute() error {
	rollbackTimeout, err := obproxyService.GetGlobalConfig(constant.OBPROXY_CONFIG_HOT_UPGRADE_ROLLBACK_TIMEOUT)
	if err != nil {
		return errors.Wrapf(err, "get %s failed", constant.OBPROXY_CONFIG_HOT_UPGRADE_ROLLBACK_TIMEOUT)
	}
	pid, err := process.FindPIDByPort(uint32(meta.OBPROXY_SQL_PORT))
	if err != nil {
		return errors.Wrapf(err, "find obproxy pid failed")
	}
	t.GetContext().SetData(PARAM_OLD_OBPROXY_PID, pid)
	t.GetContext().SetData(PARAM_HOT_UPGRADE_ROLLBACK_TIMEOUT, rollbackTimeout)
	return nil
}

type WaitHotRestartObproxyFinishTask struct {
	task.Task
	rollbackTimeout string
	oldPid          int32
	targetVersion   string
	buildNumber     string
}

func newWaitHotRestartObproxyFinishTask() *WaitHotRestartObproxyFinishTask {
	newTask := &WaitHotRestartObproxyFinishTask{
		Task: *task.NewSubTask(TASK_WAIT_HOT_RESTART_OBPROXY_FINISH),
	}
	newTask.SetCanContinue().SetCanRetry().SetCanCancel()
	return newTask
}

func (t *WaitHotRestartObproxyFinishTask) Execute() error {
	var err error
	if err = t.GetContext().GetDataWithValue(PARAM_OLD_OBPROXY_PID, &t.oldPid); err != nil {
		return err
	}
	if err = t.GetContext().GetDataWithValue(PARAM_HOT_UPGRADE_ROLLBACK_TIMEOUT, &t.rollbackTimeout); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_VERSION, &t.targetVersion); err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_BUILD_NUMBER, &t.buildNumber); err != nil {
		return err
	}

	// parse rollbackTimeout
	rollbackTimeouot, err := parse.TimeParse(t.rollbackTimeout)
	if err != nil {
		return errors.Wrapf(err, "parse rollback timeout failed")
	}

	retryTimes := rollbackTimeouot / waitPeriod
	var pid int32
	for i := 0; i < retryTimes; i++ {
		t.TimeoutCheck()
		time.Sleep(time.Duration(waitPeriod) * time.Second)
		pid, err = process.FindPIDByPort(uint32(meta.OBPROXY_SQL_PORT))
		if err != nil {
			continue
		}
		t.ExecuteLogf("obproxy %d is running", pid)
		if pid == t.oldPid {
			t.ExecuteLogf("obproxy %d is still running, waiting for it to exit...", t.oldPid)
			err = errors.New("obproxy is still running")
			continue
		}
		err = t.checkVersion()
		break
	}

	if err == nil {
		// Modify the pid file.
		if err := process.WritePidForce(path.ObproxyPidPath(), int(pid)); err != nil {
			return errors.Wrapf(err, "write obproxy pid file failed")
		}
		return nil

	}

	return errors.Wrapf(err, "wait hot restart obproxy finish timeout")
}

func (t *WaitHotRestartObproxyFinishTask) checkVersion() (err error) {
	dsConfig := config.NewObproxyDataSourceConfig().SetPort(meta.OBPROXY_SQL_PORT).SetPassword(meta.OBPROXY_SYS_PWD)
	var tempDb *gorm.DB
	defer func() {
		if tempDb != nil {
			db, _ := tempDb.DB()
			db.Close()
		}
	}()
	for retryCount := 1; retryCount <= obproxydb.WAIT_OBPROXY_CONNECTED_MAX_TIMES; retryCount++ {
		t.ExecuteLogf("retry %d times", retryCount)
		t.TimeoutCheck()
		time.Sleep(obproxydb.WAIT_OBPROXY_CONNECTED_MAX_INTERVAL)
		if tempDb, err = obproxydb.LoadTempObproxyInstance(dsConfig); err != nil {
			t.ExecuteLogf("load obproxy instance failed: %s", err.Error())
			continue
		}
		var proxyInfo bo.ObproxyInfo
		if err = tempDb.Raw("show proxyinfo binary").Scan(&proxyInfo).Error; err != nil {
			t.ExecuteLogf("show proxyconfig failed: %s", err.Error())
			continue
		}
		// parse obproxy version
		re := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+-\d+`)
		version := re.FindString(proxyInfo.Info)
		if version != strings.Join([]string{t.targetVersion, t.buildNumber}, "-") {
			t.ExecuteLogf("obproxy version is not the target version, current version: %s, target version: %s", version, t.targetVersion)
			continue
		}
		return nil
	}
	return errors.New("check obproxy version timeout...")
}

type CreateObproxyUpgradeDirTask struct {
	task.Task
	upgradeDir string
}

func newCreateObproxyUpgradeDirTask() *CreateObproxyUpgradeDirTask {
	newTask := &CreateObproxyUpgradeDirTask{
		Task: *task.NewSubTask(TASK_CREATE_UPGRADE_DIR),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanRollback().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *CreateObproxyUpgradeDirTask) Execute() (err error) {
	if err = t.GetContext().GetParamWithValue(PARAM_UPGRADE_DIR, &t.upgradeDir); err != nil {
		return err
	}
	t.ExecuteLogf("Upgrade dir is %s", t.upgradeDir)
	if err = t.checkUpgradeDir(); err != nil {
		return err
	}
	return nil
}

func (t *CreateObproxyUpgradeDirTask) checkUpgradeDir() (err error) {
	t.GetContext().SetData(PARAM_CREATE_UPGRADE_DIR_FLAG, false)

	t.ExecuteLogf("Mkdir %s ", t.upgradeDir)
	if err = os.MkdirAll(t.upgradeDir, 0755); err != nil {
		return err
	}

	isDirEmpty, err := system.IsDirEmpty(t.upgradeDir)
	if err != nil {
		return err
	}
	if !isDirEmpty {
		return fmt.Errorf("%s is not empty", t.upgradeDir)
	}
	t.GetContext().SetData(PARAM_CREATE_UPGRADE_DIR_FLAG, true)
	return nil
}

func (t *CreateObproxyUpgradeDirTask) Rollback() (err error) {
	t.ExecuteLog("Rolling back...")
	if t.GetContext().GetData(PARAM_CREATE_UPGRADE_DIR_FLAG) == nil {
		return nil
	}
	t.ExecuteLog("Remove " + t.upgradeDir)
	return os.RemoveAll(t.upgradeDir)
}

// RemoveUpgradeObproxyDirTask remove upgrade dir
type RemoveUpgradeObproxyDirTask struct {
	task.Task
	upgradeDir string
}

func newRemoveUpgradeCheckDirTask() *RemoveUpgradeObproxyDirTask {
	newTask := &RemoveUpgradeObproxyDirTask{
		Task: *task.NewSubTask(TASK_REMOVE_UPGRADE_DIR),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanPass().
		SetCanCancel()
	return newTask
}

func (t *RemoveUpgradeObproxyDirTask) Execute() (err error) {
	t.ExecuteLog("remove upgrade dir")
	if err = t.removeUpgradeDir(); err != nil {
		return
	}
	t.ExecuteLog("remove upgrade check dir finished")
	return nil
}

func (t *RemoveUpgradeObproxyDirTask) removeUpgradeDir() (err error) {
	if err := t.GetContext().GetParamWithValue(PARAM_UPGRADE_DIR, &t.upgradeDir); err != nil {
		return errors.New("get upgrade check task dir failed")
	}
	return os.RemoveAll(t.upgradeDir)
}
