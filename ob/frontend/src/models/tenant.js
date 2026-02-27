import * as ObShellTenantController from '@/service/obshell/tenant';

export const namespace = 'tenant';

const model = {
  namespace,
  state: {
    tenantData: {},
    precheckResult: {},
  },

  effects: {
    // 根据集群下的全部租户
    *getTenantPreCheck({ payload }, { call, put }) {
      const res = yield call(ObShellTenantController.tenantPreCheck, payload);

      if (res.successful) {
        yield put({
          type: 'update',
          payload: {
            precheckResult: {
              ...(res.data || {}),
              tenantName: payload?.name,
            },
          },
        });
      }
    },
    *getTenantData({ payload, onSuccess }, { call, put }) {
      const res = yield call(ObShellTenantController.getTenantInfo, payload);
      if (res.successful) {
        const resTenantData = {
          ...(res?.data || {}),
          tenantName: res?.data?.tenant_name || '',
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
  },

  reducers: {
    update(state, { payload }) {
      return { ...state, ...payload };
    },
  },
};

export default model;
