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
import { Form, Alert, Modal, message } from '@oceanbase/design';
import { ExclamationCircleFilled } from '@oceanbase/icons';
import { useRequest } from 'ahooks';
import { MODAL_FORM_ITEM_LAYOUT } from '@/constant';
import Password from '@/component/Password';
import MyInput from '@/component/MyInput';
import { persistTenantRootPassword } from '@/service/obshell/tenant';

export interface TenantAdminPasswordModalProps {
  tenantName?: string;
  errorMessage: string;
  type: string; // 'EDIT' | 'ADD'
  onSuccess: () => void;
}

const TenantAdminPasswordModal: React.FC<TenantAdminPasswordModalProps> = ({
  type,
  onSuccess,
  tenantName,
  errorMessage,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;

  const { run: createOrReplacePassword, loading } = useRequest(persistTenantRootPassword, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          type === 'ADD'
            ? formatMessage({
                id: 'ocp-express.src.component.TenantAdminPasswordModal.PasswordAddedSuccessfully',
                defaultMessage: '密码新增成功',
              })
            : formatMessage({
                id: 'ocp-express.src.component.TenantAdminPasswordModal.PasswordModifiedSuccessfully',
                defaultMessage: '密码修改成功',
              })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const { newPassword } = values;
      createOrReplacePassword({
        name: tenantName,

      },{
        password: newPassword,
      });
    });
  };

  return (
    <Modal
      title={
        type === 'ADD'
          ? formatMessage({
              id: 'ocp-express.src.component.TenantAdminPasswordModal.EnterTheTenantAdministratorPassword',
              defaultMessage: '录入租户管理员密码',
            })
          : formatMessage({
              id: 'ocp-express.src.component.TenantAdminPasswordModal.UpdateTenantAdministratorPassword',
              defaultMessage: '更新租户管理员密码',
            })
      }
      destroyOnClose={true}
      {...restProps}
      confirmLoading={ loading}
      onOk={handleSubmit}
    >
      <Alert
        icon={<ExclamationCircleFilled />}
        message={errorMessage}
        type="error"
        showIcon
        style={{
          marginBottom: 24,
        }}
      />
      <Form
        form={form}
        layout="vertical"
        hideRequiredMark
        preserve={false}
        {...MODAL_FORM_ITEM_LAYOUT}
      >
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.src.component.TenantAdminPasswordModal.Tenant',
            defaultMessage: '租户',
          })}
          name="tenantName"
          initialValue={tenantName}
        >
          <MyInput disabled={true} />
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.src.component.TenantAdminPasswordModal.Password',
            defaultMessage: '密码',
          })}
          name="newPassword"
        >
          <Password hasRule={false}/>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default TenantAdminPasswordModal;
