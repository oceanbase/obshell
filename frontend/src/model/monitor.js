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

import { listAllMetrics } from '@/service/obshell/metric';
import { DEFAULT_LIST_DATA } from '@/constant';

const model = {
  namespace: 'monitor',
  state: {
    metricGroupListData: DEFAULT_LIST_DATA,
  },
  effects: {
    *getMetricGroupListData({ payload }, { call, put }) {
      const res = yield call(listAllMetrics, payload);
      if (res.successful) {
        yield put({
          type: 'update',
          payload: {
            metricGroupListData: res.data || [],
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
