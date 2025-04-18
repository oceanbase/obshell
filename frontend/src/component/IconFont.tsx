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
import { createFromIconfontCN } from '@oceanbase/icons';

interface IconFontProps {
  type: string;
  className?: string;
  style?: React.CSSProperties;
}

const CustomIcon = createFromIconfontCN({
  // OCP Express 需要引入本地资源，保证离线环境下自定义图标的可访问性
  scriptUrl: '/js/iconfont.js',
});

export type IconFontType =
  | 'backup'
  | 'backup-colored'
  | 'cluster'
  | 'cluster-colored'
  | 'data-source'
  | 'data-source-colored'
  | 'docs'
  | 'diagnosis'
  | 'diagnosis-colored'
  | 'host'
  | 'host-colored'
  | 'log'
  | 'log-colored'
  | 'migration'
  | 'migration-colored'
  | 'monitor'
  | 'monitor-colored'
  | 'notification'
  | 'obproxy'
  | 'obproxy-colored'
  | 'package'
  | 'package-colored'
  | 'overview'
  | 'overview-colored'
  | 'property'
  | 'property-colored'
  | 'tenant'
  | 'tenant-colored'
  | 'sync'
  | 'sync-colored'
  | 'system'
  | 'system-colored'
  | 'user';

const IconFont = (props: IconFontProps) => {
  const { type, className, ...restProps } = props;
  return <CustomIcon type={type} className={className} {...restProps} />;
};

export default IconFont;
