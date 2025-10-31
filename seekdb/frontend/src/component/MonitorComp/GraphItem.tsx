import React from 'react';
import { Card } from '@oceanbase/design';
import IconTip from '../IconTip';
import LineGraph from './LineGraph';
import styles from './index.less';

interface GraphItemProps {
  graphContainer: API.MetricGroup;
  graphIdx: number;
  activeTabKey: string;
  isRefresh: boolean;
  queryRange: API.QueryRange;
  filterLabel: Monitor.MetricsLabels;
  groupLabels: Monitor.LabelKeys[];
  type: Monitor.MonitorUseTarget;
  useFor: Monitor.MonitorUseFor;
  filterData?: any[];
  filterQueryMetric?: Monitor.MetricsLabels;
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
  groupLabels,
  type,
  useFor,
  filterData,
  filterQueryMetric,
}) => {
  const graphTitle = generateGraphTitle(graphContainer, type);
  const graphId = generateGraphId(activeTabKey, graphContainer.name);

  return (
    <Card className={styles.monitorItem} key={graphIdx}>
      <div className={styles.graphHeader}>
        <IconTip
          tip={graphContainer.description}
          style={{ fontSize: 16 }}
          content={<span className={styles.graphHeaderText}>{graphTitle}</span>}
        />
      </div>
      <LineGraph
        id={graphId}
        isRefresh={isRefresh}
        queryRange={queryRange}
        metrics={graphContainer.metrics}
        labels={filterLabel}
        groupLabels={groupLabels}
        type={type}
        useFor={useFor}
        filterData={filterData}
        filterQueryMetric={filterQueryMetric}
      />
    </Card>
  );
};

export { GraphItem };
