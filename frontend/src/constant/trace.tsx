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

import { HddFilled } from '@oceanbase/icons';

// 链路节点类型列表
export const NODE_TYPE_LIST = [
  {
    value: 'CLIENT',
    label: 'Client',
    icon: <img src="/assets/icon/client.svg" width={14} height={14} />,
  },
  {
    value: 'OBPROXY',
    label: 'OBProxy',
    // 由于 DeploymentUnitOutlined 不存在对应的实底 icon，因此使用 SVG 图片代替
    icon: <img src="/assets/icon/obproxy.svg" width={14} height={14} />,
  },
  {
    value: 'OBSERVER',
    label: 'OBServer',
    icon: (
      <HddFilled
        style={{
          fontSize: 14,
          color: '#B37FEB',
        }}
      />
    ),
  },
];
