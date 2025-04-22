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
import { Input, Alert, Descriptions, Modal, message } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import * as ObUserController from '@/service/ocp-express/ObUserController';
import { dropUser } from '@/service/obshell/tenant';

/**
 * 参数说明
 * username 用户名
 * roleName 角色名
 * username / roleName 二者只需传一个
 * 删除用户 传 username
 * 删除角色 传 roleName
 *
 *  */
interface DeleteUserModalProps {
  username?: string;
  roleName?: string;
  userStats?: {
    total: number;
    active: number;
  };
  tenantData: API.TenantInfo;
  onSuccess?: () => void;
}

const DeleteUserModal: React.FC<DeleteUserModalProps> = ({
  username,
  roleName,
  userStats,
  tenantData,
  onSuccess,
  ...restProps
}) => {
  const [allowDelete, setAllowDelete] = useState(false);

  let APIName = ObUserController.deleteDbRole;
  const deleteDbRoleParams = {
    tenantId: tenantData.tenant_name,
    roleName,
  };

  const deleteDbUserParams = {
    name: tenantData.tenant_name,
    user: username,
  };

  if (username) {
    APIName = dropUser;
  } else if (roleName) {
    APIName = ObUserController.deleteDbRole;
  }
  const { run, loading } = useRequest(APIName, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          username
            ? formatMessage({ id: 'ocp-express.User.Component.DeleteUserModal.UserDeleted' })
            : formatMessage({ id: 'ocp-express.User.Component.DeleteUserModal.RoleDeleted' })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  const handleDelete = () => {
    run(roleName ? deleteDbRoleParams : deleteDbUserParams);
  };

  return (
    <Modal
      title={
        username
          ? formatMessage({
              id: 'ocp-express.User.Component.DeleteUserModal.DeleteAUser',
              defaultMessage: '删除用户',
            })
          : formatMessage({
              id: 'ocp-express.User.Component.DeleteUserModal.DeleteARole',
              defaultMessage: '删除角色',
            })
      }
      destroyOnClose={true}
      confirmLoading={loading}
      okText={formatMessage({
        id: 'ocp-express.User.Component.DeleteUserModal.Delete',
        defaultMessage: '删除',
      })}
      cancelText={formatMessage({
        id: 'ocp-express.User.Component.DeleteUserModal.Cancel',
        defaultMessage: '取消',
      })}
      onOk={handleDelete}
      {...restProps}
      okButtonProps={{
        ghost: true,
        danger: true,
        disabled: !allowDelete,
      }}
    >
      <Alert
        message={
          username
            ? formatMessage({
                id: 'ocp-express.User.Component.DeleteUserModal.DeletingAUserWillDelete',
                defaultMessage: '删除用户将删除该用户下所有对象和数据，请谨慎操作',
              })
            : formatMessage({
                id: 'ocp-express.User.Component.DeleteUserModal.DeletingARoleWillCause',
                defaultMessage:
                  '删除角色将导致引用了该角色的角色和用户失去该角色的所有权限，请谨慎操作',
              })
        }
        type="warning"
        showIcon={true}
        style={{
          marginBottom: 12,
        }}
      />

      <Descriptions column={1}>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Session.List.User',
            defaultMessage: '用户',
          })}
        >
          {username || roleName}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.User.Component.DeleteUserModal.Tenant',
            defaultMessage: '所属租户',
          })}
        >
          {tenantData?.tenant_name}
        </Descriptions.Item>
        {username && (
          <>
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.User.Component.DeleteUserModal.CurrentSessions',
                defaultMessage: '当前会话总数',
              })}
            >
              {userStats?.total || 0}
            </Descriptions.Item>
            <Descriptions.Item
              label={formatMessage({
                id: 'ocp-express.User.Component.DeleteUserModal.ActiveSessions',
                defaultMessage: '活跃会话数',
              })}
            >
              {userStats?.active || 0}
            </Descriptions.Item>
          </>
        )}
      </Descriptions>
      <div>
        {formatMessage({
          id: 'ocp-express.User.Component.DeleteUserModal.Enter',
          defaultMessage: '请输入',
        })}

        <span style={{ color: 'red' }}> delete </span>
        {formatMessage({
          id: 'ocp-express.User.Component.DeleteUserModal.ConfirmOperation',
          defaultMessage: '确认操作',
        })}
      </div>
      <Input
        style={{ width: 400, marginTop: 8 }}
        onChange={e => setAllowDelete(e.target.value === 'delete')}
      />
    </Modal>
  );
};

export default DeleteUserModal;
