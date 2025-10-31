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
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	"github.com/oceanbase/obshell/ob/agent/lib/system"
	"github.com/oceanbase/obshell/ob/agent/meta"
)

type ReinstallAndRestartObTask struct {
	task.Task
	realExecAgent  meta.AgentInfo
	rpmPkgHomePath string
	zone           string
}

func newReinstallAndRestartObNode(zone string, agents []meta.AgentInfo, idx int) *task.Node {
	ctx := task.NewTaskContext()
	ctx.SetParam(task.EXECUTE_AGENTS, agents).
		SetParam(PARAM_UPGRADE_ROUTE_INDEX, idx).
		SetParam(PARAM_ZONE, zone)
	return task.NewNodeWithContext(&ReinstallAndRestartObTask{
		Task: *task.NewSubTask(TASK_REINSTALL_AND_RESTART_OBSERVER).
			SetCanContinue().
			SetCanRetry()},
		true, ctx)
}

func (t *ReinstallAndRestartObTask) getParams() (err error) {
	_, t.realExecAgent, err = isRealExecuteAgent(t)
	if err != nil {
		return err
	}
	if err = t.GetContext().GetParamWithValue(PARAM_ZONE, &t.zone); err != nil {
		return err
	}

	var upgradeRouteIndex float64
	if err = t.GetContext().GetParamWithValue(PARAM_UPGRADE_ROUTE_INDEX, &upgradeRouteIndex); err != nil {
		return err
	}
	upgradeRoute, err := getUpgradeRouteForTask(t.GetContext())
	if err != nil {
		return err
	}
	node := upgradeRoute[int(upgradeRouteIndex)]
	var rpmPkgInfo rpmPacakgeInstallInfo
	if err = t.GetContext().GetAgentDataByAgentKeyWithValue(t.realExecAgent.String(), node.BuildVersion, &rpmPkgInfo); err != nil {
		return err
	}
	t.rpmPkgHomePath = rpmPkgInfo.RpmPkgHomepath
	return nil
}

func (t *ReinstallAndRestartObTask) Execute() (err error) {
	if err = t.getParams(); err != nil {
		return err
	}

	err = t.reinstallAndRestartOb()
	if err != nil {
		t.ExecuteErrorLog(err)
		return
	}
	t.ExecuteLog("reinstall and restart ob success")
	return nil
}

func (t *ReinstallAndRestartObTask) reinstallAndRestartOb() (err error) {
	t.ExecuteLog("build start service time map")
	t.ExecuteLog("stop ob")
	if err = stopObserver(t); err != nil {
		return err
	}
	t.ExecuteLog("reinstall ob")
	if err = t.installNewOb(t.rpmPkgHomePath); err != nil {
		return err
	}
	t.ExecuteLog("start ob")
	if err = startObserver(t, nil); err != nil {
		return err
	}
	t.ExecuteLog("wait all observer available")
	return t.waitAllObSeverAvailable()
}

func (t *ReinstallAndRestartObTask) installNewOb(upgradePath string) (err error) {
	t.ExecuteLogf("copy new ob from '%s' to '%s'", upgradePath, global.HomePath)
	if err = copyFilesForInstallObserver(upgradePath, global.HomePath); err != nil {
		return errors.Wrap(err, "copy new ob failed")
	}
	return nil
}

func copyFilesForInstallObserver(src, dest string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return system.CopyFile(src, dest)
	}
	if err = os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}
	entries, err := in.Readdir(0)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), constant.PROC_OBSHELL) {
			continue
		}
		subSrc := filepath.Join(src, entry.Name())
		subDest := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			if err = copyFilesForInstallObserver(subSrc, subDest); err != nil {
				return err
			}
		} else {
			if err = os.RemoveAll(subDest); err != nil {
				return err
			}
			if err = system.CopyFile(subSrc, subDest); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *ReinstallAndRestartObTask) waitAllObSeverAvailable() (err error) {
	log.Info("wait all observer available")
	for i := 0; i < constant.TICK_NUM_FOR_OB_STATUS_CHECK; i++ {
		allObserverIsAvailable, _ := isAllObSeverAvailable()
		if allObserverIsAvailable {
			return nil
		}
		time.Sleep(constant.TICK_INTERVAL_FOR_OB_STATUS_CHECK)
		t.TimeoutCheck()

	}
	return errors.Occur(errors.ErrObClusterAsyncOperationTimeout, "wait all observer available")
}

func isAllObSeverAvailable() (res bool, err error) {
	count, err := obclusterService.GetInactiveServerCount()
	if err != nil || count != 0 {
		return
	}
	count, err = obclusterService.GetNotInSyncServerCount()
	if err != nil || count != 0 {
		return
	}
	return true, nil
}
