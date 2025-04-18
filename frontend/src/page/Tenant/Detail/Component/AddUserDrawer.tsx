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
import { connect } from 'umi';
import React, { useEffect, useState } from 'react';
import { Form, message } from '@oceanbase/design';
import Password from '@/component/Password';
import MyDrawer from '@/component/MyDrawer';
import MyInput from '@/component/MyInput';
import MySelect from '@/component/MySelect';
import { useRequest } from 'ahooks';
import SelectAllAndClearRender from '@/component/SelectAllAndClearRender';
import DatabasePrivilegeTransfer from '@/component/DatabasePrivilegeTransfer';
import { DATABASE_USER_NAME_RULE } from '@/constant';
import { validatePassword } from '@/util';
import { DATABASE_GLOBAL_PRIVILEGE_LIST } from '@/constant/tenant';
import { differenceBy, isEqual, uniq } from 'lodash';
import {
  listDatabases,
  modifyGlobalPrivilege,
  modifyDbPrivilege,
  createUser,
} from '@/service/obshell/tenant';

const { Option } = MySelect;
interface AddOrEditUserDrawerProps {
  dbUser?: API.DbUser;
  tenantData: API.TenantInfo;
  databaseList: API.DbPrivilege[];
  charsetList: API.Charset[];
  onSuccess: () => void;
  onCancel: () => void;
}

const AddUserDrawer: React.FC<AddOrEditUserDrawerProps> = ({
  onSuccess,
  onCancel,
  dbUser,
  tenantData,
  charsetList,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields, getFieldsValue, setFieldsValue } = form;
  // TODO 这里借助一下深拷贝，来解决在组件内操作时dbUser被篡改的问题；待后期排查优化
  const dbUserInfo = JSON.parse(JSON.stringify(dbUser));

  const [passed, setPassed] = useState(true);

  const { runAsync: handleCreateUser, loading: createDbUserLoading } = useRequest(createUser, {
    manual: true,
    onSuccess: () => {
      if (onSuccess) {
        onSuccess();
      }
    },
  });

  const { runAsync: handleModifyGlobalPrivilege, loading: modifyGlobalPrivilegeLoading } =
    useRequest(modifyGlobalPrivilege, {
      manual: true,
      onSuccess: () => {
        const dbUsername = dbUser?.username;
        message.success(
          formatMessage(
            {
              id: 'ocp-express.Detail.Component.AddUserDrawer.TheSystemPermissionOfDbusername',
            },
            { dbUsername }
          )
        );
        if (onSuccess) {
          onSuccess();
        }
      },
    });

  const { runAsync: handleModifyDbPrivilege, loading: modifyDbPrivilegeLoading } = useRequest(
    modifyDbPrivilege,
    {
      manual: true,
      onSuccess: () => {
        const dbUsername = dbUser?.username;
        message.success(
          formatMessage(
            {
              id: 'ocp-express.Detail.Component.AddUserDrawer.ThePermissionOfDbusernameHas',
            },
            { dbUsername }
          )
        );
        if (onSuccess) {
          onSuccess();
        }
      },
    }
  );

  const {
    data: databaseListData,
    loading: databaseListLoading,
    refresh,
  } = useRequest(listDatabases, {
    defaultParams: [
      {
        name: tenantData?.tenant_name,
      },
    ],
    manual: false,
  });

  const databaseList = (databaseListData?.data?.contents || [])?.map(item => ({
    ...item,
    dbName: item?.db_name,
  }));

  const validateConfirmPassword = (rule, value, callback) => {
    const { password } = getFieldsValue();
    if (value && value !== password) {
      callback(
        formatMessage({
          id: 'ocp-express.Detail.Component.AddUserDrawer.TheNewPasswordEnteredTwice',
          defaultMessage: '两次输入的新密码不一致，请重新输入',
        })
      );
    } else {
      callback();
    }
  };

  const modifyPrivilege = (globalPrivileges?: string[], dbPrivileges?: API.DbPrivilege) => {
    const promiseList = [];
    if (!isEqual(uniq(globalPrivileges), uniq(dbUserInfo.globalPrivileges))) {
      promiseList.push(
        handleModifyGlobalPrivilege(
          {
            name: tenantData.tenant_name,
            user: dbUser?.username,
          },
          {
            global_privileges: globalPrivileges,
          }
        )
      );
    }
    if (dbPrivileges) {
      promiseList.push(
        handleModifyDbPrivilege(
          {
            name: tenantData.tenant_name,
            user: dbUser?.username,
          },
          {
            db_privileges: dbPrivileges,
          }
        )
      );
    }

    if (promiseList.length) {
      Promise.all(promiseList);
    } else {
      message.info(
        formatMessage({
          id: 'ocp-express.Detail.Component.AddUserDrawer.NoPermissionsHaveBeenModified',
          defaultMessage: '未修改任何权限',
        })
      );
    }
  };

  const handleSubmit = () => {
    validateFields().then(values => {
      const { username, password, globalPrivileges, dbPrivileges } = values;
      if (dbUserInfo) {
        if (!dbPrivileges) {
          modifyPrivilege(globalPrivileges);
        } else {
          const oldDbNameList = uniq(
            dbUserInfo.dbPrivileges?.map((item: API.DbPrivilege) => item.dbName)
          );

          const newDbNameList = uniq(dbPrivileges?.map((item: API.DbPrivilege) => item.dbName));
          // 比较修改前后的权限，如果只是简单修改权限，直接保存
          if (isEqual(oldDbNameList, newDbNameList)) {
            modifyPrivilege(globalPrivileges, dbPrivileges);
          } else {
            // 整体删除了某个数据库的权限，需要传递给后端 {dbName, privileges: []}，接口拿到这种参数才会整体删除数据库的权限
            dbUserInfo.dbPrivileges?.forEach((item: API.DbPrivilege) => {
              if (!newDbNameList.includes(item.dbName)) {
                dbPrivileges.push({
                  dbName: item?.dbName,
                  privileges: [],
                });
              }
            });
            modifyPrivilege(globalPrivileges, dbPrivileges);
          }
        }
      } else {
        handleCreateUser(
          {
            name: tenantData?.tenant_name,
          },
          {
            user_name: username,
            password,
            global_privileges: globalPrivileges,
            db_privileges: dbPrivileges?.map(item => ({
              ...item,
              db_name: item?.dbName,
            })),
          }
        );
      }
    });
  };

  return (
    <MyDrawer
      width={960}
      title={
        dbUser
          ? formatMessage({
              id: 'ocp-express.Detail.Component.AddUserDrawer.ModifyDatabaseUserPermissions',
              defaultMessage: '修改数据库用户的权限',
            })
          : formatMessage({
              id: 'ocp-express.Detail.Component.AddUserDrawer.CreateADatabaseUser',
              defaultMessage: '新建数据库用户',
            })
      }
      destroyOnClose={true}
      okText={formatMessage({
        id: 'ocp-express.Detail.Component.AddUserDrawer.Submitted',
        defaultMessage: '提交',
      })}
      confirmLoading={
        createDbUserLoading || modifyGlobalPrivilegeLoading || modifyDbPrivilegeLoading
      }
      onOk={handleSubmit}
      onCancel={onCancel}
      {...restProps}
    >
      <Form form={form} layout="vertical" requiredMark="optional" preserve={false}>
        <Form.Item
          wrapperCol={{ span: 9 }}
          label={formatMessage({
            id: 'ocp-express.Detail.Component.AddUserDrawer.UserName',
            defaultMessage: '用户名',
          })}
          name="username"
          initialValue={dbUser?.username}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.Detail.Component.AddUserDrawer.EnterAUsername',
                defaultMessage: '请输入用户名',
              }),
            },
            // 修改权限时，不对用户名做格式校验
            dbUser ? {} : DATABASE_USER_NAME_RULE,
          ]}
        >
          <MyInput
            disabled={dbUser}
            placeholder={formatMessage({
              id: 'ocp-express.Detail.Component.AddUserDrawer.EnterAUsername',
              defaultMessage: '请输入用户名',
            })}
          />
        </Form.Item>
        {!dbUser && (
          <>
            <Form.Item
              wrapperCol={{ span: 9 }}
              label={formatMessage({
                id: 'ocp-express.Detail.Component.AddUserDrawer.Password',
                defaultMessage: '密码',
              })}
              name="password"
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'ocp-express.Detail.Component.AddUserDrawer.EnterAPassword',
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
              wrapperCol={{ span: 9 }}
              label={formatMessage({
                id: 'ocp-express.Detail.Component.AddUserDrawer.ConfirmPassword',
                defaultMessage: '确认密码',
              })}
              name="confirmPassword"
              dependencies={['password']}
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'ocp-express.Detail.Component.AddUserDrawer.EnterThePasswordAgain',
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
                  id: 'ocp-express.Detail.Component.AddUserDrawer.EnterThePasswordAgain',
                  defaultMessage: '请再次输入密码',
                })}
              />
            </Form.Item>
          </>
        )}

        <Form.Item
          wrapperCol={{ span: 9 }}
          label={formatMessage({
            id: 'ocp-express.Detail.Component.AddUserDrawer.GlobalPermissions',
            defaultMessage: '全局权限',
          })}
          tooltip={{
            placement: 'right',
            title: formatMessage({
              id: 'ocp-express.Detail.Component.AddUserDrawer.ApplicableToAllDatabases',
              defaultMessage: '适用于所有的数据库',
            }),
          }}
          name="globalPrivileges"
          initialValue={(dbUser && dbUser.globalPrivileges) || undefined}
        >
          <MySelect
            mode="multiple"
            maxTagCount={8}
            allowClear={true}
            showSearch={true}
            dropdownRender={menu => (
              <SelectAllAndClearRender
                menu={menu}
                onSelectAll={() => {
                  setFieldsValue({
                    globalPrivileges: uniq(DATABASE_GLOBAL_PRIVILEGE_LIST?.map(item => item.value)),
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
            {DATABASE_GLOBAL_PRIVILEGE_LIST.map(item => (
              <Option key={item.value} value={item.value}>
                {item.value}
              </Option>
            ))}
          </MySelect>
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.AddUserDrawer.DatabasePermissions',
            defaultMessage: '数据库权限',
          })}
          tooltip={{
            placement: 'right',
            title: formatMessage({
              id: 'ocp-express.Detail.Component.AddUserDrawer.AppliesToAllTargetsIn',
              defaultMessage: '适用于一个给定数据库中的所有目标',
            }),
          }}
          name="dbPrivileges"
        >
          <DatabasePrivilegeTransfer
            dbUserList={(dbUser && dbUser.dbPrivileges
              ? differenceBy(databaseList, dbUser.dbPrivileges, 'dbName')
              : databaseList
            )?.filter(item => item.dbName !== 'information_schema')}
            loading={databaseListLoading}
            dbPrivilegedList={dbUser && dbUser.dbPrivileges}
          />
        </Form.Item>
      </Form>
    </MyDrawer>
  );
};

function mapStateToProps({ loading, tenant, database }) {
  return {
    tenantData: tenant.tenantData,
  };
}

export default connect(mapStateToProps)(AddUserDrawer);
