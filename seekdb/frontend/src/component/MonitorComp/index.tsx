import { formatMessage } from '@/util/intl';
import { useRequest } from 'ahooks';
import { Card } from '@oceanbase/design';

import { listAllMetrics } from '@/service/obshell/metric';
import React, { useMemo, useState } from 'react';
import styles from './index.less';
import { GraphItem } from './GraphItem';

/**
 * Queryable label:
 * ob_cluster_name
 * ob_cluster_id
 * tenant_name
 * tenant_id
 * svr_ip
 * obzone
 */

interface MonitorCompProps {
  filterLabel: Monitor.MetricsLabels;
  queryRange: Monitor.QueryRangeType;
  isRefresh?: boolean;
  queryScope: string;
  type: Monitor.MonitorUseTarget;
  groupLabels: Monitor.LabelKeys[];
  useFor?: Monitor.MonitorUseFor;
  filterData?: any[];
  filterQueryMetric?: Monitor.MetricsLabels;
}

export default function MonitorComp({
  filterLabel,
  queryRange,
  isRefresh = false,
  type,
  queryScope,
  groupLabels,
  useFor = 'cluster',
  filterData,
  filterQueryMetric,
}: MonitorCompProps) {
  const [activeTabKey, setActiveTabKey] = useState<string>('0');
  const { data: listAllMetricsRes } = useRequest(listAllMetrics, {
    defaultParams: [{ scope: queryScope }],
  });
  const allMetrics: API.MetricClass[] = listAllMetricsRes?.data?.contents || [];

  // 生成tab列表
  const tabList = useMemo(() => {
    return (
      allMetrics?.map((container: API.MetricClass, index: number) => ({
        key: index.toString(),
        label: container?.name,
      })) || []
    );
  }, [allMetrics]);

  // 生成当前激活tab的内容
  const currentTabContent = useMemo(() => {
    const currentContainer = allMetrics?.[parseInt(activeTabKey)];
    if (!currentContainer?.metric_groups) return null;

    const graphList = currentContainer?.metric_groups?.map(
      (graphContainer: API.MetricGroup, graphIdx: number) => (
        <GraphItem
          key={graphIdx}
          graphContainer={graphContainer}
          graphIdx={graphIdx}
          activeTabKey={activeTabKey}
          isRefresh={isRefresh}
          queryRange={queryRange}
          filterLabel={filterLabel}
          groupLabels={groupLabels}
          type={type}
          useFor={useFor}
          filterData={filterData}
          filterQueryMetric={filterQueryMetric}
        />
      )
    );

    return <div className={styles.monitorContainer}>{graphList}</div>;
  }, [
    activeTabKey,
    allMetrics,
    isRefresh,
    queryRange,
    filterLabel,
    groupLabels,
    type,
    useFor,
    filterData,
    filterQueryMetric,
  ]);

  return (
    <div style={{ marginTop: 16 }}>
      {allMetrics && tabList.length > 0 && (
        <Card
          tabList={tabList}
          bodyStyle={{ padding: 0 }}
          activeTabKey={activeTabKey}
          onTabChange={key => setActiveTabKey(key)}
        >
          {currentTabContent || (
            <div style={{ padding: 20, textAlign: 'center', color: '#999' }}>
              {formatMessage({
                id: 'OBShell.component.MonitorComp.NoData',
                defaultMessage: '暂无数据',
              })}
            </div>
          )}
        </Card>
      )}
    </div>
  );
}
