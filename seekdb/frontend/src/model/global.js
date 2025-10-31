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

import { getObInfo } from '@/service/obshell/observer';

export const namespace = 'global';

const model = {
  namespace,
  state: {
    // theme mode
    // 设置一个默认的主题，默认主题为 light
    themeMode: localStorage.getItem('themeMode') || 'light',
    // RSA 加密用的公钥
    publicKey: '',
    obInfoData: {},
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
    // 获取obInfo信息
    *getObInfoData(_, { call, put }) {
      const res = yield call(getObInfo);
      yield put({
        type: 'update',
        payload: {
          obInfoData: res?.data || {},
        },
      });
    },
  },
  reducers: {
    update(state, { payload }) {
      return { ...state, ...payload };
    },
  },
};

export default model;
