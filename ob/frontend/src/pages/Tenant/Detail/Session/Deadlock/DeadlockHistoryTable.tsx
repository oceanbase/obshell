import { formatMessage } from '@/util/intl';
import React from 'react';
import { Table, Tooltip, Typography } from '@oceanbase/design';
import styles from './Deadlock.less';

export interface DeadlockHistoryTableProps {
  dataSource?: API.DeadLockNode[];
}

const DeadlockHistoryTable: React.FC<DeadlockHistoryTableProps> = ({ dataSource }) => {
  const columns = [
    {
      title: formatMessage({
        id: 'OBShell.Session.Deadlock.DeadlockHistoryTable.NumberId',
        defaultMessage: '编号 ID',
      }),
      dataIndex: 'idx',
    },

    {
      title: formatMessage({
        id: 'OBShell.Session.Deadlock.DeadlockHistoryTable.TransactionHash',
        defaultMessage: '事务 Hash',
      }),
      dataIndex: 'transaction_hash',
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          {text}
        </Tooltip>
      ),
    },

    {
      title: formatMessage({
        id: 'OBShell.Session.Deadlock.DeadlockHistoryTable.InterruptOrNot',
        defaultMessage: '是否中断',
      }),
      dataIndex: 'roll_backed',
      render: (text: boolean) =>
        text
          ? formatMessage({
              id: 'OBShell.Session.Deadlock.DeadlockHistoryTable.Yes',
              defaultMessage: '是',
            })
          : formatMessage({
              id: 'OBShell.Session.Deadlock.DeadlockHistoryTable.No',
              defaultMessage: '否',
            }),
    },

    {
      title: formatMessage({
        id: 'OBShell.Session.Deadlock.DeadlockHistoryTable.RequestResources',
        defaultMessage: '请求资源',
      }),
      dataIndex: 'resource',
      render: (text: string) => (
        <Typography.Text ellipsis={{ tooltip: true }} style={{ maxWidth: 200 }}>
          {text}
        </Typography.Text>
      ),
    },

    {
      title: 'SQL',
      dataIndex: 'sql',
      render: (text: string) => (
        <Typography.Text ellipsis={{ tooltip: true }} style={{ maxWidth: 200 }}>
          {text}
        </Typography.Text>
      ),
    },
  ];

  return (
    <Table
      dataSource={dataSource}
      columns={columns}
      pagination={false}
      size="small"
      className={styles.deadlockTable}
    />
  );
};

export default DeadlockHistoryTable;
