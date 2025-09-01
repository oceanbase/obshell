import { RESTORE_TASK_STATUS_LIST } from '@/constant/backup';
import { formatMessage } from '@/util/intl';
import * as ObTenantBackupController from '@/service/obshell/restore';
import { poolDrop } from '@/service/obshell/resourcePool';
import { getTableData } from '@/util';
import { getOperationComponent } from '@/util/component';
import { formatTime, formatTimeWithMicroseconds } from '@/util/datetime';
import { getTaskDuration } from '@/util/task';
import { useLocation } from 'umi';
import React, { useState } from 'react';
import { Badge, Button, Descriptions, Empty, message, Modal, Card, Table } from '@oceanbase/design';
import { findByValue } from '@oceanbase/util';
import { useInterval, useRequest } from 'ahooks';
import './RestoreTask.less';

export interface RestoreTaskProps {
  type?: 'restoreTask' | 'samplingTask';
  tenantName?: string;
  startTime?: string;
  endTime?: string;
}

const RestoreTask: React.FC<RestoreTaskProps> = ({
  type = 'restoreTask',
  tenantName,
  startTime,
  endTime,
}) => {
  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<API.RestoreOverview | null>(null);
  const { pathname } = useLocation();

  const { tableProps, refresh } = getTableData({
    fn: ObTenantBackupController.listRestoreTasks,
    params: {
      name: tenantName,
      start_time: startTime,
      end_time: endTime,
    },
    deps: [startTime, endTime],
  });

  const { runAsync: poolDropFn } = useRequest(poolDrop, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-v2.Backup.Restore.RestoreTask.ResourcePoolCleanSucceeded',
            defaultMessage: '资源池清理成功',
          })
        );
        refresh();
      }
    },
  });

  const { runAsync: cancelRestoreTask } = useRequest(ObTenantBackupController.cancelRestoreTask, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'OBShell.Backup.Restore.RestoreTask.TaskCanceledSuccessfully',
            defaultMessage: '任务取消成功',
          })
        );
        refresh();
      }
    },
  });

  const handleOperation = (key: string, record: API.RestoreOverview) => {
    if (key === 'clear') {
      Modal.confirm({
        title: formatMessage(
          {
            id: 'OBShell.Backup.Restore.RestoreTask.AreYouSureYouWant',
            defaultMessage:
              '确定要清理 {recordBackupCluster} / {recordBackupTenant} 恢复到 {recordRestoreTenant} 创建的资源池吗？',
          },
          {
            recordBackupCluster: record.backup_cluster_name,
            recordBackupTenant: record.backup_tenant_name,
            recordRestoreTenant: record.restore_tenant_name,
          }
        ),
        okButtonProps: { danger: true, ghost: true },
        onOk: () => {
          return poolDropFn({
            name: record?.restore_tenant_name,
          });
        },
      });
    } else if (key === 'viewErrorMessage') {
      setVisible(true);
      setCurrentRecord(record);
    } else if (key === 'cancel') {
      Modal.confirm({
        title: `确定要取消 ${record.restore_tenant_name} 的恢复任务吗？`,
        okButtonProps: { danger: true, ghost: true },
        onOk: () => {
          return cancelRestoreTask(
            {
              tenantName: record?.restore_tenant_name,
            },
            {
              status: 'canceled',
            }
          );
        },
      });
    } else if (key === 'copy') {
      // history.push(
      //   `${pathname}/copy?taskId=${record.job_id}${
      //     record.restore_tenant_id ? `&targetTenantId=${record.restore_tenant_id}` : ''
      //   }`
      // );
    }
  };
  const columns = [
    { title: 'ID', dataIndex: 'job_id', fixed: 'left' },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.BackupCluster',
        defaultMessage: '备份集群',
      }),
      fixed: 'left',
      dataIndex: 'backup_cluster_name',
    },
    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.BackupTenant',
        defaultMessage: '备份租户',
      }),
      dataIndex: 'backup_tenant_name',
      render: (text: string) => <span>{text || '-'}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.BackupMethod',
        defaultMessage: '备份方式',
      }),
      dataIndex: 'backupMode',
      render: (text: string) =>
        formatMessage({
          id: 'OBShell.Backup.Restore.RestoreTask.PhysicalBackup',
          defaultMessage: '物理备份',
        }),
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.TargetTenant',
        defaultMessage: '目标租户',
      }),
      dataIndex: 'restore_tenant_name',
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.RestoreObjects',
        defaultMessage: '恢复对象',
      }),
      dataIndex: 'restoreObjectList',
      render: (text: string) => {
        return formatMessage({
          id: 'ocp-v2.Backup.Restore.RestoreTask.Tenant',
          defaultMessage: '租户',
        });
      },
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.StartTime',
        defaultMessage: '开始时间',
      }),
      dataIndex: 'start_timestamp',
      // sorter: true,
      render: (text: string) => <span>{formatTime(text)}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.RecoveryDuration',
        defaultMessage: '恢复时长',
      }),
      dataIndex: 'finish_timestamp',
      render: (text: string, record: API.RestoreOverview) => (
        <span>
          {getTaskDuration({
            start_timestamp: record?.start_timestamp,
            end_timestamp: record?.finish_timestamp,
          })}
        </span>
      ),
    },
    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.State',
        defaultMessage: '状态',
      }),
      dataIndex: 'status',
      fixed: 'right',
      filters: RESTORE_TASK_STATUS_LIST.map(item => ({
        value: item.value,
        text: item.label,
      })),
      // 拆分字符串 string 中的词为数组
      render: (text: string, record: API.RestoreOverview) => {
        const statusItem = findByValue(RESTORE_TASK_STATUS_LIST, text);
        return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
      },
    },
    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Restore.RestoreTask.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      fixed: 'right',
      render: (text: string, record: API.RestoreOverview) => {
        let operations = findByValue(RESTORE_TASK_STATUS_LIST, record.status).operations || [];
        // 抽检任务不允许复制
        if (!record?.replicable || type !== 'restoreTask')
          operations = operations.filter(opt => opt.value !== 'copy');
        return getOperationComponent({
          operations,
          handleOperation,
          record,
          displayCount: 0,
        });
      },
    },
  ];

  const getExpandedRowRender = (record: API.RestoreOverview) => {
    return (
      <Descriptions column={4} className="descriptions-in-expandable-table">
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-v2.Backup.Restore.RestoreTask.RecoveryTimePoint',
            defaultMessage: '恢复时间点',
          })}
        >
          {formatTimeWithMicroseconds(record.restore_scn_display)}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-v2.Backup.Restore.RestoreTask.EndTime',
            defaultMessage: '结束时间',
          })}
        >
          {formatTime(record.finish_timestamp)}
        </Descriptions.Item>

        <Descriptions.Item
          label={formatMessage({
            id: 'OBShell.Backup.Restore.RestoreTask.LogStorageDirectory',
            defaultMessage: '日志存储目录',
          })}
        >
          {record.backup_log_uri || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'OBShell.Backup.Restore.RestoreTask.DataStorageDirectory',
            defaultMessage: '数据存储目录',
          })}
        >
          {record.backup_data_uri || '-'}
        </Descriptions.Item>
      </Descriptions>
    );
  };

  const POLLING_STATUS_LIST = RESTORE_TASK_STATUS_LIST?.filter(
    item => item.badgeStatus === 'processing'
  ).map(item => item.value);

  // 如果存在初始化和进行中的任务
  const polling = tableProps?.dataSource?.some(item => POLLING_STATUS_LIST.includes(item.status));

  useInterval(
    () => {
      refresh();
    },
    polling ? 1000 : undefined
  );

  return (
    <Card className="card-without-padding">
      <Table
        columns={columns}
        // 存在相同 id
        rowKey={(record, index) => `${record.job_id}_${index}`}
        expandedRowRender={getExpandedRowRender}
        {...tableProps}
        scroll={{ x: 'max-content' }}
        loading={tableProps.loading && !polling}
      />

      <Modal
        title={formatMessage({
          id: 'ocp-v2.Backup.Restore.RestoreTask.CauseOfFailure',
          defaultMessage: '失败原因',
        })}
        visible={visible}
        destroyOnClose={true}
        onCancel={() => {
          setVisible(false);
          setCurrentRecord(null);
        }}
        footer={
          <Button
            type="primary"
            onClick={() => {
              setVisible(false);
            }}
          >
            {formatMessage({
              id: 'ocp-v2.Backup.Restore.RestoreTask.Closed',
              defaultMessage: '关闭',
            })}
          </Button>
        }
      >
        <div style={{ padding: 12, minHeight: 120 }}>
          {(currentRecord && currentRecord.comment) || <Empty />}
        </div>
      </Modal>
    </Card>
  );
};

export default RestoreTask;
