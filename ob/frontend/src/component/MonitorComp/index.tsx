import { formatMessage } from '@/util/intl';
import { useRequest } from 'ahooks';
import { Card } from '@oceanbase/design';

import { listAllMetrics } from '@/service/obshell/metric';
import React, { useEffect, useMemo, useState } from 'react';
import styles from './index.less';
import { Dimension } from '@/constant/monitor';
import MySegmented from '../MySegmented';
import { isEmpty } from 'lodash';
import MonitorUnitSelect from '../MonitorUnitSelect';
import { useSelector } from 'umi';
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
  queryScope: Monitor.EventObjectType;
  type: Monitor.MonitorUseTarget;
  groupLabels: Monitor.LabelKeys[];
  useFor?: Monitor.MonitorUseFor;
  filterData?: any[];
  filterQueryMetric?: Monitor.MetricsLabels;
  serverOption: Monitor.OptionType[];
  onTabChange?: (isHostPerformanceTab: boolean) => void;
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
  serverOption,
  onTabChange,
}: MonitorCompProps) {
  const [activeTabKey, setActiveTabKey] = useState<string>('0');
  const [activeDimension, setActiveDimension] = useState<string>('');
  const { isSingleMachine } = useSelector((state: DefaultRootState) => state.cluster);

  const [selectedUnit, setSelectedUnit] = useState<Monitor.OptionType>();

  const { data: listAllMetricsRes } = useRequest(listAllMetrics, {
    defaultParams: [{ scope: queryScope }],
  });
  const allMetrics: API.MetricClass[] = listAllMetricsRes?.data?.contents || [];

  // 判断当前是否为"主机性能" tab
  const currentContainer = allMetrics?.[parseInt(activeTabKey)];
  const isHostPerformanceTab = useMemo(() => {
    return (
      currentContainer?.name ===
      formatMessage({
        id: 'OBShell.component.MonitorComp.HostPerformance',
        defaultMessage: '主机性能',
      })
    );
  }, [currentContainer?.name]);

  // 根据 filterLabel 过滤 serverOption
  const filteredServerOption = useMemo(() => {
    let filteredServers = [...serverOption];

    if (filterLabel?.length) {
      // 查找 obzone 筛选条件
      const zoneFilter = filterLabel.find(label => label.key === 'obzone');
      if (zoneFilter) {
        filteredServers = filteredServers.filter(server => server.zone === zoneFilter.value);
      }

      // 查找 svr_ip 和 svr_port 筛选条件
      const serverIpFilter = filterLabel.find(label => label.key === 'svr_ip');
      const serverPortFilter = filterLabel.find(label => label.key === 'svr_port');

      if (serverIpFilter || serverPortFilter) {
        filteredServers = filteredServers.filter(server => {
          // server.value 是组合值，格式为 "ip:port"
          const [serverIp, serverPort] = String(server.value).split(':');
          const ipMatch = !serverIpFilter || serverIp === serverIpFilter.value;
          const portMatch = !serverPortFilter || serverPort === serverPortFilter.value;
          return ipMatch && portMatch;
        });
      }
    }

    return filteredServers;
  }, [serverOption, filterLabel]);

  const dimension = useMemo(() => {
    return Dimension[currentContainer?.name as keyof typeof Dimension]?.filter(item =>
      isSingleMachine ? item.singleMachineFlag : true
    );
  }, [currentContainer?.name, isSingleMachine]);

  // 生成tab列表
  const tabList = useMemo(() => {
    return (
      allMetrics?.map((container: API.MetricClass, index: number) => ({
        key: index.toString(),
        label: container?.name,
      })) || []
    );
  }, [allMetrics]);

  useEffect(() => {
    if (isEmpty(dimension)) return;
    setActiveDimension(dimension[0]?.value || '');
  }, [dimension]);

  // 当 isHostPerformanceTab 变化时，通知父组件
  useEffect(() => {
    onTabChange?.(isHostPerformanceTab);
  }, [isHostPerformanceTab]);

  // 当筛选条件变化时，检查当前选中的 unit 是否仍在过滤后的列表中
  useEffect(() => {
    if (selectedUnit && filteredServerOption.length > 0) {
      const isSelectedUnitValid = filteredServerOption.some(
        server => server.value === selectedUnit.value
      );

      if (!isSelectedUnitValid) {
        // 如果当前选中的 unit 不在过滤后的列表中，重置选择
        setSelectedUnit(undefined);
      }
    }
  }, [filteredServerOption, selectedUnit]);

  // 当 DataFilter 中选择了特定 OBServer 时，自动在 MonitorUnitSelect 中选中
  useEffect(() => {
    if (activeDimension === 'unit' && filterLabel?.length) {
      const serverIpFilter = filterLabel.find(label => label.key === 'svr_ip');
      const serverPortFilter = filterLabel.find(label => label.key === 'svr_port');

      if ((serverIpFilter || serverPortFilter) && filteredServerOption.length === 1) {
        // 如果有 svr_ip 或 svr_port 筛选且过滤后只有一个服务器，自动选中
        const targetServer = filteredServerOption[0];
        if (!selectedUnit || selectedUnit.value !== targetServer.value) {
          setSelectedUnit(targetServer);
        }
      }
    }
  }, [activeDimension, filterLabel, filteredServerOption, selectedUnit]);

  // // 根据 activeDimension 动态生成 groupLabels
  const dynamicGroupLabels: Monitor.LabelKeys[] = useMemo(() => {
    let labels: Monitor.LabelKeys[] = [];

    if (activeDimension === 'unit') {
      labels = ['svr_ip', 'svr_port', 'tenant_name'];
    } else if (activeDimension === 'svr_ip') {
      // 主机性能 tab：只包含 svr_ip，不包含 svr_port
      if (isHostPerformanceTab) {
        labels = ['svr_ip'];
      } else {
        labels = ['svr_ip', 'svr_port'];
      }
    } else {
      labels = activeDimension ? [activeDimension as Monitor.LabelKeys] : groupLabels;
    }

    // 如果 filterLabel 中包含 svr_ip 或 svr_port，确保它们在 groupLabels 中
    // 但主机性能 tab 不包含 svr_port
    const hasSvrIp = filterLabel?.some(label => label.key === 'svr_ip');
    const hasSvrPort = filterLabel?.some(label => label.key === 'svr_port');

    if (hasSvrIp && !labels.includes('svr_ip')) {
      labels.push('svr_ip');
    }
    if (hasSvrPort && !labels.includes('svr_port') && !isHostPerformanceTab) {
      labels.push('svr_port');
    }

    return labels;
  }, [activeDimension, groupLabels, filterLabel, isHostPerformanceTab]);

  // 根据选中的 unit 生成筛选条件
  const unitFilterQueryMetric: Monitor.MetricsLabels | undefined = useMemo(() => {
    // 对于 unit 维度，需要特殊处理
    if (activeDimension === 'unit') {
      // 对于 unit 维度，需要传递 svr_ip 和 svr_port 的 querymetric
      if (selectedUnit?.value) {
        const filters: Monitor.MetricsLabels = [];
        // selectedUnit.value 是组合值，格式为 "ip:port"
        const [svrIp, svrPort] = String(selectedUnit.value).split(':');
        if (svrIp) {
          filters.push({ key: 'svr_ip', value: svrIp });
        }
        if (svrPort) {
          filters.push({ key: 'svr_port', value: svrPort });
        }
        return filters.length > 0 ? filters : undefined;
      }

      // 如果没有选中 unit，检查是否有来自其他地方的 svr_ip 或 svr_port 筛选条件
      const baseFilters = [
        ...(filterQueryMetric || []),
        ...(filterLabel || []).map(label => ({ key: label.key, value: label.value })),
      ];

      // 只保留 svr_ip 和 svr_port 相关的筛选条件
      const svrFilters = baseFilters.filter(
        filter => filter.key === 'svr_ip' || filter.key === 'svr_port'
      );
      return svrFilters.length > 0 ? svrFilters : undefined;
    }

    // 对于其他维度（租户维度、OBServer维度等），只使用原始的 filterQueryMetric
    // 不合并 filterLabel，避免冲突
    return filterQueryMetric;
  }, [activeDimension, selectedUnit, filterQueryMetric, filterLabel]);

  // 渲染图表列表
  const renderGraphList = (metricGroups: API.MetricGroup[]) => {
    return metricGroups?.map((graphContainer: API.MetricGroup, graphIdx: number) => (
      <GraphItem
        key={graphIdx}
        graphContainer={graphContainer}
        graphIdx={graphIdx}
        activeTabKey={activeTabKey}
        isRefresh={isRefresh}
        queryRange={queryRange}
        filterLabel={filterLabel}
        dynamicGroupLabels={dynamicGroupLabels}
        type={type}
        useFor={useFor}
        filterData={filterData}
        filterQueryMetric={unitFilterQueryMetric}
      />
    ));
  };

  // 生成当前激活tab的内容
  const currentTabContent = useMemo(() => {
    const currentContainer = allMetrics?.[parseInt(activeTabKey)];
    if (!currentContainer?.metric_groups) return null;

    const graphList = renderGraphList(currentContainer.metric_groups);

    return (
      <>
        <div style={{ margin: !activeDimension.length ? 0 : 24 }}>
          <MySegmented
            value={activeDimension}
            options={dimension?.map(item => ({
              label: item.label,
              value: item.value,
            }))}
            onChange={value => {
              setActiveDimension(value as string);
            }}
          />
        </div>
        <div
          className={`${styles.monitorContainer} ${
            activeDimension === 'unit' ? styles.hasUnitSelect : ''
          }`}
        >
          {activeDimension === 'unit' && (
            <MonitorUnitSelect onSelect={setSelectedUnit} units={filteredServerOption} />
          )}
          {activeDimension === 'unit' ? (
            <div className={styles.monitorRightContent}>{graphList}</div>
          ) : (
            graphList
          )}
        </div>
      </>
    );
  }, [
    activeTabKey,
    allMetrics,
    isRefresh,
    queryRange,
    filterLabel,
    dynamicGroupLabels,
    type,
    useFor,
    filterData,
    unitFilterQueryMetric,
    activeDimension,
    dimension,
    filteredServerOption,
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
