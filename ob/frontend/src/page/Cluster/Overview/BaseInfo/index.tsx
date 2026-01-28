import { formatMessage } from '@/util/intl';
import MyCard from '@/component/MyCard';
import { OB_CLUSTER_STATUS_LIST } from '@/constant/oceanbase';
import { Badge, Descriptions, theme } from '@oceanbase/design';
import { findByValue } from '@oceanbase/util';
import React from 'react';

export interface BaseInfoProps {
  clusterData: API.ClusterInfo;
}

const BaseInfo: React.FC<BaseInfoProps> = ({ clusterData }) => {
  const { token } = theme.useToken();

  const statusItem = findByValue(OB_CLUSTER_STATUS_LIST, clusterData.status);

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
          contentProps={{ ellipsis: true, copyable: true }}
        >
          {clusterData.cluster_name}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'OBShell.Overview.BaseInfo.ClusterStatus',
            defaultMessage: '集群状态',
          })}
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
      </Descriptions>
    </MyCard>
  );
};

export default BaseInfo;
