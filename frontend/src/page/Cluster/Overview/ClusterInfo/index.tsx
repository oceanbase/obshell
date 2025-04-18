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
import React from 'react';
import moment from 'moment';
import {
  Col,
  Descriptions,
  Row,
  Tooltip,
  Typography,
  Badge,
  Popover,
  theme,
} from '@oceanbase/design';
import { EllipsisOutlined } from '@oceanbase/icons';
import { byte2GB, findByValue, isNullValue } from '@oceanbase/util';
import { Liquid } from '@oceanbase/charts';
import { toPercent } from '@oceanbase/charts/es/util/number';
import { useRequest } from 'ahooks';
import * as MonitorController from '@/service/ocp-express/MonitorController';
import { RFC3339_DATE_TIME_FORMAT } from '@/constant/datetime';
import { formatSize, isEnglish } from '@/util';
import tracert from '@/util/tracert';
import { getObServerCountByObCluster, getIdcCountByObCluster } from '@/util/oceanbase';
import ObClusterDeployMode from '@/component/ObClusterDeployMode';
import MyCard from '@/component/MyCard';
import { OB_CLUSTER_STATUS_LIST } from '@/constant/oceanbase';
import MouseTooltip from '@/component/MouseTooltip';
import useStyles from './index.style';

const { Text } = Typography;

export interface DetailProps {
  clusterData: API.ClusterInfo;
}

const Detail: React.FC<DetailProps> = ({ clusterData }) => {
  const { styles } = useStyles();
  const { token } = theme.useToken();

  const statusItem = findByValue(OB_CLUSTER_STATUS_LIST, clusterData.status);
  // 将 badge 状态映射为 color
  const colorMap = {
    processing: token.colorPrimary,
    success: token.colorSuccess,
    warning: token.colorWarning,
    error: token.colorError,
  };

  // 部署模式
  const deployModeString = (clusterData.zones || [])
    .map(item => ({
      regionName: item.region_name,
      serverCount: (item.servers || []).length,
    }))
    .map(item => item.serverCount)
    .join('-');

  const stats = clusterData.stats || {};

  const metricList = [
    {
      key: 'cpu',
      title: 'CPU',
      description: formatMessage({
        id: 'ocp-express.Component.ClusterInfo.Allocation',
        defaultMessage: '分配',
      }),
      percentValue: stats.cpu_core_assigned_percent,
      totalValue: stats.cpu_core_total,
      leftValue: stats.cpu_core_total - stats.cpu_core_assigned,
    },
    {
      key: 'memory',
      title: formatMessage({
        id: 'ocp-express.Component.ClusterInfo.Memory',
        defaultMessage: '内存',
      }),
      description: formatMessage({
        id: 'ocp-express.Component.ClusterInfo.Allocation',
        defaultMessage: '分配',
      }),
      percentValue: stats.memory_assigned_percent,
      totalValue: stats.memory_total,
      leftValue: byte2GB(stats.memory_in_bytes_total - stats.memory_in_bytes_assigned),
    },
    {
      key: 'disk',
      title: formatMessage({
        id: 'ocp-express.Component.ClusterInfo.Disk',
        defaultMessage: '磁盘',
      }),
      description: formatMessage({
        id: 'ocp-express.Component.ClusterInfo.Use',
        defaultMessage: '使用',
      }),
      percentValue: stats.disk_assigned_percent,
      totalValue: stats.disk_total,
      leftValue: byte2GB(stats.disk_in_bytes_total - stats.disk_in_bytes_assigned),
    },
  ];

  // 使用空字符串兜底，避免文案拼接时出现 undefined
  const clusterName = clusterData.cluster_name || '';

  return (
    <div
      data-aspm="c304183"
      data-aspm-desc="集群信息"
      data-aspm-expo
      // 扩展参数
      data-aspm-param={tracert.stringify({
        // 集群 OB 版本
        clusterObVersion: clusterData.obVersion,
        // 集群部署模式
        clusterDeployMode: deployModeString,
        // 集群机房数
        clusterIdcCount: getIdcCountByObCluster(clusterData),
        // 集群 OBServer 数
        clusterObserverCount: getObServerCountByObCluster(clusterData),
        // 集群租户数
        clusterTenantCount: clusterData.tenants?.length || 0,
      })}
    >
      <MyCard
        title={
          <div>
            <span
              style={{
                marginRight: 16,
              }}
            >
              {formatMessage(
                {
                  id: 'ocp-express.Component.ClusterInfo.ClusterClustername',
                  defaultMessage: '集群 {clusterName}',
                },
                { clusterName: clusterName }
              )}
            </span>
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
          </div>
        }
        extra={
          <Popover
            placement="bottomRight"
            arrowPointAtCenter={true}
            overlayStyle={{
              maxWidth: 200,
            }}
            content={
              <Descriptions colon={false} column={1}>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Component.ClusterInfo.ClusterName',
                    defaultMessage: '集群名称',
                  })}
                  className="descriptions-item-with-ellipsis"
                >
                  <Text
                    ellipsis={{
                      tooltip: clusterData.cluster_name,
                    }}
                  >
                    {clusterData.cluster_name}
                  </Text>
                </Descriptions.Item>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Detail.Overview.DeploymentMode',
                    defaultMessage: '部署模式',
                  })}
                  style={{
                    paddingBottom: 0,
                  }}
                >
                  {clusterData.zones?.length > 0 ? (
                    <Tooltip title={<ObClusterDeployMode clusterData={clusterData} />}>
                      <ObClusterDeployMode clusterData={clusterData} mode="text" />
                    </Tooltip>
                  ) : (
                    '-'
                  )}
                </Descriptions.Item>
              </Descriptions>
            }
            bodyStyle={{
              padding: '16px 24px',
            }}
          >
            <EllipsisOutlined
              data-aspm-click="c304183.d308772"
              data-aspm-desc="集群信息-查看更多"
              data-aspm-expo
              data-aspm-param={``}
              className="pointable"
              style={{
                fontSize: 20,
              }}
            />
          </Popover>
        }
      >
        <MouseTooltip
          style={{
            maxWidth: isEnglish() ? 600 : 500,
            padding: 16,
          }}
          overlay={
            <span>
              <Row gutter={[48, 12]}>
                {metricList.map(item => (
                  <Col key={item.key} span={8}>
                    <div
                      style={{ paddingLeft: 4, fontSize: 16, fontWeight: 500, marginBottom: 12 }}
                    >
                      {item.title}
                    </div>
                    <Descriptions
                      size="small"
                      colon={false}
                      column={1}
                      className={styles.borderRight}
                    >
                      <Descriptions.Item
                        label={formatMessage({
                          id: 'ocp-express.Component.ClusterInfo.Total',
                          defaultMessage: '总量',
                        })}
                      >
                        {isNullValue(item.totalValue)
                          ? '-'
                          : item.key === 'cpu'
                          ? `${item.totalValue} C`
                          : // 内存和磁盘需要进行单位换算
                            formatSize(item.totalValue)}
                      </Descriptions.Item>
                      <Descriptions.Item label={item.description}>
                        {/* 最多保留 1 位有效小数，需要用 toPercent 处理下 */}
                        {isNullValue(item.percentValue)
                          ? '-'
                          : `${toPercent(item.percentValue / 100, 1)}%`}
                      </Descriptions.Item>
                      <Descriptions.Item
                        label={formatMessage({
                          id: 'ocp-express.Component.ClusterInfo.Remaining',
                          defaultMessage: '剩余',
                        })}
                      >
                        {isNullValue(item.leftValue)
                          ? '-'
                          : item.key === 'cpu'
                          ? `${item.leftValue} C`
                          : // 内存和磁盘需要进行单位换算
                            formatSize(item.leftValue)}
                      </Descriptions.Item>
                    </Descriptions>
                  </Col>
                ))}
              </Row>
            </span>
          }
        >
          <Row gutter={48}>
            {metricList.map((item, index) => (
              <Col
                key={item.key}
                span={8}
                style={
                  index !== 2 ? { borderRight: `1px solid ${token.colorBorderSecondary}` } : {}
                }
              >
                <Liquid
                  height={54}
                  layout="horizontal"
                  title={item.title}
                  shape="rect"
                  // Progress 的 percent 为 0 ~ 1 的值，监控返回的 percent 是 0 ~ 100 的百分比值，需要进行换算
                  percent={item.percentValue / 100}
                  // 最多保留 1 位有效小数
                  decimal={1}
                  warningPercent={0.7}
                  dangerPercent={0.8}
                />
              </Col>
            ))}
          </Row>
        </MouseTooltip>
      </MyCard>
    </div>
  );
};

export default Detail;
