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
import { history } from 'umi';
import React from 'react';
import { max } from 'lodash';
import moment from 'moment';
import { Progress } from '@oceanbase/charts';
import { toPercent } from '@oceanbase/charts/es/util/number';
import { Col, Empty, Row, Typography, theme } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import * as MonitorController from '@/service/ocp-express/MonitorController';
import MyCard from '@/component/MyCard';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import { RFC3339_DATE_TIME_FORMAT } from '@/constant/datetime';
import { formatSize } from '@/util';
import { isNullValue, formatNumber, byte2GB, byte2MB } from '@oceanbase/util';

export interface TenantResourceTop3Props {
  clusterData: API.ClusterInfo;
}

const TenantResourceTop3: React.FC<TenantResourceTop3Props> = ({ clusterData }) => {
  const { token } = theme.useToken();

  const tenantStats = clusterData?.tenant_stats || [];

  const tenantMetricList = [
    {
      key: 'ob_cpu_percent',
      title: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.CpuUsage',
        defaultMessage: 'CPU 消耗比',
      }),
      description: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.PercentageOfCpuUsedToTenantCpu',
        defaultMessage: '已使用 CPU 占分配 CPU 的百分比',
      }),
      chartData: tenantStats
        .map(item => ({
          ...item,
          percentValue: item.cpu_used_percent,
        }))
        .sort((a, b) => b.percentValue - a.percentValue)
        .slice(0, 3),
    },

    {
      key: 'ob_memory_percent',
      title: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.MemoryUsage',
        defaultMessage: '内存消耗比',
      }),
      description: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.UsedMemoryAsAPercentageOfTenantMemory',
        defaultMessage: '分配内存总量，以及已使用内存占分配内存的百分比',
      }),
      chartData: tenantStats
        .map(item => ({
          ...item,
          percentValue: item.memory_used_percent,
        }))
        .sort((a, b) => b.percentValue - a.percentValue)
        .slice(0, 3),
    },

    {
      key: 'ob_tenant_disk_usage',
      title: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.DataVolume',
        defaultMessage: '数据量',
      }),
      description: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.TenantDataSizeAndPercentageOfClusterDisk',
        defaultMessage: '租户数据量大小，以及占集群磁盘容量的百分比',
      }),
      chartData: tenantStats
        .map(item => ({
          ...item,
          usedValue: item.data_disk_usage,
          percentValue: (item.data_disk_usage / clusterData?.stats?.disk_in_bytes_total) * 100,
        }))
        .sort((a, b) => b.percentValue - a.percentValue)
        .slice(0, 3),
    },
  ];

  const tenantCount = max(tenantMetricList.map(item => item.chartData?.length)) || 3;

  return (
    <MyCard
      title={formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.TenantResourceUsageTop',
        defaultMessage: '租户资源使用 Top3',
      })}
      headStyle={{
        marginBottom: 16,
      }}
      style={{
        height: tenantCount < 3 ? 210 - (3 - tenantCount) * 32 : 210,
      }}
      bordered={false}
    >
      <Row gutter={48}>
        {tenantMetricList.map((item, index) => {
          return (
            <Col
              key={item.key}
              span={8}
              style={
                index !== tenantMetricList.length - 1
                  ? { borderRight: `1px solid ${token.colorBorderSecondary}` }
                  : {}
              }
            >
              <MyCard
                loading={item.loading}
                title={
                  <div
                    style={{
                      fontSize: 14,
                      fontWeight: 400,
                      fontFamily: 'PingFangSC',
                      color: token.colorText,
                    }}
                  >
                    <ContentWithQuestion
                      content={item.title}
                      tooltip={{
                        title: item.description,
                      }}
                    />
                  </div>
                }
                headStyle={{
                  marginBottom: 16,
                }}
                className="card-without-padding"
              >
                {item.chartData.length > 0 ? (
                  <Row gutter={[0, 8]}>
                    {item.chartData.map(dataItem => (
                      <Col key={dataItem.tenant_name} span={24}>
                        <Progress
                          compact={true}
                          title={
                            <Typography.Text
                              data-aspm-click="c304255.d308758"
                              data-aspm-desc="租户资源使用 Top3-跳转租户详情"
                              data-aspm-param={``}
                              data-aspm-expo
                              ellipsis={{
                                tooltip: true,
                              }}
                              // TODO: 不存在监控, 先禁止跳转
                              onClick={() => {
                                // history.push(`/tenant/${dataItem.tenant_name}/monitor`);
                              }} // className="ocp-link-hover"
                              style={{ width: 70, color: token.colorTextTertiary }}
                            >
                              {dataItem.tenant_name}
                            </Typography.Text>
                          }
                          // Progress 的 percent 为 0 ~ 1 的值，监控返回的 percent 是 0 ~ 100 的百分比值，需要进行换算
                          percent={dataItem.percentValue / 100}
                          formatter={() => {
                            if (item.key === 'ob_tenant_disk_usage') {
                              // 刚部署的 OCP Express，可能存在 percentValue 为空、usedValue 不为空的情况，为了避免百分比出现 NaN，需要做下空值判断
                              return `${formatSize(dataItem.usedValue)} / ${
                                isNullValue(dataItem.percentValue)
                                  ? '-'
                                  : toPercent(dataItem.percentValue / 100, 1)
                              }%`;
                            }

                            const memory = dataItem?.memory_in_bytes_total || 0;
                            let label1 = '';
                            if (item?.key === 'ob_memory_percent') {
                              if (byte2MB(memory) >= 1024) {
                                label1 = `${formatNumber(byte2GB(memory), 2)} GB`;
                              } else {
                                label1 = `${formatNumber(byte2MB(memory))} MB`;
                              }
                            }

                            const label2 =
                              item?.key === 'ob_cpu_percent' ? `${dataItem?.cpu_core_total}C` : '';

                            const labelResult = `${label1}${label2} | ${toPercent(
                              dataItem.percentValue / 100,
                              1
                            )}%`;

                            return formatMessage(
                              {
                                id: 'ocp-v2.Overview.TenantResourceTop3.TotalLabelresult',
                                defaultMessage: '总: {labelResult}',
                              },
                              { labelResult: labelResult }
                            );
                          }}
                          percentStyle={{
                            width: item.key === 'ob_memory_percent' ? 140 : 120,
                          }}
                          warningPercent={0.7}
                          dangerPercent={0.8}
                          maxColumnWidth={10}
                        />
                      </Col>
                    ))}
                  </Row>
                ) : (
                  <Empty
                    imageStyle={{
                      height: 72,
                    }}
                    description=""
                  />
                )}
              </MyCard>
            </Col>
          );
        })}
      </Row>
    </MyCard>
  );
};

export default TenantResourceTop3;
