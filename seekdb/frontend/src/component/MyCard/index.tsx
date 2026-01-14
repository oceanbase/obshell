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
import type { CardProps } from '@oceanbase/design/es/card';
import styles from './index.less';
export interface MyCardProps extends CardProps {
  children: React.ReactNode;
  title?: React.ReactNode;
  extra?: React.ReactNode;
  loading?: boolean;
  className?: string;
  headStyle?: React.CSSProperties;
}

const MyCard = ({
  children,
  title,
  extra,
  loading = false,
  className,
  headStyle,
  bodyStyle,
  ...restProps
}: MyCardProps) => {
  return (
    <Card
      className={`${className} ${styles.card}`}
      bordered={false}
      bodyStyle={{ padding: '20px 24px', ...bodyStyle }}
      {...restProps}
    >
      {(title || extra) && (
        <div className={styles.header} style={headStyle}>
          {title && <span className={styles.title}>{title}</span>}
          {extra && <span className={styles.extra}>{extra}</span>}
        </div>
      )}

      <div style={{ width: '100%' }}>
        <Spin spinning={loading}>{children}</Spin>
      </div>
    </Card>
  );
};

export default MyCard;
