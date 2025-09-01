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
import { history, useSelector } from 'umi';
import React, { useState, useEffect } from 'react';
import {
  Col,
  Row,
  Table,
  Tooltip,
  Button,
  Space,
  Switch,
  Popconfirm,
  Descriptions,
  Modal,
  message,
  TableColumnType,
} from '@oceanbase/design';
import type { Route } from '@oceanbase/design/es/breadcrumb/Breadcrumb';
import { PageContainer } from '@oceanbase/ui';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import { uniq } from 'lodash';
import { PAGINATION_OPTION_10 } from '@/constant';
import { ORACLE_OBJECT_TYPE_LIST, ORACLE_BUILT_IN_ROLE_LIST } from '@/constant/tenant';
import { formatTime } from '@/util/datetime';
import { breadcrumbItemRender } from '@/util/component';
import MyInput from '@/component/MyInput';
import MyCard from '@/component/MyCard';
import BatchOperationBar from '@/component/BatchOperationBar';
import ModifyGlobalPrivilegeModal from '../Component/ModifyGlobalPrivilegeModal';
import ModifyRoleModal from '../Component/ModifyRoleModal';
import AddObjectPrivilegeDrawer from '../Component/AddObjectPrivilegeDrawer';
import ModifyObjectPrivilegeDrawer from '../Component/ModifyObjectPrivilegeDrawer';
import ModifyObjectPrivilegeModal from '../Component/ModifyObjectPrivilegeModal';
import ModifyDbUserPassword from '../../../Component/ModifyDbUserPassword';
import DeleteUserModal from '../../Component/DeleteUserModal';
import { DATE_FORMAT_DISPLAY } from '@/constant/datetime';
import {
  getRole,
  getUser,
  lockUser,
  revokeRoleObjectPrivilege,
  revokeUserObjectPrivilege,
  unlockUser,
  getStats,
} from '@/service/obshell/tenant';
import { UserStats } from '../UserList';

const { confirm } = Modal;

export interface UserOrRoleDetailProps {
  match: {
    params: {
      tenantName: string;
      username?: string;
      roleName?: string;
    };
  };
}

const UserOrRoleDetail: React.FC<UserOrRoleDetailProps> = ({
  match: {
    params: { username, roleName },
  },
}) => {
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const tenantName = tenantData.tenant_name;
  // 是否展示删除弹窗
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [modifyPasswordVisible, setModifyPasswordVisible] = useState(false);
  const [globalPrivVisible, setGlobalPrivVisible] = useState(false);
  const [roleVisible, setRoleVisible] = useState(false);
  const [modifydObjectVisible, setModifydObjectVisible] = useState(false);
  const [addObjectVisible, setAddObjectVisible] = useState(false);
  const [modifyObjectVisible, setModifyObjectVisible] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([]);
  const [selectedRows, setSelectedRows] = useState<API.ObjectPrivilege[]>([]);
  const [batchPopconfirmVisible, setBatchPopconfirmVisible] = useState(false);

  const [userStats, setUserStats] = useState<UserStats>({});

  const {
    run: getDbUser,
    data: dbUserData,
    refresh: refreshgGetDbUser,
  } = useRequest(getUser, {
    manual: true,
  });

  const userData: API.ObUser = dbUserData?.data || {};

  const {
    run: getDbRole,
    data: dbRoleData,
    refresh: refreshGetDbRole,
  } = useRequest(getRole, {
    manual: true,
  });

  const roleData: API.ObRole = dbRoleData?.data || {};
  const userOrRoleDetail = username ? userData : roleData;

  // 初始化默认值
  useEffect(() => {
    if (tenantName && username) {
      getDbUser({
        name: tenantName,
        user: username,
      });
    }
    if (tenantName && roleName) {
      getDbRole({
        name: tenantName,
        role: roleName,
      });
    }
  }, [tenantName, username, roleName]);

  const { run: getUserStats } = useRequest(getStats, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        setUserStats(res?.data?.session || {});
      }
    },
  });

  const { run: revokeUserObjectPrivilegeRun, loading: deleteDbRoleLoading } = useRequest(
    revokeUserObjectPrivilege,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage(
              {
                id: 'ocp-express.Oracle.UserOrRoleDetail.UsernameObjectPermissionDeleted',
                defaultMessage: '{username} 对象权限删除成功',
              },

              { username }
            )
          );

          setBatchPopconfirmVisible(false);
          setSelectedRowKeys([]);
          setSelectedRows([]);
          refreshgGetDbUser();
        }
      },
    }
  );

  const { run: revokeRoleObjectPrivilegeRun, loading: deleteDbUserLoading } = useRequest(
    revokeRoleObjectPrivilege,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage(
              {
                id: 'ocp-express.Oracle.UserOrRoleDetail.RolenameObjectPermissionDeleted',
                defaultMessage: '{roleName} 对象权限删除成功',
              },

              { roleName }
            )
          );

          setBatchPopconfirmVisible(false);
          setSelectedRowKeys([]);
          setSelectedRows([]);
          refreshGetDbRole();
        }
      },
    }
  );

  const deleteObjectPrivilege = (objectPrivileges: API.ObjectPrivilege[]) => {
    if (username) {
      revokeUserObjectPrivilegeRun(
        { name: tenantName, user: username },
        {
          object_privileges: objectPrivileges.map(item => ({
            object_name: item.object?.name,
            object_type: item.object?.type,
            owner: item.object?.owner,
            privileges: item.privileges,
          })),
        }
      );
    }
    if (roleName) {
      revokeRoleObjectPrivilegeRun(
        { name: tenantName, role: roleName },
        {
          object_privileges: objectPrivileges.map(item => ({
            object_name: item.object?.name,
            object_type: item.object?.type,
            owner: item.object?.owner,
            privileges: item.privileges,
          })),
        }
      );
    }
  };

  const { runAsync: unlockUserRun } = useRequest(unlockUser, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage(
            {
              id: 'ocp-express.Oracle.UserOrRoleDetail.TheUsernameUserIsUnlocked',
              defaultMessage: '{username} 用户解锁成功',
            },

            { username }
          )
        );

        refreshgGetDbUser();
      }
    },
  });

  const { runAsync: lockUserRun } = useRequest(lockUser, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage(
            {
              id: 'ocp-express.Oracle.UserOrRoleDetail.TheUsernameUserIsLocked',
              defaultMessage: '{username} 用户锁定成功',
            },

            { username }
          )
        );

        refreshgGetDbUser();
      }
    },
  });

  const changeLockedStatus = (record: API.ObUser) => {
    let iconType: any = null;
    if (!record.is_locked) {
      iconType = {
        title: formatMessage({
          id: 'ocp-express.Oracle.UserOrRoleDetail.LockedUsersAreNotAllowed',
          defaultMessage: '被锁定的用户将不允许登录，请谨慎操作',
        }),

        apiType: lockUserRun,
      };
    } else {
      iconType = {
        title: formatMessage({
          id: 'ocp-express.Oracle.UserOrRoleDetail.UnlockedUsersWillAllowThem',
          defaultMessage: '解锁用户将允许其登录',
        }),

        apiType: unlockUserRun,
      };
    }

    confirm({
      title: iconType.title,
      icon: iconType.icon,
      okText: formatMessage({
        id: 'ocp-express.Oracle.UserOrRoleDetail.Determine',
        defaultMessage: '确定',
      }),

      cancelText: formatMessage({
        id: 'ocp-express.Oracle.UserOrRoleDetail.Cancel',
        defaultMessage: '取消',
      }),

      content: (
        <div style={{ color: '#5C6B8A' }}>
          <div>
            {formatMessage(
              {
                id: 'ocp-express.Oracle.UserOrRoleDetail.TenantTenantdataname',
                defaultMessage: '租户：{tenantDataName}',
              },

              { tenantDataName: tenantName }
            )}
          </div>
          <div>
            {formatMessage(
              {
                id: 'ocp-express.Oracle.UserOrRoleDetail.UsernameRecordusername',
                defaultMessage: '用户名：{recordUsername}',
              },

              { recordUsername: record.user_name }
            )}
          </div>
        </div>
      ),

      onOk() {
        return iconType.apiType({ name: tenantName, user: record.user_name! });
      },
    });
  };

  const columns: TableColumnType<API.ObjectPrivilege>[] = [
    {
      title: formatMessage({
        id: 'ocp-express.Oracle.UserOrRoleDetail.ObjectName',
        defaultMessage: '对象名',
      }),

      dataIndex: 'object',
      render: (text: API.DbaObjectBo) => text?.name || '-',
    },

    {
      title: formatMessage({
        id: 'ocp-express.Oracle.UserOrRoleDetail.User',
        defaultMessage: '所属用户',
      }),

      dataIndex: 'object',
      render: (text: API.DbaObjectBo) => text?.owner || '-',
    },

    {
      title: formatMessage({
        id: 'ocp-express.Oracle.UserOrRoleDetail.Type',
        defaultMessage: '所属类型',
      }),

      dataIndex: 'object',
      filters: ORACLE_OBJECT_TYPE_LIST.map(item => ({
        text: item.label,
        value: item.value,
      })),

      onFilter: (value: string, record: API.ObjectPrivilege) => record.object?.type === value,
      render: (text: API.DbaObjectBo) => findByValue(ORACLE_OBJECT_TYPE_LIST, text?.type).label,
    },

    {
      title: formatMessage({
        id: 'ocp-express.Oracle.UserOrRoleDetail.Permissions',
        defaultMessage: '权限',
      }),

      dataIndex: 'privileges',
      render: (text: API.ObjectPrivilegeType[]) => text?.join('、') || '-',
    },

    {
      title: formatMessage({
        id: 'ocp-express.Oracle.UserOrRoleDetail.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      render: (text: string, record: API.ObjectPrivilege) => (
        <Space size="middle">
          <a
            onClick={() => {
              setSelectedRows([record]);
              setModifydObjectVisible(true);
            }}
          >
            {formatMessage({
              id: 'ocp-express.Oracle.UserOrRoleDetail.ModifyPermissions',
              defaultMessage: '修改权限',
            })}
          </a>
          <Popconfirm
            placement="topRight"
            title={formatMessage({
              id: 'ocp-express.Oracle.UserOrRoleDetail.AfterTheObjectIsDeleted',
              defaultMessage: '删除后，该用户将无此对象的权限，确定要删除吗？',
            })}
            okButtonProps={{
              loading: deleteDbUserLoading || deleteDbRoleLoading,
              danger: true,
              ghost: true,
            }}
            okText={formatMessage({
              id: 'ocp-express.Oracle.UserOrRoleDetail.Delete',
              defaultMessage: '删除',
            })}
            onConfirm={() => deleteObjectPrivilege([record])}
          >
            <a>
              {formatMessage({
                id: 'ocp-express.Oracle.UserOrRoleDetail.Delete',
                defaultMessage: '删除',
              })}
            </a>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const rowSelection = {
    selectedRows,
    selectedRowKeys,
    onChange: (keys: number[], newSelectedRows: API.ObjectPrivilege) => {
      setSelectedRowKeys(keys as number[]);
      setSelectedRows(newSelectedRows);
    },
    getCheckboxProps: record => ({
      objectName: record.objectName,
    }),
  };

  const batchModifyBtn =
    uniq(selectedRows?.map(item => item?.object?.type)).length > 1 ? (
      <Tooltip
        placement="top"
        title={
          uniq(selectedRows?.map(item => item?.object?.type)).length > 1 &&
          formatMessage({
            id: 'ocp-express.Oracle.UserOrRoleDetail.YouCannotModifyPermissionsAt',
            defaultMessage: '选择的对象类型不同，不支持批量修改权限',
          })
        }
      >
        <Button disabled={uniq(selectedRows?.map(item => item?.object?.type)).length > 1}>
          {formatMessage({
            id: 'ocp-express.Oracle.UserOrRoleDetail.ModifyPermissions.2',
            defaultMessage: '批量修改权限',
          })}
        </Button>
      </Tooltip>
    ) : (
      <Button onClick={() => setModifyObjectVisible(true)}>
        {formatMessage({
          id: 'ocp-express.Oracle.UserOrRoleDetail.ModifyPermissions.2',
          defaultMessage: '批量修改权限',
        })}
      </Button>
    );

  const routes: Route[] = [
    {
      path: `/cluster/tenant/${tenantName}/user`,
      breadcrumbName: formatMessage({
        id: 'ocp-express.Oracle.UserOrRoleDetail.UserManagement',
        defaultMessage: '用户管理',
      }),
    },

    {
      path: `/cluster/tenant/${tenantName}/user${roleName ? '/role' : ''}`,
      breadcrumbName: roleName
        ? formatMessage({
          id: 'ocp-express.Oracle.UserOrRoleDetail.RoleList',
          defaultMessage: '角色列表',
        })
        : formatMessage({
          id: 'ocp-express.Oracle.UserOrRoleDetail.UserList',
          defaultMessage: '用户列表',
        }),
    },

    {
      breadcrumbName: username || roleName,
    },
  ];

  return (
    <PageContainer
      ghost={true}
      header={{
        title: `${username || roleName}`,
        extra:
          userData?.user_name !== 'SYS' && userData?.user_name !== 'proxyro' ? (
            <Space>
              {roleName && ORACLE_BUILT_IN_ROLE_LIST.includes(roleName) ? (
                ''
              ) : (
                <Button
                  onClick={() => {
                    if (username) {
                      getUserStats({ name: tenantName, user: username });
                    }
                    setDeleteModalVisible(true);
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.Delete',
                    defaultMessage: '删除',
                  })}
                </Button>
              )}

              {username && (
                <Button
                  type="primary"
                  onClick={() => {
                    setModifyPasswordVisible(true);
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.ChangePassword',
                    defaultMessage: '修改密码',
                  })}
                </Button>
              )}
            </Space>
          ) : (
            <Space>
              {username && (
                <Button
                  type="primary"
                  onClick={() => {
                    setModifyPasswordVisible(true);
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.ChangePassword',
                    defaultMessage: '修改密码',
                  })}
                </Button>
              )}
            </Space>
          ),

        breadcrumb: { routes, itemRender: breadcrumbItemRender },
        onBack: () => {
          history.goBack();
        },
      }}
    >
      <Row gutter={[16, 16]}>
        {username && (
          <Col span={24}>
            <MyCard
              title={formatMessage({
                id: 'ocp-express.Oracle.UserOrRoleDetail.BasicInformation',
                defaultMessage: '基本信息',
              })}
              bordered={false}
            >
              <Descriptions>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.Locking',
                    defaultMessage: '锁定',
                  })}
                >
                  <Tooltip
                    placement="topLeft"
                    title={
                      userData?.user_name === 'SYS' &&
                      formatMessage({
                        id: 'ocp-express.Oracle.UserOrRoleDetail.SysUsersDoNotSupport',
                        defaultMessage: 'SYS 用户不支持此操作',
                      })
                    }
                  >
                    <Switch
                      disabled={userData?.user_name === 'SYS'}
                      checked={userData?.is_locked}
                      size="small"
                      onChange={() => {
                        changeLockedStatus(userData);
                      }}
                    />
                  </Tooltip>
                </Descriptions.Item>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.CreationTime',
                    defaultMessage: '新建时间',
                  })}
                >
                  {formatTime(userData?.create_time, DATE_FORMAT_DISPLAY)}
                </Descriptions.Item>
              </Descriptions>
            </MyCard>
          </Col>
        )}
        <Col span={24}>
          <MyCard
            title={formatMessage({
              id: 'ocp-express.Oracle.UserOrRoleDetail.HaveSystemPermissions',
              defaultMessage: '拥有系统权限',
            })}
            bordered={false}
            extra={
              userData?.user_name !== 'SYS' && (
                <Button onClick={() => setGlobalPrivVisible(true)}>
                  {formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.ModifySystemPermissions',
                    defaultMessage: '修改系统权限',
                  })}
                </Button>
              )
            }
          >
            {userOrRoleDetail?.global_privileges
              ? userOrRoleDetail?.global_privileges
                ?.map(item =>
                  item === 'PURGE_DBA_RECYCLEBIN'
                    ? 'PURGE DBA_RECYCLEBIN'
                    : item.replace(/_/g, ' ')
                )
                .join('、')
              : '-'}
          </MyCard>
        </Col>
        <Col span={24}>
          <MyCard
            title={formatMessage({
              id: 'ocp-express.Oracle.UserOrRoleDetail.HaveARole',
              defaultMessage: '拥有角色',
            })}
            bordered={false}
            extra={
              userData?.user_name !== 'SYS' && (
                <Button onClick={() => setRoleVisible(true)}>
                  {formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.ModifyARole',
                    defaultMessage: '修改角色',
                  })}
                </Button>
              )
            }
          >
            {userOrRoleDetail?.granted_roles ? userOrRoleDetail?.granted_roles.join('、') : '-'}
          </MyCard>
        </Col>
        <Col span={24}>
          <MyCard
            title={formatMessage({
              id: 'ocp-express.Oracle.UserOrRoleDetail.AccessibleObjects',
              defaultMessage: '可访问对象',
            })}
            bordered={false}
            extra={
              <Space>
                <MyInput.Search
                  allowClear={true}
                  onSearch={(value: string) => setKeyword(value)}
                  placeholder={formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.EnterAnObjectNameOr',
                    defaultMessage: '请输入对象名或用户名',
                  })}
                  className="search-input"
                />
                {userData?.user_name !== 'SYS' && (
                  <Button
                    onClick={() => {
                      setAddObjectVisible(true);
                    }}
                  >
                    {formatMessage({
                      id: 'ocp-express.Oracle.UserOrRoleDetail.AddObjects',
                      defaultMessage: '添加对象',
                    })}
                  </Button>
                )}
              </Space>
            }
          >
            {selectedRowKeys && selectedRowKeys.length > 0 && (
              <BatchOperationBar
                size="small"
                selectedCount={selectedRowKeys && selectedRowKeys.length}
                onCancel={() => {
                  setSelectedRows([]);
                  setSelectedRowKeys([]);
                }}
                actions={[
                  batchModifyBtn,
                  <Tooltip
                    key="batch-delete"
                    placement="topRight"
                    title={
                      userData?.user_name === 'SYS' &&
                      formatMessage({
                        id: 'ocp-express.Oracle.UserOrRoleDetail.SysUsersDoNotSupport',
                        defaultMessage: 'SYS 用户不支持此操作',
                      })
                    }
                  >
                    <Popconfirm
                      placement="topRight"
                      title={formatMessage(
                        {
                          id: 'ocp-express.Oracle.UserOrRoleDetail.AfterTheObjectIsDeleted',
                        },

                        { selectedRowsLength: selectedRows.length }
                      )}
                      okButtonProps={{
                        loading: deleteDbUserLoading || deleteDbRoleLoading,
                        danger: true,
                        ghost: true,
                      }}
                      okText={formatMessage({
                        id: 'ocp-express.Oracle.UserOrRoleDetail.Delete',
                        defaultMessage: '删除',
                      })}
                      visible={batchPopconfirmVisible}
                      onVisibleChange={() => {
                        setBatchPopconfirmVisible(true);
                      }}
                      onConfirm={() => deleteObjectPrivilege(selectedRows)}
                      onCancel={() => setBatchPopconfirmVisible(false)}
                    >
                      <Button disabled={userData?.user_name === 'SYS'}>
                        {formatMessage({
                          id: 'ocp-express.Oracle.UserOrRoleDetail.BatchDelete',
                          defaultMessage: '批量删除',
                        })}
                      </Button>
                    </Popconfirm>
                  </Tooltip>,
                ]}
                style={{ marginBottom: 16 }}
              />
            )}

            <Table
              columns={columns}
              rowKey={record => record.object?.full_name!}
              dataSource={userOrRoleDetail?.object_privileges?.filter(
                item =>
                  !keyword ||
                  (item?.object?.name && item.object?.name.includes(keyword)) ||
                  (item?.object?.owner && item.object?.owner.includes(keyword))
              )}
              pagination={PAGINATION_OPTION_10}
              rowSelection={userData?.user_name !== 'SYS' && rowSelection}
            />
          </MyCard>
        </Col>
        {roleName && (
          <Col span={24}>
            <MyCard
              title={formatMessage({
                id: 'ocp-express.Oracle.UserOrRoleDetail.ReferencedByTheFollowingRoles',
                defaultMessage: '被以下角色及用户引用',
              })}
              bordered={false}
            >
              <Descriptions column={1}>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.Role',
                    defaultMessage: '角色',
                  })}
                >
                  {userOrRoleDetail?.role_grantees?.length > 0
                    ? userOrRoleDetail?.role_grantees.join('、')
                    : '-'}
                </Descriptions.Item>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Oracle.UserOrRoleDetail.User.1',
                    defaultMessage: '用户',
                  })}
                >
                  {userOrRoleDetail?.user_grantees?.length > 0
                    ? userOrRoleDetail?.user_grantees.join('、')
                    : '-'}
                </Descriptions.Item>
              </Descriptions>
            </MyCard>
          </Col>
        )}
      </Row>
      <DeleteUserModal
        visible={deleteModalVisible}
        roleName={roleName}
        username={username}
        tenantData={tenantData}
        userStats={userStats}
        onCancel={() => {
          setDeleteModalVisible(false);
        }}
        onSuccess={() => {
          setDeleteModalVisible(false);
          history.goBack();
        }}
      />

      <ModifyDbUserPassword
        visible={modifyPasswordVisible}
        dbUser={userOrRoleDetail}
        tenantData={tenantData}
        onCancel={() => {
          setModifyPasswordVisible(false);
        }}
        onSuccess={() => {
          message.success(
            formatMessage({
              id: 'ocp-express.Oracle.UserOrRoleDetail.ModifyPasswordSuccess',
              defaultMessage: '修改密码成功',
            })
          );
          setModifyPasswordVisible(false);
          refreshGetDbRole();
        }}
      />

      <ModifyGlobalPrivilegeModal
        visible={globalPrivVisible}
        tenantName={tenantName}
        username={username}
        roleName={roleName}
        globalPrivileges={userOrRoleDetail?.global_privileges || []}
        onCancel={() => {
          setGlobalPrivVisible(false);
        }}
        onSuccess={() => {
          setGlobalPrivVisible(false);
          if (username) {
            refreshgGetDbUser();
          }
          if (roleName) {
            refreshGetDbRole();
          }
        }}
      />

      <ModifyRoleModal
        visible={roleVisible}
        tenantName={tenantName}
        username={username}
        roleName={roleName}
        grantedRoles={userOrRoleDetail?.granted_roles}
        onCancel={() => {
          setRoleVisible(false);
        }}
        onSuccess={() => {
          setRoleVisible(false);
          if (username) {
            refreshgGetDbUser();
          }
          if (roleName) {
            refreshGetDbRole();
          }
        }}
      />

      <ModifyObjectPrivilegeModal
        visible={modifydObjectVisible}
        tenantName={tenantName}
        username={username}
        roleName={roleName}
        dbObject={selectedRows && selectedRows[0]}
        onCancel={() => {
          setModifydObjectVisible(false);
          setSelectedRows([]);
          setSelectedRowKeys([]);
        }}
        onSuccess={() => {
          setModifydObjectVisible(false);
          setSelectedRows([]);
          setSelectedRowKeys([]);
          if (username) {
            refreshgGetDbUser();
          }
          if (roleName) {
            refreshGetDbRole();
          }
        }}
      />

      <AddObjectPrivilegeDrawer
        visible={addObjectVisible}
        tenantName={tenantName}
        username={username}
        roleName={roleName}
        addedDbObjects={userOrRoleDetail?.object_privileges}
        onCancel={() => {
          setAddObjectVisible(false);
          setSelectedRowKeys([]);
          setSelectedRows([]);
        }}
        onSuccess={() => {
          setAddObjectVisible(false);
          setSelectedRowKeys([]);
          setSelectedRows([]);
          if (username) {
            refreshgGetDbUser();
          }
          if (roleName) {
            refreshGetDbRole();
          }
        }}
      />

      <ModifyObjectPrivilegeDrawer
        visible={modifyObjectVisible}
        tenantName={tenantName}
        username={username}
        roleName={roleName}
        dbObjects={selectedRows}
        onCancel={() => {
          setModifyObjectVisible(false);
          setSelectedRowKeys([]);
          setSelectedRows([]);
        }}
        onSuccess={() => {
          setModifyObjectVisible(false);
          setSelectedRowKeys([]);
          setSelectedRows([]);
          if (username) {
            refreshgGetDbUser();
          }
          if (roleName) {
            refreshGetDbRole();
          }
        }}
      />
    </PageContainer>
  );
};

export default UserOrRoleDetail;
