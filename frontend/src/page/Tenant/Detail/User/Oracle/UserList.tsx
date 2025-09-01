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
import { history, connect } from 'umi';
import React, { useState } from 'react';
import { Table, Tooltip, Space, Switch, Modal, message, TableColumnType } from '@oceanbase/design';
import { uniq } from 'lodash';
import { sortByMoment } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import { PAGINATION_OPTION_10 } from '@/constant';
import { DATE_FORMAT_DISPLAY } from '@/constant/datetime';
import { formatTime } from '@/util/datetime';
import ModifyDbUserPassword from '../../Component/ModifyDbUserPassword';
import DeleteUserModal from '../Component/DeleteUserModal';
import OBProxyAndConnectionStringModal from '../../Component/OBProxyAndConnectionStringModal';
import RenderConnectionString from '@/component/RenderConnectionString';
import { getStats, lockUser, unlockUser } from '@/service/obshell/tenant';

const { confirm } = Modal;

export interface UserProps {
  dbUserLoading: boolean;
  dbUserList: [];
  refreshDbUser: () => void;
  tenantData: API.DbaObTenant;
}

export interface UserStats {
  total?: number;
  active?: number;
}

const UserList: React.FC<UserProps> = ({
  dbUserLoading,
  dbUserList,
  refreshDbUser,
  tenantData,
}) => {
  const [current, setCurrent] = useState<API.ObUser | null>(null);
  const [userStats, setUserStats] = useState<UserStats>({});
  // 修改密码
  const [modifyPasswordVisible, setModifyPasswordVisible] = useState(false);
  // 删除Modal
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [connectionStringModalVisible, setConnectionStringModalVisible] = useState(false);

  const tenantName = tenantData.tenant_name || '';

  const { run: getUserStats } = useRequest(getStats, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        setUserStats(res?.data?.session || {});
      }
    },
  });

  const modifyPassword = (record: API.ObUser) => {
    setModifyPasswordVisible(true);
    setCurrent(record);
  };

  const { runAsync: unlockDbUser } = useRequest(unlockUser, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.User.Oracle.UserList.TheUserIsUnlocked',
            defaultMessage: '用户解锁成功',
          })
        );
        refreshDbUser();
      }
    },
  });

  const { runAsync: lockDbUser } = useRequest(lockUser, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.User.Oracle.UserList.TheUserIsLocked',
            defaultMessage: '用户锁定成功',
          })
        );
        refreshDbUser();
      }
    },
  });

  const changeLockedStatus = (record: API.ObUser) => {
    let iconType: any = null;
    if (!record.is_locked) {
      iconType = {
        title: formatMessage({
          id: 'ocp-express.User.Oracle.UserList.LockedUsersAreNotAllowed',
          defaultMessage: '被锁定的用户将不允许登录，请谨慎操作',
        }),
        apiType: lockDbUser,
      };
    } else {
      iconType = {
        title: formatMessage({
          id: 'ocp-express.User.Oracle.UserList.UnlockedUsersWillAllowThem',
          defaultMessage: '解锁用户将允许其登录',
        }),
        apiType: unlockDbUser,
      };
    }
    confirm({
      title: iconType.title,
      icon: iconType.icon,
      okText: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.Determine',
        defaultMessage: '确定',
      }),
      cancelText: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.Cancel',
        defaultMessage: '取消',
      }),
      content: (
        <div style={{ color: '#5C6B8A' }}>
          <div>
            {formatMessage(
              {
                id: 'ocp-express.User.Oracle.UserList.TenantTenantdataname',
                defaultMessage: '租户：{tenantDataName}',
              },
              { tenantDataName: tenantName }
            )}
          </div>
          <div>
            {formatMessage(
              {
                id: 'ocp-express.User.Oracle.UserList.UsernameRecordusername',
                defaultMessage: '用户名：{recordUsername}',
              },
              { recordUsername: record.user_name }
            )}
          </div>
        </div>
      ),
      onOk() {
        return iconType.apiType({ name: tenantName, user: record.user_name });
      },
    });
  };

  const columns: TableColumnType<API.ObUser>[] = [
    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.UserName',
        defaultMessage: '用户名',
      }),
      dataIndex: 'user_name',
      render: (text: string, record: API.ObUser) => (
        <Tooltip placement="topLeft" title={text}>
          <a onClick={() => history.push(`/cluster/tenant/${tenantName}/user/${record?.user_name}`)}>
            {text}
          </a>
        </Tooltip>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.HaveSystemPermissions',
        defaultMessage: '拥有系统权限',
      }),
      dataIndex: 'global_privileges',
      render: (text: string[]) => {
        const textContent = text
          ?.map(item =>
            item === 'PURGE_DBA_RECYCLEBIN' ? 'PURGE DBA_RECYCLEBIN' : item.replace(/_/g, ' ')
          )
          .join('、');
        return textContent ? (
          <Tooltip placement="topLeft" title={textContent}>
            <span
              style={{
                maxWidth: 180,
              }}
              className="ellipsis"
            >
              {textContent}
            </span>
          </Tooltip>
        ) : (
          '-'
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.HaveARole',
        defaultMessage: '拥有角色',
      }),
      dataIndex: 'granted_roles',
      render: (text: string[]) => {
        const textContent = text?.join('、');
        return textContent ? (
          <Tooltip placement="topLeft" title={textContent}>
            <span
              style={{
                maxWidth: 180,
              }}
              className="ellipsis"
            >
              {textContent}
            </span>
          </Tooltip>
        ) : (
          '-'
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.RolesList.AccessibleObjects',
        defaultMessage: '可访问对象',
      }),
      dataIndex: 'object_privileges',
      render: (text: API.ObjectPrivilege[]) => {
        const objectPrivileges = uniq(text?.map(item => item?.object?.full_name));
        const textContent = objectPrivileges?.join('、');
        return textContent ? (
          <Tooltip placement="topLeft" title={textContent}>
            <span
              style={{
                maxWidth: 180,
              }}
              className="ellipsis"
            >
              {textContent}
            </span>
          </Tooltip>
        ) : (
          '-'
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.CreationTime',
        defaultMessage: '新建时间',
      }),
      dataIndex: 'create_time',
      defaultSortOrder: 'descend',
      sorter: (a: API.ObUser, b: API.ObUser) => sortByMoment(a, b, 'create_time'),
      render: (text: string) => formatTime(text, DATE_FORMAT_DISPLAY),
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.LogonConnectionString',
        defaultMessage: '登录连接串',
      }),
      dataIndex: 'connection_strings',
      render: (connectionStrings: API.ObproxyAndConnectionString[], record: API.ObUser) => {
        return (
          <RenderConnectionString
            callBack={() => {
              setCurrent(record);
              setConnectionStringModalVisible(true);
            }}
            connectionStrings={connectionStrings}
          />
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.Locking',
        defaultMessage: '锁定',
      }),
      dataIndex: 'is_locked',
      render: (text: boolean, record: API.ObUser) =>
        record.user_name === 'SYS' ? (
          <Tooltip
            placement="topRight"
            title={formatMessage({
              id: 'ocp-express.User.Oracle.UserList.YouCannotLockUnlockOr',
              defaultMessage: 'SYS 用户（相当于 MySQL 租户的 root 用户）不支持锁定',
            })}
          >
            <Switch size="small" checked={text} disabled={true} />
          </Tooltip>
        ) : (
          <Switch
            data-aspm-click="c304262.d308778"
            data-aspm-desc="Oracle 用户列表-切换锁定状态"
            data-aspm-param={``}
            data-aspm-expo
            checked={text}
            size="small"
            onChange={() => {
              changeLockedStatus(record);
            }}
          />
        ),
    },
    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.UserList.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      render: (text: string, record: API.ObUser) => {
        return (
          <Space size="middle">
            <a
              data-aspm-click="c304262.d308776"
              data-aspm-desc="Oracle 用户列表-修改密码"
              data-aspm-param={``}
              data-aspm-expo
              onClick={() => {
                modifyPassword(record);
              }}
            >
              {formatMessage({
                id: 'ocp-express.User.Oracle.UserList.ChangePassword',
                defaultMessage: '修改密码',
              })}
            </a>
            {record.user_name !== 'SYS' && (
              <a
                data-aspm-click="c304262.d308775"
                data-aspm-desc="Oracle 用户列表-删除用户"
                data-aspm-param={``}
                data-aspm-expo
                onClick={() => {
                  getUserStats({ name: tenantName, user: record.user_name! });
                  setCurrent(record);
                  setDeleteModalVisible(true);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.User.Oracle.UserList.Delete',
                  defaultMessage: '删除',
                })}
              </a>
            )}
          </Space>
        );
      },
    },
  ];

  return (
    <>
      <Table
        data-aspm="c304262"
        data-aspm-desc="Oracle 用户列表"
        data-aspm-param={``}
        data-aspm-expo
        columns={columns}
        rowKey="id"
        loading={dbUserLoading}
        dataSource={dbUserList}
        pagination={PAGINATION_OPTION_10}
      />

      <ModifyDbUserPassword
        visible={modifyPasswordVisible}
        dbUser={current}
        tenantData={tenantData}
        onCancel={() => {
          setModifyPasswordVisible(false);
          setCurrent(null);
        }}
        onSuccess={() => {
          setModifyPasswordVisible(false);
          setCurrent(null);
        }}
      />

      <DeleteUserModal
        visible={deleteModalVisible}
        username={current?.user_name}
        tenantData={tenantData}
        userStats={userStats}
        onCancel={() => {
          setDeleteModalVisible(false);
          setCurrent(null);
        }}
        onSuccess={() => {
          setDeleteModalVisible(false);
          setCurrent(null);
          refreshDbUser();
        }}
      />

      <OBProxyAndConnectionStringModal
        userName={current?.user_name}
        width={900}
        visible={connectionStringModalVisible}
        obproxyAndConnectionStrings={current?.connection_strings || []}
        onCancel={() => {
          setConnectionStringModalVisible(false);
        }}
        onSuccess={() => {
          setConnectionStringModalVisible(false);
        }}
      />
    </>
  );
};

function mapStateToProps({ tenant }) {
  return {
    tenantData: tenant.tenantData,
  };
}

export default connect(mapStateToProps)(UserList);
