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
import { useRequest } from 'ahooks';
import React from 'react';
import {
  Form,
  Checkbox,
  Tag,
  Space,
  Button,
  Row,
  Col,
  message,
  DrawerProps,
} from '@oceanbase/design';
import {
  ORACLE_OBJECT_TYPE_LIST,
  ORACLE_TABLE_PRIVILEGE_LIST,
  ORACLE_VIEW_PRIVILEGE_LIST,
  ORACLE_STORED_PROCEDURE_PRIVILEGE_LIST,
} from '@/constant/tenant';
import MyDrawer from '@/component/MyDrawer';
import { patchUserObjectPrivilege, patchRoleObjectPrivilege } from '@/service/obshell/tenant';

/**
 * 参数说明
 * username 用户名
 * roleName 角色名
 * username / roleName 二者只需传一个
 *  */
interface ModifyObjectPrivilegeDrawerProps extends DrawerProps {
  tenantName: string;
  username?: string;
  roleName?: string;
  dbObjects?: API.ObjectPrivilege[];
  onSuccess: () => void;
  onCancel: () => void;
}

const ModifyObjectPrivilegeDrawer: React.FC<ModifyObjectPrivilegeDrawerProps> = ({
  tenantName,
  username,
  roleName,
  dbObjects,
  onSuccess,
  onCancel,
  ...restProps
}) => {
  const objectType = dbObjects?.[0]?.object?.type;

  const [form] = Form.useForm();
  const { validateFields } = form;

  const { run: patchUserObjectPrivilegeRun, loading: userModifyObjectLoading } = useRequest(
    patchUserObjectPrivilege,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeDrawer.PermissionModificationSucceeded',
              defaultMessage: '批量修改权限成功',
            })
          );
          if (onSuccess) {
            onSuccess();
          }
        }
      },
    }
  );

  const { run: patchRoleObjectPrivilegeRun, loading: roleModifyObjectLoading } = useRequest(
    patchRoleObjectPrivilege,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeDrawer.PermissionModificationSucceeded',
              defaultMessage: '批量修改权限成功',
            })
          );
          if (onSuccess) {
            onSuccess();
          }
        }
      },
    }
  );

  const handleSubmit = () => {
    validateFields().then(values => {
      const { privileges } = values;
      // 整理参数 用户/角色 添加对象 || 用户/角色 修改对象权限
      let param = {
        name: tenantName,
      };

      const objectPrivileges = dbObjects?.map(item => ({
        object_name: item.object?.name,
        object_type: item.object?.type,
        owner: item.object?.owner,
        privileges,
      }));

      // 用户
      if (username) {
        param = { ...param, user: username };
        patchUserObjectPrivilegeRun(param, { object_privileges: objectPrivileges });
      }
      // 角色
      if (roleName) {
        param = { ...param, role: roleName };
        patchRoleObjectPrivilegeRun(param, { object_privileges: objectPrivileges });
      }
    });
  };

  let privilegeOptions: string[] = [];
  if (objectType === 'TABLE') {
    privilegeOptions = ORACLE_TABLE_PRIVILEGE_LIST;
  } else if (objectType === 'VIEW') {
    privilegeOptions = ORACLE_VIEW_PRIVILEGE_LIST;
  } else if (objectType === 'STORED_PROCEDURE') {
    privilegeOptions = ORACLE_STORED_PROCEDURE_PRIVILEGE_LIST;
  }

  const titleTextype = ORACLE_OBJECT_TYPE_LIST.filter(item => item.value === objectType)[0]?.label;
  const titleText = formatMessage(
    {
      id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeDrawer.ModifyTitletextypePermissions',
      defaultMessage: '批量修改{titleTextype}权限',
    },
    { titleTextype }
  );
  return (
    <MyDrawer
      width={680}
      title={titleText}
      destroyOnClose={true}
      {...restProps}
      onCancel={onCancel}
      footer={
        <Space>
          <Button onClick={onCancel}>
            {formatMessage({
              id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeDrawer.Cancel',
              defaultMessage: '取消',
            })}
          </Button>
          <Button
            type="primary"
            onClick={handleSubmit}
            loading={userModifyObjectLoading || roleModifyObjectLoading}
          >
            {formatMessage({
              id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeDrawer.Determine',
              defaultMessage: '确定',
            })}
          </Button>
        </Space>
      }
    >
      <Form form={form} preserve={false} layout="vertical" hideRequiredMark={true}>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeDrawer.Object',
            defaultMessage: '对象',
          })}
        >
          <Row
            gutter={[8, 8]}
            style={{ padding: 12, maxHeight: 520, background: '#FAFAFA', overflow: 'auto' }}
          >
            {dbObjects?.map(item => (
              <Col key={item?.object?.full_name} span={6}>
                <Tag>{item?.object?.full_name}</Tag>
              </Col>
            ))}
          </Row>
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeDrawer.GrantPermissions',
            defaultMessage: '授予权限',
          })}
          name="privileges"
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.Oracle.Component.ModifyObjectPrivilegeDrawer.SelectAtLeastOnePermission',
                defaultMessage: '至少选择一个权限',
              }),
            },
          ]}
        >
          <Checkbox.Group style={{ width: '100%' }}>
            <Row gutter={[8, 8]}>
              {privilegeOptions.map(item => (
                <Col key={item} span={6}>
                  <Checkbox key={item} value={item} style={{ marginLeft: 0 }}>
                    {item}
                  </Checkbox>
                </Col>
              ))}
            </Row>
          </Checkbox.Group>
        </Form.Item>
      </Form>
    </MyDrawer>
  );
};
export default ModifyObjectPrivilegeDrawer;
