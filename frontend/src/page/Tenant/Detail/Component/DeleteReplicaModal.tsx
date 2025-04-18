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
import React from 'react';
import { Alert, Descriptions, Modal } from '@oceanbase/design';
import type { ModalProps } from '@oceanbase/design/es/modal';
import { useRequest } from 'ahooks';
import * as ObTenantController from '@/service/ocp-express/ObTenantController';
import { taskSuccess } from '@/util/task';
import { tenantRemoveReplicas } from '@/service/obshell/tenant';

export interface DeleteReplicaModalProps extends ModalProps {
  dispatch: any;
  tenantData: API.TenantInfo;
  tenantZones: API.TenantZone[];
  onSuccess: () => void;
}

const DeleteReplicaModal: React.FC<DeleteReplicaModalProps> = ({
  dispatch,
  tenantData,
  tenantZones = [],
  onSuccess,
  ...restProps
}) => {
  const { run: deleteReplica, loading } = useRequest(tenantRemoveReplicas, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const taskId = res.data?.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'ocp-express.src.model.tenant.TheTaskOfDeletingThe',
            defaultMessage: '删除副本的任务提交成功',
          }),
        });
        if (onSuccess) {
          onSuccess();
        }
        dispatch({
          type: 'tenant/getTenantData',
          payload: {
            id: tenantData.clusterId,
            tenantId: tenantData.id,
          },
        });
        dispatch({
          type: 'task/update',
          payload: {
            runningTaskListDataRefreshDep: taskId,
          },
        });
      }
    },
  });

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-express.Detail.Component.DeleteReplicaModal.AreYouSureYouWant',
        defaultMessage: '确定要删除副本吗？',
      })}
      destroyOnClose={true}
      confirmLoading={loading}
      onOk={() => {
        deleteReplica(
          {
            name: tenantData.tenant_name,
          },
          {
            zones: tenantZones.map(item => item.name),
          }
        );
      }}
      okText={formatMessage({
        id: 'ocp-express.Detail.Component.DeleteReplicaModal.Delete',
        defaultMessage: '删除',
      })}
      okButtonProps={{
        danger: true,
        ghost: true,
      }}
      {...restProps}
    >
      <Alert
        type="warning"
        showIcon={true}
        message={formatMessage({
          id: 'ocp-express.Detail.Component.DeleteReplicaModal.DeletingTheCopyDataWill',
          defaultMessage: '删除副本数据将不可恢复，请谨慎操作',
        })}
        style={{ marginBottom: 24 }}
      />

      <Descriptions column={1}>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.DeleteReplicaModal.Tenant',
            defaultMessage: '所在租户',
          })}
        >
          {tenantData.tenant_name}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.BatchDeleteReplicaModal.Zone',
            defaultMessage: '所在 Zone',
          })}
        >
          {tenantZones
            .map(item => {
              // const replicaTypeLabel = findByValue(REPLICA_TYPE_LIST, item.replicaType).label;
              return item.name;
            })
            .join('、')}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.DeleteReplicaModal.ReplicaType',
            defaultMessage: '副本类型',
          })}
        >
          {formatMessage({
            id: 'ocp-express.Detail.Component.DeleteReplicaModal.AllPurposeCopy',
            defaultMessage: '全能型副本',
          })}
        </Descriptions.Item>
      </Descriptions>
    </Modal>
  );
};

export default connect(() => ({}))(DeleteReplicaModal);
