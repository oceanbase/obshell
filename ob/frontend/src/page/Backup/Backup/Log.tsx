import { LOG_BACKUP_TASK_STATUS_LIST } from '@/constant/backup';
import { formatMessage } from '@/util/intl';
import * as ObBackupController from '@/service/obshell/backup';
import { getTableData, isEnglish } from '@/util';
import { getOperationComponent } from '@/util/component';
import { formatTime } from '@/util/datetime';
import React, { useState } from 'react';
import { noop } from 'lodash';
import { Badge, Button, Empty, message, Modal, Table, Tooltip } from '@oceanbase/design';
import { findByValue } from '@oceanbase/util';
import { useInterval, useRequest } from 'ahooks';

export interface LogProps {
  tenantName?: string;
  startTime?: string;
  endTime?: string;
  onSuccess?: (logBackupTaskList: API.LogBackupTask[]) => void;
  operationsFilter?: (operations: Global.OperationItem[]) => Global.OperationItem[]; // 是否允许操作
  obVersion?: string;
}

const Log: React.FC<LogProps> = ({
  tenantName,
  startTime,
  endTime,
  onSuccess = noop,
  operationsFilter,
  obVersion,
}) => {
  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<API.LogBackupTask | null>(null);

  const { tableProps, refresh } = getTableData({
    fn: ObBackupController.getTenantArchiveLogOverview,
    params: {
      name: tenantName,
      start_time: startTime,
      end_time: endTime,
    },
    deps: [tenantName, startTime, endTime],
    options: {
      onSuccess: res => {
        if (onSuccess) {
          // 由于 getTableData 对 response 进行了 format，因此取 res.list 作为日志备份的列表数据
          onSuccess(res.list || []);
        }
      },
    },
  });

  const polling = tableProps?.dataSource?.some(item =>
    ['BEGINNING', 'STOPPING', 'SUSPENDING'].includes(item.status)
  );

  useInterval(
    () => {
      refresh();
    },
    polling ? 1000 : undefined
  );

  const { runAsync: patchTenantArchiveLog } = useRequest(ObBackupController.patchTenantArchiveLog, {
    manual: true,
  });

  const handleOperation = (key: string, record: any) => {
    const tenant_name = record.tenant_name || tenantName;
    if (key === 'start') {
      Modal.confirm({
        title: formatMessage(
          {
            id: 'OBShell.Backup.Backup.Log.AreYouSureYouWant',
            defaultMessage: '确定开启 {tenantName} 的备份调度吗？',
          },
          { tenantName: tenant_name }
        ),
        content: formatMessage({
          id: 'OBShell.Backup.Backup.Log.DataBackupConsumesACertain',
          defaultMessage: '启动备份策略后数据备份需要消耗一定数据库资源。',
        }),
        onOk: () => {
          return patchTenantArchiveLog(
            {
              name: tenant_name,
            },
            {
              status: 'DOING',
            }
          ).then(res => {
            if (res.successful) {
              message.success(
                formatMessage({
                  id: 'OBShell.Backup.Backup.Log.LogBackupTaskResumedSuccessfully',
                  defaultMessage: '日志备份任务恢复成功',
                })
              );
              refresh();
            }
          });
        },
      });
    } else if (key === 'stop') {
      Modal.confirm({
        title: formatMessage(
          {
            id: 'ocp-v2.Backup.Backup.Log.AreYouSureYouWant.1',
            defaultMessage: '确定要停止 {recordClusterName}:{recordObClusterId} 的日志备份任务吗？',
          },
          { recordClusterName: tenant_name, recordObClusterId: record.tenant_id }
        ),
        content: (
          <span>
            {formatMessage({
              id: 'ocp-v2.Backup.Backup.Log.IfYouEnableABackupPolicyOrManually',
              defaultMessage: '· 若手动发起数据备份，系统将默认开启日志备份。',
            })}
            <br />
            {formatMessage({
              id: 'ocp-v2.Backup.Backup.Log.AfterLogBackupIsStoppedTheSystemDoes',
              defaultMessage:
                '· 停止日志备份后，系统将不支持任意时间点恢复，无法保障数据安全，请谨慎操作。',
            })}
          </span>
        ),

        okButtonProps: { ghost: true, danger: true },
        okText: formatMessage({ id: 'ocp-v2.Backup.Backup.Log.Stop', defaultMessage: '停止' }),

        onOk: () => {
          return patchTenantArchiveLog(
            {
              name: tenant_name,
            },
            {
              status: 'STOP',
            }
          ).then(res => {
            if (res.successful) {
              message.success(
                formatMessage({
                  id: 'OBShell.Backup.Backup.Log.LogBackupTaskStoppedSuccessfully',
                  defaultMessage: '日志备份任务停止成功',
                })
              );
              refresh();
            }
          });
        },
      });
    } else if (key === 'pause') {
      Modal.confirm({
        title: formatMessage(
          {
            id: 'OBShell.Backup.Backup.Log.AreYouSureYouWant.1',
            defaultMessage: '确定暂停 {tenantName} 的备份调度吗？',
          },
          { tenantName: tenant_name }
        ),
        content: formatMessage({
          id: 'OBShell.Backup.Backup.Log.AutomaticBackupWillNoLonger',
          defaultMessage: '暂停后将不再自动备份，请谨慎操作。',
        }),

        okText: formatMessage({ id: 'ocp-v2.Backup.Backup.Log.Pause', defaultMessage: '暂停' }),
        okButtonProps: { ghost: true, danger: true },
        onOk: () => {
          return patchTenantArchiveLog(
            {
              name: tenant_name,
            },
            {
              status: 'SUSPEND',
            }
          ).then(res => {
            if (res.successful) {
              message.success(
                formatMessage({
                  id: 'OBShell.Backup.Backup.Log.LogBackupTaskPausedSuccessfully',
                  defaultMessage: '日志备份任务暂停成功',
                })
              );
              refresh();
            }
          });
        },
      });
    } else if (key === 'viewErrorMessage') {
      setVisible(true);
      setCurrentRecord(record);
    }
  };

  const columns = [
    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Log.BackupMethod',
        defaultMessage: '备份方式',
      }),
      dataIndex: 'backupMode',
      render: (text: API.BackupMode) => (
        <span>
          {formatMessage({
            id: 'OBShell.Backup.Backup.Log.PhysicalBackup',
            defaultMessage: '物理备份',
          })}
        </span>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Log.LatestBackupTime',
        defaultMessage: '最新备份时间点',
      }),
      dataIndex: 'checkpoint',
      render: (text: string) => <span>{formatTime(text)}</span>,
      sorter: true,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Log.LatencyS',
        defaultMessage: '延时（s）',
      }),
      dataIndex: 'delay',
      sorter: true,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Log.StorageDirectory',
        defaultMessage: '存储目录',
      }),
      dataIndex: 'path',
      ellipsis: true,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Log.StartTime',
        defaultMessage: '开始时间',
      }),
      dataIndex: 'start_time',
      render: (text: string) => <span>{formatTime(text)}</span>,
    },

    {
      title: formatMessage({ id: 'ocp-v2.Backup.Backup.Log.State', defaultMessage: '状态' }),
      dataIndex: 'status',
      // filters: LOG_BACKUP_TASK_STATUS_LIST.map(item => ({
      //   value: item.value,
      //   text: item.label,
      // })),
      width: 120,
      ...(isEnglish() ? { fixed: 'right' } : {}),
      render: (text: API.LogBackupTaskStatus) => {
        const statusItem = findByValue(LOG_BACKUP_TASK_STATUS_LIST, text);
        return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
      },
    },
    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Log.Operation',
        defaultMessage: '操作',
      }),
      width: 80,
      dataIndex: 'operation',
      ...(isEnglish() ? { fixed: 'right' } : {}),
      render: (text: string, record: API.LogBackupTask = {}) => {
        let operations = findByValue(LOG_BACKUP_TASK_STATUS_LIST, record.status).operations || [];

        operations = operationsFilter?.(operations) || [];
        return getOperationComponent({
          operations,
          handleOperation,
          record: { ...record, obVersion },
          displayCount: 0,
        });
      },
    },
  ];

  return (
    <div>
      {/* 日志备份任务没有 taskId，因此将 index 作为 rowKey */}
      <Table
        columns={columns}
        rowKey={(record, index) => index}
        {...(isEnglish()
          ? {
              scroll: {
                x: 1500,
              },
            }
          : {})}
        {...tableProps}
        loading={tableProps.loading && !polling}
      />

      <Modal
        title={formatMessage({
          id: 'ocp-v2.Backup.Backup.Log.CauseOfFailure',
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
            {formatMessage({ id: 'ocp-v2.Backup.Backup.Log.Closed', defaultMessage: '关闭' })}
          </Button>
        }
      >
        <div style={{ padding: 12, minHeight: 120 }}>
          {(currentRecord && currentRecord.errorMsg) || <Empty />}
        </div>
      </Modal>
    </div>
  );
};

export default Log;
