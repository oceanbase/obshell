import MyProgress from '@/component/MyProgress';
import TaskLabel from '@/component/TaskLabel';
import { OB_SERVER_STATUS_LIST } from '@/constant/oceanbase';
import { getAgentMainDags } from '@/service/obshell/task';
import { isEnglish } from '@/util';
import { getFailedTasksInfo } from '@/util/cluster';
import { getOperationComponent, Operation } from '@/util/component';
import { formatMessage } from '@/util/intl';
import { Badge, message, Space, Table, TableColumnType } from '@oceanbase/design';
import { byte2GB, findByValue } from '@oceanbase/util';
import { useModel } from '@umijs/max';
import { useRequest } from 'ahooks';
import { find } from 'lodash';
import React, { useState } from 'react';
import styles from './index.less';
import OBServerModal from './OBServerModal';

export interface OBServerListProps {
  zoneRecord: API.Zone;
}

const OBServerList: React.FC<OBServerListProps> = ({ zoneRecord }) => {
  const { zoneList, refreshTopologyInfo } = useModel('topologyInfo');
  const { servers } = zoneRecord;
  const [statusList, setStatusList] = useState<string[]>([]);

  const [obServerModalVisible, setObServerModalVisible] = useState(false);
  const [operation, setOperation] = useState<Operation | undefined>(undefined);
  const [server, setServer] = useState<API.Observer | undefined>(undefined);
  const isStandAloneCluster = zoneList?.length === 1 && zoneList?.[0]?.servers?.length === 1;
  const [failedTasksInfo, setFailedTasksInfo] = useState<{
    content: string;
    forcePassDag: { id: any[] };
  } | null>(null);
  const { runAsync: runAgentMainDags } = useRequest(getAgentMainDags, {
    manual: true,
  });

  const handleServerOperation = async (
    key: string,
    record: API.Observer & { zoneName?: string },
    operations: Operation[]
  ) => {
    const currentOperation = find(operations, item => item.value === key);
    const agentMainDags = await runAgentMainDags();
    const agentMainDagsData: API.DagDetailDTO[] = agentMainDags.data.contents;

    const failedData = getFailedTasksInfo(agentMainDagsData || []);

    if (agentMainDagsData?.some(item => item.state !== 'FAILED')) {
      message.warning(
        formatMessage({
          id: 'OBShell.ZoneListOrTopo.OBServerList.ThereAreCurrentlyObserverStart',
          defaultMessage: '当前有正在进行的OBServer启停任务，此时禁止继续启停',
        })
      );
      return;
    }
    setFailedTasksInfo(failedData);

    setObServerModalVisible(true);
    setServer(record);
    setOperation(currentOperation);
  };

  const expandColumns: TableColumnType<API.Observer>[] = [
    {
      title: 'IP',
      dataIndex: 'ip',
      ellipsis: true,
      width: 115,
    },

    {
      title: formatMessage({
        id: 'ocp-express.Component.ZoneListOrTopo.ZoneList.SqlPort',
        defaultMessage: 'SQL 端口',
      }),
      dataIndex: 'sql_port',
      width: 100,
      ellipsis: true,
    },

    {
      title: formatMessage({
        id: 'ocp-express.Component.ZoneListOrTopo.ZoneList.RpcPort',
        defaultMessage: 'RPC 端口',
      }),
      dataIndex: 'svr_port',
      width: 100,
      ellipsis: true,
    },

    {
      title: formatMessage({
        id: 'ocp-express.Component.ZoneListOrTopo.ZoneList.HardwareArchitecture',
        defaultMessage: '硬件架构',
      }),
      dataIndex: 'architecture',
      width: 100,
    },
    {
      title: formatMessage({
        id: 'OBShell.ZoneListOrTopo.OBServerList.DataDiskPath',
        defaultMessage: '数据盘路径',
      }),
      width: 110,
      ellipsis: true,
      dataIndex: 'data_dir',
    },
    {
      title: formatMessage({
        id: 'OBShell.ZoneListOrTopo.OBServerList.LogDiskPath',
        defaultMessage: '日志盘路径',
      }),
      width: 110,
      ellipsis: true,
      dataIndex: 'redo_dir',
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Component.WaterLevel.ResourceWaterLevel',
        defaultMessage: '资源水位',
      }),
      width: 290,
      dataIndex: 'stats',
      render: (text: API.ServerResourceStats) => {
        const {
          cpu_core_assigned: cpuCoreAssigned = 0,
          cpu_core_total: cpuCoreTotal = 0,
          cpu_core_assigned_percent: cpuCoreAssignedPercent = 0,
          memory_in_bytes_assigned: memoryInBytesAssigned = 0,
          memory_in_bytes_total: memoryInBytesTotal = 0,
          memory_assigned_percent: memoryAssignedPercent = 0,
          disk_used: diskUsed = 0,
          disk_total: diskTotal = 0,
          disk_used_percent: diskUsedPercent = 0,
        } = text || {};
        const prefixWidth = isEnglish() ? 50 : 30;
        return (
          <span className={styles.stats}>
            <MyProgress
              strokeLinecap="round"
              showInfo={false}
              prefix="CPU"
              prefixWidth={prefixWidth}
              affix={formatMessage(
                {
                  id: 'ocp-express.Component.OBServerList.CpucoreassignedCpucoretotalCore',
                  defaultMessage: '{cpuCoreAssigned}/{cpuCoreTotal} 核',
                },

                { cpuCoreAssigned, cpuCoreTotal }
              )}
              affixWidth={120}
              percent={cpuCoreAssignedPercent}
            />

            <MyProgress
              strokeLinecap="round"
              showInfo={false}
              prefix={formatMessage({
                id: 'ocp-express.Component.OBServerList.Memory',
                defaultMessage: '内存',
              })}
              prefixWidth={prefixWidth}
              affix={`${byte2GB(memoryInBytesAssigned).toFixed(1)}/${byte2GB(
                memoryInBytesTotal
              ).toFixed(1)} GB`}
              affixWidth={120}
              percent={memoryAssignedPercent}
            />

            <MyProgress
              strokeLinecap="round"
              showInfo={false}
              prefix={formatMessage({
                id: 'ocp-express.Component.OBServerList.Disk',
                defaultMessage: '磁盘',
              })}
              prefixWidth={prefixWidth}
              // 磁盘使用率一般不高，已使用量和总量可能差距较大，因此这里展示时不统一单位
              affix={`${diskUsed || 0}/${diskTotal || 0}`}
              affixWidth={120}
              percent={diskUsedPercent}
            />
          </span>
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.Component.OBServerList.State',
        defaultMessage: '状态',
      }),
      width: 120,
      dataIndex: 'status',
      render: text => {
        const statusItem = findByValue(OB_SERVER_STATUS_LIST, text);
        return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
      },
    },
    {
      title: formatMessage({
        id: 'OBShell.ZoneListOrTopo.OBServerList.Operation',
        defaultMessage: '操作',
      }),
      width: 136,
      render: (_, record) => {
        if (isStandAloneCluster) return null;
        const taskId = record.local_task_id || '';
        const operations = findByValue(OB_SERVER_STATUS_LIST, record.status).operations || [];

        return (
          <Space>
            {taskId && <TaskLabel taskId={taskId} />}
            {getOperationComponent({
              operations,
              handleOperation: handleServerOperation,
              record: { ...record, zoneName: zoneRecord.name },
              displayCount: 0,
            })}
          </Space>
        );
      },
    },
  ];

  return (
    <>
      <Table
        columns={expandColumns}
        dataSource={(servers || []).filter(
          item => statusList.length === 0 || statusList.includes(item.status as string)
        )}
        rowKey={(item: API.Observer) => item.ip as string}
        pagination={false}
        onChange={(_, filters) => {
          setStatusList((filters.status as string[]) || []);
        }}
        scroll={{ x: 1000 }}
      />

      <OBServerModal
        operation={operation || {}}
        visible={obServerModalVisible}
        server={server || {}}
        failedTasksInfo={failedTasksInfo}
        onCancel={() => {
          setObServerModalVisible(false);
          setOperation(undefined);
          setServer(undefined);
          setFailedTasksInfo(null);
        }}
        onSuccess={() => {
          setObServerModalVisible(false);
          setOperation(undefined);
          setServer(undefined);
          setFailedTasksInfo(null);
          refreshTopologyInfo();
        }}
      />
    </>
  );
};

export default OBServerList;
