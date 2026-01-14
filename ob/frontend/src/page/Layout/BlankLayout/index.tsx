import React from 'react';
import { Layout } from '@oceanbase/design';
import styles from './index.less';

const BlankLayout: React.FC = ({ children, ...restProps }) => {
  return (
    <div className={styles.main} {...restProps}>
      <Layout className={styles.layout}>{children as React.ReactNode}</Layout>
    </div>
  );
};

export default BlankLayout;
