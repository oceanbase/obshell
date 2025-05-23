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

import React from 'react';

export interface ClusterLabelProps {
  cluster?: API.ClusterInfo | API.Cluster | API.BasicClusterInfo;
  tenant?: API.TenantInfo | API.Tenant;
}

const ClusterLabel: React.FC<ClusterLabelProps> = ({ tenant, cluster }) => {
  // 优先级: cluster > tenant
  const clusterName = cluster?.name || tenant?.clusterName;
  return <span>{clusterName}</span>;
};

export default ClusterLabel;
