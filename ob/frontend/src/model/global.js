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

import * as InfoController from '@/service/ocp-express/InfoController';
import * as IamController from '@/service/ocp-express/IamController';
import * as PropertyController from '@/service/ocp-express/PropertyController';
import { obclusterInfo } from '@/service/obshell/obcluster';

export const namespace = 'global';

const model = {
  namespace,
  state: {
    // theme mode
    // 设置一个默认的主题，默认主题为 light
    themeMode: localStorage.getItem('themeMode') || 'light',
    // RSA 加密用的公钥
    publicKey: '',
    uiMode: '',
    // 应用信息
    appInfo: {},
    // 系统配置
    systemInfo: {},
    showTenantAdminPasswordModal: false,
    showCredentialModal: false,
    tenantAdminPasswordErrorData: {},
  },
  effects: {
    *setThemeMode({ payload }, { put }) {
      const themeMode = payload.themeMode || 'light';
      localStorage.setItem('themeMode', themeMode);
      yield put({
        type: 'update',
        payload: {
          themeMode,
        },
      });
    },
    *setUiMode({ payload }, { put }) {
      yield put({
        type: 'update',
        payload: { uiMode: payload.uiMode },
      });
    },
    *getPublicKey(_, { call, put }) {
      // const res = yield call(IamController.getLoginKey);
      // yield put({
      //   type: 'update',
      //   payload: {
      //     publicKey: res.data?.publicKey || '',
      //   },
      // });
    },
    // 获取应用信息
    *getAppInfo(_, { call, put }) {
      // const res = yield call(InfoController.info);
      // yield put({
      //   type: 'update',
      //   payload: {
      //     // /api/v1/info 接口的 res 并没有 data 字段，res 本身就包含实际数据
      //     appInfo: res || {},
      //   },
      // });
    },
    // 获取系统配置
    *getSystemInfo(_, { call, put }) {
      // const res = yield call(PropertyController.getSystemInfo);
      // yield put({
      //   type: 'update',
      //   payload: {
      //     systemInfo: res.data || {},
      //   },
      // });
    },
  },
  reducers: {
    update(state, { payload }) {
      return { ...state, ...payload };
    },
  },
};

export default model;
