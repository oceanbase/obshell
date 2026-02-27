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
import { Form, Modal, message, ModalProps } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import MySelect from '@/component/MySelect';
import { listRoles, modifyRole, modifyUserRoles } from '@/service/obshell/tenant';

const { Option } = MySelect;

/**
 * 参数说明
 * username 用户名
 * roleName 角色名
 * username / roleName 二者只需传一个
 *  */
interface ModifyRoleModalProps extends ModalProps {
  tenantName: string;
  username?: string;
  roleName?: string;
  grantedRoles?: string[];
  onSuccess: () => void;
}

const ModifyRoleModal: React.FC<ModifyRoleModalProps> = ({
  visible,
  tenantName,
  username,
  roleName,
  grantedRoles,
  onSuccess,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;

  const { data } = useRequest(listRoles, {
    ready: visible,
    defaultParams: [
      {
        name: tenantName,
      },
    ],

    refreshDeps: [tenantName],
  });

  const dbRoleList: API.ObRole[] = data?.data?.contents || [];

  const { run: modifyRoleRun, loading: roleLoading } = useRequest(modifyRole, {
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
  });

  const { run: modifyUserRoleRun, loading: userLoading } = useRequest(modifyUserRoles, {
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
  });

  const submitFn = () => {
    validateFields().then(values => {
      const { roles } = values;
      if (roleName) {
        modifyRoleRun(
          {
            name: tenantName,
            role: roleName,
          },

          { roles }
        );
      } else {
        modifyUserRoleRun(
          {
            name: tenantName,
            user: username!,
          },

          { roles }
        );
      }
    });
  };

  return (
    <Modal
      open={visible}
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
              .filter(role => role.role !== roleName)
              .map((item: API.ObRole) => (
                <Option key={item.role} value={item.role}>
                  {item.role}
                </Option>
              ))}
          </MySelect>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ModifyRoleModal;
