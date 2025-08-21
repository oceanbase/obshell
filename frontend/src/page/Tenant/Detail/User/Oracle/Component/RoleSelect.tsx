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
import { useRequest } from 'ahooks';
import { Form } from '@oceanbase/design';
import type { SelectProps } from '@oceanbase/design/es/select';
import MyDrawer from '@/component/MyDrawer';
import MyInput from '@/component/MyInput';
import MySelect from '@/component/MySelect';
import SelectAllAndClearRender from '@/component/SelectAllAndClearRender';
import SelectDropdownRender from '@/component/SelectDropdownRender';
import { ORACLE_DATABASE_USER_NAME_RULE, MODAL_FORM_ITEM_LAYOUT } from '@/constant';
import { ORACLE_SYS_PRIVS } from '@/constant/tenant';
import { createRole, listRoles } from '@/service/obshell/tenant';
import { useSelector } from 'umi';

const { Option } = MySelect;

interface RoleSelectProps extends SelectProps<string> {
  onAddSuccess: (roles: API.ObRole[]) => void;
  onCancel?: () => void;
}

const RoleSelect: React.FC<RoleSelectProps> = ({ onAddSuccess, onCancel, ...restProps }) => {
  const [visible, setVisible] = useState(false);
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const tenantName = tenantData.tenant_name || '';
  const [form] = Form.useForm();
  const { validateFields, setFieldsValue } = form;

  const { data, refresh } = useRequest(listRoles, {
    defaultParams: [
      {
        name: tenantName,
      },
    ],

    refreshDeps: [tenantName],
  });

  const dbRoleList: API.ObRole[] = data?.data?.contents || [];

  const { run, loading } = useRequest(createRole, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        setVisible(false);
        refresh();
        if (onAddSuccess) {
          onAddSuccess([res?.data]);
        }
      }
    },
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const { roleName, globalPrivileges, roles } = values;
      run(
        {
          name: tenantName,
        },

        {
          role_name: roleName,
          global_privileges: globalPrivileges,
          roles: roles,
        }
      );
    });
  };

  return (
    <span>
      <MySelect
        width={520}
        showSearch={true}
        mode="multiple"
        maxTagCount={5}
        allowClear={true}
        // ID 作为值，但要根据 name(作为 children) 进行过滤
        optionFilterProp="children"
        dropdownRender={menu => (
          <SelectDropdownRender
            menu={menu}
            text={formatMessage({
              id: 'ocp-express.Oracle.Component.RoleSelect.CreateARole',
              defaultMessage: '新建角色',
            })}
            onClick={() => {
              setVisible(true);
            }}
          />
        )}
        {...restProps}
      >
        {dbRoleList.map(item => (
          <Option key={item.role} value={item.role}>
            {item.role}
          </Option>
        ))}
      </MySelect>
      <MyDrawer
        width={520}
        title={formatMessage({
          id: 'ocp-express.Oracle.Component.RoleSelect.CreateARole',
          defaultMessage: '新建角色',
        })}
        visible={visible}
        confirmLoading={loading}
        destroyOnClose={true}
        onCancel={() => {
          setVisible(false);
          onCancel?.();
        }}
        okText={formatMessage({
          id: 'ocp-express.Oracle.Component.RoleSelect.Submitted',
          defaultMessage: '提交',
        })}
        onOk={() => handleSubmit()}
      >
        <Form
          form={form}
          layout="vertical"
          requiredMark="optional"
          preserve={false}
          {...MODAL_FORM_ITEM_LAYOUT}
        >
          <Form.Item
            label={formatMessage({
              id: 'ocp-express.Oracle.Component.RoleSelect.Role',
              defaultMessage: '角色名',
            })}
            name="roleName"
            rules={[
              {
                required: true,
                message: formatMessage({
                  id: 'ocp-express.Oracle.Component.RoleSelect.EnterARoleName',
                  defaultMessage: '请输入角色名',
                }),
              },

              ORACLE_DATABASE_USER_NAME_RULE,
            ]}
          >
            <MyInput
              placeholder={formatMessage({
                id: 'ocp-express.Oracle.Component.RoleSelect.EnterARoleName',
                defaultMessage: '请输入角色名',
              })}
            />
          </Form.Item>
          <Form.Item
            label={formatMessage({
              id: 'ocp-express.Oracle.Component.RoleSelect.HaveSystemPermissions',
              defaultMessage: '拥有系统权限',
            })}
            tooltip={{
              placement: 'right',
              title: formatMessage({
                id: 'ocp-express.Oracle.Component.RoleSelect.ApplicableToAllDatabases',
                defaultMessage: '适用于所有的数据库',
              }),
            }}
            name="globalPrivileges"
          >
            <MySelect
              mode="multiple"
              maxTagCount={5}
              allowClear={true}
              showSearch={true}
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
                  {
                    item === 'PURGE_DBA_RECYCLEBIN'
                      ? 'PURGE DBA_RECYCLEBIN'
                      : item.replace(/_/g, ' ')
                  }
                </Option>
              ))}
            </MySelect>
          </Form.Item>
          <Form.Item
            label={formatMessage({
              id: 'ocp-express.Oracle.Component.RoleSelect.HaveARole',
              defaultMessage: '拥有角色',
            })}
            name="roles"
            tooltip={{
              placement: 'right',
              title: formatMessage({
                id: 'ocp-express.Oracle.Component.RoleSelect.HaveARole',
                defaultMessage: '拥有角色',
              }),
            }}
          >
            <MySelect mode="multiple" allowClear={true} showSearch={true}>
              {dbRoleList.map(item => (
                <Option key={item.role} value={item.role}>
                  {item.role}
                </Option>
              ))}
            </MySelect>
          </Form.Item>
        </Form>
      </MyDrawer>
    </span>
  );
};

export default RoleSelect;
