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
import { Form, Modal, message } from '@oceanbase/design';
import * as ObUserController from '@/service/ocp-express/ObUserController';
import { useRequest } from 'ahooks';
import MySelect from '@/component/MySelect';

const { Option } = MySelect;

/**
 * 参数说明
 * username 用户名
 * roleName 角色名
 * username / roleName 二者只需传一个
 *  */
interface ModifyRoleModalProps {
  tenantId?: number;
  username?: string;
  roleName?: string;
  grantedRoles?: string[];
  onSuccess: () => void;
}

const ModifyRoleModal: React.FC<ModifyRoleModalProps> = ({
  tenantId,
  username,
  roleName,
  grantedRoles,
  onSuccess,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;

  const { data } = useRequest(ObUserController.listDbRoles, {
    defaultParams: [
      {
        tenantId,
      },
    ],

    refreshDeps: [tenantId],
  });

  const dbRoleList = data?.data?.contents || [];

  const { run: modifyRoleFromRole, loading: roleLoading } = useRequest(
    ObUserController.modifyRoleFromRole,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage(
              {
                id: 'ocp-express.Oracle.Component.ModifyRoleModal.RolenameRoleModified',
                defaultMessage: '{roleName} 角色修改成功',
              },
              { roleName }
            )
          );
          if (onSuccess) {
            onSuccess();
          }
        }
      },
    }
  );

  const { run: modifyRoleFromUser, loading: userLoading } = useRequest(
    ObUserController.modifyRoleFromUser,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage(
              {
                id: 'ocp-express.Oracle.Component.ModifyRoleModal.TheUsernameRoleHasBeen',
                defaultMessage: '{username} 角色修改成功',
              },
              { username }
            )
          );
          if (onSuccess) {
            onSuccess();
          }
        }
      },
    }
  );

  const submitFn = () => {
    validateFields().then(values => {
      const { roles } = values;

      if (roleName) {
        modifyRoleFromRole(
          {
            tenantId,
            roleName,
          },

          { roles }
        );
      } else {
        modifyRoleFromUser(
          {
            tenantId,
            username,
          },

          { roles }
        );
      }
    });
  };

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-express.Oracle.Component.ModifyRoleModal.ModifyARole',
        defaultMessage: '修改角色',
      })}
      destroyOnClose={true}
      confirmLoading={userLoading || roleLoading}
      onOk={submitFn}
      {...restProps}
    >
      <Form form={form} layout="vertical" requiredMark="optional" preserve={false}>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Oracle.Component.ModifyRoleModal.HaveARole',
            defaultMessage: '拥有角色',
          })}
          name="roles"
          tooltip={{
            placement: 'right',
            title: formatMessage({
              id: 'ocp-express.Oracle.Component.ModifyRoleModal.HaveARole',
              defaultMessage: '拥有角色',
            }),
          }}
          initialValue={grantedRoles}
        >
          <MySelect mode="multiple" maxTagCount={5} allowClear={true} showSearch={true}>
            {dbRoleList
              // 只能拥有其他角色，不能拥有自身，否则内核会出现死循环错误
              .filter(role => role.name !== roleName)
              .map((item: API.DbRole) => (
                <Option key={item.name} value={item.name}>
                  {item.name}
                </Option>
              ))}
          </MySelect>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ModifyRoleModal;
