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
import { DrawerProps, Form, message } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import Password from '@/component/Password';
import MyDrawer from '@/component/MyDrawer';
import MyInput from '@/component/MyInput';
import MySelect from '@/component/MySelect';
import SelectAllAndClearRender from '@/component/SelectAllAndClearRender';
import {
  ORACLE_DATABASE_USER_NAME_RULE,
  ORACLE_DATABASE_ROLE_NAME_RULE,
  MODAL_FORM_ITEM_LAYOUT,
} from '@/constant';
import { validatePassword } from '@/util';
import { ORACLE_SYS_PRIVS } from '@/constant/tenant';
import RoleSelect from './RoleSelect';
import { createUser } from '@/service/obshell/tenant';
import { createRole } from '@/service/obshell/tenant';
import { useSelector } from 'umi';

const { Option } = MySelect;

interface AddOracleUserOrRoleDrawerProps extends DrawerProps {
  visible: boolean;
  addIsUser: string; // user, role
  onSuccess: () => void;
}

const AddOracleUserOrRoleDrawer: React.FC<AddOracleUserOrRoleDrawerProps> = ({
  addIsUser,
  onSuccess,
  ...restProps
}) => {
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const tenantName = tenantData.tenant_name;

  const [passed, setPassed] = useState(true);
  const [form] = Form.useForm();
  const { validateFields, setFieldsValue, getFieldsValue } = form;

  const validateConfirmPassword = (rule: any, value: any, callback: any) => {
    const { password } = getFieldsValue();
    if (value && value !== password) {
      callback(
        formatMessage({
          id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.TheNewPasswordEnteredTwice',
          defaultMessage: '两次输入的新密码不一致，请重新输入',
        })
      );
    } else {
      callback();
    }
  };
  const { run: createDbUser, loading: createDbUserLoading } = useRequest(createUser, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.UserCreated',
            defaultMessage: '用户新建成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  const { run: createDbRole, loading: createDbRoleLoading } = useRequest(createRole, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.RoleCreated',
            defaultMessage: '角色新建成功',
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
      const { username, password, roleName, globalPrivileges, roles } = values;
      if (addIsUser === 'user') {
        createDbUser(
          {
            name: tenantName,
          },

          {
            user_name: username,
            password,
            global_privileges: globalPrivileges,
            roles: roles,
          }
        );
      } else {
        createDbRole(
          {
            name: tenantName,
          },

          {
            role_name: roleName,
            global_privileges: globalPrivileges,
            roles: roles,
          }
        );
      }
    });
  };

  const titleText =
    addIsUser === 'user'
      ? formatMessage({
          id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.User',
          defaultMessage: '用户',
        })
      : formatMessage({
          id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.Role',
          defaultMessage: '角色',
        });
  return (
    <MyDrawer
      width={520}
      title={formatMessage(
        {
          id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.CreateTitletext',
          defaultMessage: '新建{titleText}',
        },
        { titleText }
      )}
      destroyOnClose={true}
      confirmLoading={createDbUserLoading || createDbRoleLoading}
      okText={formatMessage({
        id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.Submitted',
        defaultMessage: '提交',
      })}
      onOk={handleSubmit}
      {...restProps}
    >
      <Form
        form={form}
        layout="vertical"
        requiredMark="optional"
        preserve={false}
        {...MODAL_FORM_ITEM_LAYOUT}
      >
        <Form.Item
          wrapperCol={{ span: 16 }}
          label={formatMessage(
            {
              id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.Titletext',
              defaultMessage: '{titleText}名',
            },
            { titleText }
          )}
          extra={
            addIsUser === 'user'
              ? formatMessage({
                  id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.LowercaseLettersInTheUsername',
                  defaultMessage: '用户名中的小写字母将会默认转为大写',
                })
              : formatMessage({
                  id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.LowercaseLettersInTheRole',
                  defaultMessage: '角色名中的小写字母将会默认转为大写',
                })
          }
          name={addIsUser === 'user' ? 'username' : 'roleName'}
          rules={[
            {
              required: true,
              message: formatMessage(
                {
                  id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.EnterTitletext',
                  defaultMessage: '请输入{titleText}名',
                },
                { titleText }
              ),
            },
            addIsUser === 'user' ? ORACLE_DATABASE_USER_NAME_RULE : ORACLE_DATABASE_ROLE_NAME_RULE,
          ]}
        >
          <MyInput
            placeholder={formatMessage(
              {
                id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.EnterTitletext',
                defaultMessage: '请输入{titleText}名',
              },
              { titleText }
            )}
          />
        </Form.Item>
        {addIsUser === 'user' && (
          <>
            <Form.Item
              wrapperCol={{ span: 16 }}
              label={formatMessage({
                id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.Password',
                defaultMessage: '密码',
              })}
              name="password"
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.EnterAPassword',
                    defaultMessage: '请输入密码',
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
              wrapperCol={{ span: 16 }}
              label={formatMessage({
                id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.ConfirmPassword',
                defaultMessage: '确认密码',
              })}
              name="confirmPassword"
              dependencies={['password']}
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.EnterThePasswordAgain',
                    defaultMessage: '请再次输入密码',
                  }),
                },

                {
                  validator: validateConfirmPassword,
                },
              ]}
            >
              <MyInput.Password
                placeholder={formatMessage({
                  id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.EnterThePasswordAgain',
                  defaultMessage: '请再次输入密码',
                })}
              />
            </Form.Item>
          </>
        )}

        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.HaveSystemPermissions',
            defaultMessage: '拥有系统权限',
          })}
          name="globalPrivileges"
          tooltip={{
            placement: 'right',
            title: formatMessage({
              id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.ApplicableToAllDatabases',
              defaultMessage: '适用于所有的数据库',
            }),
          }}
        >
          <MySelect
            mode="multiple"
            allowClear={true}
            showSearch={true}
            maxTagCount={5}
            style={{ width: 440 }}
            dropdownRender={menu => (
              <SelectAllAndClearRender
                menu={menu}
                onSelectAll={() => {
                  setFieldsValue({
                    globalPrivileges: ORACLE_SYS_PRIVS,
                  });
                }}
                onClearAll={() => {
                  setFieldsValue({
                    globalPrivileges: [],
                  });
                }}
              />
            )}
          >
            {ORACLE_SYS_PRIVS.map((item: string) => (
              <Option key={item} value={item}>
                {item.replace(/_/g, ' ')}
              </Option>
            ))}
          </MySelect>
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.HaveARole',
            defaultMessage: '拥有角色',
          })}
          name="roles"
          tooltip={{
            placement: 'right',
            title: formatMessage({
              id: 'ocp-express.Oracle.Component.AddOracleUserOrRoleDrawer.HaveARole',
              defaultMessage: '拥有角色',
            }),
          }}
        >
          <RoleSelect
            style={{ width: 440 }}
            onAddSuccess={role => {
              if (role.length > 0) {
                setFieldsValue({
                  roles: role.map(item => item.role),
                });
              }
            }}
          />
        </Form.Item>
      </Form>
    </MyDrawer>
  );
};

export default AddOracleUserOrRoleDrawer;
