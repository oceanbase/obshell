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
import { connect } from 'umi';
import React, { useEffect } from 'react';
import { Select } from '@oceanbase/design';
import type { SelectProps } from '@oceanbase/design/es/select';
import { isNullValue } from '@oceanbase/util';
import { flatten } from 'lodash';
import { DEFAULT_LIST_DATA } from '@/constant';

export interface ServerSelectProps extends SelectProps<number | string> {
  dispatch?: any;
  loading: boolean;
  clusterId?: number;
  zoneName?: string;
  valueProp?: 'id' | 'ip';
  showAllOption?: boolean;
  // 上层组件传入的 server 列表
  realServerList?: API.Server[];
  // 根据集群 ID 获取的 server 列表
  serverList: API.Server[];
  clusterList: API.ClusterInfo[];
  allLabel?: string;
}

const { Option } = Select;

const ServerSelect: React.FC<ServerSelectProps> = ({
  dispatch,
  loading,
  clusterId,
  zoneName,
  // 用作 value 的属性字段，默认为 id
  valueProp = 'id',
  showAllOption = false,
  realServerList,
  serverList,
  clusterList,
  allLabel = formatMessage({
    id: 'ocp-express.component.common.ServerSelect.All',
    defaultMessage: '全部',
  }),
  ...restProps
}) => {
  useEffect(() => {
    // clusterId 为空，则 server 列表的获取依赖于 clusterList，因此需要额外请求集群列表，重置掉 serverList
    if (isNullValue(clusterId) && !realServerList) {
      dispatch({
        type: 'cluster/getClusterListData',
        payload: {},
      });
      resetServerList();
    } else {
      getServerList();
    }
  }, [clusterId]);

  function resetServerList() {
    dispatch({
      type: 'cluster/update',
      payload: {
        serverListData: DEFAULT_LIST_DATA,
      },
    });
  }

  function getServerList() {
    dispatch({
      type: 'cluster/getServerListData',
      payload: {
        id: clusterId,
      },
    });
  }

  // 外部传入的 server 列表优先级高于内部请求的 server 列表
  let myRealServerList = [];
  if (isNullValue(clusterId) && !realServerList) {
    const realClusterList = flatten(clusterList.map(item => [item]));
    myRealServerList = realClusterList
      ?.filter(item => !isNullValue(item?.zones))
      ?.flatMap(item => item?.zones)
      ?.filter(item => !isNullValue(item?.servers))
      ?.flatMap(item => item?.servers);
  } else if (isNullValue(clusterId) && realServerList) {
    // 没有传递集群 id，但是外部传入了 serverList ,使用外部的 serverList
    myRealServerList = realServerList;
  } else {
    myRealServerList = serverList;
  }

  return (
    <Select
      loading={loading}
      showSearch={true}
      optionFilterProp="children"
      placeholder={formatMessage({
        id: 'ocp-express.component.common.ServerSelect.SelectAServer',
        defaultMessage: '请选择服务器',
      })}
      {...restProps}
    >
      {showAllOption && <Option value="all">{allLabel}</Option>}

      {myRealServerList
        .filter(item => !zoneName || item.zoneName === zoneName)
        .map(item => (
          <Option key={item.id} value={item[valueProp]}>
            {item[valueProp]}
          </Option>
        ))}
    </Select>
  );
};

function mapStateToProps({ loading, cluster }) {
  return {
    clusterList: (cluster.clusterListData && cluster.clusterListData.contents) || [],
    serverList: (cluster.serverListData && cluster.serverListData.contents) || [],
    loading: loading.effects['cluster/getServerListData'],
  };
}

export default connect(mapStateToProps)(ServerSelect);
