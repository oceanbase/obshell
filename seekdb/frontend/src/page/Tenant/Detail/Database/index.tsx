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
import React, { useState } from 'react';
import {
  Table,
  Col,
  Row,
  Card,
  Descriptions,
  Tooltip,
  Button,
  Alert,
  Input,
  Space,
  Modal,
  message,
} from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { sortByMoment } from '@oceanbase/util';
import { PAGINATION_OPTION_10 } from '@/constant';
import { FORBID_OPERATION_DBLIST } from '@/constant/tenant';
import { formatTime } from '@/util/datetime';
import { useRequest } from 'ahooks';
import ContentWithReload from '@/component/ContentWithReload';
import MyInput from '@/component/MyInput';
import AddDatabaseDrawer from '../Component/AddDatabaseDrawer';
import OBProxyAndConnectionStringModal from '../Component/OBProxyAndConnectionStringModal';
import RenderConnectionString from '@/component/RenderConnectionString';
import { listDatabases, deleteDatabase } from '@/service/obshell/database';

const Database: React.FC = () => {
  const [connectionStringModalVisible, setConnectionStringModalVisible] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [currentDatabase, setCurrentDatabase] = useState<API.Database | null>(null);
  const [IsAllowDel, setIsAllowDel] = useState(false);

  // 修改参数值的抽屉是否可见
  const [valueVisible, setValueVisible] = useState(false);
  // 删除 Modal
  const [deleteDatabaseModalVisible, setDeleteDatabaseModalVisible] = useState(false);

  const { data, loading, refresh } = useRequest(listDatabases, {
    defaultParams: [{}],
  });

  const dataSource = (data?.data?.contents || [])?.map(item => ({
    ...item,
    createTime: item?.create_time,
    dbName: item?.db_name,
    readonly: item?.read_only,
    connectionUrls: item?.connection_urls?.map(url => ({
      ...url,
      connectionString: url?.connection_string,
      obProxyAddress: url?.obproxy_address,
      obProxyPort: url?.obproxy_port,
    })),
  }));

  const { runAsync, loading: deleteDatabaseLoading } = useRequest(deleteDatabase, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.Detail.Database.TheDatabaseWasDeleted',
            defaultMessage: '数据库删除成功',
          })
        );
        refresh();
        setIsAllowDel(false);
        setCurrentDatabase(null);
        setDeleteDatabaseModalVisible(false);
      }
    },
  });

  const editDatabase = (record: API.Database) => {
    setValueVisible(true);
    setCurrentDatabase(record);
  };

  const confirmOperation = (e: any) => {
    setIsAllowDel(e.target.value === 'delete');
  };

  const handleDeleteDatabase = () => {
    runAsync({
      database: currentDatabase?.db_name,
    });
  };

  const columns = [
    {
      title: formatMessage({
        id: 'ocp-express.Detail.Database.DatabaseName',
        defaultMessage: '数据库名',
      }),

      dataIndex: 'dbName',
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Database.CharacterSet',
        defaultMessage: '字符集',
      }),
      dataIndex: 'charset',
    },

    {
      title: 'Collation',
      dataIndex: 'collation',
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          {text}
        </Tooltip>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Database.JdbcConnectionString',
        defaultMessage: 'JDBC 连接串',
      }),

      dataIndex: 'connectionUrls',
      render: (connectionUrls: API.ObproxyAndConnectionString[], record: API.Database) => {
        return (
          <RenderConnectionString
            callBack={() => {
              setCurrentDatabase(record);
              setConnectionStringModalVisible(true);
            }}
            connectionStrings={connectionUrls}
          />
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Database.Created',
        defaultMessage: '创建时间',
      }),
      dataIndex: 'createTime',
      sorter: (a: API.DbUser, b: API.DbUser) => sortByMoment(a, b, 'create_time'),
      render: (text: string) => formatTime(text),
    },

    {
      title: formatMessage({ id: 'ocp-express.Detail.Database.ReadOnly', defaultMessage: '只读' }),
      dataIndex: 'readonly',
      render: (text: boolean) =>
        text
          ? formatMessage({ id: 'ocp-express.Detail.Database.Is', defaultMessage: '是' })
          : formatMessage({ id: 'ocp-express.Detail.Database.No', defaultMessage: '否' }),
    },
    {
      title: formatMessage({
        id: 'ocp-express.Detail.Database.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      render: (text: string, record: API.Database) => {
        if ([...FORBID_OPERATION_DBLIST, 'information_schema'].includes(record?.dbName)) {
          return '';
        }
        return (
          <Space size="middle">
            <a
              onClick={() => {
                editDatabase(record);
              }}
            >
              {formatMessage({
                id: 'ocp-express.Detail.Database.Modify',
                defaultMessage: '编辑',
              })}
            </a>
            <a
              onClick={() => {
                setDeleteDatabaseModalVisible(true);
                setCurrentDatabase(record);
              }}
            >
              {formatMessage({
                id: 'ocp-express.Detail.Database.Delete',
                defaultMessage: '删除',
              })}
            </a>
          </Space>
        );
      },
    },
  ];

  return (
    <PageContainer
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'ocp-express.Detail.Database.DatabaseManagement',
              defaultMessage: '数据库管理',
            })}
            spin={loading}
            onClick={() => {
              refresh();
            }}
          />
        ),

        extra: (
          <Space>
            <MyInput.Search
              allowClear={true}
              onSearch={(value: string) => setKeyword(value)}
              placeholder={formatMessage({
                id: 'ocp-express.Detail.Database.SearchDatabaseName',
                defaultMessage: '搜索数据库名',
              })}
              className="search-input"
            />
            <Button
              type="primary"
              onClick={() => {
                setValueVisible(true);
              }}
            >
              {formatMessage({
                id: 'ocp-express.Detail.Database.CreateADatabase',
                defaultMessage: '新建数据库',
              })}
            </Button>
          </Space>
        ),
      }}
    >
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card bordered={false} className="card-without-padding">
            <Table
              loading={loading}
              dataSource={dataSource.filter(
                item => !keyword || (item.dbName && item.dbName.includes(keyword))
              )}
              columns={columns}
              pagination={PAGINATION_OPTION_10}
            />

            <AddDatabaseDrawer
              visible={valueVisible}
              database={currentDatabase}
              onCancel={() => {
                setCurrentDatabase(null);
                setValueVisible(false);
              }}
              onSuccess={() => {
                setCurrentDatabase(null);
                setValueVisible(false);
                refresh();
              }}
            />

            <Modal
              width={540}
              title={formatMessage({
                id: 'ocp-express.Detail.Database.DeleteADatabase',
                defaultMessage: '删除数据库',
              })}
              destroyOnClose={true}
              visible={deleteDatabaseModalVisible}
              footer={
                <>
                  <Button
                    onClick={() => {
                      setCurrentDatabase(null);
                      setDeleteDatabaseModalVisible(false);
                    }}
                  >
                    {formatMessage({
                      id: 'ocp-express.Detail.Database.Cancel',
                      defaultMessage: '取消',
                    })}
                  </Button>
                  <Button
                    type="primary"
                    danger={true}
                    ghost={true}
                    disabled={!IsAllowDel}
                    loading={deleteDatabaseLoading}
                    onClick={handleDeleteDatabase}
                  >
                    {formatMessage({
                      id: 'ocp-express.Detail.Database.Delete',
                      defaultMessage: '删除',
                    })}
                  </Button>
                </>
              }
              onCancel={() => {
                setDeleteDatabaseModalVisible(false);
              }}
            >
              <Alert
                message={formatMessage({
                  id: 'ocp-express.Detail.Database.DataCannotBeRecoveredAfter',
                  defaultMessage: '数据库删除后数据将不可恢复，请谨慎操作',
                })}
                type="warning"
                showIcon
                style={{
                  marginBottom: 12,
                }}
              />

              <Descriptions column={1}>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Detail.Session.Statistics.Database',
                    defaultMessage: '数据库',
                  })}
                >
                  {currentDatabase?.dbName}
                </Descriptions.Item>
              </Descriptions>
              <div>
                {formatMessage({
                  id: 'ocp-express.Detail.Database.Enter',
                  defaultMessage: '请输入',
                })}
                <span style={{ color: 'red' }}> delete </span>
                {formatMessage({
                  id: 'ocp-express.Detail.Database.ConfirmOperation',
                  defaultMessage: '确认操作',
                })}
              </div>
              <Input style={{ width: 400, marginTop: 8 }} onChange={confirmOperation} />
            </Modal>

            <OBProxyAndConnectionStringModal
              width={900}
              visible={connectionStringModalVisible}
              obproxyAndConnectionStrings={currentDatabase?.connectionUrls || []}
              onCancel={() => {
                setConnectionStringModalVisible(false);
              }}
              onSuccess={() => {
                setConnectionStringModalVisible(false);
              }}
            />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};
export default Database;
