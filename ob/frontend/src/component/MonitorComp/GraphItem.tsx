import { Card, theme } from '@oceanbase/design';
import { find, isEmpty } from 'lodash';
import React, { useState } from 'react';
import IconTip from '../IconTip';
import EditMetrics from './EditMetrics';
import styles from './index.less';
import LineGraph from './LineGraph';

interface GraphItemProps {
  graphContainer: API.MetricGroup;
  graphIdx: number;
  activeTabKey: string;
  isRefresh: boolean;
  queryRange: API.QueryRange;
  filterLabel: Monitor.MetricsLabels;
  dynamicGroupLabels: Monitor.LabelKeys[];
  type: Monitor.MonitorUseTarget;
  useFor: Monitor.MonitorUseFor;
  filterData?: any[];
  filterQueryMetric?: Monitor.MetricsLabels;
  activeDimension?: string;
}

// 生成图表标题文本
const generateGraphTitle = (
  graphContainer: API.MetricGroup,
  type: Monitor.MonitorUseTarget
): string => {
  const { name, metrics } = graphContainer;
  const firstMetric = metrics?.[0];

  if (!firstMetric?.unit) {
    return name;
  }

  const unitText = firstMetric.unit;
  const isOverview = type === 'OVERVIEW';
  const keyText = isOverview ? firstMetric.key : '';
  const separator = unitText && isOverview ? ',' : '';

  return `${name}(${unitText}${separator}${keyText})`;
};

// 生成图表唯一ID
const generateGraphId = (activeTabKey: string, graphName: string): string => {
  return `monitor-${activeTabKey}-${graphName.replace(/\s+/g, '')}`;
};

const GraphItem: React.FC<GraphItemProps> = ({
  graphContainer,
  graphIdx,
  activeTabKey,
  isRefresh,
  queryRange,
  filterLabel,
  dynamicGroupLabels,
  type,
  useFor,
  filterData,
  filterQueryMetric,
  activeDimension,
}) => {
  const { token } = theme.useToken();

  // 管理当前显示的指标
  const [metrics, setMetrics] = useState(graphContainer.metrics);

  const graphTitle = generateGraphTitle(graphContainer, type);
  const graphId = generateGraphId(activeTabKey, graphContainer.name);

  // 当前展示的指标名称列表
  const metricsName = metrics.map(item => item.name as string);

  return (
    <Card
      className={styles.monitorItem}
      key={graphIdx}
      styles={{
        body: {
          padding: 16,
        },
      }}
    >
      <div className={styles.graphHeader}>
        <IconTip
          tip={graphContainer.description}
          iconProps={{ style: { color: token.colorTextSecondary } }}
          content={<span className={styles.graphHeaderText}>{graphTitle}</span>}
        />
        {!isEmpty(graphContainer.metrics) && (
          <EditMetrics
            key={graphId}
            groupMetrics={graphContainer.metrics}
            initialKeys={metricsName}
            onOk={selectedMetrics => {
              const newMetrics = selectedMetrics.map(item => {
                return find(graphContainer.metrics, metric => metric.name === item)!;
              });
              setMetrics(newMetrics);
            }}
          />
        )}
      </div>
      <LineGraph
        id={graphId}
        isRefresh={isRefresh}
        queryRange={queryRange}
        metrics={metrics}
        labels={filterLabel}
        groupLabels={dynamicGroupLabels}
        type={type}
        useFor={useFor}
        filterData={filterData}
        filterQueryMetric={filterQueryMetric}
        activeDimension={activeDimension}
      />
    </Card>
  );
};

export { GraphItem };
