import TaskLabel from '@/component/TaskLabel';
import { ZONE_STATUS_LIST } from '@/constant/oceanbase';
import { obStart, obStopZone } from '@/service/obshell/ob';
import { getAgentMainDags } from '@/service/obshell/task';
import { getFailedTasksInfo } from '@/util/cluster';
import { getOperationComponent } from '@/util/component';
import { formatMessage } from '@/util/intl';
import { taskSuccess } from '@/util/task';
import { Badge, Checkbox, message, Modal, Space, Table, TableColumnType } from '@oceanbase/design';
import { ContentWithIcon } from '@oceanbase/ui';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import React, { useState } from 'react';
import { useModel } from 'umi';
import styles from './index.less';
import OBServerList from './OBServerList';

export interface Operation {
  label: string;
  modalTitle: string;
  value: string;
  isDanger: boolean;
}

const ZoneList: React.FC = () => {
  const { zoneList, refreshTopologyInfo, setPendingOperation } = useModel('topologyInfo');

  const expandable = (zoneList?.filter(item => (item?.servers?.length || 0) > 0).length || 0) > 0;

  const [expandedRowKeys, setExpandedRowKeys] = useState<React.Key[]>(
    // 默认展开全部 Zone
    zoneList.map(item => item.name!) || []
  );

  const { runAsync: runAgentMainDags } = useRequest(getAgentMainDags, {
    manual: true,
  });

  const { run: startZone, loading: startZoneLoading } = useRequest(obStart, {
    manual: true,
    onSuccess: (res, params) => {
      if (res.successful && res.data?.id) {
        const body = params?.[0] as API.StartObParam | undefined;
        const zoneName = body?.scope?.target?.[0];
        if (zoneName) {
          setPendingOperation(zoneName, res.data.id);
        }
        taskSuccess({
          taskId: res.data.id,
          message: formatMessage({
            id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.TheTaskOfStartingZone',
            defaultMessage: '启动 Zone 的任务提交成功',
          }),
        });
        refreshTopologyInfo();
      }
    },
  });
  const { run: stopZone, loading: stopZoneLoading } = useRequest(obStopZone, {
    manual: true,
    onSuccess: (res, params) => {
      if (res.successful && res.data?.id) {
        const body = params?.[0] as API.ObZoneStopParam | undefined;
        const zoneName = body?.zone;
        if (zoneName) {
          setPendingOperation(zoneName, res.data.id);
        }
        taskSuccess({
          taskId: res.data.id,
          message: formatMessage({
            id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.TaskSubmissionOfZoneStopped',
            defaultMessage: '停止 Zone 的任务提交成功',
          }),
        });
        refreshTopologyInfo();
      }
    },
  });

  const isStandAloneCluster = zoneList?.length === 1 && zoneList?.[0]?.servers?.length === 1;

  // 停止 Zone 确认弹窗
  const showStopZoneModal = (
    record: API.Zone,
    forcePassDag: { id: any[] },
    failedTaskContent: string
  ) => {
    const freezeServerValueRef = { current: true };

    Modal.confirm({
      title: formatMessage(
        {
          id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.AreYouSureYouWant',
          defaultMessage: '确定要停止 {recordName} 吗？',
        },
        { recordName: record.name }
      ),
      content: (
        <Space direction="vertical">
          <div>
            {formatMessage({
              id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.IfYouStopAZone',
              defaultMessage: '停止 Zone 会停止当前 Zone 中的所有 OBserver，请谨慎操作。',
            })}
          </div>
          <div>
            <ContentWithIcon
              iconType="question"
              content={formatMessage({
                id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.PerformADumpOperationBefore',
                defaultMessage: '停止进程前执行转储操作',
              })}
              tooltip={{
                title: formatMessage({
                  id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.PerformingThisActionWillIncrease',
                  defaultMessage:
                    '执行本动作会延长停止进程的响应时间，但可以显著缩短 OBServer 恢复时间。',
                }),
              }}
            />

            <Checkbox
              style={{ marginLeft: '8px' }}
              defaultChecked={freezeServerValueRef.current}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                freezeServerValueRef.current = e.target.checked;
              }}
            />
          </div>
          {failedTaskContent && <div style={{ marginBottom: '8px' }}>{failedTaskContent}</div>}
        </Space>
      ),

      okText: formatMessage({
        id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.Stop',
        defaultMessage: '停止',
      }),
      okButtonProps: { danger: true, ghost: true, loading: stopZoneLoading },
      onOk: () => {
        const params = {
          zone: record.name!,
          forcePassDag,
          freeze_server: freezeServerValueRef.current,
        };
        stopZone(params);
      },
    });
  };

  // 启动 Zone 确认弹窗
  const showStartZoneModal = (
    record: API.Zone,
    forcePassDag: { id: any[] },
    failedTaskContent: string
  ) => {
    Modal.confirm({
      title: formatMessage(
        {
          id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.AreYouSureYouWant.1',
          defaultMessage: '确定要启动 {recordName} 吗？',
        },
        { recordName: record.name }
      ),
      content: (
        <Space direction="vertical">
          <div>
            {formatMessage({
              id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.StartingAZoneStartsAll',
              defaultMessage:
                '启动 Zone 会启动当前 Zone 中的所有 OBServer，并同步数据，可能需要一段时间。',
            })}
          </div>
          {failedTaskContent && <div style={{ marginBottom: '8px' }}>{failedTaskContent}</div>}
        </Space>
      ),

      okText: formatMessage({
        id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.Start',
        defaultMessage: '启动',
      }),
      okButtonProps: { loading: startZoneLoading },
      onOk: () => {
        startZone({
          scope: {
            type: 'ZONE' as const,
            target: [record.name!],
          },
          forcePassDag,
        });
      },
    });
  };

  // 处理 Zone 操作
  const handleZoneOperation = async (key: string, record: API.Zone) => {
    const res = await runAgentMainDags();
    const dagList: API.DagDetailDTO[] = res.data?.contents || [];

    if (dagList.some(item => item.state !== 'FAILED')) {
      message.warning(
        formatMessage({
          id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.ThereIsCurrentlyAnOngoing',
          defaultMessage: '当前有正在进行的Zone启停任务，此时禁止继续启停',
        })
      );
      return;
    }

    const { content, forcePassDag } = getFailedTasksInfo(dagList);

    if (key === 'stop') {
      showStopZoneModal(record, forcePassDag, content);
    } else if (key === 'start') {
      showStartZoneModal(record, forcePassDag, content);
    }
  };

  const columns: TableColumnType<API.Zone>[] = [
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
    {
      title: formatMessage({
        id: 'ocp-express.Component.ZoneList.NumberOfMachines',
        defaultMessage: 'OBServer 数量',
      }),
      key: 'serverCount',
      dataIndex: 'servers',
      render: (text?: API.Observer[]) => <span>{text?.length || 0}</span>,
    },

    {
      title: 'Root Server',
      key: 'rootServer',
      dataIndex: 'root_server',
      render: (text?: API.RootServer) => {
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
      width: 110,
      dataIndex: 'status',
      filters: ZONE_STATUS_LIST.map(item => ({
        value: item.value,
        text: item.label,
      })),
      onFilter: (value, record: API.Zone) => record.status === value,
      render: (text: string) => {
        const statusItem = findByValue(ZONE_STATUS_LIST, text);
        return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
      },
    },
    {
      title: formatMessage({
        id: 'OBShell.Overview.ZoneListOrTopo.ZoneList.Operation',
        defaultMessage: '操作',
      }),
      width: 136,
      render: (_, record) => {
        if (isStandAloneCluster) return null;
        const taskId = record.local_task_id || '';
        const operations = findByValue(ZONE_STATUS_LIST, record.status)?.operations || [];
        return (
          <Space onClick={e => e.stopPropagation()}>
            {taskId && <TaskLabel taskId={taskId} />}
            {getOperationComponent({
              mode: 'link',
              operations: operations,
              record,
              displayCount: 0,
              handleOperation: handleZoneOperation,
            })}
          </Space>
        );
      },
    },
  ];

  const getExpandedRowRender = (zoneRecord: API.Zone) => {
    return <OBServerList zoneRecord={zoneRecord} />;
  };

  return (
    <Table
      id="ocp-zone-table"
      columns={columns}
      dataSource={zoneList}
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
      scroll={{ x: 1000 }}
    />
  );
};

export default ZoneList;
