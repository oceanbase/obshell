import ContentWithReload from '@/component/ContentWithReload';
import Empty from '@/component/Empty';
import RenderConnectionString from '@/component/RenderConnectionString';
import { getDateFormatDisplay } from '@/constant/datetime';
import { TENANT_MODE_LIST, TENANT_STATUS_LIST } from '@/constant/tenant';
import useDocumentTitle from '@/hook/useDocumentTitle';
import useUiMode from '@/hook/useUiMode';
import { getTenantOverView, tenantLock, tenantUnlock } from '@/service/obshell/tenant';
import { getBooleanLabel, getYESorNOLabel } from '@/util';
import { formatMessage } from '@/util/intl';
import {
  Badge,
  BadgeProps,
  Button,
  Card,
  Col,
  Modal,
  Row,
  Space,
  Table,
  TableColumnType,
  Tooltip,
  message,
} from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { findByValue, formatTime, sortByMoment } from '@oceanbase/util';
import { history } from '@umijs/max';
import { useInterval, useRequest } from 'ahooks';
import React from 'react';

const Tenant: React.FC = () => {
  useDocumentTitle(
    formatMessage({
      id: 'ocp-express.Detail.Tenant.TenantManagement',
      defaultMessage: '租户管理',
    })
  );

  const { isDesktopMode } = useUiMode();

  // 预先获取租户列表
  const {
    data: tenantListData,
    loading: tenantListLoading,
    refresh: listTenantsRefresh,
  } = useRequest(getTenantOverView, {
    defaultParams: [{}],
  });

  const tenantList: API.DbaObTenant[] = tenantListData?.data?.contents || [];

  // 如果存在进行中的任务
  const polling = tenantList?.some(
    item =>
      item.status === 'CREATING' ||
      item.status === 'MODIFYING' ||
      item.status === 'RESTORE' ||
      item.status === 'DROPPING'
  );
  useInterval(
    () => {
      listTenantsRefresh();
    },
    polling ? 2000 : undefined
  );

  const { runAsync: tenantLockFn } = useRequest(tenantLock, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        listTenantsRefresh();
        message.success(
          formatMessage({
            id: 'ocp-v2.page.Tenant.TenantLockSucceeded',
            defaultMessage: '租户锁定成功',
          })
        );
      }
    },
  });

  const { runAsync: tenantUnlockFn } = useRequest(tenantUnlock, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        listTenantsRefresh();
        message.success(
          formatMessage({
            id: 'ocp-v2.page.Tenant.TenantUnlockedSuccessfully',
            defaultMessage: '租户解锁成功',
          })
        );
      }
    },
  });

  const handleLock = (record: API.DbaObTenant) => {
    // 锁定
    Modal.confirm({
      title: formatMessage(
        {
          id: 'ocp-express.component.TenantList.AreYouSureYouWant',
          defaultMessage: '确定要锁定租户 {recordName} 吗？',
        },

        { recordName: record.tenant_name }
      ),

      content: formatMessage({
        id: 'ocp-express.component.TenantList.LockingATenantWillCause',
        defaultMessage: '锁定租户会导致用户无法访问该租户，请谨慎操作',
      }),

      okText: formatMessage({
        id: 'ocp-express.component.TenantList.Locking',
        defaultMessage: '锁定',
      }),
      okButtonProps: {
        danger: true,
        ghost: true,
      },

      onOk: () => {
        return tenantLockFn({ name: record?.tenant_name });
      },
    });
  };

  const handleUnlock = (record: API.DbaObTenant) => {
    // 解锁
    Modal.confirm({
      title: formatMessage(
        {
          id: 'ocp-express.component.TenantList.AreYouSureYouWant.1',
          defaultMessage: '确定要解锁租户 {recordName} 吗？',
        },

        { recordName: record.tenant_name }
      ),

      content: formatMessage({
        id: 'ocp-express.component.TenantList.UnlockingTheTenantRestoresUser',
        defaultMessage: '解锁租户将恢复用户对租户的访问',
      }),

      okText: formatMessage({
        id: 'ocp-express.component.TenantList.Unlock',
        defaultMessage: '解锁',
      }),
      okButtonProps: {
        danger: true,
        ghost: true,
      },

      onOk: () => {
        return tenantUnlockFn({ name: record?.tenant_name });
      },
    });
  };

  const columns: TableColumnType<API.DbaObTenant>[] = [
    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.NameOfTheTenant',
        defaultMessage: '租户名',
      }),

      dataIndex: 'tenant_name',
      render: (text: string, record: API.DbaObTenant) => {
        if (record.status === 'CREATING') {
          return record.tenant_name;
        }
        if (record.status === 'RESTORE') {
          return (
            <Tooltip
              title={formatMessage({
                id: 'OBShell.page.Tenant.TheTenantIsBeingRestored',
                defaultMessage: '租户正在恢复中，不支持查看详情',
              })}
              placement="right"
            >
              {record.tenant_name}
            </Tooltip>
          );
        }
        const pathname = `/cluster/tenant/${record.tenant_name}`;
        return (
          <a
            onClick={() => {
              history.push({
                pathname,
              });
            }}
          >
            {record.tenant_name}
          </a>
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.TenantMode',
        defaultMessage: '租户模式',
      }),

      dataIndex: 'mode',
      render: (text: string) => <span>{findByValue(TENANT_MODE_LIST, text).label}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.ConnectionString',
        defaultMessage: '连接字符串',
      }),

      dataIndex: 'connection_strings',
      render: (obproxyAndConnectionStrings: API.ObproxyAndConnectionString[]) => {
        return (
          <RenderConnectionString maxWidth={360} connectionStrings={obproxyAndConnectionStrings} />
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.ReadOnly',
        defaultMessage: '只读',
      }),
      dataIndex: 'read_only',
      render: (text: boolean) => <span>{getBooleanLabel(text)}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.Locked',
        defaultMessage: '锁定',
      }),
      dataIndex: 'locked',
      render: text => <span>{getYESorNOLabel(text)}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.State',
        defaultMessage: '状态',
      }),
      dataIndex: 'status',
      filters: TENANT_STATUS_LIST.map(item => ({
        text: item.label,
        value: item.value,
      })),
      onFilter: (value, record) => record.status === value,
      render: (text: string) => {
        const statusItem = findByValue(TENANT_STATUS_LIST, text);
        return (
          <Badge status={statusItem.badgeStatus as BadgeProps['status']} text={statusItem.label} />
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.CreationTime',
        defaultMessage: '创建时间',
      }),

      dataIndex: 'created_time',
      sorter: (a, b) => sortByMoment(a, b, 'created_time'),
      render: (text: string) => <span>{formatTime(text, getDateFormatDisplay())}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      render: (text: string, record: API.DbaObTenant) => {
        if (record.status === 'CREATING') {
          return (
            <Button size="small" onClick={() => history.push(`/task`)}>
              {formatMessage({
                id: 'ocp-express.component.TenantList.ViewTasks',
                defaultMessage: '查看任务',
              })}
            </Button>
          );
        } else if (record.status === 'FAILED') {
          return null;
        } else {
          return (
            <Space size="middle">
              {record.locked === 'YES' ? (
                <Tooltip
                  title={
                    record.tenant_name === 'sys' &&
                    formatMessage({
                      id: 'ocp-express.component.TenantList.TheSysTenantDoesNot',
                      defaultMessage: 'sys 租户不支持解锁',
                    })
                  }
                  placement="topRight"
                >
                  <Button
                    size="small"
                    disabled={record.tenant_name === 'sys'}
                    onClick={() => handleUnlock(record)}
                  >
                    {formatMessage({
                      id: 'ocp-express.component.TenantList.Unlock',
                      defaultMessage: '解锁',
                    })}
                  </Button>
                </Tooltip>
              ) : (
                <Tooltip
                  title={
                    record.tenant_name === 'sys' &&
                    formatMessage({
                      id: 'ocp-express.component.TenantList.SysTenantsCannotBeLocked',
                      defaultMessage: 'sys 租户不支持锁定',
                    })
                  }
                  placement="topRight"
                >
                  <Button
                    size="small"
                    disabled={record.tenant_name === 'sys' || record.status === 'RESTORE'}
                    onClick={() => handleLock(record)}
                  >
                    {formatMessage({
                      id: 'ocp-express.component.TenantList.Locking',
                      defaultMessage: '锁定',
                    })}
                  </Button>
                </Tooltip>
              )}
            </Space>
          );
        }
      },
    },
  ];

  return !tenantListLoading && tenantList.length === 0 ? (
    <Empty
      title={formatMessage({ id: 'ocp-express.page.Tenant.NoTenant', defaultMessage: '暂无租户' })}
      description={formatMessage({
        id: 'ocp-express.page.Tenant.ThereIsNoDataRecord',
        defaultMessage: '暂无任何数据记录，立即新建一个租户吧！',
      })}
    >
      <Button
        type="primary"
        onClick={() => {
          history.push('/tenantCreate/new');
        }}
      >
        {formatMessage({ id: 'ocp-express.page.Tenant.NewTenant', defaultMessage: '新建租户' })}
      </Button>
    </Empty>
  ) : (
    <PageContainer
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            spin={tenantListLoading}
            content={formatMessage({
              id: 'ocp-express.Detail.Tenant.TenantManagement',
              defaultMessage: '租户管理',
            })}
            onClick={() => {
              // 刷新租户列表 (集群数据)
              listTenantsRefresh();
              // 无需刷新租户监控 Top5，TenantMonitorTop5 重新卸载、挂载后会自动发送请求
            }}
          />
        ),

        extra: (
          <Space>
            {!isDesktopMode && (
              <Button
                onClick={() => {
                  history.push(`/tenant/restore`);
                }}
              >
                {formatMessage({
                  id: 'ocp-v2.Detail.Backup.InitiateRecovery',
                  defaultMessage: '发起恢复',
                })}
              </Button>
            )}
            <Button
              type="primary"
              onClick={() => {
                history.push('/tenantCreate/new');
              }}
            >
              {formatMessage({
                id: 'ocp-express.page.Tenant.NewTenant',
                defaultMessage: '新建租户',
              })}
            </Button>
          </Space>
        ),
      }}
    >
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card
            bordered={false}
            title={formatMessage({
              id: 'ocp-express.component.TenantList.TenantList',
              defaultMessage: '租户列表',
            })}
            className={`card-without-padding`}
          >
            <div>
              <Table
                loading={tenantListLoading}
                dataSource={tenantList}
                columns={columns}
                rowKey={(record: API.DbaObTenant) => record.tenant_id}
              />
            </div>
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Tenant;
