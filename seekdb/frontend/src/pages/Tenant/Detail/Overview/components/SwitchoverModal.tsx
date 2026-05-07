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

import { standbySwitchover } from '@/service/obshell/seekdb';
import { taskSuccess } from '@/util/task';
import { Alert, Form, Modal, Select, Typography } from '@oceanbase/design';
import { RightOutlined } from '@oceanbase/icons';
import { useRequest } from 'ahooks';
import React from 'react';

const { Text } = Typography;

export interface SwitchoverModalProps {
  visible?: boolean;
  peers: API.PeerStandbyStatus[];
  defaultPeer?: API.PeerStandbyStatus;
  localInstanceName?: string;
  onCancel: () => void;
  onSuccess: () => void;
}

const buildPeerValue = (peer?: API.PeerStandbyStatus) =>
  `${peer?.peer_host || ''}:${peer?.peer_obshell_port || 0}`;

const getPeerDisplayName = (peer?: API.PeerStandbyStatus) =>
  peer?.instance_name || `${peer?.peer_host || '-'}:${peer?.peer_obshell_port || '-'}`;

const SwitchoverModal: React.FC<SwitchoverModalProps> = ({
  visible,
  peers,
  defaultPeer,
  localInstanceName,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const isSingleStandby = peers.length <= 1;

  const { runAsync, loading } = useRequest(standbySwitchover, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        taskSuccess({
          taskId: res.data?.id,
          message: formatMessage({
            id: 'seekdb.Overview.components.SwitchoverModal.DailySwitchingTaskHasBeen',
            defaultMessage: '日常切换的任务已提交',
          }),
        });
        onSuccess();
      }
    },
  });

  const handleSubmit = async () => {
    let peer: API.PeerStandbyStatus | undefined;
    if (isSingleStandby) {
      peer = defaultPeer || peers[0];
    } else {
      const values = await form.validateFields();
      peer = peers.find(item => buildPeerValue(item) === values.peerKey);
    }
    if (!peer?.peer_host || !peer?.peer_obshell_port) {
      return;
    }

    await runAsync({
      peer_host: peer.peer_host,
      peer_obshell_port: peer.peer_obshell_port,
    });
  };

  const singlePeer = defaultPeer || peers[0];

  return (
    <Modal
      width={696}
      visible={visible}
      destroyOnClose={true}
      title={formatMessage({
        id: 'seekdb.Overview.components.SwitchoverModal.DailySwitching',
        defaultMessage: '日常切换',
      })}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      okText={formatMessage({
        id: 'seekdb.Overview.components.SwitchoverModal.Switch',
        defaultMessage: '切换',
      })}
      cancelText={formatMessage({
        id: 'seekdb.Overview.components.SwitchoverModal.Cancel',
        defaultMessage: '取消',
      })}
    >
      <Alert
        type="info"
        showIcon={true}
        style={{ marginBottom: 24 }}
        message={
          isSingleStandby
            ? formatMessage({
                id: 'seekdb.Overview.components.SwitchoverModal.AfterTheSwitchIsSuccessful',
                defaultMessage: '切换成功后，原主实例将自动变更为备实例。',
              })
            : formatMessage({
                id: 'seekdb.Overview.components.SwitchoverModal.SelectASecondaryInstanceAs',
                defaultMessage:
                  '请选择一个备实例作为主实例；切换成功后，原主实例将自动变更为备实例。',
              })
        }
      />

      <Form form={form} preserve={false}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
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
                  id: 'seekdb.Overview.components.SwitchoverModal.OriginalMasterInstance',
                  defaultMessage: '原主实例',
                })}
              </Text>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <Text style={{ fontSize: 14, fontWeight: 400, color: '#5C6B8A' }}>
                {formatMessage({
                  id: 'seekdb.Overview.components.SwitchoverModal.InstanceName',
                  defaultMessage: '实例名',
                })}
              </Text>
              <Text style={{ fontSize: 14, fontWeight: 400, color: '#132039' }}>
                {localInstanceName ||
                  formatMessage({
                    id: 'seekdb.Overview.components.SwitchoverModal.CurrentInstance',
                    defaultMessage: '当前实例',
                  })}
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
                  id: 'seekdb.Overview.components.SwitchoverModal.NewPrimaryInstance',
                  defaultMessage: '新主实例',
                })}
              </Text>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <Text style={{ fontSize: 14, fontWeight: 400, color: '#5C6B8A' }}>
                {formatMessage({
                  id: 'seekdb.Overview.components.SwitchoverModal.InstanceName',
                  defaultMessage: '实例名',
                })}
              </Text>
              {isSingleStandby ? (
                <Text style={{ fontSize: 14, fontWeight: 400, color: '#132039' }}>
                  {getPeerDisplayName(singlePeer)}
                </Text>
              ) : (
                <Form.Item
                  name="peerKey"
                  initialValue={defaultPeer ? buildPeerValue(defaultPeer) : undefined}
                  rules={[
                    {
                      required: true,
                      message: formatMessage({
                        id: 'seekdb.Overview.components.SwitchoverModal.PleaseSelectAnInstance',
                        defaultMessage: '请选择实例',
                      }),
                    },
                  ]}
                  style={{ marginBottom: 0 }}
                >
                  <Select
                    style={{ width: '100%' }}
                    placeholder={formatMessage({
                      id: 'seekdb.Overview.components.SwitchoverModal.PleaseSelect',
                      defaultMessage: '请选择',
                    })}
                    options={peers.map(item => ({
                      value: buildPeerValue(item),
                      label: getPeerDisplayName(item),
                    }))}
                  />
                </Form.Item>
              )}
            </div>
          </div>
        </div>
      </Form>
    </Modal>
  );
};

export default SwitchoverModal;
