import { getObServersByObCluster } from '@/util/oceanbase';

export const getFilterData = (clusterData: API.ClusterInfo): Monitor.FilterDataType => {
  const serverList: Monitor.OptionType[] = getObServersByObCluster(clusterData).map(item => {
    return {
      label: `${item.ip!}:${item.svr_port!}`,
      value: `${item.ip!}:${item.svr_port!}`,
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
 * 对服务器列表按 IP 去重（用于主机性能监控）
 * @param servers 服务器列表
 * @returns 去重后的服务器列表，每个 IP 只保留一条记录
 */
export const deduplicateServersByIp = (servers: Monitor.OptionType[]): Monitor.OptionType[] => {
  const ipMap = new Map<string, Monitor.OptionType>();
  servers.forEach(server => {
    const [serverIp] = String(server.value).split(':');
    if (!ipMap.has(serverIp)) {
      // 创建一个新的 option，value 和 label 只包含 IP
      ipMap.set(serverIp, {
        ...server,
        value: serverIp,
        label: serverIp,
      });
    }
  });
  return Array.from(ipMap.values());
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
