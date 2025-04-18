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
import { history } from 'umi';
import { message } from '@oceanbase/design';
import * as ObClusterController from '@/service/ocp-express/ObClusterController';
import { obclusterInfo } from '@/service/obshell/obcluster';
import { DEFAULT_LIST_DATA } from '@/constant';
import { taskSuccess } from '@/util/task';

export const namespace = 'cluster';

const model = {
  namespace,
  state: {
    clusterListData: DEFAULT_LIST_DATA,
    clusterData: {},
    zoneListData: DEFAULT_LIST_DATA,
    serverListData: DEFAULT_LIST_DATA,
    compactionLatestData: {},
    compactionSettings: {},
    parameterListData: DEFAULT_LIST_DATA,
    importPreCheckResult: {},
    importResult: {},
    clusterConfig: {},
    startupParameterList: [],
  },

  effects: {
    *getClusterData({ onSuccess }, { call, put }) {
      const res = yield call(obclusterInfo, {});
      if (res.successful) {
        const clusterData = res?.data || {};

        yield put({
          type: 'update',
          payload: { clusterData },
        });

        // 成功的回调函数需要在 update 之后执行，以在组件中拿到 clusterData 的最新值
        if (onSuccess) {
          onSuccess(res.data || {});
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
