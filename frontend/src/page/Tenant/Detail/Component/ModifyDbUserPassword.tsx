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
import { Form, Modal, message } from '@oceanbase/design';
import { validatePassword } from '@/util';
import { MODAL_FORM_ITEM_LAYOUT } from '@/constant';
import Password from '@/component/Password';
import MyInput from '@/component/MyInput';
import { useRequest } from 'ahooks';
import { changePassword } from '@/service/obshell/tenant';
import useStyles from './ModifyPasswordModal.style';

export interface ModifyDbUserPasswordProps {
  ClusterInfo: API.ClusterInfo;
  tenantData: API.TenantInfo;
  dbUser: API.DbUser;
  onSuccess: () => void;
}

const ModifyDbUserPassword: React.FC<ModifyDbUserPasswordProps> = ({
  onSuccess,
  tenantData,
  dbUser,
  ...restProps
}) => {
  const { styles } = useStyles();
  const [passed, setPassed] = useState(true);

  const [form] = Form.useForm();
  const { getFieldsValue, validateFields } = form;

  const validateConfirmPassword = (rule, value, callback) => {
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

  const { run: handleChangePassword, loading } = useRequest(changePassword, {
    manual: true,
    onSuccess: res => {
      if (res?.successful) {
        if (onSuccess) {
          message.success(
            formatMessage(
              {
                id: 'ocp-express.Detail.Component.ModifyDbUserPassword.DbuserusernamePasswordChanged',
                defaultMessage: '{dbUserUsername} 密码修改成功',
              },

              { dbUserUsername: dbUser.username }
            )
          );

          onSuccess();
        }
      }
    },
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const { newPassword } = values;
      handleChangePassword(
        {
          name: tenantData.tenant_name,
          user: dbUser.username,
        },
        {
          new_password: newPassword,
        }
      );
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
