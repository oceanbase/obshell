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
}: MonitorCompProps) {
  const [activeTabKey, setActiveTabKey] = useState<string>('0');
  const [activeDimension, setActiveDimension] = useState<string>('');
  const { isSingleMachine } = useSelector((state: DefaultRootState) => state.cluster);

  const [selectedUnit, setSelectedUnit] = useState<Monitor.OptionType>();

  const { data: listAllMetricsRes } = useRequest(listAllMetrics, {
    defaultParams: [{ scope: queryScope }],
  });
  const allMetrics: API.MetricClass[] = listAllMetricsRes?.data?.contents || [];

  // 根据 filterLabel 过滤 serverOption
  const filteredServerOption = useMemo(() => {
    if (!filterLabel?.length) {
      return serverOption;
    }

    let filteredServers = [...serverOption];

    // 查找 obzone 筛选条件
    const zoneFilter = filterLabel.find(label => label.key === 'obzone');
    if (zoneFilter) {
      filteredServers = filteredServers.filter(server => server.zone === zoneFilter.value);
    }

    // 查找 svr_ip 筛选条件
    const serverFilter = filterLabel.find(label => label.key === 'svr_ip');
    if (serverFilter) {
      filteredServers = filteredServers.filter(server => server.value === serverFilter.value);
    }

    return filteredServers;
  }, [serverOption, filterLabel]);

  const currentContainer = allMetrics?.[parseInt(activeTabKey)];
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
      const serverFilter = filterLabel.find(label => label.key === 'svr_ip');

      if (serverFilter && filteredServerOption.length === 1) {
        // 如果有 svr_ip 筛选且过滤后只有一个服务器，自动选中
        const targetServer = filteredServerOption[0];
        if (!selectedUnit || selectedUnit.value !== targetServer.value) {
          setSelectedUnit(targetServer);
        }
      }
    }
  }, [activeDimension, filterLabel, filteredServerOption, selectedUnit]);

  // 根据 activeDimension 动态生成 groupLabels
  const dynamicGroupLabels: Monitor.LabelKeys[] = useMemo(() => {
    if (activeDimension === 'unit') {
      return ['svr_ip', 'tenant_name'];
    }

    return activeDimension ? [activeDimension as Monitor.LabelKeys] : groupLabels;
  }, [activeDimension, groupLabels]);

  // 根据选中的 unit 生成筛选条件
  const unitFilterQueryMetric: Monitor.MetricsLabels | undefined = useMemo(() => {
    // 对于 unit 维度，需要特殊处理
    if (activeDimension === 'unit') {
      // 对于 unit 维度，只需要传递 key 为 svr_ip 的 querymetric
      if (selectedUnit?.value) {
        return [{ key: 'svr_ip', value: String(selectedUnit.value) }];
      }

      // 如果没有选中 unit，检查是否有来自其他地方的 svr_ip 筛选条件
      const baseFilters = [
        ...(filterQueryMetric || []),
        ...(filterLabel || []).map(label => ({ key: label.key, value: label.value })),
      ];

      // 只保留 svr_ip 相关的筛选条件
      const svrIpFilters = baseFilters.filter(filter => filter.key === 'svr_ip');
      return svrIpFilters.length > 0 ? svrIpFilters : undefined;
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
