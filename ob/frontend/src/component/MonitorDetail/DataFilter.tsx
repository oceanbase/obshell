import { formatMessage } from '@/util/intl';
import { POINT_NUMBER } from '@/constant/monitor';
import { useUpdateEffect } from 'ahooks';
import { Card, Select } from 'antd';
import React, { useEffect, useState } from 'react';
import { caculateStep, deduplicateServersByIp } from './helper';
import AutoFresh, { FREQUENCY_TYPE } from '../MonitorSearch/AutoFresh';
import dayjs from 'dayjs';
import { useSelector } from 'umi';
import { getZonesFromTenant } from '@/util/tenant';

interface DataFilterProps {
  isRefresh: boolean;
  filterData: Monitor.FilterDataType;
  filterLabel: Monitor.LabelType[];
  queryRange: Monitor.QueryRangeType;
  setQueryRange: React.Dispatch<React.SetStateAction<Monitor.QueryRangeType>>;
  setIsRefresh: React.Dispatch<React.SetStateAction<boolean>>;
  setFilterLabel: React.Dispatch<React.SetStateAction<Monitor.LabelType[]>>;
  serverOption: Monitor.OptionType[];
  setServerOption: React.Dispatch<React.SetStateAction<Monitor.OptionType[]>>;
  isHostPerformanceTab?: boolean; // 是否为主机性能 tab
}

export default function DataFilter({
  isRefresh,
  filterData,
  filterLabel, //发送请求的label
  queryRange: initialQueryRange, //defaultVAlue
  setIsRefresh,
  setFilterLabel,
  setQueryRange,
  serverOption,
  setServerOption,
  isHostPerformanceTab = false,
}: DataFilterProps) {
  const [zoneOption, setZoneOption] = useState<Monitor.OptionType[]>([]);
  const [tenantOption, setTenantOption] = useState<Monitor.OptionType[]>([]);
  const [selectZone, setSelectZone] = useState<string>();
  const [selectServer, setSelectServer] = useState<string>();
  const [selectTenant, setSelectTenant] = useState<string>();
  const [frequency, setFrequency] = useState<FREQUENCY_TYPE>('off');
  const [queryRangeState, setQueryRangeState] = useState([
    dayjs(initialQueryRange.start_timestamp * 1000),
    dayjs(initialQueryRange.end_timestamp * 1000),
  ]);
  const [isExternalUpdate, setIsExternalUpdate] = useState(false);

  const { isSingleMachine, tenantListData } = useSelector(
    (state: DefaultRootState) => state.cluster
  );

  const clearLabel = (current: Monitor.LabelType[], key: Monitor.Label): Monitor.LabelType[] => {
    const newLabel = [...current];
    const idx = newLabel.findIndex((item: Monitor.LabelType) => item.key === key);
    if (idx !== -1) {
      newLabel.splice(idx, 1);
    }
    return newLabel;
  };

  //替换 或者 添加
  const updateLabel = (
    current: Monitor.LabelType[],
    key: Monitor.Label,
    value: string
  ): Monitor.LabelType[] => {
    const newLabel = [...current];
    const idx = newLabel.findIndex((item: Monitor.LabelType) => item.key === key);
    if (idx !== -1) {
      newLabel[idx] = { key, value };
    } else {
      newLabel.push({
        key,
        value,
      });
    }
    return newLabel;
  };

  const handleLabel = (val: string | undefined): Monitor.LabelType[] => {
    const isClear: boolean = !val;
    let currentLabel = [...filterLabel];
    if (isClear) {
      //clear obzone&svr_ip&svr_port
      currentLabel = clearLabel(
        clearLabel(clearLabel(filterLabel, 'obzone'), 'svr_ip'),
        'svr_port'
      );
    } else {
      //clear the server after updating the zone
      currentLabel = clearLabel(
        clearLabel(updateLabel(filterLabel, 'obzone', val!), 'svr_ip'),
        'svr_port'
      );
    }
    return currentLabel;
  };

  const tenantSelectChange = (val: string | undefined) => {
    setSelectTenant(val);
    setSelectZone(undefined);
    setSelectServer(undefined);
    setFilterLabel(updateLabel(filterLabel, 'tenant_name', val!));
  };

  const zoneSelectChange = (val: string | undefined) => {
    setSelectZone(val);
    setSelectServer(undefined);
    setFilterLabel(handleLabel(val));
    //clear
    if (filterData.serverList) {
      let serversList: Monitor.OptionType[] = [];

      if (typeof val === 'undefined') {
        if (selectTenant) {
          const tenant = tenantListData.find(
            (tenant: API.TenantInfo) => tenant.tenant_name === selectTenant
          );
          const zonesName = getZonesFromTenant(tenant).map(zone => zone.name) || [];
          serversList = filterData.serverList.filter(server =>
            zonesName.includes(server.zone as string)
          );
        } else {
          serversList = filterData.serverList;
        }
      } else {
        serversList = filterData.serverList.filter((server: Monitor.OptionType) => {
          return server.zone === val;
        });
      }

      // 如果是主机性能 tab，对服务器列表按 IP 去重
      setServerOption(isHostPerformanceTab ? deduplicateServersByIp(serversList) : serversList);
    }
  };

  const serverSelectChange = (val: string | undefined) => {
    const isClear: boolean = !val;
    let label: Monitor.LabelType[] = [...filterLabel];
    if (isClear) {
      // 清除时同时清除 svr_ip 和 svr_port
      label = clearLabel(clearLabel(label, 'svr_ip'), 'svr_port');
    } else {
      if (isHostPerformanceTab) {
        // 主机性能 tab：val 就是 IP，不需要拆分
        label = updateLabel(label, 'svr_ip', val!);
        // 清除可能存在的 svr_port
        label = clearLabel(label, 'svr_port');
      } else {
        // 将组合值拆分为 svr_ip 和 svr_port
        const [svrIp, svrPort] = val!.split(':');
        if (svrIp && svrPort) {
          label = updateLabel(label, 'svr_ip', svrIp);
          label = updateLabel(label, 'svr_port', svrPort);
        }
      }
    }
    setFilterLabel(label);
    setSelectServer(val);
  };

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

  // 初始化和更新 filterData 中的选项数据
  useEffect(() => {
    if (filterData?.zoneList?.length) {
      setZoneOption(filterData.zoneList);
    }
    if (filterData?.serverList?.length) {
      // 如果是主机性能 tab，对 serverList 按 IP 去重
      setServerOption(
        isHostPerformanceTab ? deduplicateServersByIp(filterData.serverList) : filterData.serverList
      );
    }
    if (filterData?.tenantList?.length) {
      setTenantOption(filterData.tenantList);
    }
  }, [filterData, isHostPerformanceTab]);

  // 同步 filterLabel 中的 tenant_name 到 selectTenant 状态
  useEffect(() => {
    const tenantLabel = filterLabel.find(label => label.key === 'tenant_name')?.value;
    if (tenantLabel) {
      setSelectTenant(tenantLabel);
    }
  }, [filterLabel]);

  useEffect(() => {
    if (selectTenant && filterData?.zoneList?.length) {
      const tenant = tenantListData.find(
        (tenant: API.TenantInfo) => tenant.tenant_name === selectTenant
      );
      const zonesName = getZonesFromTenant(tenant).map(zone => zone.name) || [];
      const zonesList = filterData.zoneList.filter(zone =>
        zonesName.includes(zone.value as string)
      );
      setZoneOption(zonesList);
      if (filterData.serverList?.length) {
        const serversList = filterData.serverList.filter(server =>
          zonesName.includes(server.zone as string)
        );

        // 如果是主机性能 tab，对服务器列表按 IP 去重
        setServerOption(isHostPerformanceTab ? deduplicateServersByIp(serversList) : serversList);
      }
    }
  }, [selectTenant, isHostPerformanceTab]);

  // 当 tab 切换时（isHostPerformanceTab 变化），清空选中的 OBServer
  useEffect(() => {
    if (selectServer) {
      setSelectServer(undefined);
      // 清除 filterLabel 中的 svr_ip 和 svr_port
      let label = clearLabel(clearLabel([...filterLabel], 'svr_ip'), 'svr_port');
      setFilterLabel(label);
    }
  }, [isHostPerformanceTab]);

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

        {filterData.tenantList && (
          <div>
            <span style={{ marginRight: 8 }}>
              {formatMessage({
                id: 'OBShell.component.MonitorDetail.DataFilter.Tenant',
                defaultMessage: '租户:',
              })}
            </span>
            <Select
              variant="borderless"
              value={selectTenant}
              onChange={tenantSelectChange}
              showSearch
              placeholder={formatMessage({
                id: 'OBShell.component.MonitorDetail.DataFilter.All',
                defaultMessage: '全部',
              })}
              options={tenantOption}
              dropdownMatchSelectWidth={false}
            />
          </div>
        )}
        {filterData.zoneList && !isSingleMachine && (
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <span style={{ marginRight: 8 }}>Zone:</span>
            <Select
              variant="borderless"
              value={selectZone}
              onChange={zoneSelectChange}
              allowClear
              showSearch
              placeholder={formatMessage({
                id: 'OBShell.component.MonitorDetail.DataFilter.All',
                defaultMessage: '全部',
              })}
              options={zoneOption}
              dropdownMatchSelectWidth={false}
            />
          </div>
        )}

        {filterData.serverList && !isSingleMachine && (
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <span style={{ marginRight: 8 }}>OBServer:</span>
            <Select
              variant="borderless"
              value={selectServer}
              onChange={serverSelectChange}
              allowClear
              showSearch
              placeholder={formatMessage({
                id: 'OBShell.component.MonitorDetail.DataFilter.All',
                defaultMessage: '全部',
              })}
              options={serverOption}
              dropdownMatchSelectWidth={false}
            />
          </div>
        )}
      </div>
    </Card>
  );
}
