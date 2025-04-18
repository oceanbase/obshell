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

import React, { useEffect } from 'react';
import { connect } from 'umi';
import { Dropdown, Menu } from '@oceanbase/design';
import type { DropDownProps } from '@oceanbase/design/es/dropdown';
import { isNullValue, findBy } from '@oceanbase/util';
import { DEFAULT_LIST_DATA } from '@/constant';

export interface TenantDropdownProps extends DropDownProps {
  dispatch: any;
  clusterId?: number;
  cluster?: string;
  valueProp?: 'id' | 'name';
  onMenuClick?: (key: string) => void;
  children: React.ReactNode;
  tenantList: API.TenantInfo[];
  clusterList: API.ClusterInfo[];
}

const TenantDropdown: React.FC<TenantDropdownProps> = ({
  dispatch,
  clusterId,
  // 集群名
  cluster,
  // 用作 value 的属性字段，默认为 id
  valueProp = 'id',
  onMenuClick,
  children,
  tenantList,
  clusterList,
  ...restProps
}) => {
  const realClusterId = isNullValue(clusterId)
    ? findBy(clusterList, 'name', cluster).id
    : clusterId;
  useEffect(() => {
    if (isNullValue(realClusterId)) {
      resetTenantList();
    } else {
      getTenantList();
    }
  }, [realClusterId]);

  function resetTenantList() {
    dispatch({
      type: 'tenant/update',
      payload: {
        tenantListData: DEFAULT_LIST_DATA,
      },
    });
  }

  function getTenantList() {
    dispatch({
      type: 'tenant/getTenantListData',
      payload: {
        id: realClusterId,
      },
    });
  }
  const menu = (
    <Menu
      onClick={({ key }) => {
        if (onMenuClick) {
          onMenuClick(key);
        }
      }}
    >
      {tenantList.map(item => (
        <Menu.Item key={item[valueProp]}>{item.name}</Menu.Item>
      ))}
    </Menu>
  );
  return (
    <Dropdown overlay={menu} {...restProps}>
      {children}
    </Dropdown>
  );
};

function mapStateToProps({ cluster, tenant }) {
  return {
    clusterList: (cluster.clusterListData && cluster.clusterListData.contents) || [],
    tenantList: (tenant.tenantListData && tenant.tenantListData.contents) || [],
  };
}

export default connect(mapStateToProps)(TenantDropdown);
