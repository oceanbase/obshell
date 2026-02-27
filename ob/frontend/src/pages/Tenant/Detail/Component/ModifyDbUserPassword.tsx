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
import { Form, Modal, message, ModalProps } from '@oceanbase/design';
import { validatePassword } from '@/util';
import { MODAL_FORM_ITEM_LAYOUT } from '@/constant';
import Password from '@/component/Password';
import MyInput from '@/component/MyInput';
import { useRequest } from 'ahooks';
import { changePassword, persistTenantRootPassword } from '@/service/obshell/tenant';

export interface ModifyDbUserPasswordProps extends ModalProps {
  tenantData: API.DbaObTenant;
  dbUser: API.ObUser | null;
  onSuccess: () => void;
}

const ModifyDbUserPassword: React.FC<ModifyDbUserPasswordProps> = ({
  onSuccess,
  tenantData,
  dbUser,
  ...restProps
}) => {
  const [passed, setPassed] = useState(true);

  const [form] = Form.useForm();
  const { getFieldsValue, validateFields } = form;

  const validateConfirmPassword = (rule: any, value: any, callback: any) => {
    const { newPassword } = getFieldsValue();
    if (value && value !== newPassword) {
      callback(
        formatMessage({
          id: 'ocp-express.Detail.Component.ModifyDbUserPassword.TheNewPasswordEnteredTwice',
          defaultMessage: '两次输入的新密码不一致，请重新输入',
        })
      );
    } else {
      callback();
    }
  };

  const { runAsync: handleChangePassword, loading } = useRequest(changePassword, {
    manual: true,
    onSuccess: res => {
      if (res?.successful) {
        message.success(
          formatMessage(
            {
              id: 'ocp-express.Detail.Component.ModifyDbUserPassword.DbuserusernamePasswordChanged',
              defaultMessage: '{dbUserUsername} 密码修改成功',
            },

            { dbUserUsername: dbUser?.user_name || '' }
          )
        );
        onSuccess?.();
      }
    },
  });

  const { run: createOrReplacePassword } = useRequest(persistTenantRootPassword, {
    manual: true,
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const { newPassword } = values;
      handleChangePassword(
        {
          name: tenantData.tenant_name || '',
          user: dbUser?.user_name || '',
        },
        {
          new_password: newPassword,
        }
      ).then(res => {
        if (res.successful && (dbUser?.user_name === 'root' || dbUser?.user_name === 'SYS')) {
          createOrReplacePassword(
            {
              name: tenantData?.tenant_name || '',
            },
            {
              password: newPassword,
            }
          );
        }
      });
    });
  };

  return (
    <Modal
      width={480}
      title={formatMessage({
        id: 'ocp-express.Detail.Component.ModifyDbUserPassword.ChangePassword',
        defaultMessage: '修改密码',
      })}
      destroyOnClose={true}
      {...restProps}
      confirmLoading={loading}
      onOk={handleSubmit}
    >
      <Form
        form={form}
        layout="vertical"
        hideRequiredMark
        preserve={false}
        {...MODAL_FORM_ITEM_LAYOUT}
      >
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.ModifyDbUserPassword.NewPassword',
            defaultMessage: '新密码',
          })}
          name="newPassword"
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.Detail.Component.ModifyDbUserPassword.EnterANewPassword',
                defaultMessage: '请输入新密码',
              }),
            },
            {
              validator: validatePassword(passed),
            },
          ]}
        >
          <Password onValidate={setPassed} />
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.ModifyDbUserPassword.ConfirmPassword',
            defaultMessage: '确认密码',
          })}
          name="confirmPassword"
          dependencies={['newPassword']}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.Detail.Component.ModifyDbUserPassword.EnterANewPasswordAgain',
                defaultMessage: '请再次输入新密码',
              }),
            },

            {
              validator: validateConfirmPassword,
            },
          ]}
        >
          <MyInput.Password
            autoComplete="new-password"
            placeholder={formatMessage({
              id: 'ocp-express.Detail.Component.ModifyDbUserPassword.EnterANewPasswordAgain',
              defaultMessage: '请再次输入新密码',
            })}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ModifyDbUserPassword;
