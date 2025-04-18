/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { formatMessage } from '@/util/intl';
import React from 'react';
import { Typography, Table, Button, Modal } from '@oceanbase/design';
import type { ModalProps } from '@oceanbase/design/es/modal';
import type { ColumnProps } from '@oceanbase/design/es/table';

const { Text } = Typography;

interface OBProxyAndConnectionStringModalProps extends ModalProps {
  obproxyAndConnectionStrings: API.ObproxyAndConnectionString[];
  onSuccess: () => void;
  onCancel: () => void;
  userName?: string;
}

const OBProxyAndConnectionStringModal: React.FC<OBProxyAndConnectionStringModalProps> = ({
  obproxyAndConnectionStrings,
  onSuccess,
  onCancel,
  userName,
  ...restProps
}) => {
  const columns: ColumnProps<API.ObproxyAndConnectionString>[] = [
    {
      title: 'OBProxy',
      dataIndex: 'OBProxy',
      width: 270,
      render: (value: string, record) => {
        const text = `${record?.obProxyAddress}:${record?.obProxyPort}`;
        return (
          <Text style={{ width: 240 }} ellipsis={{ tooltip: text }} copyable={{ text }}>
            {text}
          </Text>
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Component.OBProxyAndConnectionStringModal.ConnectionString',
        defaultMessage: '连接串',
      }),
      dataIndex: 'connectionString',
      width: 570,
      render: (connectionString: string) => (
        <Text
          style={{ width: 540 }}
          ellipsis={{ tooltip: connectionString }}
          copyable={{ text: connectionString }}
        >
          {connectionString}
        </Text>
      ),
    },
  ];

  return (
    <>
      <Modal
        title={
          userName
            ? formatMessage(
                {
                  id: 'ocp-express.Detail.Component.OBProxyAndConnectionStringModal.ConnectionStringOfUserUsername',
                  defaultMessage: '用户 {userName} 的连接串',
                },
                { userName }
              )
            : formatMessage({
                id: 'ocp-express.Detail.Component.OBProxyAndConnectionStringModal.ObproxyConnectionString',
                defaultMessage: 'OBProxy/连接串',
              })
        }
        destroyOnClose={true}
        onOk={() => {
          onSuccess();
        }}
        onCancel={onCancel}
        footer={
          <Button type="primary" onClick={onCancel}>
            {formatMessage({
              id: 'ocp-express.Detail.Component.OBProxyAndConnectionStringModal.Closed',
              defaultMessage: '关闭',
            })}
          </Button>
        }
        {...restProps}
      >
        <Table
          columns={columns}
          pagination={{
            size: 'small',
            pageSize: 5,
            showTotal: total =>
              formatMessage(
                {
                  id: 'ocp-express.Detail.Component.OBProxyAndConnectionStringModal.TotalTotal',
                  defaultMessage: '共 {total} 条',
                },
                { total }
              ),
          }}
          dataSource={obproxyAndConnectionStrings}
          rowKey={(record: API.ObproxyAndConnectionString) => record?.connectionString}
        />
      </Modal>
    </>
  );
};
export default OBProxyAndConnectionStringModal;
