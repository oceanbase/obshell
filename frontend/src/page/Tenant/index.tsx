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
import { history } from 'umi';
// import { EllipsisOutlined } from '@oceanbase/icons';
import React from 'react';
import { PageContainer } from '@oceanbase/ui';
import { Badge, Button, Card, Col, Modal, Row, Space, Table, Tooltip } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import useDocumentTitle from '@/hook/useDocumentTitle';
import ContentWithReload from '@/component/ContentWithReload';
import Empty from '@/component/Empty';
import { getTenantOverView, tenantLock, tenantUnlock } from '@/service/obshell/tenant';
import { directTo, findByValue, formatTime, sortByMoment } from '@oceanbase/util';
import RenderConnectionString from '@/component/RenderConnectionString';
import { DATE_FORMAT_DISPLAY } from '@/constant/datetime';
import { TENANT_MODE_LIST, TENANT_STATUS_LIST } from '@/constant/tenant';
import { getBooleanLabel, getYESorNOLabel } from '@/util';
import useStyles from './index.style';
import { message } from 'antd';

interface TenantProps {
  location: {
    query: {
      status?: string;
    };
  };
}

const Tenant: React.FC<TenantProps> = ({
  location: {
    query: { status },
  },
}: TenantProps) => {
  const { styles } = useStyles();
  const statusList = status?.split(',') || [];

  useDocumentTitle(
    formatMessage({
      id: 'ocp-express.Detail.Tenant.TenantManagement',
      defaultMessage: '租户管理',
    })
  );

  // 预先获取租户列表
  const {
    data: tenantListData,
    loading: tenantListLoading,
    refresh: listTenantsRefresh,
  } = useRequest(getTenantOverView, {
    defaultParams: [
      {
        mode: 'MYSQL',
      },
    ],
  });

  const tenantList = tenantListData?.data?.contents || [];

  const { runAsync: tenantLockFn } = useRequest(tenantLock, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        listTenantsRefresh();
        message.success('租户锁定成功');
      }
    },
  });

  const { runAsync: tenantUnlockFn } = useRequest(tenantUnlock, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        listTenantsRefresh();
        message.success('租户解锁成功');
      }
    },
  });

  // const handleMenuClick = (key: string) => {
  //   if (key === 'parameterTemplate') {
  //     history.push('/tenant/parameterTemplate');
  //   }
  //   if (key === 'unitSpec') {
  //     history.push('/tenant/unitSpec');
  //   }
  // };

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

  const columns = [
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
        const pathname = `/tenant/${record.tenant_name}`;
        return (
          <a
            onClick={() => {
              history.push(pathname);
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
      filters: TENANT_MODE_LIST.map(item => ({
        text: item.label,
        value: item.value,
      })),
      // filteredValue: modeList,
      render: (text: API.TenantMode) => <span>{findByValue(TENANT_MODE_LIST, text).label}</span>,
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.ConnectionString',
        defaultMessage: '连接字符串',
      }),

      dataIndex: 'connection_strings',
      render: (obproxyAndConnectionStrings: API.ObproxyAndConnectionString[]) => {
        return (
          <RenderConnectionString
            data-aspm-click="c304184.d308808"
            data-aspm-desc="租户列表-跳转租户详情"
            data-aspm-param={``}
            data-aspm-expo
            maxWidth={360}
            connectionStrings={obproxyAndConnectionStrings}
          />
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
      render: (text: boolean) => <span>{getYESorNOLabel(text)}</span>,
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
      onFilter: (value: string, record: API.TenantStatus) => record.status === value,
      render: (text: API.TenantStatus) => {
        const statusItem = findByValue(TENANT_STATUS_LIST, text);
        return <Badge status={statusItem.badgeStatus} text={statusItem.label} />;
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.TenantList.CreationTime',
        defaultMessage: '创建时间',
      }),

      dataIndex: 'created_time',
      sorter: (a, b) => sortByMoment(a, b, 'created_time'),
      render: (text: string) => <span>{formatTime(text, DATE_FORMAT_DISPLAY)}</span>,
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
            <a
              data-aspm-click="c304184.d308812"
              data-aspm-desc="租户列表-查看创建任务"
              data-aspm-param={``}
              data-aspm-expo
              onClick={() => history.push(`/task`)}
            >
              {formatMessage({
                id: 'ocp-express.component.TenantList.ViewTasks',
                defaultMessage: '查看任务',
              })}
            </a>
          );
        } else if (record.status === 'FAILED') {
          return null;
          // return (
          //   <Space size="middle">
          //     <a
          //       data-aspm-click="c304184.d308811"
          //       data-aspm-desc="租户列表-查看失败任务"
          //       data-aspm-param={``}
          //       data-aspm-expo
          //       onClick={() => history.push(`/task`)}
          //     >
          //       {formatMessage({
          //         id: 'ocp-express.component.TenantList.ViewTasks',
          //         defaultMessage: '查看任务',
          //       })}
          //     </a>
          //     <Popconfirm
          //       placement="topRight"
          //       title={formatMessage({
          //         id: 'ocp-express.component.TenantList.AreYouSureYouWantToDeleteThis',
          //         defaultMessage: '确定要删除此租户吗？',
          //       })}
          //       onConfirm={() =>
          //         ObTenantController.deleteTenant({
          //           tenantId: record.obTenantId,
          //         }).then(() => refresh())
          //       }
          //     >
          //       <a
          //         data-aspm-click="c304184.d308806"
          //         data-aspm-desc="租户列表-删除失败租户"
          //         data-aspm-param={``}
          //         data-aspm-expo
          //       >
          //         {formatMessage({
          //           id: 'ocp-express.component.TenantList.Delete',
          //           defaultMessage: '删除',
          //         })}
          //       </a>
          //     </Popconfirm>
          //   </Space>
          // );
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
                  <a
                    data-aspm-click="c304184.d308810"
                    data-aspm-desc="租户列表-解锁租户"
                    data-aspm-param={``}
                    data-aspm-expo
                    disabled={record.tenant_name === 'sys'}
                    onClick={() => handleUnlock(record)}
                  >
                    {formatMessage({
                      id: 'ocp-express.component.TenantList.Unlock',
                      defaultMessage: '解锁',
                    })}
                  </a>
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
                    data-aspm-click="c304184.d308807"
                    data-aspm-desc="租户列表-锁定租户"
                    data-aspm-param={``}
                    data-aspm-expo
                    type="link"
                    disabled={record.tenant_name === 'sys'}
                    onClick={() => handleLock(record)}
                  >
                    {formatMessage({
                      id: 'ocp-express.component.TenantList.Locking',
                      defaultMessage: '锁定',
                    })}
                  </Button>
                </Tooltip>
              )}

              {/* <a
                data-aspm-click="c304184.d308813"
                data-aspm-desc="租户列表-复制租户"
                data-aspm-param={``}
                data-aspm-expo
                onClick={() => {
                  directTo(`/tenant/new?tenantId=${record.obTenantId}`);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.component.TenantList.Copy',
                  defaultMessage: '复制',
                })}
              </a> */}
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
      {/* <Button
        style={{ marginRight: 12 }}
        onClick={() => {
          history.push('/tenant/unitSpec');
        }}
      >
        {formatMessage({
          id: 'ocp-express.page.Tenant.UnitSpecificationManagement',
          defaultMessage: 'Unit 规格管理',
        })}
      </Button> */}
      {/* <Button
        style={{ marginRight: 12 }}
        onClick={() => {
          history.push('/tenant/parameterTemplate');
        }}
      >
        {formatMessage({
          id: 'ocp-express.page.Tenant.TenantParameterTemplateManagement',
          defaultMessage: '租户参数模板管理',
        })}
      </Button> */}
      <Button
        data-aspm-click="c318544.d343271"
        data-aspm-desc="暂无租户-新建租户"
        data-aspm-param={``}
        data-aspm-expo
        type="primary"
        onClick={() => {
          history.push('/tenant/new');
        }}
      >
        {formatMessage({ id: 'ocp-express.page.Tenant.NewTenant', defaultMessage: '新建租户' })}
      </Button>
    </Empty>
  ) : (
    <PageContainer
      className={styles.container}
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
          <>
            <Button
              data-aspm-click="c304184.d308814"
              data-aspm-desc="租户列表-新建租户"
              data-aspm-param={``}
              data-aspm-expo
              type="primary"
              onClick={() => {
                history.push('/tenant/new');
              }}
            >
              {formatMessage({
                id: 'ocp-express.page.Tenant.NewTenant',
                defaultMessage: '新建租户',
              })}
            </Button>
            {/* 二期 */}
            {/* <Dropdown
              overlay={
                <Menu
                  onClick={({ key }) => {
                    handleMenuClick(key);
                  }}
                >
                  <Menu.Item key="parameterTemplate">
                    <a>
                      {formatMessage({
                        id: 'ocp-express.page.Tenant.TenantParameterTemplateManagement',
                        defaultMessage: '租户参数模板管理',
                      })}
                    </a>
                  </Menu.Item>
                  <Menu.Item key="unitSpec">
                    <a>Unit 规格管理</a>
                  </Menu.Item>
                </Menu>
              }
            >
              <Button>
                <EllipsisOutlined />
              </Button>
            </Dropdown> */}
          </>
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
                dataSource={tenantList}
                columns={columns}
                rowKey={(record: API.Tenant) => record.id}
              />
            </div>
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Tenant;
