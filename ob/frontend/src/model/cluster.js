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
import { flatten } from 'lodash';

export const namespace = 'cluster';

const model = {
  namespace,
  state: {
    clusterListData: [],
    tenantListData: [],
    zoneListData: [],
    serverListData: [],
    clusterData: {},
    isSingleMachine: false,
  },

  effects: {
    *getClusterData({ onSuccess }, { call, put }) {
      const res = yield call(obclusterInfo, {
        HIDE_ERROR_MESSAGE: true,
      });
      if (res.successful) {
        const clusterData = res?.data || {};
        // 是否单机版判断，如果clusterData里面zones只有一条数据，而且zones里面的servers只有一条数据，则认为单集群
        const isSingleMachine =
          clusterData.zones.length === 1 && clusterData.zones[0].servers.length === 1;
        const tenantListData = clusterData.tenants;
        const zoneListData = clusterData.zones;
        const serverListData = flatten(
          zoneListData.map(item =>
            item.servers.map(server => ({
              ...server,
              zoneName: item.name,
            }))
          )
        );

        yield put({
          type: 'update',
          payload: {
            clusterData,
            clusterListData: [clusterData],
            isSingleMachine,
            tenantListData,
            zoneListData,
            serverListData,
          },
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
