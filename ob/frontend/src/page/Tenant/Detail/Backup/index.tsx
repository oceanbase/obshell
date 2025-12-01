import ContentWithReload from '@/component/ContentWithReload';
import Empty from '@/component/Empty';
import MyCard from '@/component/MyCard';
import { BACKUP_SCHEDULE_STATUS_LIST } from '@/constant/backup';
import { formatMessage } from '@/util/intl';
import * as ObTenantBackupController from '@/service/obshell/backup';
import { getOperationComponent } from '@/util/component';
import { formatTime } from '@/util/datetime';
import { history, useSelector } from 'umi';
import React, { useEffect } from 'react';
import { Button, Descriptions, message, Modal, Space, Tooltip } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { findByValue, isNullValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import useReload from '@/hook/useReload';
import Overview from './Overview';
import styles from './index.less';

interface BackupProps {
  location: {
    pathname: string;
    query: {
      tab?: string;
      taskTab?: string;
      taskSubTab?: string;
    };
  };

  match: {
    params: {
      tenantName: string;
    };
  };
}

const Backup: React.FC<BackupProps> = ({
  location: {
    query: { tab = 'detail', taskTab, taskSubTab },
  },

  match: {
    params: { tenantName },
  },
}) => {
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
  const backupScheduleStatusItem = findByValue(
    BACKUP_SCHEDULE_STATUS_LIST,
    backupTenant.backupScheduleStatus
  );

  const operations = backupScheduleStatusItem.operations || [];

  // 启动租户的备份调度
  const { runAsync: startBackupStrategySchedule } = useRequest(
    ObTenantBackupController.startTenantBackupStrategySchedule,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'ocp-v2.Detail.Backup.BackupSchedulingStarted',
              defaultMessage: '备份调度启动成功',
            })
          );

          getBackupOverviewRefresh();
        }
      },
    }
  );

  // 暂停租户的备份调度
  const { runAsync: stopBackupStrategySchedule } = useRequest(
    ObTenantBackupController.stopTenantBackupStrategySchedule,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'ocp-v2.Detail.Backup.BackupSchedulingPaused',
              defaultMessage: '备份调度暂停成功',
            })
          );

          getBackupOverviewRefresh();
        }
      },
    }
  );

  // 删除租户的备份策略
  const { runAsync: deleteObTenantBackupStrategy } = useRequest(
    ObTenantBackupController.deleteObTenantBackupStrategy,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'ocp-v2.Detail.Backup.TheBackupPolicyHasBeen',
              defaultMessage: '备份策略删除成功',
            })
          );

          getBackupOverviewRefresh();
        }
      },
    }
  );

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
  // !getBackupOverviewLoading && backupTenant.backupStatus === 'HAS_BACKUP_STRATEGY';

  const handleOperation = (key: string) => {
    if (key === 'new') {
      // history.push(`/cluster/${clusterId}/tenant/${tenantName}/backup/addBackupStrategy`);
    } else if (key === 'backupNow') {
      // history.push(`/cluster/${clusterId}/tenant/${tenantName}/backup/backup/nowBackupTenant`);
    } else if (key === 'restoreNow') {
      // history.push(`/cluster/${clusterId}/tenant/${tenantName}/backup/restoreNow`);
    } else if (key === 'start') {
      Modal.confirm({
        title: formatMessage(
          {
            id: 'ocp-v2.Detail.Backup.AreYouSureYouWantToStartThe',
            defaultMessage: '确定启动租户 {tenantDataName} 的备份调度吗？',
          },
          { tenantDataName: tenantData?.tenant_name }
        ),
        content: formatMessage({
          id: 'ocp-v2.Detail.Backup.StartingABackupPolicyRequires',
          defaultMessage: '启动备份策略需要消耗一定的资源。',
        }),
        okText: formatMessage({ id: 'ocp-v2.Detail.Backup.Start', defaultMessage: '启动' }),
        onOk: () => {
          return startBackupStrategySchedule({
            // id: clusterId,
            // tenantName,
          });
        },
      });
    } else if (key === 'stop') {
      Modal.confirm({
        title: formatMessage(
          {
            id: 'ocp-v2.Detail.Backup.AreYouSureYouWantToSuspendBackup',
            defaultMessage: '确定暂停租户 {tenantDataName} 的备份调度吗？',
          },
          { tenantDataName: tenantData?.tenant_name }
        ),
        content: formatMessage({
          id: 'ocp-v2.Detail.Backup.TheBackupWillNotBeAutomaticallyBackedUp',
          defaultMessage: '暂停后将不再自动备份，请谨慎操作。',
        }),
        okText: formatMessage({ id: 'ocp-v2.Detail.Backup.Pause', defaultMessage: '暂停' }),
        okButtonProps: {
          danger: true,
          ghost: true,
        },
        onOk: () => {
          return stopBackupStrategySchedule({
            // id: clusterId,
            // tenantName,
          });
        },
      });
    } else if (key === 'edit') {
      // history.push({
      //   pathname: `/cluster/${clusterId}/tenant/${tenantName}/backup/editBackupStrategy`,
      //   query: {
      //     strategyId: backupTenant.id,
      //   },
      // });
    } else if (key === 'delete') {
      Modal.confirm({
        title: formatMessage(
          {
            id: 'ocp-v2.Detail.Backup.AreYouSureYouWantToDeleteThe',
            defaultMessage: '确定删除租户 {tenantDataName} 的备份策略吗？',
          },
          { tenantDataName: tenantData?.tenant_name }
        ),
        content: formatMessage({
          id: 'ocp-v2.Detail.Backup.AfterDeletionTheTenantDoesNotHaveA',
          defaultMessage:
            '删除后将导致租户无备份策略，且默认关闭日志备份，系统无法保障数据安全，请谨慎操作。',
        }),

        okText: formatMessage({ id: 'ocp-v2.Detail.Backup.Delete', defaultMessage: '删除' }),
        okButtonProps: { danger: true, ghost: true },
        onOk: () => {
          return deleteObTenantBackupStrategy({
            name: tenantName,
          });
        },
      });
    }
  };

  const addTenantBackupStrategyButton = (type = 'default') => (
    <Tooltip
      title={
        (backupTenant.obClusterType === 'STANDBY' &&
          formatMessage({
            id: 'ocp-v2.Detail.Backup.TheStandbyTenantDoesNotSupportCreatingA',
            defaultMessage: '备租户不支持新建备份策略',
          })) ||
        (tenantData?.tenant_name === 'sys' &&
          formatMessage({
            id: 'ocp-v2.Detail.Backup.SysTenantsDoNotSupportNewBackupPolicies',
            defaultMessage: 'sys 租户不支持新建备份策略',
          }))
      }
    >
      <Button
        type={type}
        disabled={backupTenant.obClusterType === 'STANDBY' || tenantData?.tenant_name === 'sys'}
        onClick={() => {
          // history.push(`/cluster/${clusterId}/tenant/${tenantName}/backup/addBackupStrategy`);
        }}
      >
        {formatMessage({
          id: 'ocp-v2.Detail.Backup.CreateATenantLevelBackupPolicy',
          defaultMessage: '新建租户级备份策略',
        })}
      </Button>
    </Tooltip>
  );

  const getOperations = () => {
    const title = {
      key: 'title',
      label: (
        <span style={{ fontSize: 12, color: '#8592AD' }}>
          {formatMessage({ id: 'ocp-v2.Detail.Backup.TenantLevel', defaultMessage: '租户级别' })}
        </span>
      ),

      disabled: true,
    };
    const opts = noBackupStrategy
      ? operations.filter(item => item?.value !== 'restoreNow')
      : operations || [];
    return opts?.length > 2 ? [...opts.slice(0, 2), title, ...opts.slice(2)] : [...opts];
  };

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
          <Space>
            {backupNowTenantButton}
            {/* {addTenantBackupStrategyButton()} */}
          </Space>
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
                  // queryObTenantBackupStrategyRefresh();
                  getBackupOverviewRefresh();
                }}
              />
            ),

            // subTitle: <PageDocLink href={TENANT_BACKUP_LINK} />,
            extra:
              // 有组户级备份策略
              // 根据备份调度状态进行操作映射的条件: 租户设置了租户级备份策略，即备份维度为租户，仅 OB 版本 >= 4.0 支持
              backupTenant.backupDimension === 'TENANT' ? (
                getOperationComponent({
                  mode: 'button',
                  record: backupTenant,
                  operations: getOperations(),
                  handleOperation,
                  displayCount: 2,
                })
              ) : (
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
                  {/* {addTenantBackupStrategyButton(noBackupStrategy ? 'default' : 'primary')} */}
                </Space>
              ),
          }}
        >
          {/* {backupTenant.backupStatus === 'HAS_BACKUP_STRATEGY' && (
           <Tabs
             activeKey={tab}
             onChange={key => {
               history.push({
                 // pathname: `/cluster/${clusterId}/tenant/${tenantName}/backup`,
                 query: {
                   tab: key,
                 },
               });
             }}
           >
             <TabPane
               key="detail"
               tab={formatMessage({ id: 'ocp-v2.Detail.Backup.Details', defaultMessage: '详情' })}
             />
              <TabPane
               key="backupStrategy"
               tab={formatMessage({
                 id: 'ocp-v2.Detail.Backup.BackupPolicy',
                 defaultMessage: '备份策略',
               })}
             />
           </Tabs>
          )} */}

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
                  {/* {hasBackupStrategy && (
                 <>
                   <Descriptions.Item
                     label={formatMessage({
                       id: 'ocp-v2.Detail.Backup.BackupMethod',
                       defaultMessage: '备份方式',
                     })}
                   >
                     {
                       findByValue(BACKUP_MODE_LIST, backupTenant.obBackupAbility?.backupMode)
                         .label
                     }
                   </Descriptions.Item>
                   <Descriptions.Item
                     label={formatMessage({
                       id: 'ocp-v2.Detail.Backup.BackupFiles',
                       defaultMessage: '备份文件',
                     })}
                   >
                     {isNullValue(backupTenant.file_count)
                       ? '-'
                       : `${byte2GB(backupTenant.file_count || 0)}GiB`}
                   </Descriptions.Item>
                 </>
                )} */}

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
                backupMode={backupTenant.backupMode}
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
