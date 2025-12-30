import { formatMessage } from '@/util/intl';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import * as ObTenantSessionController from '@/service/obshell/tenant';
import { formatSql, getTableData, isEnglish } from '@/util';
import { getColumnSearchProps, getOperationComponent } from '@/util/component';
import { useSelector } from 'umi';
import React, { useEffect, useState } from 'react';
import { flatten, uniq } from 'lodash';
import {
  Button,
  Card,
  Checkbox,
  Input,
  message,
  Modal,
  Space,
  Table,
  TableColumnType,
  Tooltip,
  Typography,
} from '@oceanbase/design';
import { CheckOutlined, CopyOutlined, SyncOutlined } from '@oceanbase/icons';
import { FooterToolbar, Highlight } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
const { Text } = Typography;

export const getUniqKey = (record: API.TenantSession) => `${record.id}-${record.svr_ip}`;

export interface ListProps {
  tenantName: string;
  query?: {
    user: string;
    db: string;
    host: string;
  };

  showReload?: boolean;
  mode?: 'page' | 'component';
}

const List: React.FC<ListProps> = ({ tenantName, query, showReload = true, mode = 'page' }) => {
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);

  const [activeOnly, setActiveOnly] = useState(false);
  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<API.TenantSession | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [sessionId, setSessionId] = useState('');
  const { user, db, host } = query || {};

  const { tableProps, run, refresh, loading, params } = getTableData({
    fn: ObTenantSessionController.getTenantSessions,
    params: {
      name: tenantName,
      active_only: activeOnly,
      id: sessionId,
    },

    deps: [activeOnly, sessionId],
  });

  const { sorter = {} } = params[0] || ({} as any);

  const { dataSource = [] } = tableProps;
  // 当前页的会话 ID 列表
  const currentPageSessionIds = dataSource.map(item => getUniqKey(item));

  // user、db 仅在组件挂载时会变化一次，因此下拉函数最多只运行一次
  // 用于会话统计页面，根据不同维度对象进行初始化请求
  // 由于 ahooks useRequest 并不支持 defaultFilters，因此这里通过手动 run 的方式来模拟
  useEffect(() => {
    if (user || db || host) {
      // 需要异步发起，以保证当前请求的值是最新的，否则会被默认的请求结果覆盖
      setTimeout(() => {
        // 这里的 run 方法并不是 listTenantSessions，而是 getTableData 的 service 方法，传参均为 antd Table 相关的参数
        run({
          pageSize: 10,
          current: 1,
          filters: {
            user: user ? [user] : [],
            db: db ? [db] : [],
            host: host ? [host] : [],
          },
        });
      }, 0);
    }
  }, [user || db || host]);

  // 批量关闭查询
  const { runAsync: closeTenantQuery } = useRequest(
    ObTenantSessionController.killTenantSessionQuery,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'OBShell.Detail.Session.List.QueryClosedSuccessfully',
              defaultMessage: '查询关闭成功',
            })
          );

          setSelectedRowKeys([]);
          refresh();
        }
      },
    }
  );

  // 批量关闭会话
  const { runAsync: closeTenantSession } = useRequest(
    ObTenantSessionController.killTenantSessions,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'OBShell.Detail.Session.List.SessionClosedSuccessfully',
              defaultMessage: '会话关闭成功',
            })
          );

          setSelectedRowKeys([]);
          refresh();
        }
      },
    }
  );

  const handleOperation = (key: string, record: API.TenantSession) => {
    if (key === 'closeSession') {
      Modal.confirm({
        title: formatMessage({
          id: 'OBShell.Detail.Session.List.AreYouSureYouWant',
          defaultMessage: '确定要关闭会话吗？',
        }),
        onOk: () => {
          return closeTenantSession(
            {
              name: tenantName,
            },

            {
              session_ids: [record.id as number],
            }
          );
        },
      });
    } else if (key === 'closeQuery') {
      Modal.confirm({
        title: formatMessage({
          id: 'OBShell.Detail.Session.List.AreYouSureYouWant.2',
          defaultMessage: '确定要关闭查询吗？',
        }),
        onOk: () => {
          return closeTenantQuery(
            {
              name: tenantName,
            },
            {
              session_ids: [record.id as number],
            }
          );
        },
      });
    }
  };

  const obServerList: string[] = flatten(
    tenantData.pools?.map((item: any) => item.observer_list.split(',')) || []
  );

  const columns: TableColumnType<API.TenantSession>[] = [
    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.List.SessionId',
        defaultMessage: '会话 ID',
      }),

      dataIndex: 'id',
      fixed: 'left',
      width: 120,
      ellipsis: true,
    },

    {
      title: 'SQL',
      dataIndex: 'info',
      fixed: 'left',
      ellipsis: true,
      width: 250,
      render: (text: string, record: API.TenantSession) =>
        text ? (
          <Tooltip placement="topLeft" title={text}>
            <a
              onClick={() => {
                setVisible(true);
                setCurrentRecord(record);
              }}
            >
              {text}
            </a>
          </Tooltip>
        ) : (
          '-'
        ),
    },

    {
      title: 'SQL ID',
      dataIndex: 'sql_id',
      ellipsis: true,
      render: text => text || '-',
    },

    {
      title: formatMessage({ id: 'OBShell.Detail.Session.List.User', defaultMessage: '用户' }),
      dataIndex: 'user',
      ...(user ? { defaultFilteredValue: [user] } : {}),
      ...getColumnSearchProps({
        frontEndSearch: false,
      }),
    },

    {
      title: formatMessage({ id: 'OBShell.Detail.Session.List.Source', defaultMessage: '来源' }),
      dataIndex: 'host',
      ...(host ? { defaultFilteredValue: [host] } : {}),
      ...getColumnSearchProps({
        frontEndSearch: false,
      }),
    },

    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.List.DatabaseName',
        defaultMessage: '数据库名',
      }),

      dataIndex: 'db',
      ...(db ? { defaultFilteredValue: [db] } : {}),
      ...getColumnSearchProps({
        frontEndSearch: false,
      }),
    },

    {
      title: formatMessage({ id: 'OBShell.Detail.Session.List.Command', defaultMessage: '命令' }),
      dataIndex: 'command',
    },

    {
      title: formatMessage({
        id: 'OBShell.Detail.Session.List.ExecutionTimeS',
        defaultMessage: '执行时间（s）',
      }),

      dataIndex: 'time',
      sorter: true,
      sortOrder: sorter.field === 'time' && sorter.order,
      render: (text: number) => <span>{text?.toFixed(2) || '-'}</span>,
    },

    {
      title: formatMessage({ id: 'OBShell.Detail.Session.List.Status', defaultMessage: '状态' }),
      dataIndex: 'state',
    },
    {
      title: 'OBServer',
      dataIndex: 'observer_list',
      filters: obServerList.map(item => ({
        text: item,
        value: item,
      })),
      render: (_, record) => `${record.svr_ip}:${record.svr_port}`,
    },
    {
      title: 'OBProxy',
      dataIndex: 'proxy_ip',
      render: (text: string) => <span>{text || '-'}</span>,
    },
    {
      title: (
        <ContentWithQuestion
          content={formatMessage({
            id: 'OBShell.Detail.Session.List.Operation',
            defaultMessage: '操作',
          })}
          tooltip={{
            title: formatMessage({
              id: 'OBShell.Detail.Session.List.ForASessionThatIs',
              defaultMessage:
                '对正在执行事务、且事务耗时超过 500ms 或者单个参与者日志生成量超过 0.5MiB 的会话，提供了查看事务的入口，便于排查',
            }),

            placement: 'topRight',
          }}
        />
      ),

      dataIndex: 'operation',
      fixed: 'right',
      width: isEnglish() ? 180 : 150,
      render: (_, record: API.TenantSession) => {
        const operations = [
          {
            value: 'closeSession',
            label: formatMessage({
              id: 'OBShell.Detail.Session.List.CloseSession',
              defaultMessage: '关闭会话',
            }),
          },
          {
            value: 'closeQuery',
            label: formatMessage({
              id: 'OBShell.Detail.Session.List.CloseCurrentQuery',
              defaultMessage: '关闭当前查询',
            }),
          },
        ];

        return getOperationComponent({
          operations,
          handleOperation,
          record,
          displayCount: 1,
        });
      },
    },
  ];

  const currentSessionId = currentRecord && currentRecord.id;

  return (
    <div>
      <Card
        bordered={false}
        className={`card-without-padding ${selectedRowKeys.length > 0 ? 'mb-14' : 'mb-0'}`}
        extra={
          <Space>
            <Checkbox
              checked={activeOnly}
              onChange={(e: any) => {
                setActiveOnly(e.target.checked);
              }}
            >
              {formatMessage({
                id: 'OBShell.Detail.Session.List.ViewOnlyActiveSessions',
                defaultMessage: '仅查看活跃会话',
              })}
            </Checkbox>
            <Input.Search
              allowClear
              placeholder={formatMessage({
                id: 'OBShell.Detail.Session.List.PleaseEnterASessionId',
                defaultMessage: '请输入会话 ID',
              })}
              onSearch={v => {
                setSessionId(v);
              }}
            />

            {showReload && (
              <Tooltip
                title={formatMessage({
                  id: 'OBShell.Detail.Session.List.Refresh',
                  defaultMessage: '刷新',
                })}
              >
                <div className="px-2 py-1">
                  <SyncOutlined
                    spin={loading}
                    onClick={() => {
                      refresh();
                    }}
                    className="cursor-pointer"
                  />
                </div>
              </Tooltip>
            )}
          </Space>
        }
      >
        <Table<API.TenantSession>
          columns={columns}
          rowKey={(record: API.TenantSession) => getUniqKey(record)}
          rowSelection={{
            selectedRowKeys,
            onSelect: (record, selected) => {
              setSelectedRowKeys(
                selected
                  ? [...selectedRowKeys, getUniqKey(record)]
                  : selectedRowKeys.filter(item => item !== getUniqKey(record))
              );
            },
            onSelectAll: selected => {
              if (selected) {
                setSelectedRowKeys(uniq([...selectedRowKeys, ...currentPageSessionIds]));
              } else {
                setSelectedRowKeys(
                  selectedRowKeys.filter(
                    selectedRowKey => !currentPageSessionIds.includes(selectedRowKey)
                  )
                );
              }
            },
          }}
          scroll={{ x: '1950px' }}
          {...tableProps}
          hiddenCancelBtn
          toolAlertRender={false}
        />
      </Card>
      {selectedRowKeys && selectedRowKeys.length > 0 && (
        <FooterToolbar
          className={mode === 'component' ? 'w-[1232px] px-6 py-3 leading-normal shadow-none' : ''}
          extra={
            <div className="flex items-center">
              <Checkbox
                // 与 antd Table 的全选逻辑保持一致，即只全选当前页的数据
                checked={
                  selectedRowKeys.filter(item => currentPageSessionIds.includes(item)).length ===
                  dataSource.length
                }
                indeterminate={
                  selectedRowKeys.filter(item => currentPageSessionIds.includes(item)).length <
                  dataSource.length
                }
                onChange={(e: any) => {
                  const { checked } = e.target;
                  if (checked) {
                    setSelectedRowKeys(uniq([...selectedRowKeys, ...currentPageSessionIds]));
                  } else {
                    setSelectedRowKeys(
                      selectedRowKeys.filter(
                        selectedRowKey => !currentPageSessionIds.includes(selectedRowKey)
                      )
                    );
                  }
                }}
              />

              <span className="ml-2">
                {formatMessage({
                  id: 'OBShell.Detail.Session.List.Selected',
                  defaultMessage: '已选',
                })}
                {selectedRowKeys && selectedRowKeys.length}
                {formatMessage({ id: 'OBShell.Detail.Session.List.Item', defaultMessage: '项' })}
              </span>
              <a
                onClick={() => {
                  setSelectedRowKeys([]);
                }}
                className="ml-3"
              >
                {formatMessage({
                  id: 'OBShell.Detail.Session.List.Cancel',
                  defaultMessage: '取消',
                })}
              </a>
            </div>
          }
        >
          <Space>
            <Button
              onClick={() => {
                Modal.confirm({
                  title: formatMessage({
                    id: 'OBShell.Detail.Session.List.AreYouSureYouWant.2',
                    defaultMessage: '确定要关闭查询吗？',
                  }),
                  onOk: () => {
                    return closeTenantQuery(
                      {
                        name: tenantName,
                      },

                      {
                        session_ids: selectedRowKeys.map(uniqKey => Number(uniqKey.split('-')[0])),
                      }
                    );
                  },
                });
              }}
            >
              {formatMessage({
                id: 'OBShell.Detail.Session.List.CloseCurrentQuery',
                defaultMessage: '关闭当前查询',
              })}
            </Button>
            <Button
              onClick={() => {
                Modal.confirm({
                  title: formatMessage({
                    id: 'OBShell.Detail.Session.List.AreYouSureYouWant',
                    defaultMessage: '确定要关闭会话吗？',
                  }),
                  onOk: () => {
                    return closeTenantSession(
                      {
                        name: tenantName,
                      },

                      {
                        session_ids: selectedRowKeys.map(uniqKey => Number(uniqKey.split('-')[0])),
                      }
                    );
                  },
                });
              }}
            >
              {formatMessage({
                id: 'OBShell.Detail.Session.List.CloseSession',
                defaultMessage: '关闭会话',
              })}
            </Button>
          </Space>
        </FooterToolbar>
      )}
      <Modal
        width={760}
        title={formatMessage(
          {
            id: 'OBShell.Detail.Session.List.SqlDetailsCurrentsessionid',
            defaultMessage: 'SQL 详情 - {currentSessionId}',
          },
          { currentSessionId: currentSessionId }
        )}
        visible={visible}
        footer={
          <Space>
            <Text
              copyable={{
                text: formatSql(currentRecord && currentRecord.info),
                icon: [
                  <Button key="copy" icon={<CopyOutlined />}>
                    <span>
                      {formatMessage({
                        id: 'OBShell.Detail.Session.List.Copy',
                        defaultMessage: '复制',
                      })}
                    </span>
                  </Button>,
                  // CheckOutlined 不能设置为 Button.icon，否则复制后 Tooltip 不会主动消失
                  <Button key="copy-success">
                    <CheckOutlined />
                  </Button>,
                ],
              }}
            />

            <Button
              type="primary"
              onClick={() => {
                setVisible(false);
                setCurrentRecord(null);
              }}
            >
              {formatMessage({ id: 'OBShell.Detail.Session.List.Close', defaultMessage: '关闭' })}
            </Button>
          </Space>
        }
        onCancel={() => {
          setVisible(false);
          setCurrentRecord(null);
        }}
        className="[&_.ant-modal-body]:max-h-[300px] [&_.ant-modal-body]:overflow-auto"
      >
        <Highlight language="sql">{formatSql(currentRecord && currentRecord.info)}</Highlight>
      </Modal>
    </div>
  );
};

export default List;
