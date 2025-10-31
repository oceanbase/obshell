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
	"github.com/oceanbase/obshell/ob/agent/config"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/repository/db/oceanbase"
	"github.com/oceanbase/obshell/ob/agent/secure"
)

type ModifyPwdTask struct {
	task.Task
	password       string
	cipherPassword string
}

func newModifyPwdTask() *ModifyPwdTask {
	newTask := &ModifyPwdTask{
		Task: *task.NewSubTask(TASK_NAME_MODIFY_PWD),
	}
	newTask.
		SetCanRetry().
		SetCanContinue().
		SetCanRollback()
	return newTask
}

func (t *ModifyPwdTask) Execute() error {
	if err := t.setRootPWD(); err != nil {
		return err
	}
	dsConfig := config.NewObMysqlDataSourceConfig().SetPassword(t.password)
	if err := loadOceanbaseInstanceWithoutDBName(t); err != nil {
		t.ExecuteLog("Failed to connect to OceanBase. Try connecting with a new password.")
		if err := oceanbase.LoadOceanbaseInstance(dsConfig); err != nil {
			return err
		}
		t.ExecuteLog("Successfully connected to OceanBase. Not need to modify password.")
		return secure.UpdateObPassword(t.password)
	}
	t.ExecuteLog("Successfully connected to OceanBase. Try to modify password.")
	if err := obclusterService.ModifyUserPwd(constant.DB_USERNAME, t.password); err != nil {
		return errors.Wrap(err, "modify root password failed")
	}
	t.ExecuteLog("Save root password")
	if err := secure.UpdateObPassword(t.cipherPassword); err != nil {
		return errors.Wrap(err, "save root password failed")
	}
	t.ExecuteLog("Reload oceanbase connection")
	if err := oceanbase.LoadOceanbaseInstance(dsConfig); err != nil {
		return errors.Wrap(err, "reload oceanbase connection failed")
	}
	t.ExecuteLog("Reload oceanbase connection successfully")
	return nil
}

func (t *ModifyPwdTask) setRootPWD() (err error) {
	cipherPassword := t.GetContext().GetData(PARAM_ROOT_PWD)
	if cipherPassword != nil {
		t.cipherPassword = cipherPassword.(string)
		t.password, err = secure.Decrypt(t.cipherPassword)
	}
	return
}
