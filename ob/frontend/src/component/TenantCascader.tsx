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
import React from 'react';
import { Cascader } from '@oceanbase/design';
import type { CascaderProps, DefaultOptionType } from '@oceanbase/design/es/cascader';
import { some, uniq } from 'lodash';
import { useRequest } from 'ahooks';
import * as ObTenantController from '@/service/ocp-express/ObTenantController';
import ClusterLabel from '@/component/ClusterLabel';

export interface DataNodeType {
  value: number;
  label: React.ReactNode;
  filterLabel: string;
  children: DataNodeType[];
}

export type TenantCascaderProps = CascaderProps<DataNodeType> & {
  /* 租户模式 */
  mode?: 'ORACLE' | 'MYSQL';
  /* 自定义租户筛选 */
  filter?: (tenant?: API.Tenant) => boolean;
};

const TenantCascader: React.FC<TenantCascaderProps> = ({
  children,
  mode,
  filter,
  ...restProps
}) => {
  // 获取租户列表
  const { data, loading } = useRequest(ObTenantController.listTenants, {
    defaultParams: [{}],
  });
  const tenantList = (data?.data?.contents || [])
    .filter(item => !mode || item.mode === mode)
    .filter(item => !filter || filter(item));

  const clusterIdList = uniq(tenantList.map(item => item.clusterId));
  // 根据集群 ID 分组后的租户数据
  const options = clusterIdList
    .map(clusterId => {
      // 对应集群下的其中一个租户，用户获取集群相关信息
      const tenantOfClusterId = tenantList.find(item => item.clusterId === clusterId);
      // 对应集群下的租户列表
      const tenantListOfClusterId = tenantList.filter(item => item.clusterId === clusterId);

      return {
        value: tenantOfClusterId?.clusterId,
        label: (
          <ClusterLabel
            cluster={{
              name: tenantOfClusterId?.clusterName,
              status: tenantOfClusterId?.clusterStatus,
            }}
          />
        ),
        // 扩展出 filterLabel 属性用于筛选
        filterLabel: tenantOfClusterId?.clusterName,
        children: tenantListOfClusterId.map(item => ({
          value: item.id,
          label: item.name,
          filterLabel: item.name,
        })),
      };
    })
    .filter(item => {
      // 只显示集群下存在租户的选项
      return item.children.length > 0;
    });
  const displayRender = (labels: string[], selectedOptions: DefaultOptionType[]) => {
    return labels.map((label, i) => {
      const option = selectedOptions[i];
      if (i === labels.length - 1) {
        return <span key={option?.value}>{label}</span>;
      }

      return <span key={option?.value}>{label} / </span>;
    });
  };

  return (
    <Cascader
      loading={loading}
      expandTrigger="hover"
      showSearch={{
        filter: (inputValue, path) => {
          return some(path, item =>
            item.filterLabel?.toLowerCase().includes(inputValue?.toLowerCase())
          );
        },
      }}
      options={options}
      placeholder={formatMessage({
        id: 'ocp-express.src.component.TenantCascader.SelectATenant',
        defaultMessage: '请选择租户',
      })}
      displayRender={displayRender}
      {...restProps}
    >
      {children}
    </Cascader>
  );
};

export default TenantCascader;
