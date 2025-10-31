import { POINT_NUMBER } from '@/constant/monitor';
import { useUpdateEffect } from 'ahooks';
import { Card } from 'antd';
import React, { useEffect, useState } from 'react';
import { caculateStep } from './helper';
import AutoFresh, { FREQUENCY_TYPE } from '../MonitorSearch/AutoFresh';
import dayjs from 'dayjs';

interface DataFilterProps {
  isRefresh: boolean;
  queryRange: Monitor.QueryRangeType;
  setQueryRange: React.Dispatch<React.SetStateAction<Monitor.QueryRangeType>>;
  setIsRefresh: React.Dispatch<React.SetStateAction<boolean>>;
}

export default function DataFilter({
  isRefresh,
  queryRange: initialQueryRange, //defaultVAlue
  setIsRefresh,
  setQueryRange,
}: DataFilterProps) {
  const [frequency, setFrequency] = useState<FREQUENCY_TYPE>('off');
  const [queryRangeState, setQueryRangeState] = useState([
    dayjs(initialQueryRange.start_timestamp * 1000),
    dayjs(initialQueryRange.end_timestamp * 1000),
  ]);
  const [isExternalUpdate, setIsExternalUpdate] = useState(false);

  const handleRefresh = (mergeDefault?: boolean) => {
    // AutoFresh 自动刷新时，更新时间范围
    if (mergeDefault && isRefresh) {
      const currentRange = queryRangeState;
      if (currentRange?.length === 2) {
        const [startDate, endDate] = currentRange;
        const duration = endDate.valueOf() - startDate.valueOf();
        const newEndTime = dayjs();
        const newStartTime = dayjs(newEndTime.valueOf() - duration);

        setIsExternalUpdate(true); // 标记为外部更新
        setQueryRangeState([newStartTime, newEndTime]);
      }
    }
  };

  const handleMenuClick = (key: string) => {
    const newFrequency = key === 'off' ? 'off' : Number(key);
    setFrequency(newFrequency);
    // 同步父组件的 isRefresh 状态
    setIsRefresh(newFrequency !== 'off');
  };

  const handleRefreshClick = () => {
    // 手动刷新时更新时间范围为最近1小时
    const endTime = dayjs();
    const startTime = dayjs().subtract(1, 'hours');
    setIsExternalUpdate(false); // 用户手动刷新
    setQueryRangeState([startTime, endTime]);
  };

  const handleRangeChange = (range: dayjs.Dayjs[] | null) => {
    if (range) {
      setIsExternalUpdate(false); // 用户手动更改
      setQueryRangeState(range);
    }
  };

  // 同步外部 queryRange 变化到内部状态
  useUpdateEffect(() => {
    setIsExternalUpdate(true);
    setQueryRangeState([
      dayjs(initialQueryRange.start_timestamp * 1000),
      dayjs(initialQueryRange.end_timestamp * 1000),
    ]);
  }, [initialQueryRange.start_timestamp, initialQueryRange.end_timestamp]);

  useEffect(() => {
    if (queryRangeState?.length === 2 && !isExternalUpdate) {
      const [startDate, endDate] = queryRangeState;
      if (startDate && endDate) {
        const startTimestamp = Math.ceil(startDate.valueOf() / 1000);
        const endTimestamp = Math.ceil(endDate.valueOf() / 1000);
        setQueryRange({
          start_timestamp: startTimestamp,
          end_timestamp: endTimestamp,
          step: caculateStep(startTimestamp, endTimestamp, POINT_NUMBER),
        });
      }
    }
    if (isExternalUpdate) {
      setIsExternalUpdate(false);
    }
  }, [queryRangeState, setQueryRange, isExternalUpdate]);

  return (
    <Card
      styles={{
        body: {
          padding: '16px 24px',
        },
      }}
    >
      <div style={{ display: 'flex', flexDirection: 'row', gap: 24 }}>
        <AutoFresh
          onRefresh={handleRefresh}
          isRealtime={isRefresh}
          queryRange={queryRangeState}
          frequency={frequency}
          currentMoment={dayjs()}
          onMenuClick={handleMenuClick}
          onRefreshClick={handleRefreshClick}
          onRangeChange={handleRangeChange}
          withRanger={true}
        />
      </div>
    </Card>
  );
}
