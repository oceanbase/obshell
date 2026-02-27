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

import { MAX_FORM_ITEM_LAYOUT } from '@/constant';
import {
  ORACLE_STORED_PROCEDURE_PRIVILEGE_LIST,
  ORACLE_TABLE_PRIVILEGE_LIST,
  ORACLE_VIEW_PRIVILEGE_LIST,
} from '@/constant/tenant';
import { patchRoleObjectPrivilege, patchUserObjectPrivilege } from '@/service/obshell/tenant';
import { formatMessage } from '@/util/intl';
import { Checkbox, Col, Form, message, Modal, ModalProps, Row } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import React from 'react';

/**
 * 参数说明
 * username 用户名
 * roleName 角色名
 * username / roleName 二者只需传一个
 *  */
interface ModifyObjectPrivilegeModalProps extends ModalProps {
  tenantName: string;
  username?: string;
  roleName?: string;
  dbObject: API.ObjectPrivilege;
  onSuccess: () => void;
}

const ModifyObjectPrivilegeModal: React.FC<ModifyObjectPrivilegeModalProps> = ({
  visible,
  tenantName,
  username,
  roleName,
  dbObject,
  onSuccess,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;

  const { run: patchUserObjectPrivilegeRun, loading: userLoading } = useRequest(
    patchUserObjectPrivilege,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          const objectName = dbObject?.object?.name;
          message.success(
            formatMessage(
              {
                id: 'OBShell.Oracle.Component.ModifyObjectPrivilegeModal.ObjectnameObjectPermissionsModifiedSuccessfully',
                defaultMessage: '{objectName} 的对象权限修改成功',
              },
              { objectName: objectName }
            )
          );
          if (onSuccess) {
            onSuccess();
          }
        }
      },
    }
  );

  const { run: patchRoleObjectPrivilegeRun, loading: roleLoading } = useRequest(
    patchRoleObjectPrivilege,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          const objectName = dbObject?.object?.name;
          message.success(
            formatMessage(
              {
                id: 'OBShell.Oracle.Component.ModifyObjectPrivilegeModal.ObjectnameObjectPermissionsModifiedSuccessfully',
                defaultMessage: '{objectName} 的对象权限修改成功',
              },
              { objectName: objectName }
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
      const { privileges } = values;
      const objectPrivileges = [
        {
          object_name: dbObject?.object?.name || '',
          object_type: dbObject?.object?.type || '',
          owner: dbObject?.object?.owner || '',
          privileges,
        },
      ];

      if (username) {
        patchUserObjectPrivilegeRun(
          {
            name: tenantName,
            user: username,
          },

          { object_privileges: objectPrivileges }
        );
      } else if (roleName) {
        patchRoleObjectPrivilegeRun(
          {
            name: tenantName,
            role: roleName,
          },

          { object_privileges: objectPrivileges }
        );
      }
    });
  };

  let privilegeOptions: string[] = ORACLE_TABLE_PRIVILEGE_LIST || [];
  if (dbObject?.object?.type === 'TABLE') {
    privilegeOptions = ORACLE_TABLE_PRIVILEGE_LIST;
    if (roleName) {
      privilegeOptions = ORACLE_TABLE_PRIVILEGE_LIST.filter(
        item => item !== 'INDEX' && item !== 'REFERENCES'
      );
    }
  } else if (dbObject?.object?.type === 'VIEW') {
    privilegeOptions = ORACLE_VIEW_PRIVILEGE_LIST;
  } else if (dbObject?.object?.type === 'PROCEDURE') {
    privilegeOptions = ORACLE_STORED_PROCEDURE_PRIVILEGE_LIST;
  }

  return (
    <Modal
      open={visible}
      title={formatMessage({
        id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeModal.ModifyPermissions',
        defaultMessage: '修改权限',
      })}
      destroyOnClose={true}
      confirmLoading={userLoading || roleLoading}
      onOk={submitFn}
      okText={formatMessage({
        id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeModal.Submitted',
        defaultMessage: '提交',
      })}
      {...restProps}
    >
      <Form
        form={form}
        layout="vertical"
        hideRequiredMark
        preserve={false}
        {...MAX_FORM_ITEM_LAYOUT}
      >
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeModal.Object',
            defaultMessage: '对象',
          })}
          name="object"
        >
          {dbObject?.object?.owner}.{dbObject?.object?.name}
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeModal.GrantPermissions',
            defaultMessage: '授予权限',
          })}
          name="privileges"
          initialValue={dbObject?.privileges}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeModal.SelectAtLeastOnePermission',
                defaultMessage: '至少选择一个权限',
              }),
            },
          ]}
        >
          <Checkbox.Group>
            <Row>
              {privilegeOptions.map(item => (
                <Col span={6}>
                  <Checkbox key={item} value={item}>
                    {item}
                  </Checkbox>
                </Col>
              ))}
            </Row>
          </Checkbox.Group>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ModifyObjectPrivilegeModal;
