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

import { obclusterInfo } from '@/service/obshell/obcluster';
import { useRequest } from 'ahooks';
import { flatten } from 'lodash';
import { useState } from 'react';

interface ClusterState {
  clusterListData: API.ClusterInfo[];
  tenantListData: API.TenantInfo[];
  zoneListData: API.Zone[];
  serverListData: (API.Observer & { zoneName?: string })[];
  clusterData: API.ClusterInfo;
  isSingleMachine: boolean;
  isStandalone: boolean;
  isCommunityEdition: boolean;
  isDistributedBusiness: boolean;
  isSharedStorage: boolean;
}

const clusterModel = () => {
  const [clusterState, setClusterState] = useState<ClusterState>({
    clusterListData: [],
    tenantListData: [],
    zoneListData: [],
    serverListData: [],
    clusterData: {} as API.ClusterInfo,
    isSingleMachine: false,
    isStandalone: false,
    isCommunityEdition: false,
    isDistributedBusiness: false,
    isSharedStorage: false, // 是否共享存储模式
  });

  const { run: getObclusterInfo, refresh, loading } = useRequest(obclusterInfo, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const clusterData = res?.data || ({} as API.ClusterInfo);
        // 是否单机版判断，如果clusterData里面zones只有一条数据，而且zones里面的servers只有一条数据，则认为单集群
        const isSingleMachine =
          clusterData.zones?.length === 1 && clusterData.zones[0].servers?.length === 1;

        // 计算是否为单机版
        const isStandalone = clusterData?.is_standalone || false;
        // 获取是否为社区版
        const isCommunityEdition = clusterData?.is_community_edition || false;
        // 获取是否是分布式商业版
        const isDistributedBusiness = !isStandalone && !isCommunityEdition;
        // 是否共享存储模式
        const isSharedStorage = clusterData?.is_shared_storage ?? false;

        const tenantListData = clusterData.tenants || [];
        const zoneListData = clusterData.zones || [];
        const serverListData: (API.Observer & { zoneName?: string })[] = flatten(
          zoneListData.map((item: API.Zone) =>
            (item.servers || []).map((server: API.Observer) => ({
              ...server,
              zoneName: item.name,
            }))
          )
        );

        setClusterState({
          clusterData,
          clusterListData: [clusterData],
          isSingleMachine,
          isStandalone,
          isCommunityEdition,
          isDistributedBusiness,
          isSharedStorage,
          tenantListData,
          zoneListData,
          serverListData,
        });
      }
    },
  });

  return {
    ...clusterState,
    getObclusterInfo,
    refreshClusterData: refresh,
    clusterDataLoading: loading,
  };
};

export default clusterModel;
