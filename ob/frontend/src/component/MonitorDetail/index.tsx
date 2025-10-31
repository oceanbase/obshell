import dayjs from 'dayjs';
import React, { useEffect, useState } from 'react';

import MonitorComp from '@/component/MonitorComp';
import { DEFAULT_QUERY_RANGE, REFRESH_FREQUENCY } from '@/constant/monitor';
import { DATE_TIME_FORMAT } from '@/constant/datetime';
import DataFilter from './DataFilter';

interface MonitorDetailProps {
  filterData: Monitor.FilterDataType;
  setFilterData: React.Dispatch<React.SetStateAction<Monitor.FilterDataType>>;
  filterLabel: Monitor.LabelType[];
  setFilterLabel: React.Dispatch<React.SetStateAction<Monitor.LabelType[]>>;
  groupLabels: Monitor.LabelKeys[];
  queryScope: Monitor.EventObjectType;
  useFor: Monitor.MonitorUseFor;
}

const getDate = () => {
  return dayjs.unix(Math.ceil(new Date().valueOf() / 1000)).format(DATE_TIME_FORMAT);
};

//Query is somewhat similar to sql statement, label is equivalent to filter condition，for example: where label1=xxx and label2 = xxx
export default function MonitorDetail({
  filterData,
  setFilterData,
  filterLabel,
  setFilterLabel,
  groupLabels,
  queryScope,
  useFor,
}: MonitorDetailProps) {
  const [isRefresh, setIsRefresh] = useState<boolean>(false);
  const [realTime, setRealTime] = useState<string>(getDate());
  const [queryRange, setQueryRange] = useState<Monitor.QueryRangeType>(DEFAULT_QUERY_RANGE);
  const [serverOption, setServerOption] = useState<Monitor.OptionType[]>([]);

  //Real time update time
  useEffect(() => {
    const updateTimer = setInterval(() => {
      setRealTime(getDate());
    }, REFRESH_FREQUENCY * 1000);

    return () => {
      clearInterval(updateTimer);
    };
  }, []);

  return (
    <>
      <DataFilter
        realTime={realTime}
        isRefresh={isRefresh}
        setIsRefresh={setIsRefresh}
        filterLabel={filterLabel}
        setFilterLabel={setFilterLabel}
        filterData={filterData}
        setFilterData={setFilterData}
        queryRange={queryRange}
        setQueryRange={setQueryRange}
        serverOption={serverOption}
        setServerOption={setServerOption}
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
      />
    </>
  );
}
