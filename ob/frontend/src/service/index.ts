import { findBy, formatNumber } from '@oceanbase/util';
import { flatten } from 'lodash';
import { queryMetrics } from './obshell/metric';

type QueryMetricsType = {
  groupLabels: string[];
  labels: Monitor.LabelType[];
  metrics: API.MetricMeta[];
  queryRange: { endTimestamp: number; startTimestamp: number; step: number };
  type: Monitor.MonitorUseTarget;
  useFor: Monitor.MonitorUseFor;
  filterData?: any[];
  filterQueryMetric?: Monitor.MetricsLabels;
  activeDimension?: string;
};

type LabelType = {
  key: string;
  value: string;
};

// 定义指标数据项的类型
interface MetricDataItem {
  metric?: {
    labels?: LabelType[];
    name?: string;
  };
  values?: Array<{
    timestamp: number;
    value: number;
    date?: number;
    name?: string;
  }>;
}

// 定义过滤数据类型
interface FilterDataItem {
  clusterName?: string;
  tenantName?: string;
}

const filterMetricsData = (
  type: 'tenant' | 'cluster',
  metricsData: MetricDataItem[],
  filterData: FilterDataItem[]
) => {
  return metricsData.filter((item: MetricDataItem) => {
    const targetName =
      item?.metric?.labels?.find(
        (label: LabelType) => label.key === (type === 'tenant' ? 'tenant_name' : 'ob_cluster_name')
      )?.value || '';
    if (type === 'cluster') {
      return !!filterData.find((cluster: FilterDataItem) => cluster.clusterName === targetName);
    } else {
      return !!filterData.find((tenant: FilterDataItem) => tenant.tenantName === targetName);
    }
  });
};

const setMetricNameFromLabels = (labels: Monitor.MetricsLabels) => {
  const tenantName = labels.find(label => label.key === 'tenant_name')?.value;
  const clustetName = labels
    .filter(label => label.key === 'ob_cluster_name')
    .map(label => label.value)
    .join(',');

  return `${tenantName}(${clustetName})`;
};

// 定义图例排序规则
const sortLegendNames = (names: string[]): string[] => {
  return names.sort((a, b) => {
    // 同类型或都不是特殊类型时，按字母顺序排序
    return a.localeCompare(b, 'zh-CN', { numeric: true });
  });
};

export async function queryMetricsReq({
  useFor,
  type,
  metrics,
  filterData,
  filterQueryMetric,
  activeDimension,
  ...data
}: QueryMetricsType) {
  const metricsKeys =
    type === 'DETAIL' ? metrics.map((metric: API.MetricMeta) => metric.key) : [metrics[0].key];
  const r = await queryMetrics({ metrics: metricsKeys, ...data });

  let metricsData: MetricDataItem[] = r.data?.contents || [];
  if (r.successful) {
    if (filterQueryMetric && metricsData) {
      metricsData = metricsData.filter((item: MetricDataItem) => {
        const labels: Monitor.MetricsLabels = item.metric?.labels as Monitor.MetricsLabels;
        if (!labels || !labels.length) return false;
        // 必须满足 filterQueryMetric 中的所有条件
        return filterQueryMetric.every((queryMetric: { key: string; value: string }) =>
          labels.some(
            (label: LabelType) => label.key === queryMetric.key && label.value === queryMetric.value
          )
        );
      });
    }
    if (filterData) {
      metricsData = filterMetricsData(useFor, metricsData, filterData as FilterDataItem[]);
    }
    if (!metricsData || !metricsData.length) return [];

    // 首先对metricsData按照metric.name排序，确保基础顺序稳定
    metricsData.sort((a: MetricDataItem, b: MetricDataItem) => {
      const nameA = a.metric?.name || '';
      const nameB = b.metric?.name || '';
      return nameA.localeCompare(nameB);
    });

    metricsData.forEach((metric: MetricDataItem) => {
      metric.values?.forEach((item: any) => {
        item.date = item.timestamp * 1000;
        item.value = formatNumber(item.value, 2);
        if (type === 'OVERVIEW') {
          if (useFor === 'tenant') {
            const metricLabels = metric.metric?.labels;
            if (metricLabels && metricLabels.length > 1) {
              item.name = setMetricNameFromLabels(metricLabels as Monitor.MetricsLabels);
            } else {
              item.name = metricLabels?.[0]?.value || '';
            }
          } else {
            item.name =
              metric.metric?.labels?.find(
                (label: LabelType) => label.key === 'ob_cluster_name' || label.key === 'cluster'
              )?.value || '';
          }
        } else {
          // 处理标签：如果同时存在 svr_ip 和 svr_port，拼接为 ip:port
          const labels = metric.metric?.labels || [];
          const svrIpLabel = labels.find((label: LabelType) => label.key === 'svr_ip');
          const svrPortLabel = labels.find((label: LabelType) => label.key === 'svr_port');

          // 根据监控类型过滤标签
          let processedLabels = labels
            .filter((label: LabelType) => {
              // 过滤掉 svr_ip 和 svr_port
              if (label.key === 'svr_ip' || label.key === 'svr_port') return false;
              // 如果是集群监控，过滤掉 ob_cluster_name
              if (useFor === 'cluster' && label.key === 'ob_cluster_name') return false;
              // 如果是租户监控，过滤掉 tenant_name
              if (useFor === 'tenant' && label.key === 'tenant_name') return false;
              return true;
            })
            .sort((a: LabelType, b: LabelType) => a.key.localeCompare(b.key));

          // 如果不是 unit 维度，且同时存在 svr_ip 和 svr_port，添加拼接后的服务器地址
          if (activeDimension !== 'unit') {
            if (svrIpLabel && svrPortLabel) {
              processedLabels = [
                ...processedLabels,
                { key: 'server', value: `${svrIpLabel.value}:${svrPortLabel.value}` },
              ];
            } else {
              // 如果只有其中一个，保留原标签
              if (svrIpLabel) processedLabels.push(svrIpLabel);
              if (svrPortLabel) processedLabels.push(svrPortLabel);
            }
          }

          const tenantLabel = processedLabels.map((label: LabelType) => label.value).join('，');
          const tenantName = tenantLabel || '';
          const metricName = findBy(metrics, 'key', metric.metric?.name)?.name || '';
          item.name = `${metricName}${tenantName ? ` (${tenantName})` : ''}`;
        }
      });
    });

    // 展平数据并按图例名称排序，确保图例顺序固定
    const flattenedData = flatten(metricsData.map((metric: MetricDataItem) => metric.values || []));

    // 收集所有唯一的图例名称
    const uniqueNames = Array.from(new Set(flattenedData.map((item: any) => item.name || '')));
    // 对图例名称进行排序
    const sortedNames = sortLegendNames(uniqueNames);

    // 创建名称到排序索引的映射
    const nameToIndex = new Map(sortedNames.map((name, index) => [name, index]));
    // 根据排序后的索引对数据进行排序
    const sortedData = flattenedData.sort((a: any, b: any) => {
      const indexA = nameToIndex.get(a.name || '') ?? 999999;
      const indexB = nameToIndex.get(b.name || '') ?? 999999;
      return indexA - indexB;
    });
    return sortedData;
  }
  return [];
}
