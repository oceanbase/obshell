import { formatMessage } from '@/util/intl';
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

import { DATE_TIME_FORMAT } from '@/constant/datetime';
import { standbyActivate } from '@/service/obshell/seekdb';
import { taskSuccess } from '@/util/task';
import { Alert, Collapse, Modal, Table, Typography } from '@oceanbase/design';
import { RightOutlined } from '@oceanbase/icons';
import { useRequest } from 'ahooks';
import dayjs from 'dayjs';
import React from 'react';

/** SYNC_SCN 与备库延迟计算一致，约 1e9 对应 1 秒，按纳秒时间线转为本地时间 */
const formatSyncScnToDateTime = (scn?: number) => {
  if (scn == null || scn <= 0) return '-';
  const ms = Math.floor(scn / 1_000_000);
  const d = dayjs(ms);
  return d.isValid() ? d.format(DATE_TIME_FORMAT) : String(scn);
};

const { Text } = Typography;
const { Panel } = Collapse;

export interface ActivateModalProps {
  visible?: boolean;
  localInstanceName?: string;
  localStatus?: API.LocalStandbyStatus;
  upstreamPeer?: API.PeerStandbyStatus;
  onCancel: () => void;
  onSuccess: () => void;
}

const ActivateModal: React.FC<ActivateModalProps> = ({
  visible,
  localInstanceName,
  localStatus,
  upstreamPeer,
  onCancel,
  onSuccess,
}) => {
  const { runAsync, loading } = useRequest(standbyActivate, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        taskSuccess({
          taskId: res.data?.id,
          message: formatMessage({
            id: 'seekdb.Overview.components.ActivateModal.TheTaskOfDisasterRecovery',
            defaultMessage: '容灾切换的任务已提交',
          }),
        });
        onSuccess();
      }
    },
  });

  const handleSubmit = async () => {
    await runAsync();
  };

  const upstreamDisplayName =
    upstreamPeer?.instance_name ||
    (upstreamPeer
      ? `${upstreamPeer.peer_host || '-'}:${upstreamPeer.peer_obshell_port || '-'}`
      : '-');

  const consistencyColumns = [
    {
      title: formatMessage({
        id: 'seekdb.Overview.components.ActivateModal.NameOfTheSpareTenant',
        defaultMessage: '备租户名',
      }),
      dataIndex: 'tenantName',
    },
    {
      title: formatMessage({
        id: 'seekdb.Overview.components.ActivateModal.SynchronizationPointInTime',
        defaultMessage: '同步时间点',
      }),
      dataIndex: 'syncScn',
    },
    {
      title: formatMessage({
        id: 'seekdb.Overview.components.ActivateModal.ConsistentSite',
        defaultMessage: '一致位点',
      }),
      dataIndex: 'readableScn',
    },
  ];

  const consistencyData = localStatus
    ? [
        {
          key: 'local',
          tenantName: localInstanceName || localStatus.instance_name,
          syncScn: formatSyncScnToDateTime(localStatus.sync_scn),
          readableScn: localStatus.readable_scn ?? '-',
        },
      ]
    : [];

  return (
    <Modal
      width={696}
      visible={visible}
      destroyOnClose={true}
      title={formatMessage({
        id: 'seekdb.Overview.components.ActivateModal.DisasterRecoverySwitchover',
        defaultMessage: '容灾切换',
      })}
      onCancel={onCancel}
      onOk={handleSubmit}
      okButtonProps={{ ghost: true, danger: true }}
      confirmLoading={loading}
      okText={formatMessage({
        id: 'seekdb.Overview.components.ActivateModal.Switch',
        defaultMessage: '切换',
      })}
      cancelText={formatMessage({
        id: 'seekdb.Overview.components.ActivateModal.Cancel',
        defaultMessage: '取消',
      })}
    >
      <Alert
        showIcon={true}
        style={{ marginBottom: 24 }}
        message={formatMessage({
          id: 'seekdb.Overview.components.ActivateModal.AfterTheSwitchIsSuccessful',
          defaultMessage: '切换成功后，原主实例将自动变更为备实例。',
        })}
      />

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          marginBottom: 16,
        }}
      >
        <div
          style={{
            flex: 1,
            minWidth: 0,
            padding: 16,
            border: '1px solid #E2E8F3',
            borderRadius: 8,
          }}
        >
          <div style={{ marginBottom: 24 }}>
            <Text style={{ fontSize: 14, fontWeight: 500, color: '#132039' }}>
              {formatMessage({
                id: 'seekdb.Overview.components.ActivateModal.OriginalPrimaryInstance',
                defaultMessage: '原主实例',
              })}
            </Text>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Text style={{ fontSize: 14, fontWeight: 400, color: '#5C6B8A' }}>
              {formatMessage({
                id: 'seekdb.Overview.components.ActivateModal.InstanceName',
                defaultMessage: '实例名',
              })}
            </Text>
            <Text style={{ fontSize: 14, fontWeight: 400, color: '#132039' }}>
              {upstreamDisplayName}
            </Text>
          </div>
        </div>
        <RightOutlined style={{ fontSize: 16, color: '#132039', flexShrink: 0 }} />
        <div
          style={{
            flex: 1,
            minWidth: 0,
            padding: 16,
            border: '1px solid #E2E8F3',
            borderRadius: 8,
          }}
        >
          <div style={{ marginBottom: 24 }}>
            <Text style={{ fontSize: 14, fontWeight: 500, color: '#132039' }}>
              {formatMessage({
                id: 'seekdb.Overview.components.ActivateModal.NewPrimaryInstance',
                defaultMessage: '新主实例',
              })}
            </Text>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Text style={{ fontSize: 14, fontWeight: 400, color: '#5C6B8A' }}>
              {formatMessage({
                id: 'seekdb.Overview.components.ActivateModal.InstanceName',
                defaultMessage: '实例名',
              })}
            </Text>
            <Text style={{ fontSize: 14, fontWeight: 400, color: '#132039' }}>
              {localInstanceName ||
                formatMessage({
                  id: 'seekdb.Overview.components.ActivateModal.CurrentInstance',
                  defaultMessage: '当前实例',
                })}
            </Text>
          </div>
        </div>
      </div>

      {consistencyData.length > 0 && (
        <Collapse
          ghost={true}
          defaultActiveKey={['consistency']}
          style={{ marginLeft: -16, marginRight: -16 }}
        >
          <Panel
            key="consistency"
            header={formatMessage({
              id: 'seekdb.Overview.components.ActivateModal.StandbyTenantConsistentPoint',
              defaultMessage: '备租户一致点位',
            })}
          >
            <Table
              rowKey="key"
              columns={consistencyColumns}
              dataSource={consistencyData}
              pagination={false}
              size="small"
            />
          </Panel>
        </Collapse>
      )}
    </Modal>
  );
};

export default ActivateModal;
