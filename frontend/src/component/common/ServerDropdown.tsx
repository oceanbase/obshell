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
import React, { useEffect, useState } from 'react';
import { connect } from 'umi';
import { Dropdown, Menu } from '@oceanbase/design';
import { isNullValue } from '@oceanbase/util';
import { DEFAULT_LIST_DATA, ALL } from '@/constant';
import { DownOutlined } from '@oceanbase/icons';
import { noop } from 'lodash';

export interface ServerDropdownProps {
  dispatch?: any;
  clusterId?: number;
  // 上层组件传入的 server 列表
  realServerList?: API.Server[];
  // 根据集群 ID 获取的 server 列表
  serverList: API.Server[];
  onChange?: (server: API.Server | null) => void;
  value?: { id: string };
  allLabel?: string;
}

const ServerDropdown: React.FC<ServerDropdownProps> = ({
  dispatch,
  clusterId,
  realServerList,
  serverList,
  onChange = noop,
  value = {},
  allLabel = formatMessage({
    id: 'ocp-express.component.common.ServerDropdown.All',
    defaultMessage: '全部',
  }),
}) => {
  const [serverId, setServerId] = useState<string>(ALL);

  useEffect(() => {
    if (!realServerList) {
      if (isNullValue(clusterId)) {
        resetServerList();
      } else {
        getServerList();
      }
    }
  }, [clusterId]);

  useEffect(() => {
    if (value?.id) {
      setServerId(`${value.id}`);
    }
  }, [value]);

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

  const serverChange = (server: API.Server | null) => {
    setServerId(server ? `${server.id}` : ALL);
    onChange(server);
  };

  const myRealServerList = realServerList || serverList;

  return (
    <Dropdown
      overlayStyle={{ marginRight: 24 }}
      overlay={
        <Menu selectedKeys={[serverId]}>
          <Menu.Item key={ALL} onClick={() => serverChange(null)}>
            {allLabel}
          </Menu.Item>
          {myRealServerList.map(server => (
            <Menu.Item key={`${server.id}`} onClick={() => serverChange(server)}>
              {server.ip}
            </Menu.Item>
          ))}
        </Menu>
      }
    >
      <span style={{ cursor: 'pointer' }}>
        {serverId === ALL
          ? allLabel
          : myRealServerList.find(server => `${server.id}` === serverId)?.ip || '-'}
        <DownOutlined style={{ marginLeft: 8 }} />
      </span>
    </Dropdown>
  );
};

function mapStateToProps({ cluster }) {
  return {
    serverList: (cluster.serverListData && cluster.serverListData.contents) || [],
  };
}

export default connect(mapStateToProps)(ServerDropdown);
