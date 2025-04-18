/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
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
