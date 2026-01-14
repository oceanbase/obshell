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
    <MyCard title="基本信息">
      <Descriptions column={3}>
        <Descriptions.Item label="集群名" contentProps={{ ellipsis: true, copyable: true }}>
          {clusterData.cluster_name}
        </Descriptions.Item>
        <Descriptions.Item label="集群状态">
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
        <Descriptions.Item label="OceanBase 版本号">{clusterData.ob_version}</Descriptions.Item>
        <Descriptions.Item label="软件安装路径">-</Descriptions.Item>
        <Descriptions.Item label="数据盘路径">-</Descriptions.Item>
        <Descriptions.Item label="日志盘路径">-</Descriptions.Item>
        <Descriptions.Item label="日志路径">-</Descriptions.Item>
      </Descriptions>
    </MyCard>
  );
};

export default BaseInfo;
