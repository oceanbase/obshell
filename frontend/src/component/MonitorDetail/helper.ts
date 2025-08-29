import { getObServersByObCluster } from '@/util/oceanbase';

export const getFilterData = (clusterData: API.ClusterInfo): Monitor.FilterDataType => {
  const serverList: Monitor.OptionType[] = getObServersByObCluster(clusterData).map(item => {
    return {
      label: item.ip!,
      value: item.ip!,
      zone: item.zoneName!,
    };
  });
  const zoneList: Monitor.OptionType[] =
    clusterData?.zones?.map(item => {
      return {
        label: item.name!,
        value: item.name!,
      };
    }) || [];
  const res: Monitor.FilterDataType = {
    serverList,
    zoneList,
    date: '',
  };

  return res;
};

/**
 * step: the interval between each point unit:s  for example: half an hour 1800s, interval 30s, return sixty points
 * @param pointNumber points，default 15
 * @param startTimeStamp unit:s
 * @param endTimeStamp
 * @returns
 */
export const caculateStep = (
  startTimeStamp: number,
  endTimeStamp: number,
  pointNumber: number
): number => {
  return Math.ceil((endTimeStamp - startTimeStamp) / pointNumber);
};
