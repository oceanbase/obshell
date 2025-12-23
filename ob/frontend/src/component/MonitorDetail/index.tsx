import React, { useState } from 'react';

import MonitorComp from '@/component/MonitorComp';
import { DEFAULT_QUERY_RANGE } from '@/constant/monitor';
import DataFilter from './DataFilter';

interface MonitorDetailProps {
  filterData: Monitor.FilterDataType;
  filterLabel: Monitor.LabelType[];
  setFilterLabel: React.Dispatch<React.SetStateAction<Monitor.LabelType[]>>;
  groupLabels: Monitor.LabelKeys[];
  queryScope: Monitor.EventObjectType;
  useFor: Monitor.MonitorUseFor;
}

//Query is somewhat similar to sql statement, label is equivalent to filter condition，for example: where label1=xxx and label2 = xxx
export default function MonitorDetail({
  filterData,
  filterLabel,
  setFilterLabel,
  groupLabels,
  queryScope,
  useFor,
}: MonitorDetailProps) {
  const [isRefresh, setIsRefresh] = useState<boolean>(false);
  const [queryRange, setQueryRange] = useState<Monitor.QueryRangeType>(DEFAULT_QUERY_RANGE);
  const [serverOption, setServerOption] = useState<Monitor.OptionType[]>([]);
  const [isHostPerformanceTab, setIsHostPerformanceTab] = useState<boolean>(false); // 是否为主机性能 tab， 只有 OceanBase集群才有

  const handleTabChange = (isHostPerformance: boolean) => {
    // 只有 OceanBase集群才有主机性能tab，需要通知 DataFilter 更新 serverOption
    if (queryScope === 'OBCLUSTER') {
      setIsHostPerformanceTab(isHostPerformance);
    }
  };

  return (
    <>
      <DataFilter
        isRefresh={isRefresh}
        setIsRefresh={setIsRefresh}
        filterLabel={filterLabel}
        setFilterLabel={setFilterLabel}
        filterData={filterData}
        queryRange={queryRange}
        setQueryRange={setQueryRange}
        serverOption={serverOption}
        setServerOption={setServerOption}
        isHostPerformanceTab={isHostPerformanceTab}
      />

      <MonitorComp
        isRefresh={isRefresh}
        queryRange={queryRange}
        filterLabel={filterLabel}
        type="DETAIL"
        groupLabels={groupLabels}
        queryScope={queryScope}
        serverOption={serverOption}
        useFor={useFor}
        onTabChange={handleTabChange}
      />
    </>
  );
}
