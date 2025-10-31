import { DATA_BACKUP_TYPE_LIST, DATA_BACKUP_TASK_STATUS_LIST } from '@/constant/backup';
import { formatMessage } from '@/util/intl';
import * as ObTenantBackupController from '@/service/obshell/backup';
import { getTableData } from '@/util';
import { getOperationComponent } from '@/util/component';
import { formatTime } from '@/util/datetime';
import { getTaskDuration } from '@/util/task';
import { history, useSelector } from 'umi';
import React, { useState } from 'react';
import { words, join } from 'lodash';
import { Badge, Button, Descriptions, Empty, Modal, Table, message } from '@oceanbase/design';
import type { TableProps } from '@oceanbase/design/es/table';
import { findByValue, isNullValue } from '@oceanbase/util';
import { useInterval, useRequest } from 'ahooks';
export interface DataProps extends TableProps<API.CdbObBackupTask> {
  tenantName?: string;
  backupSetId?: number;
  obVersion?: string;
  startTime?: string;
  endTime?: string;
  // 是否嵌套在抽屉组件中
  isNestInDrawer?: boolean;
  status?: string;
}

const Data: React.FC<DataProps> = ({
  tenantName,
  backupSetId,
  obVersion,
  startTime,
  endTime,
  isNestInDrawer,
  status,
  ...restProps
}) => {
  const { clusterData } = useSelector((state: DefaultRootState) => state.cluster);

  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<API.CdbObBackupTask>();

  const { tableProps, refresh } = getTableData({
    fn: ObTenantBackupController.listTenantBackupTasks,
    params: {
      name: tenantName,
      backupSetId,
      start_time: startTime,
      end_time: endTime,
    },
    options: {
      defaultParams: [
        {
          current: 1,
          pageSize: 10,
          sorter: {},
          filters: {
            status: status?.split(',') || [],
          },
        },
      ],
    },
    deps: [tenantName, backupSetId, startTime, endTime],
  });

  const { run: patchTenantBackup } = useRequest(ObTenantBackupController.patchTenantBackup, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'OBShell.Backup.Backup.Data.TaskCanceledSuccessfully',
            defaultMessage: '任务取消成功',
          })
        );
        refresh();
      }
    },
  });

  const handleOperation = (key: string, record: API.CdbObBackupTask) => {
    if (key === 'viewErrorMessage') {
      setVisible(true);
      setCurrentRecord(record);
    } else if (key === 'cancel') {
      Modal.confirm({
        title: formatMessage({
          id: 'obshell.Backup.Backup.Data.AreYouSureYouWantToCancelThe',
          defaultMessage: '确定要取消备份调度吗？',
        }),
        okButtonProps: { danger: true, ghost: true },
        onOk: () => {
          return patchTenantBackup(
            {
              name: tenantName,
            },
            {
              status: 'canceled',
            }
          );
        },
      });
    }
  };

  const columns = [
    {
      title: isNestInDrawer
        ? formatMessage({
            id: 'ocp-v2.Backup.Backup.Data.TheIdOfTheData',
            defaultMessage: '数据备份任务 ID',
          })
        : 'ID',
      dataIndex: 'job_id',
      fixed: 'left',
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Data.DataBackupMethod',
        defaultMessage: '数据备份方式',
      }),

      dataIndex: 'backup_type',
      render: (text: API.CdbObBackupTask) => (
        <span>{findByValue(DATA_BACKUP_TYPE_LIST, text).label}</span>
      ),
    },
    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Data.BackupSetId',
        defaultMessage: '备份集 ID',
      }),

      dataIndex: 'backup_set_id',
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Data.StartTime',
        defaultMessage: '开始时间',
      }),

      dataIndex: 'start_timestamp',
      sorter: true,
      render: (text: string) => <span>{formatTime(text)}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Data.EndTime',
        defaultMessage: '结束时间',
      }),

      dataIndex: 'end_timestamp',
      sorter: true,
      render: (text: string) => <span>{formatTime(text)}</span>,
    },

    {
      title: formatMessage({ id: 'ocp-v2.Backup.Backup.Data.State', defaultMessage: '状态' }),
      dataIndex: 'status',
      filters: DATA_BACKUP_TASK_STATUS_LIST.map(item => ({
        value: item.value,
        text: item.label,
      })),
      fixed: 'right',
      // 拆分字符串 string 中的词为数组
      ...(status ? { defaultFilteredValue: words(status) } : {}),
      render: (text: string, record: API.CdbObBackupTask) => {
        const statusItem = findByValue(DATA_BACKUP_TASK_STATUS_LIST, text);
        return (
          <Badge
            status={statusItem.badgeStatus}
            text={`${statusItem.label}${
              isNullValue(record.progress) ? '' : ` (${record.progress}%)`
            }`}
          />
        );
      },
    },
    {
      title: formatMessage({
        id: 'ocp-v2.Backup.Backup.Data.Operation',
        defaultMessage: '操作',
      }),

      dataIndex: 'operation',
      fixed: 'right',
      render: (text: string, record: API.CdbObBackupTask) => {
        const operations =
          findByValue(DATA_BACKUP_TASK_STATUS_LIST, record.status).operations || [];
        return getOperationComponent({
          operations,
          handleOperation,
          record,
          displayCount: 1,
        });
      },
    },
  ];

  const getExpandedRowRender = (record: API.CdbObBackupTask) => {
    return (
      <Descriptions column={4} className="descriptions-in-expandable-table">
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-v2.Backup.Backup.Data.BackupMethod',
            defaultMessage: '备份方式',
          })}
        >
          {formatMessage({
            id: 'OBShell.Backup.Backup.Data.PhysicalBackup',
            defaultMessage: '物理备份',
          })}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-v2.Backup.Backup.Data.BackupDuration',
            defaultMessage: '备份时长',
          })}
        >
          {getTaskDuration(record)}
        </Descriptions.Item>
        {/* OB 4.0 会返回 -1，表示没有数据版本号 */}
        {/* issue: https://aone.alipay.com/issue/44373048 */}
        {record.dataVersion === -1 ? undefined : (
          <Descriptions.Item
            label={formatMessage({
              id: 'ocp-v2.Backup.Backup.Data.DataVersion',
              defaultMessage: '数据版本',
            })}
          >
            {record.dataVersion || '-'}
          </Descriptions.Item>
        )}

        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-v2.Backup.Backup.Data.StorageDirectory',
            defaultMessage: '存储目录',
          })}
        >
          {record.path}
        </Descriptions.Item>
      </Descriptions>
    );
  };

  const handleTableChange = (newPagination, filters, sorter) => {
    if (tableProps.onChange) {
      tableProps.onChange(newPagination, filters, sorter);
    }
    history.replace({
      query: {
        status: join(filters?.status, ',') || '',
      },
    });
  };

  // 如果存在初始化和进行中的任务
  const polling = tableProps?.dataSource?.some(
    item => item.status === 'BEGINNING' || item.status === 'DOING' || item.status === 'CANCELING'
  );

  useInterval(
    () => {
      refresh();
    },
    polling ? 1000 : undefined
  );

  return (
    <div>
      <Table
        columns={columns}
        // 由于 taskId 可能不唯一，因此将 taskId_index 拼接成 rowKey
        // issue: https://aone.alipay.com/issue/32903538
        rowKey={(record, index) => `${record.taskId}_${index}`}
        expandedRowRender={getExpandedRowRender}
        {...restProps}
        {...tableProps}
        loading={tableProps.loading && !polling}
        onChange={handleTableChange}
        scroll={{ x: 'max-content' }}
      />

      <Modal
        title={formatMessage({
          id: 'ocp-v2.Backup.Backup.Data.CauseOfFailure',
          defaultMessage: '失败原因',
        })}
        visible={visible}
        destroyOnClose={true}
        onCancel={() => {
          setVisible(false);
          setCurrentRecord(undefined);
        }}
        footer={
          <Button
            type="primary"
            onClick={() => {
              setVisible(false);
            }}
          >
            {formatMessage({ id: 'ocp-v2.Backup.Backup.Data.Closed', defaultMessage: '关闭' })}
          </Button>
        }
      >
        <div style={{ padding: 12, minHeight: 120 }}>
          {(currentRecord && currentRecord.comment) || <Empty />}
        </div>
      </Modal>
    </div>
  );
};

export default Data;
