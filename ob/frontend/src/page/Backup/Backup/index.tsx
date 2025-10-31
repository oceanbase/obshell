import ContentWithReload from '@/component/ContentWithReload';
import useDocumentTitle from '@/hook/useDocumentTitle';
import useReload from '@/hook/useReload';
import { formatMessage } from '@/util/intl';
import { history } from 'umi';
import React from 'react';
import { PageContainer } from '@oceanbase/ui';
import styles from './index.less';
import Task from './Task';
export interface BackupProps {
  location: {
    pathname: string;
    query: {
      tab?: string;
      subTab?: string;
      status?: string;
    };
  };
}

const Backup: React.FC<BackupProps> = ({ location: { pathname, query } }) => {
  useDocumentTitle(formatMessage({ id: 'ocp-v2.Backup.Backup.Backup', defaultMessage: '备份' }));

  const { tab = 'backup', subTab = 'data', status } = query;
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
          history.push({
            pathname,
            query: {
              ...query,
              tab: key,
              subTab: 'data',
            },
          });
        }}
        logOperationsFilter={opts => {
          // 全局视图日志备份不允许操作，只能查看错误信息
          return opts.filter(opt => opt.value === 'viewErrorMessage');
        }}
        onSubTabChange={(key: string) => {
          history.push({
            pathname,
            query: {
              ...query,
              subTab: key,
            },
          });
        }}
      />
    </PageContainer>
  );
};

export default Backup;
