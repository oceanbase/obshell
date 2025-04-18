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
import { Alert, Descriptions, Modal, message, theme } from '@oceanbase/design';
import type { ModalProps } from '@oceanbase/design/es/modal';
import MyInput from '@/component/MyInput';
import { useRequest } from 'ahooks';
import * as ObTenantController from '@/service/ocp-express/ObTenantController';
import { tenantDrop } from '@/service/obshell/tenant';

export interface DeleteTenantModalProps extends ModalProps {
  tenantData: API.TenantInfo;
  onSuccess: () => void;
}

const DeleteTenantModal: React.FC<DeleteTenantModalProps> = ({
  tenantData,
  onSuccess,
  ...restProps
}) => {
  const { token } = theme.useToken();

  const [confirmCode, setConfirmCode] = useState('');

  const { run, loading } = useRequest(tenantDrop, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.Detail.Component.DeleteTenantModal.TenantDeletedSuccessfully',
            defaultMessage: '租户删除成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-express.Detail.Component.DeleteTenantModal.DeleteATenant',
        defaultMessage: '删除租户',
      })}
      destroyOnClose={true}
      confirmLoading={loading}
      onOk={() => run({ name: tenantData.tenant_name })}
      okText={formatMessage({
        id: 'ocp-express.Detail.Component.DeleteTenantModal.Delete',
        defaultMessage: '删除',
      })}
      okButtonProps={{
        danger: true,
        ghost: true,
        disabled: confirmCode !== 'delete',
      }}
      {...restProps}
    >
      <Alert
        type="warning"
        showIcon={true}
        message={formatMessage({
          id: 'ocp-express.Detail.Component.DeleteTenantModal.DataCannotBeRecoveredAfter',
          defaultMessage: '租户删除后数据将不可恢复，请谨慎操作',
        })}
        style={{ marginBottom: 24 }}
      />

      <Descriptions column={1}>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.DeleteTenantModal.Tenant',
            defaultMessage: '租户',
          })}
        >
          {tenantData.tenant_name}
        </Descriptions.Item>
      </Descriptions>
      <div>
        {formatMessage({
          id: 'ocp-express.Detail.Component.DeleteTenantModal.PleaseEnter',
          defaultMessage: '请输入',
        })}

        <span style={{ color: token.colorError }}> delete </span>
        {formatMessage({
          id: 'ocp-express.Detail.Component.DeleteTenantModal.ConfirmOperation',
          defaultMessage: '确认操作',
        })}
      </div>
      <MyInput
        style={{ width: 400, marginTop: 8 }}
        onChange={e => setConfirmCode(e.target.value)}
      />
    </Modal>
  );
};

export default DeleteTenantModal;
