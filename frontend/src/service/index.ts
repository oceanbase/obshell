import { formatMessage } from '@/util/intl';
import { flatten } from 'lodash';
import { queryMetrics } from './obshell/metric';

type QueryMetricsType = {
  groupLabels: string[];
  labels: Monitor.LabelType[];
  metrics: string[];
  queryRange: { endTimestamp: number; startTimestamp: number; step: number };
  type: Monitor.MonitorUseTarget;
  useFor: Monitor.MonitorUseFor;
  filterData?: any[];
  filterQueryMetric?: Monitor.MetricsLabels;
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
    // 优先级规则：
    // 1. 集群名称在前
    // 2. 租户名称在后
    // 3. 同类型按字母顺序排序

    const isClusterA =
      a.includes('cluster') ||
      a.includes(formatMessage({ id: 'OBShell.src.service.Cluster', defaultMessage: '集群' }));
    const isClusterB =
      b.includes('cluster') ||
      b.includes(formatMessage({ id: 'OBShell.src.service.Cluster', defaultMessage: '集群' }));
    const isTenantA =
      a.includes('tenant') ||
      a.includes(formatMessage({ id: 'OBShell.src.service.Tenant', defaultMessage: '租户' }));
    const isTenantB =
      b.includes('tenant') ||
      b.includes(formatMessage({ id: 'OBShell.src.service.Tenant', defaultMessage: '租户' }));

    // 集群相关的排在前面
    if (isClusterA && !isClusterB) return -1;
    if (!isClusterA && isClusterB) return 1;

    // 租户相关的排在后面
    if (isTenantA && !isTenantB) return 1;
    if (!isTenantA && isTenantB) return -1;

    // 同类型或都不是特殊类型时，按字母顺序排序
    return a.localeCompare(b, 'zh-CN', { numeric: true });
  });
};

export async function queryMetricsReq({
  useFor,
  type,
  filterData,
  filterQueryMetric,
  ...data
}: QueryMetricsType) {
  const r = await queryMetrics(data);

  let metricsData: MetricDataItem[] = r.data?.contents || [];
  if (r.successful) {
    if (filterQueryMetric && metricsData) {
      metricsData = metricsData.filter((item: MetricDataItem) => {
        const labels: Monitor.MetricsLabels = item.metric?.labels as Monitor.MetricsLabels;
        if (!labels || !labels.length) return false;
        return filterQueryMetric.some((queryMetric: { key: string; value: string }) =>
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
          const tenantLabel = metric.metric?.labels
            ?.sort((a: LabelType, b: LabelType) => a.key.localeCompare(b.key))
            .map((label: LabelType) => label.value)
            .join('，');
          const tenantName = tenantLabel || '';
          item.name = `${metric.metric?.name} (${tenantName})`;
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
    return flattenedData.sort((a: any, b: any) => {
      const indexA = nameToIndex.get(a.name || '') || 999999;
      const indexB = nameToIndex.get(b.name || '') || 999999;
      return indexA - indexB;
    });
  }
  return [];
}
