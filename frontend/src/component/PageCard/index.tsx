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
import { Card, Spin } from '@oceanbase/design';
import useStyles from './index.style';

export interface TaskSuccessProps {
  children: React.ReactNode;
  loading?: boolean;
  style?: React.CSSProperties;
  className?: string;
}

const PageCard: React.FC<TaskSuccessProps> = ({
  children,
  loading = false,
  className,
  ...restProps
}) => {
  const { styles } = useStyles();
  return (
    <Card bordered={false} className={`${styles.card} ${className}`} {...restProps}>
      <Spin spinning={loading}>{children}</Spin>
    </Card>
  );
};

export default PageCard;
