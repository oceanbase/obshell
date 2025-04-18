/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

import { formatMessage } from '@/util/intl';
import { history } from 'umi';
import { message } from '@oceanbase/design';
import * as ObShellTenantController from '@/service/obshell/tenant';
import { DEFAULT_LIST_DATA } from '@/constant';
import { taskSuccess } from '@/util/task';

export const namespace = 'tenant';

const model = {
  namespace,
  state: {
    allTenantListData: DEFAULT_LIST_DATA,
    tenantListData: DEFAULT_LIST_DATA,
    tenantData: {},
    charsetListData: DEFAULT_LIST_DATA,
    parameterListData: DEFAULT_LIST_DATA,
    unitSpecList: [],
    unitSpecLimitRule: null,
  },

  effects: {
    // 根据集群下的全部租户
    *getTenantListData ({ payload }, { call, put }) {
      const res = yield call(ObShellTenantController.getTenantOverView, payload);
      if (res.successful) {
        yield put({
          type: 'update',
          payload: {
            tenantListData: res.data || DEFAULT_LIST_DATA,
          },
        });
      }
    },
    *getTenantData ({ payload, onSuccess }, { call, put }) {
      const res = yield call(ObShellTenantController.getTenantInfo, payload);
      if (res.successful) {

        const resTenantData = {
          ...(res?.data || {}),
          tenantName: res?.data?.tenant_name || ''
        };

        if (onSuccess) {
          onSuccess(resTenantData);
        }

        yield put({
          type: 'update',
          payload: {
            tenantData: resTenantData,
          },
        });
      }
    },
    *addTenant ({ payload, onSuccess }, { call }) {
      const { id, ...rest } = payload;
      const res = yield call(
        ObShellTenantController.tenantCreate,
        {
          id,
        },

        rest
      );

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.TenantCreatedSuccessfully',
            defaultMessage: '租户新建成功',
          })
        );

        if (onSuccess) {
          onSuccess();
        }
        // 由于创建租户是同步的，因此创建成功后应该跳转到租户详情页
        const tenantId = res.data && res.data.id;
        if (tenantId) {
          history.push(`/cluster/${id}/tenant/${tenantId}`);
        }
      }
    },
    *deleteTenant ({ payload, onSuccess }, { call }) {
      const res = yield call(ObShellTenantController.tenantDrop, payload);
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.TenantDeletedSuccessfully',
            defaultMessage: '租户删除成功',
          })
        );

        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *lockTenant ({ payload, onSuccess }, { call }) {
      const res = yield call(ObShellTenantController.tenantLock, payload);
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.TenantLockedSuccessfully',
            defaultMessage: '租户锁定成功',
          })
        );

        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *unlockTenant ({ payload, onSuccess }, { call }) {
      const res = yield call(ObShellTenantController.tenantUnlock, payload);
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.TenantUnlockedSuccessfully',
            defaultMessage: '租户解锁成功',
          })
        );

        if (onSuccess) {
          onSuccess();
        }
      }
    },
    *changePassword ({ payload, onSuccess }, { call, put }) {
      const { id, tenantId, ...rest } = payload;
      const res = yield call(
        ObShellTenantController.tenantModifyPassword,
        {
          id,
          tenantId,
        },

        rest
      );

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.TheTenantPasswordHasBeen',
            defaultMessage: '租户密码修改成功',
          })
        );

        if (onSuccess) {
          onSuccess();
        }
        yield put({
          type: 'getTenantData',
          payload: {
            id,
            tenantId,
          },
        });
      }
    },
    *modifyPrimaryZone ({ payload, onSuccess }, { call, put }) {
      const { id, tenantId, ...rest } = payload;
      const res = yield call(
        ObShellTenantController.tenantModifyPrimaryZone,
        {
          id,
          tenantId,
        },

        rest
      );

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.ZonePriorityModifiedSuccessfully',
            defaultMessage: 'Zone 优先级修改成功',
          })
        );

        if (onSuccess) {
          onSuccess();
        }
        yield put({
          type: 'getTenantData',
          payload: {
            id,
            tenantId,
          },
        });
      }
    },
    *modifyWhitelist ({ payload, onSuccess }, { call, put }) {
      const { id, tenantId, ...rest } = payload;
      const res = yield call(
        ObShellTenantController.tenantModifyWhitelist,
        {
          id,
          tenantId,
        },

        rest
      );

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.IpWhitelistModifiedSuccessfully',
            defaultMessage: 'IP 白名单修改成功',
          })
        );

        if (onSuccess) {
          onSuccess();
        }
        yield put({
          type: 'getTenantData',
          payload: {
            id,
            tenantId,
          },
        });
      }
    },
    *modifyReplica ({ payload, onSuccess }, { call, put }) {
      const { id, tenantId, body } = payload;
      const res = yield call(
        ObShellTenantController.tenantAddReplicas,
        {
          tenantId,
        },

        body
      );

      if (res.successful) {
        const taskId = res.data && res.data.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'ocp-express.src.model.tenant.TheTaskOfModifyingThe',
            defaultMessage: '修改副本的任务提交成功',
          }),
        });

        if (onSuccess) {
          onSuccess();
        }
        yield put({
          type: 'getTenantData',
          payload: {
            id,
            tenantId,
          },
        });

        yield put({
          type: 'task/update',
          payload: {
            runningTaskListDataRefreshDep: taskId,
          },
        });
      }
      return res;
    },
    // TODO：缺少描述
    *modifyTenantDescription ({ payload, onSuccess }, { call }) {
      const { id, tenantId, ...rest } = payload;
      const res = yield call(
        ObShellTenantController?.modifyTenantDescription,
        {
          id,
          tenantId,
        },

        rest
      );

      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.RemarksModified',
            defaultMessage: '备注修改成功',
          })
        );

        if (onSuccess) {
          onSuccess();
        }
      }
    },
  },

  reducers: {
    update (state, { payload }) {
      return { ...state, ...payload };
    },
  },
};

export default model;
