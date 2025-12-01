import PageCard from '@/component/PageCard';
import React from 'react';
import { Empty } from '@oceanbase/design';
import type { EmptyProps as ObEmptyProps } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import styles from './index.less';

export interface EmptyProps extends ObEmptyProps {
  children?: React.ReactNode;
  // 展示模式: 页面模式 | 组件模式
  mode?: 'page' | 'pageCard' | 'component';
  size?: 'default' | 'small';
}

export default ({
  children,
  // 默认为页面模式
  mode = 'page',
  size = 'default',
  // 对于 page 和 pageCard 模式，使用彩色图片
  image = mode === 'page' || mode === 'pageCard' ? Empty.PRESENTED_IMAGE_COLORED : undefined,
  className,
  style,
  ...restProps
}: EmptyProps) => {
  const empty = (
    <Empty
      className={`${styles.empty} ${size === 'small' ? styles.small : ''} ${className}`}
      image={image}
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
