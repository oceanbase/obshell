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

import React, { useState, useRef } from 'react';
import { formatMessage } from '@/util/intl';
import { history, connect } from 'umi';
import { Button, Card, Checkbox, Tooltip, Space } from '@oceanbase/design';
import { SyncOutlined } from '@oceanbase/icons';
import { PageContainer } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import useDocumentTitle from '@/hook/useDocumentTitle';
import * as ObTenantController from '@/service/ocp-express/ObTenantController';
import MyDropdown from '@/component/MyDropdown';
import Empty from '@/component/Empty';
import List from './List';
import Statistics from './Statistics';

export interface IndexProps {
  location: {
    query: {
      tab: string;
      tenantId?: number;
    };
    pathname: string;
  };
  tenantData: API.TenantInfo;
}

const Index: React.FC<IndexProps> = ({ location, tenantData }: IndexProps) => {
  useDocumentTitle(
    formatMessage({
      id: 'ocp-express.Diagnosis.Session.SessionManagement',
      defaultMessage: '会话诊断',
    })
  );
  const [activeOnly, setActiveOnly] = useState(false);
  const [loading, setLoading] = useState(false);
  const listRef = useRef({
    refreshListTenantSessions: () => {},
  });

  const { tab = 'list', tenantId: propTenantId } = location?.query || {};

  const { data, loading: listTenantsLoading } = useRequest(ObTenantController.listTenants, {
    defaultParams: [{}],
  });

  const tenants = data?.data?.contents || [];
  // const tenants = (data?.data?.contents || []).filter(tenant => tenant.status !== 'CREATING');
  const tenantId = (propTenantId && Number(propTenantId)) || tenants[0]?.obTenantId;

  return (
    <PageContainer
      ghost={true}
      header={{
        title: formatMessage({
          id: 'ocp-express.Diagnosis.Session.SessionManagement',
          defaultMessage: '会话诊断',
        }),
        subTitle: (
          <Space style={{ fontWeight: 500, fontSize: 16, marginLeft: 12, color: '#364563' }}>
            <span>
              {formatMessage({
                id: 'ocp-express.Diagnosis.Session.Tenant',
                defaultMessage: '租户:',
              })}
            </span>
            {!listTenantsLoading && (
              <MyDropdown
                menuList={tenants.map(item => {
                  return {
                    value: item.obTenantId,
                    label: item.name,
                  };
                })}
                isSolidIcon={true}
                defaultMenuKey={tenantId}
                onChange={(v: string) => {
                  history.replace({
                    pathname: location.pathname,
                    query: {
                      ...location.query,
                      tenantId: v,
                    },
                  });
                }}
              />
            )}
          </Space>
        ),
      }}
    >
      {tenantData.status === 'CREATING' ? (
        <Empty
          title={formatMessage({
            id: 'ocp-express.Detail.Session.NoData',
            defaultMessage: '暂无数据',
          })}
          description={formatMessage({
            id: 'ocp-express.Detail.Session.TheTenantIsBeingCreated',
            defaultMessage: '租户正在创建中，请等待租户创建完成',
          })}
        >
          <Button
            type="primary"
            onClick={() => {
              history.push(`/cluster/tenant/${tenantId}`);
            }}
          >
            {formatMessage({
              id: 'ocp-express.Detail.Session.AccessTheOverviewPage',
              defaultMessage: '访问总览页',
            })}
          </Button>
        </Empty>
      ) : (
        <Card
          bordered={false}
          activeTabKey={tab}
          onTabChange={key => {
            history.push({
              pathname: `/diagnosis/session`,
              query: {
                tab: key,
                tenantId: String(tenantId),
              },
            });
          }}
          tabList={[
            {
              tab: formatMessage({
                id: 'ocp-express.Diagnosis.Session.TenantSession',
                defaultMessage: '租户会话',
              }),
              key: 'list',
            },
            {
              tab: formatMessage({
                id: 'ocp-express.Diagnosis.Session.SessionStatistics',
                defaultMessage: '会话统计',
              }),
              key: 'statistics',
            },
          ]}
          tabBarExtraContent={
            tab === 'list' && (
              <span>
                <Checkbox
                  checked={activeOnly}
                  onChange={e => {
                    setActiveOnly(e.target.checked);
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.Detail.Session.List.ViewOnlyActiveSessions',
                    defaultMessage: '仅查看活跃会话',
                  })}
                </Checkbox>
                <Tooltip
                  title={formatMessage({
                    id: 'ocp-express.Detail.Session.List.Refresh',
                    defaultMessage: '刷新',
                  })}
                >
                  <SyncOutlined
                    spin={loading}
                    onClick={() => {
                      setLoading(true);
                      listRef.current.refreshListTenantSessions()?.then(() => setLoading(false));
                    }}
                    style={{ marginLeft: 16, cursor: 'pointer' }}
                  />
                </Tooltip>
              </span>
            )
          }
        >
          {tab === 'list' && <List ref={listRef} tenantId={tenantId} activeOnly={activeOnly} />}
          {tab === 'statistics' && <Statistics tenantId={tenantId} />}
        </Card>
      )}
    </PageContainer>
  );
};

function mapStateToProps({ tenant }) {
  return {
    tenantData: tenant.tenantData,
  };
}

export default connect(mapStateToProps)(Index);
