import ContentWithReload from '@/component/ContentWithReload';
import { scenarioList, statusList } from '@/constant/inspection';
import * as InspectionController from '@/service/obshell/obcluster';
import { getTableData } from '@/util';
import { formatTime } from '@/util/datetime';
import { formatMessage } from '@/util/intl';
import {
  Button,
  Dropdown,
  message,
  Modal,
  Space,
  Table,
  TableColumnType,
  Tag,
  theme,
} from '@oceanbase/design';
import { DownOutlined } from '@oceanbase/icons';
import { PageContainer } from '@oceanbase/ui';
import { findBy, findByValue } from '@oceanbase/util';
import { Link } from '@umijs/max';
import { useInterval, useRequest } from 'ahooks';
import { isEmpty } from 'lodash';
import React, { useRef, useState } from 'react';

export default function HistoryList() {
  const { token } = theme.useToken();
  const [showLoading, setShowLoading] = useState(false);
  const isAutoRefresh = useRef(false);

  const { tableProps, loading, refresh } = getTableData({
    fn: InspectionController.getInspectionHistory,
    params: {},
  });

  // 自动刷新 - 不显示 loading
  useInterval(() => {
    isAutoRefresh.current = true;
    refresh();
  }, 10000);

  const { run: triggerInspection } = useRequest(InspectionController.triggerInspection, {
    manual: true,
    onSuccess: (res, params) => {
      if (res.successful) {
        const name = findBy(scenarioList, 'key', params[0]?.scenario)?.label;
        message.success(
          formatMessage(
            {
              id: 'OBShell.page.Inspection.SuccessfullyLaunchedName',
              defaultMessage: '发起{name}成功',
            },
            { name: name }
          )
        );
        isAutoRefresh.current = false;
        refresh();
      }
    },
  });

  // 监听 loading 状态变化
  React.useEffect(() => {
    if (isAutoRefresh.current) {
      // 自动刷新时不显示 loading
      setShowLoading(false);
      if (!loading) {
        isAutoRefresh.current = false;
      }
    } else {
      // 手动刷新时显示 loading
      setShowLoading(loading);
    }
  }, [loading]);

  const { runAsync: fetchReport } = useRequest(InspectionController.getInspectionReport, {
    manual: true,
  });

  const columns: TableColumnType<API.InspectionReportBriefInfo>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
    },
    {
      title: formatMessage({
        id: 'OBShell.page.Inspection.InspectionScene',
        defaultMessage: '巡检场景',
      }),
      dataIndex: 'scenario',
      filters: scenarioList.map(item => ({
        text: item.label,
        value: item.key,
      })),
      render: text => {
        const color = text === 'BASIC' ? 'success' : 'processing';
        const content =
          text === 'BASIC'
            ? formatMessage({
                id: 'OBShell.page.Inspection.BasicInspection',
                defaultMessage: '基础巡检',
              })
            : formatMessage({
                id: 'OBShell.page.Inspection.PerformanceInspection',
                defaultMessage: '性能巡检',
              });

        return <Tag color={color}>{content}</Tag>;
      },
    },
    {
      title: formatMessage({ id: 'OBShell.page.Inspection.StartTime', defaultMessage: '开始时间' }),
      dataIndex: 'start_time',
      sorter: true,
      render: text => {
        return formatTime(text);
      },
    },
    {
      title: formatMessage({ id: 'OBShell.page.Inspection.EndTime', defaultMessage: '结束时间' }),
      dataIndex: 'finish_time',
      sorter: true,
      render: text => {
        return formatTime(text);
      },
    },
    {
      title: formatMessage({
        id: 'OBShell.page.Inspection.TaskStatus',
        defaultMessage: '任务状态',
      }),
      dataIndex: 'status',
      width: 100,
      render: text => {
        const content = findByValue(statusList, text);
        if (isEmpty(content)) {
          return <Tag color="default">-</Tag>;
        }
        return <Tag color={content.color}>{content.label}</Tag>;
      },
    },
    {
      title: formatMessage({
        id: 'OBShell.page.Inspection.InspectionResults',
        defaultMessage: '巡检结果',
      }),
      width: 120,
      render: (_, record) => {
        if (record?.status !== 'SUCCEED') {
          return <div style={{ lineHeight: '66px' }}>-</div>;
        }
        return (
          <div>
            <div style={{ color: token.colorError }}>{`失败:${record?.failed_count || 0}`}</div>
            <div style={{ color: 'rgba(166,29,36,1)' }}>{`高风险:${
              record?.critical_count || 0
            }`}</div>
            <div style={{ color: 'orange' }}>{`中风险:${record?.warning_count || 0}`}</div>
          </div>
        );
      },
    },
    {
      title: formatMessage({ id: 'OBShell.page.Inspection.Operation', defaultMessage: '操作' }),
      dataIndex: 'opeation',
      width: 130,
      render: (_, record) => {
        if (record?.status === 'SUCCEED') {
          return (
            <Link to={`/inspection/report/${record?.id}`}>
              {formatMessage({
                id: 'OBShell.page.Inspection.ViewReport',
                defaultMessage: '查看报告',
              })}
            </Link>
          );
        } else if (record?.status === 'FAILED') {
          return (
            <a
              onClick={() => {
                fetchReport({ id: String(record.id) }).then(res => {
                  const errorMessage = res?.data?.error_message;
                  Modal.error({
                    title: formatMessage({
                      id: 'OBShell.page.Inspection.ErrorMessage',
                      defaultMessage: '错误信息',
                    }),
                    content: errorMessage,
                  });
                });
              }}
            >
              {formatMessage({
                id: 'OBShell.page.Inspection.ErrorMessage',
                defaultMessage: '错误信息',
              })}
            </a>
          );
        }
        return null;
      },
    },
  ];

  // 手动刷新
  const handleManualRefresh = () => {
    isAutoRefresh.current = false;
    refresh();
  };

  return (
    <PageContainer
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'OBShell.page.Inspection.InspectionService',
              defaultMessage: '巡检服务',
            })}
            spin={showLoading}
            onClick={handleManualRefresh}
          />
        ),
      }}
      ghost={true}
      extra={
        <Dropdown
          menu={{
            items: scenarioList,
            onClick: ({ key }) => {
              triggerInspection({ scenario: key });
            },
          }}
        >
          <Button type="primary">
            <Space>
              {formatMessage({
                id: 'OBShell.page.Inspection.InitiatePatrolInspection',
                defaultMessage: '发起巡检',
              })}

              <DownOutlined />
            </Space>
          </Button>
        </Dropdown>
      }
    >
      <Table {...tableProps} columns={columns} loading={showLoading} />
    </PageContainer>
  );
}
