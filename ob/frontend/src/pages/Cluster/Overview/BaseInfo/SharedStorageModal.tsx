import { formatMessage } from '@/util/intl';
import { formatterNumber } from '@/util';
import { Alert, Descriptions, DescriptionsProps, Modal } from '@oceanbase/design';
import React, { useState } from 'react';
import RotatingKeyModal from './RotatingKeyModal';

export interface SharedStorageModalProps {
  visible: boolean;
  sharedStorageInfo?: API.SharedStorageInfo;
  setSharedStorageVisible: (visible: boolean) => void;
  onCancel: () => void;
}

const SharedStorageModal: React.FC<SharedStorageModalProps> = ({
  visible,
  sharedStorageInfo,
  setSharedStorageVisible,
  onCancel,
}) => {
  const [rotatingKeyVisible, setRotatingKeyVisible] = useState(false);
  const baseItemList: DescriptionsProps['items'] = [
    {
      label: formatMessage({
        id: 'OBShell.Overview.BaseInfo.SharedStorageModal.StorageDirectory',
        defaultMessage: '存储目录',
      }),
      children: (
        <div style={{ wordBreak: 'break-all', whiteSpace: 'pre-wrap' }}>
          {sharedStorageInfo?.storage_path}
        </div>
      ),
    },
    {
      label: formatMessage({
        id: 'OBShell.Overview.BaseInfo.SharedStorageModal.AccessDomainName',
        defaultMessage: '访问域名',
      }),
      children: sharedStorageInfo?.access_domain,
    },
    {
      label: formatMessage({
        id: 'OBShell.Overview.BaseInfo.SharedStorageModal.Region',
        defaultMessage: '地域',
      }),
      children: sharedStorageInfo?.region,
    },
    {
      label: 'AK',
      children: sharedStorageInfo?.access_key,
    },
  ];

  const advancedItemList: DescriptionsProps['items'] = [
    {
      label: 'delete_mode',
      children: sharedStorageInfo?.delete_mode,
    },
    {
      label: 'checksum_type',
      children: sharedStorageInfo?.checksum_type,
    },
    {
      label: 'addressing_model',
      children: sharedStorageInfo?.addressing_model,
    },
    {
      label: 'MAX-IOPS',
      children: formatterNumber(sharedStorageInfo?.max_iops),
    },
    {
      label: 'MAX-BANDWITH',
      children: sharedStorageInfo?.max_bandwidth,
    },
  ];

  const titleStyle: React.CSSProperties = {
    fontSize: 14,
  };

  return (
    <Modal
      visible={visible}
      onCancel={onCancel}
      title={formatMessage({
        id: 'OBShell.Overview.BaseInfo.SharedStorageModal.SharedStorage',
        defaultMessage: '共享存储',
      })}
      footer={null}
    >
      <Descriptions
        column={1}
        size="small"
        title={formatMessage({
          id: 'OBShell.Overview.BaseInfo.SharedStorageModal.BasicConfiguration',
          defaultMessage: '基础配置',
        })}
        items={baseItemList}
        styles={{ title: titleStyle }}
      />

      <Alert
        showIcon
        message={
          <>
            {formatMessage({
              id: 'OBShell.Overview.BaseInfo.SharedStorageModal.WeRecommendThatYouPeriodically',
              defaultMessage:
                '建议您定期更换共享存储的 AK/SK，进而有效防止密钥泄漏，提升存储系统整体安全性和可管控性。',
            })}

            <a
              onClick={() => {
                setRotatingKeyVisible(true);
              }}
            >
              {formatMessage({
                id: 'OBShell.Overview.BaseInfo.SharedStorageModal.RotaryKey',
                defaultMessage: '轮转密钥',
              })}
            </a>
          </>
        }
        style={{
          margin: '8px 0 24px',
        }}
        type="info"
      />

      <Descriptions
        column={1}
        size="small"
        title={formatMessage({
          id: 'OBShell.Overview.BaseInfo.SharedStorageModal.AdvancedConfiguration',
          defaultMessage: '高级配置',
        })}
        items={advancedItemList}
        styles={{ title: titleStyle }}
      />

      <RotatingKeyModal
        visible={rotatingKeyVisible}
        setSharedStorageVisible={setSharedStorageVisible}
        onCancel={() => setRotatingKeyVisible(false)}
      />
    </Modal>
  );
};

export default SharedStorageModal;
