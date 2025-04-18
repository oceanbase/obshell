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

import React from 'react';
import moment from 'moment';
import { DATE_TIME_FORMAT_DISPLAY, TIME_FORMAT_WITHOUT_SECOND } from '@/constant/datetime';
import { formatValueForChart } from '@/util';
import type { ChartProps } from '@/component/Chart';
import type { TooltipScroll } from '@/component/Chart';
import Chart from '@/component/Chart';
import { useComparisonMetrichartDataList } from '@/hook/sqlDiagnosis';
import { Card } from '@oceanbase/design';

export interface MetricGroupWithChartConfig extends API.MetricGroup {
  chartConfig?: ChartProps;
  drilldownable?: string;
  tooltipScroll?: TooltipScroll;
  scope?: Global.MonitorScope;
  metrics: any[];
  name: string;
  description: string;
}

export interface CommonProps {
  // 监控卡片的顺序索引，用于判断是否为第一个卡片
  index?: number;
  /* 一级下钻维度，为空时不展示下钻入口，只支持集群、租户和主机 */
  drilldownScope?: 'ob_cluster_id' | 'tenant_name' | 'svr_ip';
  metricClass?: API.MetricClass;
  chartConfig?: ChartProps;
  scope?: Global.MonitorScope;
  app?: Global.MonitorApp;
  clusterName?: string;
  tenantName?: string;
  isRealtime?: boolean;
  interval?: number;
  baselineStartTime?: string;
  baselineEndTime?: string;
  startTime?: string;
  endTime?: string;
  zoneName?: string;
  serverIp?: string;
  mount_point?: string;
  mount_label?: string;
  task_type?: string;
  process?: string;
  device?: string;
  cpu?: string;
  zoneNameList?: string[];
  serverList?: MonitorServer[];
  tenantList?: API.TenantInfo[];
  className?: string;
  style?: React.CSSProperties;
}

export interface ComparisonMetricChartItemProps extends CommonProps {
  metricGroup: MetricGroupWithChartConfig;
}
const Item: React.FC<ComparisonMetricChartItemProps> = ({
  metricGroup,
  // 默认获取 OB 的监控数据
  app = 'OB',
  clusterName,
  // 租户
  tenantName,
  baselineStartTime,
  baselineEndTime,
  startTime,
  endTime,
  zoneNameList,
  serverList,
}) => {
  const { chartDataList = [] } = useComparisonMetrichartDataList({
    app,
    metricGroup,
    clusterName,
    tenantName,
    zoneNameList,
    serverList,
    interval: 60,
    baselineStartTime,
    baselineEndTime,
    startTime,
    endTime,
  });

  const getConfig = (chartData: any[]) => {
    return {
      height: 300,
      type: 'Line',
      xField: 'timestamp',
      yField: 'value',
      seriesField: 'metric',
      animation: false,
      meta: {
        timestamp: {
          formatter: (value: moment.MomentInput) => {
            return moment(value).format(DATE_TIME_FORMAT_DISPLAY);
          },
        },
        value: {
          formatter: (value: number) => {
            return formatValueForChart(chartData, value, chartData?.[0]?.unit as string);
          },
        },
      },
      xAxis: {
        type: 'time',
        tickCount: 4,
        label: {
          formatter: (value: moment.MomentInput) => {
            return moment(value, DATE_TIME_FORMAT_DISPLAY).format(TIME_FORMAT_WITHOUT_SECOND);
          },
        },
      },
      yAxis: {
        nice: true,
        tickCount: 3,
      },
      tooltip: {
        fields: ['value', 'timestamp', 'ComparisonTimestamp', 'metric'],
        showTitle: false,
        formatter: (datum: Global.ChartData) => {
          const { ComparisonTimestamp, timestamp } = datum;
          let value = formatValueForChart(chartData, datum.value, chartData?.[0]?.unit as string);
          value += ` (${moment(ComparisonTimestamp || timestamp).format(
            DATE_TIME_FORMAT_DISPLAY
          )})`;
          return { name: datum.metric, value };
        },
      },
    };
  };

  const { metrics = [] } = metricGroup || {};

  return (
    <Card bordered={false}>
      {metrics.map(metric => {
        const chartData =
          chartDataList.find(item => {
            return item?.[0]?.key === metric.key;
          }) || [];

        const config = getConfig(chartData);

        return (
          <Card.Grid key={metric.key} hoverable={false} style={{ width: '50%' }}>
            <div style={{ fontSize: 14, fontWeight: 600 }}>{metric.name}</div>
            <Chart data={chartData} {...config} />
          </Card.Grid>
        );
      })}
    </Card>
  );
};

export default Item;
