import { obclusterTopologyInfo } from '@/service/obshell/obcluster';
import { useDeepCompareMemo } from '@oceanbase/ui';
import { useRafInterval, useRequest } from 'ahooks';
import { useState, useCallback, useEffect } from 'react';

const topologyInfoModel = () => {
  const {
    data: obclusterTopologyInfoData,
    run,
    refresh,
  } = useRequest(obclusterTopologyInfo, {
    manual: true,
  });

  const zoneList: API.Zone[] = obclusterTopologyInfoData?.data?.zones || [];

  // 存储待处理的操作：Map<zoneName, taskId>
  const [pendingOperations, setPendingOperations] = useState<Map<string, string>>(new Map());

  // 设置待处理操作
  const setPendingOperation = useCallback((zoneName: string, taskId: string) => {
    setPendingOperations(prev => {
      const newMap = new Map(prev);
      newMap.set(zoneName, taskId);
      return newMap;
    });
  }, []);

  // 清除待处理操作
  const clearPendingOperation = useCallback((zoneName: string) => {
    setPendingOperations(prev => {
      const newMap = new Map(prev);
      newMap.delete(zoneName);
      return newMap;
    });
  }, []);

  // 检查是否需要轮询
  const needPolling = useDeepCompareMemo(() => {
    // 检查 zone 是否有 local_task_id
    const hasZoneTask = zoneList.some(zone => zone.local_task_id);

    // 检查 zone 的 servers 是否有 local_task_id
    const hasServerTask = zoneList.some(zone => zone.servers?.some(server => server.local_task_id));

    // 检查是否有待处理的操作
    const hasPendingOperation = pendingOperations.size > 0;

    return hasZoneTask || hasServerTask || hasPendingOperation;
  }, [zoneList, pendingOperations]);

  // 当检测到 local_task_id 存在时，清除对应的 pendingOperation
  useEffect(() => {
    if (zoneList.length === 0 || pendingOperations.size === 0) {
      return;
    }

    setPendingOperations(prev => {
      const newMap = new Map(prev);
      let changed = false;

      // 检查每个 zone 是否已有 local_task_id
      zoneList.forEach(zone => {
        if (zone.name && newMap.has(zone.name) && zone.local_task_id) {
          newMap.delete(zone.name);
          changed = true;
        }

        // 检查每个 server 是否已有 local_task_id
        zone.servers?.forEach(server => {
          if (server.ip && server.obshell_port) {
            const serverKey = `${server.ip}:${server.obshell_port}`;
            if (newMap.has(serverKey) && server.local_task_id) {
              newMap.delete(serverKey);
              changed = true;
            }
          }
        });
      });

      return changed ? newMap : prev;
    });
  }, [zoneList, pendingOperations.size]);

  // 使用 useRafInterval 实现轮询
  useRafInterval(
    () => {
      refresh();
    },
    needPolling ? 3000 : undefined
  );

  return {
    zoneList,
    getTopologyInfo: run,
    refreshTopologyInfo: refresh,
    setPendingOperation,
    clearPendingOperation,
  };
};
export default topologyInfoModel;
