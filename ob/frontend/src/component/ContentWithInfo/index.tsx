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
import type { SpaceProps } from '@oceanbase/design';
import { Space } from '@oceanbase/design';
import { InfoCircleFilled } from '@oceanbase/icons';
import styles from './index.less';

export interface ContentWithInfoProps extends SpaceProps {
  content: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

const ContentWithInfo: React.FC<ContentWithInfoProps> = ({ content, className, ...restProps }) => {
  return (
    <Space className={`${styles.container} ${className}`} {...restProps}>
      <InfoCircleFilled className={styles.icon} />
      <span className={styles.content}>{content}</span>
    </Space>
  );
};

export default ContentWithInfo;
