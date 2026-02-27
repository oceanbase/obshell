import ContentWithReload from '@/component/ContentWithReload';
import Empty from '@/component/Empty';
import MyCard from '@/component/MyCard';
import { useQueryParams } from '@/hook/useQueryParams';
import useReload from '@/hook/useReload';
import * as ObTenantBackupController from '@/service/obshell/backup';
import { formatTime } from '@/util/datetime';
import { formatMessage } from '@/util/intl';
import { Button, Descriptions, Space, Tooltip } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { isNullValue } from '@oceanbase/util';
import { history, useParams, useSelector } from '@umijs/max';
import { useRequest } from 'ahooks';
import React, { useEffect } from 'react';
import Overview from './Overview';
import styles from './index.less';

const Backup: React.FC = () => {
  const { tenantName = '' } = useParams<{ tenantName: string }>();
  const { tab, taskTab, taskSubTab } = useQueryParams({
    tab: 'detail',
    taskTab: '',
    taskSubTab: '',
  });
  const [loading, reload] = useReload(false);

  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);

  const { data: tenantBackupConfigData, run } = useRequest(
    ObTenantBackupController.getTenantBackupConfig,
    {
      manual: true,
    }
  );

  const tenantBackupConfigInfo = tenantBackupConfigData?.data || {};

  useEffect(() => {
    if (tenantName) {
      run({
        name: tenantName,
      });
    }
  }, [tenantName]);

  // 获取租户的备份详情
  const {
    data: backupData,
    loading: getBackupOverviewLoading,
    refresh: getBackupOverviewRefresh,
  } = useRequest(ObTenantBackupController.getTenantBackupInfo, {
    ready: !!tenantName,
    defaultParams: [
      {
        name: tenantName,
      },
    ],
  });

  const backupTenant = backupData?.data || {};

  const backupNowTenantButton = (
    <Tooltip
      placement="topRight"
      title={
        (backupTenant.obClusterType === 'STANDBY' &&
          formatMessage({
            id: 'ocp-v2.Detail.Backup.BackupIsNotSupportedBy',
            defaultMessage: '备集群下的租户不支持备份',
          })) ||
        (tenantData?.tenant_name === 'sys' &&
          formatMessage({
            id: 'ocp-v2.Detail.Backup.SysTenantsDoNotSupport',
            defaultMessage: 'sys 租户不支持备份',
          }))
      }
    >
      <Button
        disabled={backupTenant.obClusterType === 'STANDBY' || tenantData?.tenant_name === 'sys'}
        onClick={() => {
          history.push(`/cluster/tenant/${tenantName}/backup/nowBackupTenant`);
        }}
      >
        {formatMessage({ id: 'ocp-v2.Detail.Backup.BackUpNow', defaultMessage: '立即备份' })}
      </Button>
    </Tooltip>
  );

  // 无备份策略
  const noBackupStrategy =
    !getBackupOverviewLoading && backupTenant.backupStatus === 'NO_BACKUP_STRATEGY';
  // 有备份策略
  const hasBackupStrategy = true;

  return (
    <>
      {/* 无策略且无备份数据展示空白页 */}
      {noBackupStrategy && !backupTenant.lastDataBackupTime ? (
        <Empty
          title={formatMessage({
            id: 'ocp-v2.Detail.Backup.NoBackupPolicyIsAvailable',
            defaultMessage: '暂无备份策略',
          })}
          description={formatMessage({
            id: 'OBShell.Detail.Backup.TheBackupPolicySupportsPerforming',
            defaultMessage:
              '备份策略支持在指定的周期定时执行全量备份，同步发起日志备份以及自动清理过期的备份文件。',
          })}
        >
          <Space>{backupNowTenantButton}</Space>
        </Empty>
      ) : (
        <PageContainer
          className={styles.container}
          loading={loading || getBackupOverviewLoading}
          ghost={true}
          header={{
            title: (
              <ContentWithReload
                spin={getBackupOverviewLoading}
                content={formatMessage({
                  id: 'ocp-v2.Detail.Backup.BackupAndRestoration',
                  defaultMessage: '备份恢复',
                })}
                onClick={() => {
                  reload();
                  getBackupOverviewRefresh();
                }}
              />
            ),

            extra: (
              // 无租户级备份策略
              <Space>
                {/* 有集群级备份策略 */}
                {hasBackupStrategy && (
                  <Tooltip
                    placement="topRight"
                    title={
                      tenantName === 'sys' &&
                      formatMessage({
                        id: 'ocp-v2.Detail.Backup.TheSysTenantDoesNot',
                        defaultMessage: 'sys 租户不支持恢复',
                      })
                    }
                  >
                    <Button
                      disabled={
                        tenantName === 'sys' ||
                        isNullValue(tenantBackupConfigInfo?.archive_base_uri) ||
                        isNullValue(tenantBackupConfigInfo?.data_base_uri)
                      }
                      onClick={() => {
                        history.push(`/cluster/tenant/${tenantName}/backup/restoreNow`);
                      }}
                    >
                      {/* 需要有备份策略才能发起恢复 */}
                      {formatMessage({
                        id: 'ocp-v2.Detail.Backup.InitiateRecovery',
                        defaultMessage: '发起恢复',
                      })}
                    </Button>
                  </Tooltip>
                )}
                {backupNowTenantButton}
              </Space>
            ),
          }}
        >
          {tab === 'detail' && (
            <>
              <MyCard
                title={formatMessage({
                  id: 'ocp-v2.Detail.Backup.BasicInformation',
                  defaultMessage: '基本信息',
                })}
                style={{ marginBottom: 16 }}
              >
                <Descriptions>
                  <Descriptions.Item
                    label={formatMessage({
                      id: 'ocp-v2.Detail.Backup.LastDataBackup',
                      defaultMessage: '最近一次数据备份',
                    })}
                  >
                    {formatTime(backupTenant.lastest_data_backup_time)}
                  </Descriptions.Item>
                  <Descriptions.Item
                    label={formatMessage({
                      id: 'ocp-v2.Detail.Backup.LogBackupOffset',
                      defaultMessage: '日志备份位点',
                    })}
                  >
                    {formatTime(backupTenant.lastest_archive_log_checkpoint)}
                  </Descriptions.Item>
                  {hasBackupStrategy && (
                    <Descriptions.Item
                      label={formatMessage({
                        id: 'ocp-v2.Detail.Backup.LogLatency',
                        defaultMessage: '日志延时',
                      })}
                    >
                      {isNullValue(backupTenant.archive_log_delay)
                        ? '-'
                        : `${backupTenant.archive_log_delay}s`}
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </MyCard>
              <Overview
                tenantName={tenantName}
                // 有备份策略才展示可恢复时间区间
                hasBackupStrategy={hasBackupStrategy}
                taskTab={taskTab}
                taskSubTab={taskSubTab}
              />
            </>
          )}
        </PageContainer>
      )}
    </>
  );
};

export default Backup;
