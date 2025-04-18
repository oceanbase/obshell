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

import { formatMessage } from '@/util/intl';
import { message } from '@oceanbase/design';
import * as DatabaseService from '@/service/ocp-express/ObDatabaseController';
import * as ObUserController from '@/service/ocp-express/ObUserController';
import * as ObDatabaseController from '@/service/ocp-express/ObDatabaseController';

export const namespace = 'database';

const model = {
  namespace,
  state: {
    databaseList: [],
    dbUserList: [],
  },

  effects: {
    *listDatabases({ payload }, { call, put }) {
      const res = yield call(DatabaseService.listDatabases, payload);
      if (res.successful) {
        yield put({
          type: 'update',
          payload: {
            databaseList: (res.data && res.data.contents) || [],
          },
        });
      }
    },
    *createDatabase({ payload, onSuccess }, { call }) {
      const { tenantId, ...rest } = payload;
      const res = yield call(
        ObDatabaseController.createDatabase,
        {
          tenantId,
        },

        rest
      );

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.database.DatabaseCreated',
            defaultMessage: '数据库新建成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *modifyDatabase({ payload, onSuccess }, { call }) {
      const { tenantId, dbName, ...rest } = payload;
      const res = yield call(
        DatabaseService.modifyDatabase,
        {
          tenantId,
          dbName,
        },

        rest
      );

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.database.DatabaseModified',
            defaultMessage: '数据库修改成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *deleteReplica({ payload, onSuccess }, { call }) {
      const { tenantId, dbName } = payload;
      const res = yield call(ObDatabaseController.deleteDatabase, {
        tenantId,
        dbName,
      });

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.database.DatabaseDeleted',
            defaultMessage: '数据库删除成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *dbUserList({ payload }, { call, put }) {
      const res = yield call(ObUserController.listDbUsers, payload);
      if (res.successful) {
        yield put({
          type: 'update',
          payload: {
            dbUserList: (res.data && res.data.contents) || [],
          },
        });
      }
    },
    *createDbUser({ payload, onSuccess }, { call }) {
      const { tenantId, ...rest } = payload;
      const res = yield call(
        ObUserController.createDbUser,
        {
          tenantId,
        },

        rest
      );

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.database.UserCreated',
            defaultMessage: '用户新建成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *modifyGlobalPrivilege({ payload, onSuccess }, { call }) {
      const { tenantId, username, ...rest } = payload;
      const res = yield call(
        ObUserController.modifyGlobalPrivilege,
        {
          tenantId,
          username,
        },

        rest
      );

      if (res.successful) {
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *modifyDbPrivilege({ payload, onSuccess }, { call }) {
      const { tenantId, username, ...rest } = payload;
      const res = yield call(
        ObUserController.modifyDbPrivilege,
        {
          tenantId,
          username,
        },

        rest
      );

      if (res.successful) {
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *changeDbUserPassword({ payload, onSuccess }, { call }) {
      const { tenantId, username, ...rest } = payload;
      const res = yield call(
        ObUserController.changeDbUserPassword,
        {
          tenantId,
          username,
        },

        rest
      );

      if (res.successful) {
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *deleteDbUser({ payload, onSuccess }, { call }) {
      const { tenantId, username } = payload;
      const res = yield call(ObUserController.deleteDbUser, {
        tenantId,
        username,
      });

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.database.UserDeleted',
            defaultMessage: '用户删除成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *unlockDbUser({ payload, onSuccess }, { call }) {
      const { tenantId, username } = payload;
      const res = yield call(ObUserController.unlockDbUser, {
        tenantId,
        username,
      });

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.database.TheUserIsUnlocked',
            defaultMessage: '用户解锁成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *lockDbUser({ payload, onSuccess }, { call }) {
      const { tenantId, username } = payload;
      const res = yield call(ObUserController.lockDbUser, {
        tenantId,
        username,
      });

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.database.TheUserIsLocked',
            defaultMessage: '用户锁定成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  },

  reducers: {
    update(state, { payload }) {
      return { ...state, ...payload };
    },
  },
};

export default model;
