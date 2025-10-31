import { formatMessage } from '@/util/intl';
import { PageContainer } from '@oceanbase/ui';
import { history, useLocation } from 'umi';
import { useRequest } from 'ahooks';
import type { TabsProps } from '@oceanbase/design';
import { Divider, Tabs } from '@oceanbase/design';
import React from 'react';
import styles from './index.less';
import useReload from '@/hook/useReload';
import ContentWithReload from '@/component/ContentWithReload';
import { getSystemExternalAlertmanager } from '@/service/obshell/system';
import AlertConfig from './AlertConfig';

const TAB_KEYS = ['event', 'shield', 'rules', 'channel', 'subscriptions'];

const TAB_ITEMS: TabsProps['items'] = [
  {
    key: 'event',
    label: formatMessage({ id: 'OBShell.page.Alert.AlarmEvent', defaultMessage: '告警事件' }),
  },
  {
    key: 'divider-1',
    label: <Divider type="vertical" />,
  },
  {
    key: 'shield',
    label: formatMessage({ id: 'OBShell.page.Alert.AlarmShielding', defaultMessage: '告警屏蔽' }),
  },
  {
    key: 'divider-2',
    label: <Divider type="vertical" />,
  },
  {
    key: 'rules',
    label: formatMessage({ id: 'OBShell.page.Alert.AlarmRules', defaultMessage: '告警规则' }),
  },
  // {
  //   key: 'divider-3',
  //   label: <Divider type="vertical" />,
  // },
  // {
  //   key: 'channel',
  //   label: '告警通道',
  // },
  // {
  //   key: 'divider-4',
  //   label: <Divider type="vertical" />,
  // },
  // {
  //   key: 'subscriptions',
  //   label: '告警推送',
  // },
];

export default function Alert({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const searchParams = location.pathname.split('/');
  const currentKey = searchParams[searchParams.length - 1] || 'event';

  const [reloading, reload] = useReload(false);

  const {
    data: alertmanagerData,
    loading: alertmanagerLoading,
    refresh: refreshAlertmanagerData,
  } = useRequest(getSystemExternalAlertmanager, {
    defaultParams: [{ HIDE_ERROR_MESSAGE: true }],
  });

  // 判断Alertmanager是否已配置：成功响应且有数据表示已配置
  const isAlertmanagerConfigured =
    alertmanagerData?.successful && alertmanagerData.status === 200 && !!alertmanagerData?.data;

  const onChange = (key: string) => {
    history.push(`/alert/${key}`);
  };

  return (
    <PageContainer
      loading={reloading || alertmanagerLoading}
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            spin={reloading || alertmanagerLoading}
            content={formatMessage({
              id: 'OBShell.page.Alert.AlarmCenter',
              defaultMessage: '告警中心',
            })}
            onClick={reload}
          />
        ),
      }}
    >
      {!isAlertmanagerConfigured && !alertmanagerLoading ? (
        <AlertConfig
          onConfigSuccess={() => {
            // 配置成功后刷新 Alertmanager 状态
            refreshAlertmanagerData();
          }}
        />
      ) : (
        <>
          <Tabs
            activeKey={currentKey}
            className={styles.tabContent}
            items={TAB_ITEMS}
            onChange={onChange}
          />
          {children}
        </>
      )}
    </PageContainer>
  );
}
