import Access from '@/component/Access';
import ContentWithInfo from '@/component/ContentWithInfo';
import { PAGINATION_OPTION_10 } from '@/constant';
import { REPORT_STATUS_LIST } from '@/constant/report';
import { formatMessage } from '@/util/intl';
import * as ObAshController from '@/service/obshell/tenant';
import * as ObClusterConfigPropertiesController from '@/service/obshell/obcluster';
import { getOperationComponent } from '@/util/component';
import { formatTime } from '@/util/datetime';
import { download, insertLocaleScript } from '@/util/export';
import { history, useSelector } from 'umi';
import React, { useState } from 'react';
import { some } from 'lodash';
import { Alert, Badge, Button, Card, Input, Space, Table } from '@oceanbase/design';
import { byte2MB, directTo, findByValue, sortByMoment } from '@oceanbase/util';
import { useInterval, useRequest } from 'ahooks';
import GenerateReportModal from './GenerateReportModal';
export interface IndexProps {
  clusterId: number;
  tenantId: number;
}

const Index: React.FC<IndexProps> = ({ clusterId, tenantId }) => {
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const { canReadTenant, canReadTask } = useSelector(
    (state: DefaultRootState) => state.auth.authData
  );
  const readTenant = canReadTenant(tenantId);
  const readTask = canReadTask();

  const [keyword, setKeyword] = useState('');
  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<API.AshReportDescription>();
  // 获取报告列表
  const { data, loading, refresh } = useRequest(ObAshController.listAshReports, {
    ready: !!clusterId && !!tenantId,
    defaultParams: [
      {
        id: clusterId,
        tenantId,
      },
    ],
  });

  // 前端分页，直接取全量数据，并做筛选处理
  const dataSource = (data?.data?.contents || []).filter(
    // 支持根据报告名和创建人进行搜索
    item => !keyword || item.name?.includes(keyword) || item.creator?.includes(keyword)
  );

  const polling = some(dataSource, item => item.status === 'CREATING');
  useInterval(
    () => {
      refresh();
    },
    polling ? 1000 : null
  );

  // 下载报告
  const { runAsync: getAshReport, loading: getAshReportLoading } = useRequest(
    ObAshController.getAshReport,
    {
      manual: true,
    }
  );

  const { data: collectionSwitchStatusData } = useRequest(
    ObClusterConfigPropertiesController.getCollectionSwitchStatus,
    {
      ready: !!clusterId,
      defaultParams: [
        {
          id: clusterId,
          parameterType: 'SQL_SESSION',
        },
      ],
    }
  );

  const collectionSwitchStatus = collectionSwitchStatusData?.data;

  const handleOperation = (key: string, record: API.AshReportDescription) => {
    if (key === 'download') {
      setCurrentRecord(record);
      getAshReport({
        id: clusterId,
        tenantId,
        reportId: record.id,
      }).then(html => {
        // 下载前插入 locale 信息
        download(insertLocaleScript(html as string), `${record?.name}.html`);
        setCurrentRecord(undefined);
      });
    } else if (key === 'view') {
      directTo(`/cluster/${clusterId}/tenant/${tenantId}/session/report/${record.id}`);
    } else if (key === 'viewTask') {
      directTo(`/task/${record.taskInstanceId}?backUrl=/task`);
    }
  };

  const columns = [
    {
      title: formatMessage({ id: 'ocp-v2.Session.Report.ReportName', defaultMessage: '报告名' }),
      dataIndex: 'name',
      width: '20%',
      render: (text: string) => <span style={{ wordBreak: 'break-all' }}>{text}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Session.Report.StartAnalysisTime',
        defaultMessage: '起始分析时间',
      }),
      dataIndex: 'beginReportTime',
      width: 150,
      sorter: (a: API.AshReportDescription, b: API.AshReportDescription) =>
        sortByMoment(a, b, 'beginReportTime'),
      render: (text: string) => <span>{formatTime(text)}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Session.Report.EndAnalysisTime',
        defaultMessage: '结束分析时间',
      }),
      dataIndex: 'endReportTime',
      width: 150,
      sorter: (a: API.AshReportDescription, b: API.AshReportDescription) =>
        sortByMoment(a, b, 'endReportTime'),
      render: (text: string) => <span>{formatTime(text)}</span>,
    },

    {
      title: formatMessage({ id: 'ocp-v2.Session.Report.Founder', defaultMessage: '创建人' }),
      dataIndex: 'creator',
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Session.Report.GenerationStatus',
        defaultMessage: '生成状态',
      }),
      dataIndex: 'status',
      filters: REPORT_STATUS_LIST.map(item => ({
        value: item.value,
        text: item.label,
      })),

      onFilter: (value: API.ReportStatus, record: API.AshReportDescription) =>
        record.status === value,
      render: (text: API.ReportStatus) => {
        const statusItem = findByValue(REPORT_STATUS_LIST, text);
        return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
      },
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Session.Report.GenerationTime',
        defaultMessage: '生成时间',
      }),
      dataIndex: 'generateTime',
      width: 150,
      sorter: (a: API.AshReportDescription, b: API.AshReportDescription) =>
        sortByMoment(a, b, 'generateTime'),
      render: (text: string) => <span>{formatTime(text)}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-v2.Session.Report.ReportSizeMb',
        defaultMessage: '报告大小（MiB）',
      }),
      dataIndex: 'reportSize',
      align: 'right',
      render: (text: number, record: API.AshReportDescription) =>
        record.status === 'SUCCESSFUL' ? byte2MB(text) : '-',
    },

    {
      title: formatMessage({ id: 'ocp-v2.Session.Report.Operation', defaultMessage: '操作' }),
      dataIndex: 'operation',
      render: (text: string, record: API.AshReportDescription) => {
        const operations = (findByValue(REPORT_STATUS_LIST, record.status).operations || []).map(
          item => {
            if (item.value === 'download') {
              return {
                ...item,
                buttonProps: {
                  type: 'link',
                  style: {
                    padding: 0,
                  },

                  ...(record.id === currentRecord?.id
                    ? {
                        // 正在下载: 增加 loading 效果
                        loading: getAshReportLoading,
                      }
                    : {
                        // 其他下载: 置灰处理，避免同时下载多个报告
                        disabled: getAshReportLoading,
                      }),
                },
              };
            }
            return {
              ...item,
              buttonProps: {
                type: 'link',
                style: {
                  padding: 0,
                },
              },
            };
          }
        );
        return getOperationComponent({
          operations,
          handleOperation,
          record,
          mode: 'button',
          displayCount: 2,
          accessible: {
            readAccessible: readTenant,
            readTaskAccessible: readTask,
          },
        });
      },
    },
  ];

  return (
    <>
      {collectionSwitchStatus === 'INACTIVE' && (
        <Alert
          type="info"
          showIcon={true}
          style={{ marginBottom: 16 }}
          message={formatMessage({
            id: 'ocp-v2.Session.Report.ActiveSessionHistoryCollectionIsNotEnabledPlease',
            defaultMessage: '活跃会话历史采集功能未开启，请前往所属集群开启',
          })}
          action={
            <Button
              size="small"
              type="link"
              onClick={() => {
                history.push(`/cluster/${clusterId}/overview?openConfig=${1}`);
              }}
            >
              {formatMessage({ id: 'ocp-v2.Session.Report.GoNow', defaultMessage: '立即前往' })}
            </Button>
          }
        />
      )}

      <Card
        className="card-without-padding"
        bordered={false}
        title={
          <Space size={16}>
            {formatMessage({ id: 'ocp-v2.Session.Report.ReportList', defaultMessage: '报告列表' })}

            <ContentWithInfo
              content={formatMessage({
                id: 'ocp-v2.Session.Report.TheSystemOnlyRetainsData',
                defaultMessage: '系统只保留近 90 天的数据',
              })}
              size={4}
            />
          </Space>
        }
        extra={
          <Space>
            <Input.Search
              allowClear={true}
              onSearch={value => {
                setKeyword(value);
              }}
              placeholder={formatMessage({
                id: 'ocp-v2.Session.Report.EnterTheReportNameAnd',
                defaultMessage: '请输入报告名、创建人',
              })}
              className="search-input"
            />

            <Access accessible={readTenant}>
              <Button
                type="primary"
                onClick={() => {
                  setVisible(true);
                }}
              >
                {formatMessage({
                  id: 'ocp-v2.Session.Report.GenerateReports',
                  defaultMessage: '生成报告',
                })}
              </Button>
            </Access>
          </Space>
        }
      >
        <Table
          loading={loading && !polling}
          dataSource={dataSource}
          columns={columns}
          rowKey={record => record.id}
          pagination={PAGINATION_OPTION_10}
        />

        <GenerateReportModal
          visible={visible}
          tenantData={tenantData}
          onCancel={() => {
            setVisible(false);
          }}
          onSuccess={() => {
            setVisible(false);
            refresh();
          }}
        />
      </Card>
    </>
  );
};

export default Index;
