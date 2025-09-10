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
import { useSelector } from 'umi';
import React, { useState } from 'react';
import {
  Table,
  Col,
  Row,
  Card,
  Tooltip,
  Button,
  Space,
  Switch,
  Modal,
  message,
} from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import { PAGINATION_OPTION_10 } from '@/constant';
import ContentWithReload from '@/component/ContentWithReload';
import MyInput from '@/component/MyInput';
import AddUserDrawer from '../../Component/AddUserDrawer';
import ModifyDbUserPassword from '../../Component/ModifyDbUserPassword';
import DeleteUserModal from '../Component/DeleteUserModal';
import OBProxyAndConnectionStringModal from '../../Component/OBProxyAndConnectionStringModal';
import RenderConnectionString from '@/component/RenderConnectionString';
import { listUsers, lockUser, unlockUser, getStats } from '@/service/obshell/tenant';
import useUiMode from '@/hook/useUiMode';

export interface IndexProps {
  tenantName: string;
}

const Index: React.FC<IndexProps> = ({ tenantName }) => {
  const { tenantData, precheckResult } = useSelector((state: DefaultRootState) => state.tenant);

  const [keyword, setKeyword] = useState('');
  const [connectionStringModalVisible, setConnectionStringModalVisible] = useState(false);
  const [dbUser, setDbUser] = useState<API.DbUser | null>(null);
  const [userStats, setUserStats] = useState<API.SessionUserStats[]>({});

  // 修改参数值的抽屉是否可见
  const [valueVisible, setValueVisible] = useState(false);
  // 修改密码
  const [modifyPasswordVisible, setModifyPasswordVisible] = useState(false);
  // 删除Modal
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);

  const { isDesktopMode } = useUiMode();

  const ready =
    tenantData?.tenant_name === 'sys' ||
    (Object.keys(precheckResult)?.length > 0 && precheckResult?.is_connectable);

  const {
    data: dbUserListData,
    refresh,
    loading,
  } = useRequest(listUsers, {
    ready,
    defaultParams: [
      {
        name: tenantName,
      },
    ],
    refreshDeps: [ready],
  });

  const dbUserList = (dbUserListData?.data?.contents || [])?.map(item => ({
    ...item,
    username: item?.user_name,
    isLocked: item?.is_locked,
    accessibleDatabases: item?.accessible_databases,
    connectionStrings: item?.connection_strings,
    dbPrivileges: item?.db_privileges?.map(db_privilege => ({
      ...db_privilege,
      dbName: db_privilege?.db_name,
    })),
    globalPrivileges: item?.global_privileges,
  }));

  const dataSource = dbUserList?.filter(
    item => !keyword || (item.username && item.username.includes(keyword))
  );

  const { run } = useRequest(getStats, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        setUserStats(res?.data?.session || {});
      }
    },
  });

  const modifyPassword = (record: API.DbUser) => {
    setModifyPasswordVisible(true);
    setDbUser(record);
  };

  const modifyPrivileges = (record: API.DbUser) => {
    setValueVisible(true);
    setDbUser(record);
  };

  const { runAsync: unlockDbUser } = useRequest(unlockUser, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.User.MySQL.TheUserIsUnlocked',
            defaultMessage: '用户解锁成功',
          })
        );

        refresh();
      }
    },
  });

  const { runAsync: lockDbUser } = useRequest(lockUser, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.User.MySQL.TheUserIsLocked',
            defaultMessage: '用户锁定成功',
          })
        );

        refresh();
      }
    },
  });

  const changeLockedStatus = (record: API.DbUser) => {
    if (record.isLocked) {
      Modal.confirm({
        title: formatMessage({
          id: 'ocp-express.User.MySQL.UnlockedUsersWillAllowThem',
          defaultMessage: '解锁用户将允许其登录',
        }),

        content: (
          <div style={{ color: '#5C6B8A' }}>
            <div>
              {formatMessage(
                {
                  id: 'ocp-express.User.MySQL.TenantTenantdataname',
                  defaultMessage: '租户：{tenantDataName}',
                },

                { tenantDataName: tenantData.tenant_name }
              )}
            </div>
            <div>
              {formatMessage(
                {
                  id: 'ocp-express.User.MySQL.UsernameRecordusername',
                  defaultMessage: '用户名：{recordUsername}',
                },

                { recordUsername: record.username }
              )}
            </div>
          </div>
        ),

        onOk: () => {
          return unlockDbUser({ name: tenantName, user: record.username });
        },
      });
    } else {
      Modal.confirm({
        title: formatMessage({
          id: 'ocp-express.User.MySQL.LockedUsersAreNotAllowed',
          defaultMessage: '被锁定的用户将不允许登录，请谨慎操作',
        }),
        content: (
          <div style={{ color: '#5C6B8A' }}>
            <div>
              {formatMessage(
                {
                  id: 'ocp-express.User.MySQL.TenantTenantdataname',
                  defaultMessage: '租户： {tenantDataName}',
                },

                { tenantDataName: tenantData.tenant_name }
              )}
            </div>
            <div>
              {formatMessage(
                {
                  id: 'ocp-express.User.MySQL.UsernameRecordusername',
                  defaultMessage: '用户名： {recordUsername}',
                },

                { recordUsername: record.username }
              )}
            </div>
          </div>
        ),

        onOk: () => {
          return lockDbUser({ name: tenantName, user: record.username });
        },
      });
    }
  };

  const columns = [
    {
      title: formatMessage({ id: 'ocp-express.User.MySQL.UserName', defaultMessage: '用户名' }),
      dataIndex: 'username',
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.MySQL.AccessibleDatabases',
        defaultMessage: '可访问数据库',
      }),

      dataIndex: 'accessibleDatabases',
      render: (text: string[], record: API.DbUser) => {
        if (record.username === 'root' || record.username === 'proxyro') {
          return '*';
        }
        const textContent = text?.join('、');
        return text?.length > 0 ? (
          <Tooltip placement="topLeft" title={textContent}>
            {textContent}
          </Tooltip>
        ) : (
          '-'
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.MySQL.LogonConnectionString',
        defaultMessage: '登录连接串',
      }),

      dataIndex: 'connectionStrings',
      render: (connectionStrings: API.ObproxyAndConnectionString[], record: API.DbUser) => {
        return (
          <RenderConnectionString
            callBack={() => {
              setDbUser(record);
              setConnectionStringModalVisible(true);
            }}
            connectionStrings={connectionStrings}
          />
        );
      },
    },

    {
      title: formatMessage({ id: 'ocp-express.User.MySQL.Locking', defaultMessage: '锁定' }),
      dataIndex: 'isLocked',
      render: (text: boolean, record: API.DbUser) =>
        record.username === 'root' || record.username === 'proxyro' ? (
          <Switch size="small" checked={text} disabled={true} />
        ) : (
          <Switch
            data-aspm-click="c304264.d308789"
            data-aspm-desc="MySQL 用户列表-切换锁定状态"
            data-aspm-param={``}
            data-aspm-expo
            onClick={() => changeLockedStatus(record)}
            checked={text}
            size="small"
          />
        ),
    },
    {
      title: formatMessage({
        id: 'ocp-express.User.MySQL.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      render: (text: string, record: API.DbUser) => {
        return (
          <Space size="middle">
            {(!isDesktopMode ||
              (record.username !== 'root' && tenantData?.tenant_name !== 'sys')) && (
              <>
                <a
                  data-aspm-click="c304264.d308787"
                  data-aspm-desc="MySQL 用户列表-修改密码"
                  data-aspm-param={``}
                  data-aspm-expo
                  onClick={() => {
                    modifyPassword(record);
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.User.MySQL.ChangePassword',
                    defaultMessage: '修改密码',
                  })}
                </a>
              </>
            )}
            {record.username !== 'root' && (
              <>
                <a
                  data-aspm-click="c304264.d308784"
                  data-aspm-desc="MySQL 用户列表-修改权限"
                  data-aspm-param={``}
                  data-aspm-expo
                  onClick={() => {
                    modifyPrivileges(record);
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.User.MySQL.ModifyPermissions',
                    defaultMessage: '修改权限',
                  })}
                </a>
                <a
                  data-aspm-click="c304264.d308786"
                  data-aspm-desc="MySQL 用户列表-删除用户"
                  data-aspm-param={``}
                  data-aspm-expo
                  onClick={() => {
                    run({
                      name: tenantName,
                      user: record?.username,
                    });
                    setDbUser(record);
                    setDeleteModalVisible(true);
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.User.MySQL.Delete',
                    defaultMessage: '删除',
                  })}
                </a>
              </>
            )}
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
              id: 'ocp-express.User.MySQL.UserManagement',
              defaultMessage: '用户管理',
            })}
            spin={loading}
            onClick={() => {
              refresh();
            }}
          />
        ),
        extra: (
          <Button
            data-aspm-click="c304264.d308782"
            data-aspm-desc="MySQL 用户列表-新建用户"
            data-aspm-param={``}
            data-aspm-expo
            type="primary"
            onClick={() => {
              setValueVisible(true);
            }}
          >
            {formatMessage({
              id: 'ocp-express.User.MySQL.CreateUser',
              defaultMessage: '新建用户',
            })}
          </Button>
        ),
      }}
    >
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <div
            data-aspm="c304265"
            data-aspm-desc="MySQL 用户列表"
            data-aspm-param={``}
            data-aspm-expo
          >
            <Card
              title={formatMessage({
                id: 'ocp-express.User.MySQL.UserList',
                defaultMessage: '用户列表',
              })}
              bordered={false}
              className="card-without-padding"
              extra={
                <MyInput.Search
                  data-aspm-click="c304264.d308781"
                  data-aspm-desc="MySQL 用户列表-搜索用户"
                  data-aspm-param={``}
                  data-aspm-expo
                  allowClear={true}
                  onSearch={(value: string) => setKeyword(value)}
                  placeholder={formatMessage({
                    id: 'ocp-express.User.MySQL.SearchUserName',
                    defaultMessage: '搜索用户名',
                  })}
                  className="search-input"
                />
              }
            >
              <Table
                columns={columns}
                rowKey="id"
                loading={loading}
                dataSource={dataSource}
                pagination={PAGINATION_OPTION_10}
              />

              <AddUserDrawer
                visible={valueVisible}
                dbUser={dbUser}
                onCancel={() => {
                  setDbUser(null);
                  setValueVisible(false);
                }}
                onSuccess={() => {
                  setDbUser(null);
                  setValueVisible(false);
                  refresh();
                }}
              />

              <ModifyDbUserPassword
                visible={modifyPasswordVisible}
                dbUser={dbUser}
                tenantData={tenantData}
                onCancel={() => {
                  setModifyPasswordVisible(false);
                  setDbUser(null);
                }}
                onSuccess={() => {
                  setModifyPasswordVisible(false);
                  setDbUser(null);
                }}
              />

              <DeleteUserModal
                visible={deleteModalVisible}
                username={dbUser?.username}
                tenantData={tenantData}
                userStats={userStats}
                onCancel={() => {
                  setDeleteModalVisible(false);
                  setDbUser(null);
                }}
                onSuccess={() => {
                  setDeleteModalVisible(false);
                  setDbUser(null);
                  refresh();
                }}
              />

              <OBProxyAndConnectionStringModal
                width={900}
                userName={dbUser?.username}
                visible={connectionStringModalVisible}
                obproxyAndConnectionStrings={dbUser?.connectionStrings || []}
                onCancel={() => {
                  setConnectionStringModalVisible(false);
                }}
                onSuccess={() => {
                  setConnectionStringModalVisible(false);
                }}
              />
            </Card>
          </div>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Index;
