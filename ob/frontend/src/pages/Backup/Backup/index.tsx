import ContentWithReload from '@/component/ContentWithReload';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { useQueryParams } from '@/hook/useQueryParams';
import useReload from '@/hook/useReload';
import { formatMessage } from '@/util/intl';
import { PageContainer } from '@oceanbase/ui';
import { history } from '@umijs/max';
import React from 'react';
import styles from './index.less';
import Task from './Task';

const Backup: React.FC = () => {
  useDocumentTitle(formatMessage({ id: 'ocp-v2.Backup.Backup.Backup', defaultMessage: '备份' }));

  const { tab, subTab, status } = useQueryParams({
    tab: 'backup',
    subTab: 'data',
    status: '',
  });
  const [loading, reload] = useReload(false);

  return (
    <PageContainer
      className={styles.container}
      loading={loading}
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            spin={loading}
            content={formatMessage({ id: 'ocp-v2.Backup.Backup.Backup', defaultMessage: '备份' })}
            onClick={() => {
              reload();
            }}
          />
        ),
      }}
    >
      <Task
        taskTab={tab}
        taskSubTab={subTab}
        status={status}
        showRestoreTask={false}
        onTabChange={(key: string) => {
          const params = new URLSearchParams({
            tab: key,
            subTab: 'data',
          });
          history.push(`/backup?${params.toString()}`);
        }}
        logOperationsFilter={opts => {
          // 全局视图日志备份不允许操作，只能查看错误信息
          return opts.filter(opt => opt.value === 'viewErrorMessage');
        }}
        onSubTabChange={(key: string) => {
          const params = new URLSearchParams({
            subTab: key,
          });
          history.push(`/backup?${params.toString()}`);
        }}
      />
    </PageContainer>
  );
};

export default Backup;
