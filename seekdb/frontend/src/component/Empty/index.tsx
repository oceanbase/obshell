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
import { Empty } from '@oceanbase/design';
import type { EmptyProps as AntdEmptyProps } from '@oceanbase/design/es/empty';
import { PageContainer } from '@oceanbase/ui';
import PageCard from '@/component/PageCard';
import styles from './index.less';

export interface EmptyProps extends AntdEmptyProps {
  image?: React.ReactNode;
  title?: React.ReactNode;
  description?: React.ReactNode;
  children?: React.ReactNode;
  // 展示模式: 页面模式 | 组件模式
  mode?: 'page' | 'pageCard' | 'component';
  size?: 'default' | 'small';
}

export default ({
  image = '/assets/common/empty.svg',
  title,
  description,
  children,
  // 默认为页面模式
  mode = 'page',
  size = 'default',
  className,
  style,
  ...restProps
}: EmptyProps) => {
  const empty = (
    <Empty
      className={`${styles.empty} ${size === 'small' ? styles.small : ''}`}
      image={image}
      description={
        <div>
          {title && <h2 className={styles.title}>{title}</h2>}
          <span className={styles.description}>{description}</span>
        </div>
      }
      {...restProps}
    >
      {children}
    </Empty>
  );

  const pageCard = (
    <PageCard
      className={`${mode === 'page' ? styles.page : styles.component} ${className}`}
      style={style}
    >
      {empty}
    </PageCard>
  );

  if (mode === 'page') {
    return <PageContainer>{pageCard}</PageContainer>;
  }
  if (mode === 'pageCard') {
    return pageCard;
  }
  if (mode === 'component') {
    return empty;
  }
  return <PageContainer>{pageCard}</PageContainer>;
};
