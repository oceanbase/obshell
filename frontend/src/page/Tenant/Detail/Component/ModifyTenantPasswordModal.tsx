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
import { useSelector } from 'umi';
import React, { useState } from 'react';
import { Form, Alert, Modal, message } from '@oceanbase/design';
import { ExclamationCircleFilled } from '@oceanbase/icons';
import { useRequest } from 'ahooks';
import { MODAL_FORM_ITEM_LAYOUT } from '@/constant';
import Password from '@/component/Password';
import MyInput from '@/component/MyInput';
import { tenantModifyPassword, persistTenantRootPassword } from '@/service/obshell/tenant';

export interface TenantAdminPasswordModalProps {
  onSuccess: () => void;
}

const ModifyTenantPasswordModal: React.FC<TenantAdminPasswordModalProps> = ({
  onSuccess,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields, getFieldsValue } = form;

  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const [newPasswordPassed, setNewPasswordPassed] = useState(false);
  const [confirmPasswordPassed, setConfirmPasswordPassed] = useState(false);

  const validateConfirmPassword = (rule, value, callback) => {
    const { newPassword } = getFieldsValue();
    if (value && value !== newPassword) {
      callback(
        formatMessage({
          id: 'ocp-express.Detail.Component.ModifyPasswordModal.TheNewPasswordEnteredTwice',
          defaultMessage: '两次输入的新密码不一致，请重新输入',
        })
      );
    } else {
      callback();
    }
  };

  const { runAsync: tenantModifyPasswordFn } = useRequest(tenantModifyPassword, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.Detail.Component.ModifyPasswordModal.PasswordModifiedSuccessfully',
            defaultMessage: '密码修改成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  const { run: createOrReplacePassword, loading } = useRequest(persistTenantRootPassword, {
    manual: true,
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const { newPassword } = values;
      tenantModifyPasswordFn(
        {
          name: tenantData?.tenantName,
          // username: tenantData?.mode === 'ORACLE' ? 'SYS' : 'root',
        },
        {
          old_password: '',
          new_password: newPassword,
        }
      ).then(res => {
        if (res.successful) {
          createOrReplacePassword(
            {
              name: tenantData?.tenantName,
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
      title={formatMessage({
        id: 'ocp-express.Detail.Component.ModifyDbUserPassword.ChangePassword',
        defaultMessage: '修改密码',
      })}
      width={520}
      destroyOnClose={true}
      {...restProps}
      confirmLoading={loading}
      onOk={handleSubmit}
    >
      <Alert
        icon={<ExclamationCircleFilled />}
        message={formatMessage({
          id: 'ocp-express.Detail.Component.ModifyTenantPasswordModal.TheRootPasswordOfTheCurrentTenantIs',
          defaultMessage: '当前租户 root 密码为空，存在安全隐患，请修改 root 密码',
        })}
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
        {tenantData?.tenantName && (
          <Form.Item
            label={formatMessage({
              id: 'ocp-express.src.component.TenantAdminPasswordModal.Tenant',
              defaultMessage: '租户',
            })}
            name="tenantName"
            initialValue={tenantData?.tenantName}
          >
            <MyInput disabled={true} />
          </Form.Item>
        )}
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.src.component.TenantAdminPasswordModal.Password',
            defaultMessage: '密码',
          })}
          name="newPassword"
          validateFirst={true}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'obshell.Detail.Component.ModifyTenantPasswordModal.PleaseEnterANewPassword',
                defaultMessage: '请输入新密码',
              }),
            },

            {
              validator: (rule, value, callback) => {
                if (value && !newPasswordPassed) {
                  callback(
                    formatMessage({
                      id: 'obshell.Detail.Component.ModifyTenantPasswordModal.ThePasswordSettingDoesNot',
                      defaultMessage: '密码设置不符合要求',
                    })
                  );
                } else {
                  callback();
                }
              },
            },
          ]}
        >
          <Password onValidate={setNewPasswordPassed} />
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.ModifyPasswordModal.ConfirmPassword',
            defaultMessage: '确认密码',
          })}
          name="confirmPassword"
          dependencies={['newPassword']}
          validateFirst={true}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.Detail.Component.ModifyPasswordModal.EnterANewPasswordAgain',
                defaultMessage: '请再次输入新密码',
              }),
            },

            {
              validator: (rule, value, callback) => {
                if (value && !confirmPasswordPassed) {
                  callback(
                    formatMessage({
                      id: 'obshell.Detail.Component.ModifyTenantPasswordModal.ThePasswordSettingDoesNot',
                      defaultMessage: '密码设置不符合要求',
                    })
                  );
                } else {
                  callback();
                }
              },
            },

            {
              validator: validateConfirmPassword,
            },
          ]}
        >
          <Password
            autoComplete="new-password"
            onValidate={setConfirmPasswordPassed}
            placeholder={formatMessage({
              id: 'ocp-express.Detail.Component.ModifyPasswordModal.EnterANewPasswordAgain',
              defaultMessage: '请再次输入新密码',
            })}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ModifyTenantPasswordModal;
