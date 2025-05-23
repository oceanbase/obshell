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
import SelectAllAndClearRender from '@/component/SelectAllAndClearRender';
import { ORACLE_SYS_PRIVS } from '@/constant/tenant';

const { Option } = MySelect;

/**
 * 参数说明
 * username 用户名
 * roleName 角色名
 * username / roleName 二者只需传一个
 *  */
interface ModifyGlobalPrivilegeModalProps {
  tenantId?: number;
  dispatch: any;
  username?: string;
  roleName?: string;
  globalPrivileges: string[];
  onSuccess: () => void;
}

const ModifyGlobalPrivilegeModal: React.FC<ModifyGlobalPrivilegeModalProps> = ({
  tenantId,
  username,
  roleName,
  globalPrivileges,
  onSuccess,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields, setFieldsValue } = form;

  const { run: modifyGlobalPrivilege, loading: userLoading } = useRequest(
    ObUserController.modifyGlobalPrivilege,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage(
              {
                id: 'ocp-express.Oracle.Component.ModifyGlobalPrivilegeModal.TheSystemPermissionForUsername',
                defaultMessage: '{username} 的系统权限修改成功',
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

  const { run: modifyGlobalPrivilegeFromRole, loading: roleLoading } = useRequest(
    ObUserController.modifyGlobalPrivilegeFromRole,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage(
              {
                id: 'ocp-express.Oracle.Component.ModifyGlobalPrivilegeModal.TheSystemPermissionOfRolename',
                defaultMessage: '{roleName} 的系统权限修改成功',
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

  const submitFn = () => {
    validateFields().then(values => {
      if (username) {
        modifyGlobalPrivilege(
          {
            tenantId,
            username,
          },

          { globalPrivileges: values?.globalPrivileges }
        );
      } else {
        modifyGlobalPrivilegeFromRole(
          {
            tenantId,
            roleName,
          },

          { globalPrivileges: values.globalPrivileges }
        );
      }
    });
  };

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-express.Oracle.Component.ModifyGlobalPrivilegeModal.ModifySystemPermissions',
        defaultMessage: '修改系统权限',
      })}
      destroyOnClose={true}
      confirmLoading={userLoading || roleLoading}
      onOk={submitFn}
      {...restProps}
    >
      <Form form={form} layout="vertical" requiredMark="optional" preserve={false}>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Oracle.Component.ModifyGlobalPrivilegeModal.SystemPermissions',
            defaultMessage: '系统权限',
          })}
          tooltip={{
            placement: 'right',
            title: formatMessage({
              id: 'ocp-express.Oracle.Component.ModifyGlobalPrivilegeModal.ApplicableToAllDatabases',
              defaultMessage: '适用于所有的数据库',
            }),
          }}
          name="globalPrivileges"
          initialValue={globalPrivileges}
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
                {item.replace(/_/g, ' ')}
              </Option>
            ))}
          </MySelect>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ModifyGlobalPrivilegeModal;
