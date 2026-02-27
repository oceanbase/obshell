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

import { PAGINATION_OPTION_10 } from '@/constant';
import { ORACLE_BUILT_IN_ROLE_LIST } from '@/constant/tenant';
import { formatMessage } from '@/util/intl';
import { Table, TableColumnType, Tooltip } from '@oceanbase/design';
import { connect, history } from '@umijs/max';
import { uniq } from 'lodash';
import React, { useState } from 'react';
import DeleteUserModal from '../Component/DeleteUserModal';

export interface RoleListProps {
  dbRoleLoading?: boolean;
  dbRoleList: [];
  refreshDbRole?: () => void;
  tenantData: API.DbaObTenant;
}

const RoleList: React.FC<RoleListProps> = ({
  dbRoleLoading,
  dbRoleList,
  refreshDbRole,
  tenantData,
}) => {
  const tenantName = tenantData.tenant_name || '';
  const [current, setCurrent] = useState<API.ObRole | null>(null);
  // 删除 Modal
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);

  const columns: TableColumnType<API.ObRole>[] = [
    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.RolesList.Role',
        defaultMessage: '角色名',
      }),
      dataIndex: 'role',
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          <a onClick={() => history.push(`/cluster/tenant/${tenantName}/user/role/${text}`)}>
            {text}
          </a>
        </Tooltip>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.RolesList.HaveSystemPermissions',
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
        id: 'ocp-express.User.Oracle.RolesList.HaveARole',
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
        id: 'ocp-express.User.Oracle.RolesList.ReferencedByTheFollowingRoles',
        defaultMessage: '被以下角色和用户引用',
      }),
      dataIndex: 'role_grantees',
      render: (text, record: API.ObRole) => {
        const roleGrantees = record.role_grantees?.join('、');
        const userGrantees = record.user_grantees?.join('、');
        return (
          <>
            <div
              className="ellipsis"
              style={{
                display: 'block',
              }}
            >
              {formatMessage({
                id: 'ocp-express.User.Oracle.RolesList.Role.1',
                defaultMessage: '角色：',
              })}

              <Tooltip placement="topLeft" title={roleGrantees}>
                <span>{roleGrantees || '-'}</span>
              </Tooltip>
            </div>
            <div
              className="ellipsis"
              style={{
                display: 'block',
              }}
            >
              {formatMessage({
                id: 'ocp-express.User.Oracle.RolesList.User',
                defaultMessage: '用户：',
              })}

              <Tooltip placement="topLeft" title={userGrantees}>
                <span>{userGrantees || '-'}</span>
              </Tooltip>
            </div>
          </>
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.User.Oracle.RolesList.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      render: (text: string, record: API.ObRole) => {
        // 内置角色不支持删除操作
        if (ORACLE_BUILT_IN_ROLE_LIST.includes(record?.role || '')) {
          return '';
        }
        return (
          <a
            onClick={() => {
              setDeleteModalVisible(true);
              setCurrent(record);
            }}
          >
            {formatMessage({
              id: 'ocp-express.User.Oracle.RolesList.Delete',
              defaultMessage: '删除',
            })}
          </a>
        );
      },
    },
  ];

  return (
    <>
      <Table
        columns={columns}
        rowKey="id"
        loading={dbRoleLoading}
        dataSource={dbRoleList}
        pagination={PAGINATION_OPTION_10}
      />

      <DeleteUserModal
        visible={deleteModalVisible}
        roleName={current?.role}
        tenantData={tenantData}
        onCancel={() => {
          setDeleteModalVisible(false);
          setCurrent(null);
        }}
        onSuccess={() => {
          setDeleteModalVisible(false);
          setCurrent(null);
          refreshDbRole?.();
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

export default connect(mapStateToProps)(RoleList);
