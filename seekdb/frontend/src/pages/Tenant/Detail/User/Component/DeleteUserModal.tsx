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
import React, { useEffect, useState } from 'react';
import { Input, Alert, Descriptions, Modal, message, ModalProps } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import { dropUser } from '@/service/obshell/user';

/**
 * 参数说明
 * username 用户名
 * 删除用户 传 username
 *
 *  */
interface DeleteUserModalProps extends ModalProps {
  username?: string;
  userStats?: API.ObUserSessionStats;
  onSuccess?: () => void;
}

const DeleteUserModal: React.FC<DeleteUserModalProps> = ({
  visible,
  username,
  userStats,
  onSuccess,
  ...restProps
}) => {
  const [allowDelete, setAllowDelete] = useState(false);

  const { run, loading } = useRequest(dropUser, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'OBShell.User.Component.DeleteUserModal.UserDeletedSuccessfully',
            defaultMessage: '用户删除成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  const handleDelete = () => {
    run({
      user: username,
    });
  };

  useEffect(() => {
    if (!visible) setAllowDelete(false);
  }, [visible]);

  return (
    <Modal
      visible={visible}
      title={formatMessage({
        id: 'ocp-express.User.Component.DeleteUserModal.DeleteAUser',
        defaultMessage: '删除用户',
      })}
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
        message={formatMessage({
          id: 'ocp-express.User.Component.DeleteUserModal.DeletingAUserWillDelete',
          defaultMessage: '删除用户将删除该用户下所有对象和数据，请谨慎操作',
        })}
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
          {username}
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
