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
import { Tag } from '@oceanbase/design';
import useStyles from './index.style';

interface ObClusterDeployModeProps {
  clusterData: API.ClusterInfo;
  // 集群部署模式的展示样式: 标签 | 文本
  mode?: 'tag' | 'text';
  className?: string;
}

const ObClusterDeployMode: React.FC<ObClusterDeployModeProps> = ({
  clusterData,
  mode = 'tag',
  className,
}: ObClusterDeployModeProps) => {
  const { styles } = useStyles();
  const deployMode = (clusterData.zones || []).map((item) => ({
    regionName: item.regionName,
    serverCount: (item.servers || []).length,
  }));
  return (
    <span className={`${styles.container} ${className}`}>
      {mode === 'tag'
        ? deployMode.map((item) => (
            <Tag color="blue">{`${item.regionName} ${item.serverCount}`}</Tag>
          ))
        : deployMode.map((item) => item.serverCount).join('-')}
    </span>
  );
};

export default ObClusterDeployMode;
