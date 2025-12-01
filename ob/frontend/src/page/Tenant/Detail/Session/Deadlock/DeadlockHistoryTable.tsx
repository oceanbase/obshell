import { formatMessage } from '@/util/intl';
import React from 'react';
import { Table, Tooltip, Typography } from '@oceanbase/design';

export interface DeadlockHistoryTableProps {
  dataSource?: API.DeadLockNode[];
}

const DeadlockHistoryTable: React.FC<DeadlockHistoryTableProps> = ({ dataSource }) => {
  const columns = [
    {
      title: formatMessage({
        id: 'ocp-v2.Session.Deadlock.DeadlockHistoryTable.Id',
        defaultMessage: '编号 ID',
      }),
      dataIndex: 'idx',
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Session.Deadlock.DeadlockHistoryTable.TransactionHash',
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
        id: 'ocp-v2.Session.Deadlock.DeadlockHistoryTable.Interrupted',
        defaultMessage: '是否中断',
      }),
      dataIndex: 'roll_backed',
      render: (text: boolean) =>
        text
          ? formatMessage({
              id: 'ocp-v2.Session.Deadlock.DeadlockHistoryTable.Yes',
              defaultMessage: '是',
            })
          : formatMessage({
              id: 'ocp-v2.Session.Deadlock.DeadlockHistoryTable.No',
              defaultMessage: '否',
            }),
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Session.Deadlock.DeadlockHistoryTable.RequestResources',
        defaultMessage: '请求资源',
      }),
      dataIndex: 'resource',
      render: (text: string) => (
        <Typography.Text ellipsis={{ tooltip: true }} className="max-w-200">
          {text}
        </Typography.Text>
      ),
    },

    {
      title: 'SQL',
      dataIndex: 'sql',
      render: (text: string) => (
        <Typography.Text ellipsis={{ tooltip: true }} className="max-w-200">
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
      className="border border-solid border-colorBorder border-b-0"
    />
  );
};

export default DeadlockHistoryTable;
