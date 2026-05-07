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

import { standbyPairDelete } from '@/service/obshell/seekdb';
import { Alert, Modal, message } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import React from 'react';

export interface DecoupleModalProps {
  visible?: boolean;
  upstreamPeer?: API.PeerStandbyStatus;
  onCancel: () => void;
  onSuccess: () => void;
}

const DecoupleModal: React.FC<DecoupleModalProps> = ({
  visible,
  upstreamPeer,
  onCancel,
  onSuccess,
}) => {
  const { runAsync, loading } = useRequest(standbyPairDelete, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'seekdb.Overview.components.DecoupleModal.PrimaryStandbyDecouplingOperationSubmitted',
            defaultMessage: '主备解耦操作已提交',
          })
        );
        onSuccess();
      }
    },
  });

  const handleSubmit = async () => {
    if (!upstreamPeer?.peer_host || !upstreamPeer?.peer_obshell_port) {
      return;
    }
    await runAsync({
      peer_host: upstreamPeer.peer_host,
      peer_obshell_port: upstreamPeer.peer_obshell_port,
      notify_peer: true,
    });
  };

  return (
    <Modal
      visible={visible}
      destroyOnClose={true}
      title={formatMessage({
        id: 'seekdb.Overview.components.DecoupleModal.MainStandbyDecoupling',
        defaultMessage: '主备解耦',
      })}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      okText={formatMessage({
        id: 'seekdb.Overview.components.DecoupleModal.Decoupling',
        defaultMessage: '解耦',
      })}
      okButtonProps={{ danger: true, ghost: true }}
      cancelText={formatMessage({
        id: 'seekdb.Overview.components.DecoupleModal.Cancel',
        defaultMessage: '取消',
      })}
    >
      <Alert
        type="warning"
        showIcon={true}
        message={formatMessage({
          id: 'seekdb.Overview.components.DecoupleModal.AreYouSureYouWant',
          defaultMessage: '确定要执行主备解耦吗？',
        })}
        description={formatMessage({
          id: 'seekdb.Overview.components.DecoupleModal.TheDecoupledSecondaryInstanceWill',
          defaultMessage: '解耦后的备实例将作为独立实例存在，不再从原主实例同步数据，请谨慎操作。',
        })}
      />
    </Modal>
  );
};

export default DecoupleModal;
