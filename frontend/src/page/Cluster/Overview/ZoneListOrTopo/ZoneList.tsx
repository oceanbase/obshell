/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { formatMessage } from '@/util/intl';
import React, { useState, useImperativeHandle } from 'react';
import { Badge } from '@oceanbase/design';
import { Table } from '@oceanbase/design';
import { byte2GB, findByValue } from '@oceanbase/util';
import { isEnglish } from '@/util';
import MyProgress from '@/component/MyProgress';
import { ZONE_STATUS_LIST, OB_SERVER_STATUS_LIST } from '@/constant/oceanbase';
import useStyles from './index.style';
import { getServerStatus } from '..';

export interface ZoneListRef {
  expandAll: () => void;
  setStatusList: (statusList: API.ObServerStatus[]) => void;
}

export interface ZoneListProps {
  clusterData: API.ClusterInfo;
}

const ZoneList = React.forwardRef<ZoneListRef, ZoneListProps>(({ clusterData }, ref) => {
  const { styles } = useStyles();
  const dataSource = clusterData?.zones || [];
  const expandable =
    clusterData?.zones?.filter(item => (item?.servers?.length || 0) > 0).length > 0;

  const [expandedRowKeys, setExpandedRowKeys] = useState<React.Key[]>(
    // 默认展开全部 Zone
    dataSource.map(item => item.name) || []
  );
  const [statusList, setStatusList] = useState<string[]>([]);

  // 向组件外部暴露 refresh 属性函数，可通过 ref 引用
  useImperativeHandle(ref, () => ({
    expandAll: () => {
      setExpandedRowKeys(dataSource.map(item => item.name as string));
    },
    setStatusList: (newStatusList: API.ObServerStatus[]) => {
      setStatusList(newStatusList);
    },
  }));

  const columns = [
    {
      title: formatMessage({
        id: 'ocp-express.Component.ZoneListOrTopo.ZoneList.ZoneName',
        defaultMessage: 'Zone 名',
      }),
      dataIndex: 'name',
    },

    {
      title: formatMessage({
        id: 'ocp-express.Component.ZoneList.Region',
        defaultMessage: '所属 Region',
      }),
      dataIndex: 'region_name',
    },

    // {
    //   title: formatMessage({
    //     id: 'ocp-express.Component.ZoneList.DataCenter',
    //     defaultMessage: '所在机房',
    //   }),

    //   dataIndex: 'idcName',
    //   render: (text: string) => <span>{text || '-'}</span>,
    // },

    {
      title: formatMessage({
        id: 'ocp-express.Component.ZoneList.NumberOfMachines',
        defaultMessage: 'OBServer 数量',
      }),
      key: 'serverCount',
      dataIndex: 'servers',
      render: (text?: API.Server[]) => <span>{text?.length || 0}</span>,
    },

    {
      title: 'Root Server',
      key: 'rootServer',
      dataIndex: 'root_server',
      render: (text?: API.RootServer, record: API.Zone) => {
        const roleLabel =
          text && text.role === 'LEADER'
            ? formatMessage({
                id: 'ocp-express.Component.ZoneList.Main',
                defaultMessage: '（主）',
              })
            : '';

        return <span>{text ? `${text?.ip}:${text?.svr_port}${roleLabel}` : '-'}</span>;
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.Component.ZoneList.State',
        defaultMessage: '状态',
      }),
      dataIndex: 'status',
      filters: ZONE_STATUS_LIST.map(item => ({
        value: item.value,
        text: item.label,
      })),
      onFilter: (value: API.ObZoneStatus, record: API.Zone) => record.status === value,
      render: (text: API.ObZoneStatus) => {
        const statusItem = findByValue(ZONE_STATUS_LIST, text);
        return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
      },
    },
  ];

  function getExpandedRowRender(record: API.Zone) {
    const { servers } = record;
    const expandColumns = [
      {
        title: 'IP',
        dataIndex: 'ip',
        render: (text: string, expandedRecord: API.Server) => {
          return text;
          // return (
          //   <a
          //     data-aspm-click="c304256.d308759"
          //     data-aspm-desc="集群拓扑列表-跳转 OBServer 详情"
          //     data-aspm-param={``}
          //     data-aspm-expo
          //     onClick={() => {
          //       history.push(`/overview/server/${text}/${expandedRecord.port}`);
          //     }}
          //   >
          //     {text}
          //   </a>
          // );
        },
      },

      {
        title: formatMessage({
          id: 'ocp-express.Component.ZoneListOrTopo.ZoneList.SqlPort',
          defaultMessage: 'SQL 端口',
        }),
        dataIndex: 'sql_port',
      },

      {
        title: formatMessage({
          id: 'ocp-express.Component.ZoneListOrTopo.ZoneList.RpcPort',
          defaultMessage: 'RPC 端口',
        }),
        dataIndex: 'svr_port',
      },

      {
        title: formatMessage({
          id: 'ocp-express.Component.ZoneListOrTopo.ZoneList.HardwareArchitecture',
          defaultMessage: '硬件架构',
        }),
        dataIndex: 'architecture',
      },

      {
        title: formatMessage({
          id: 'ocp-express.Detail.Component.WaterLevel.ResourceWaterLevel',
          defaultMessage: '资源水位',
        }),

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
                affixWidth={110}
                percent={cpuCoreAssignedPercent}
              />

              <MyProgress
                showInfo={false}
                prefix={formatMessage({
                  id: 'ocp-express.Component.OBServerList.Memory',
                  defaultMessage: '内存',
                })}
                prefixWidth={prefixWidth}
                affix={`${byte2GB(memoryInBytesAssigned).toFixed(1)}/${byte2GB(
                  memoryInBytesTotal
                ).toFixed(1)} GB`}
                affixWidth={110}
                percent={memoryAssignedPercent}
              />

              <MyProgress
                showInfo={false}
                prefix={formatMessage({
                  id: 'ocp-express.Component.OBServerList.Disk',
                  defaultMessage: '磁盘',
                })}
                prefixWidth={prefixWidth}
                // 磁盘使用率一般不高，已使用量和总量可能差距较大，因此这里展示时不统一单位
                affix={`${diskUsed}/${diskTotal}`}
                affixWidth={110}
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
        dataIndex: 'inner_status',
        // filters: OB_SERVER_STATUS_LIST.map(item => ({
        //   value: item.value,
        //   text: item.label,
        // })),
        // filteredValue: statusList,
        // 这里不用设置 onFilter，dataSource 已经根据 statusList 做了筛选
        render: (text: API.ObServerStatus, record: API.ObServer) => {
          const value = getServerStatus(record);
          const statusItem = findByValue(OB_SERVER_STATUS_LIST, value);
          return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
        },
      },
    ];

    return (
      <Table
        columns={expandColumns}
        dataSource={(servers || []).filter(
          item => statusList.length === 0 || statusList.includes(item.status as API.ObServerStatus)
        )}
        rowKey={item => item.id}
        pagination={false}
        onChange={(pagination, filters) => {
          setStatusList(filters.status || []);
        }}
      />
    );
  }

  return (
    <Table
      id="ocp-zone-table"
      columns={columns}
      dataSource={dataSource}
      rowKey={(record: API.Zone) => record.name as string}
      rowClassName={(record: API.Zone) =>
        expandable && !record.servers?.length ? 'table-row-hide-expand-icon' : ''
      }
      expandedRowKeys={expandedRowKeys}
      onExpandedRowsChange={newExpandedRowKeys => {
        setExpandedRowKeys(newExpandedRowKeys);
      }}
      expandable={{
        expandRowByClick: true,
        expandedRowRender: expandable ? getExpandedRowRender : undefined,
      }}
      pagination={false}
      className={styles.table}
    />
  );
});

export default ZoneList;
