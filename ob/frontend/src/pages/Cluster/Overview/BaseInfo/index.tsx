import MyCard from '@/component/MyCard';
import { formatMessage } from '@/util/intl';
import { Badge, Descriptions, theme } from '@oceanbase/design';
import { useModel } from '@umijs/max';
import React, { useState } from 'react';
import SharedStorageModal from './SharedStorageModal';

const unavailableStatus = {
  label: formatMessage({
    id: 'ocp-express.src.constant.oceanbase.Unavailable',
    defaultMessage: '不可用',
  }),

  value: 'UNAVAILABLE',
  badgeStatus: 'error',
};

const availableStatus = {
  label: formatMessage({ id: 'ocp-v2.src.constant.oceanbase.Available', defaultMessage: '可用' }),
  value: 'AVAILABLE',
  badgeStatus: 'success',
};

const BaseInfo: React.FC = () => {
  const { token } = theme.useToken();
  const { clusterData, isSharedStorage } = useModel('cluster');
  const [sharedStorageVisible, setSharedStorageVisible] = useState(false);

  const statusItem = clusterData.status === 'AVAILABLE' ? availableStatus : unavailableStatus;

  // 将 badge 状态映射为 color
  const colorMap: Record<string, string> = {
    processing: token.colorPrimary,
    success: token.colorSuccess,
    warning: token.colorWarning,
    error: token.colorError,
  };

  return (
    <MyCard
      title={formatMessage({
        id: 'OBShell.Overview.BaseInfo.BasicInformation',
        defaultMessage: '基本信息',
      })}
    >
      <Descriptions column={3}>
        <Descriptions.Item
          label={formatMessage({
            id: 'OBShell.Overview.BaseInfo.ClusterName',
            defaultMessage: '集群名',
          })}
        >
          {clusterData.cluster_name}
        </Descriptions.Item>
        {isSharedStorage && (
          <Descriptions.Item
            label={formatMessage({
              id: 'OBShell.Overview.BaseInfo.ClusterId',
              defaultMessage: '集群ID',
            })}
          >
            {clusterData.cluster_id}
          </Descriptions.Item>
        )}
        <Descriptions.Item
          label={formatMessage({ id: 'OBShell.Overview.BaseInfo.Status', defaultMessage: '状态' })}
        >
          <Badge
            status={statusItem.badgeStatus}
            text={
              <span
                style={{
                  color: colorMap[statusItem.badgeStatus as string],
                }}
              >
                {statusItem.label}
              </span>
            }
          />
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'OBShell.Overview.BaseInfo.OceanbaseVersionNumber',
            defaultMessage: 'OceanBase 版本号',
          })}
        >
          {clusterData.ob_version}
        </Descriptions.Item>
        {isSharedStorage && (
          <>
            <Descriptions.Item
              label={formatMessage({
                id: 'OBShell.Overview.BaseInfo.NumberOfRegions',
                defaultMessage: 'Region 数量',
              })}
            >
              {clusterData.zones?.length}
            </Descriptions.Item>
            <Descriptions.Item
              label={formatMessage({
                id: 'OBShell.Overview.BaseInfo.SharedStorage',
                defaultMessage: '共享存储',
              })}
            >
              <a onClick={() => setSharedStorageVisible(true)}>
                {clusterData.shared_storage_info?.storage_path}
              </a>
            </Descriptions.Item>
          </>
        )}
      </Descriptions>
      <SharedStorageModal
        visible={sharedStorageVisible}
        setSharedStorageVisible={setSharedStorageVisible}
        sharedStorageInfo={clusterData?.shared_storage_info}
        onCancel={() => setSharedStorageVisible(false)}
      />
    </MyCard>
  );
};

export default BaseInfo;
