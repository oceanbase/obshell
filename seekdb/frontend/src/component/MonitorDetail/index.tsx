import React, { useState } from 'react';

import MonitorComp from '@/component/MonitorComp';
import { DEFAULT_QUERY_RANGE } from '@/constant/monitor';
import DataFilter from './DataFilter';

interface MonitorDetailProps {
  filterLabel: Monitor.LabelType[];
  groupLabels: Monitor.LabelKeys[];
  queryScope: string;
  useFor: Monitor.MonitorUseFor;
}

//Query is somewhat similar to sql statement, label is equivalent to filter condition，for example: where label1=xxx and label2 = xxx
export default function MonitorDetail({
  filterLabel,
  groupLabels,
  queryScope,
  useFor,
}: MonitorDetailProps) {
  const [isRefresh, setIsRefresh] = useState<boolean>(false);
  const [queryRange, setQueryRange] = useState<Monitor.QueryRangeType>(DEFAULT_QUERY_RANGE);

  return (
    <>
      <DataFilter
        isRefresh={isRefresh}
        setIsRefresh={setIsRefresh}
        queryRange={queryRange}
        setQueryRange={setQueryRange}
      />

      <MonitorComp
        isRefresh={isRefresh}
        queryRange={queryRange}
        filterLabel={filterLabel}
        type="DETAIL"
        groupLabels={groupLabels}
        queryScope={queryScope}
        useFor={useFor}
      />
    </>
  );
}
