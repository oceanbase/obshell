import { formatMessage } from '@/util/intl';
import { Button, Drawer, DrawerProps, Space } from '@oceanbase/design';
import { Outlet } from '@umijs/max';
import React from 'react';

import styles from './index.less';

type AlertRuleDrawerProps = {
  onSubmit: () => void;
} & DrawerProps;
export default function AlertDrawer({ onClose, onSubmit, footer, ...props }: AlertRuleDrawerProps) {
  return (
    <Drawer
      style={{ paddingBottom: 60 }}
      closeIcon={false}
      onClose={onClose}
      maskClosable={false}
      {...props}
    >
      <Outlet />
      <div className={styles.drawerFooter}>
        {footer ? (
          footer
        ) : (
          <Space>
            <Button onClick={onSubmit} type="primary">
              {formatMessage({
                id: 'OBShell.component.AlertDrawer.Submission',
                defaultMessage: '提交',
              })}
            </Button>
            <Button onClick={onClose}>
              {formatMessage({
                id: 'OBShell.component.AlertDrawer.Cancel',
                defaultMessage: '取消',
              })}
            </Button>
          </Space>
        )}
      </div>
    </Drawer>
  );
}
