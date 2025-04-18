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
import { connect, history } from 'umi';
import React, { useEffect } from 'react';
import { Card, Col, Row } from '@oceanbase/design';
import { flatten } from 'lodash';
import { PageContainer } from '@oceanbase/ui';
import { findBy } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import * as ObTenantController from '@/service/ocp-express/ObTenantController';
import ContentWithReload from '@/component/ContentWithReload';
import MonitorSearch from '@/component/MonitorSearch';
import MetricChart from '@/component/MetricChart';

interface MonitorProps {
  location: {
    query: {
      tab: string;
    };
  };

  match: {
    params: {
      tenantId: number;
    };
  };

  dispatch: any;
  loading: boolean;
  metricGroupList: API.MetricClass[];
}

const Monitor: React.FC<MonitorProps> = ({
  location,
  match: {
    params: { tenantId },
  },
  dispatch,
  loading,
  metricGroupList,
}) => {
  const { query } = location || {};
  const { tab } = query || {};
  const activeKey = tab || metricGroupList[0]?.key;
  const queryData = MonitorSearch.getQueryData(query);

  const { data, loading: listTenantsLoading } = useRequest(ObTenantController.listTenants, {
    defaultParams: [{}],
  });

  const tenants = data?.data?.contents || [];

  const tenantData = tenants.find(item => item.obTenantId === Number(tenantId)) || {};

  const zoneNameList = tenantData.zones?.map(item => item.name);
  const serverList = flatten(
    tenantData.zones?.map(
      item =>
        // 根据租户的 unit 分布获取 server 列表
        item.units?.map(unit => ({
          // 需要 ip 加 port 来区分 server
          ip: `${unit.serverIp}:${unit.serverPort}`,
          zoneName: unit.zoneName,
          // TODO: 待接口扩展 hostId 信息
          hostId: unit.hostId,
        })) || []
    ) || []
  );

  useEffect(() => {
    getMetricGroupList();
  }, []);

  function getMetricGroupList() {
    dispatch({
      type: 'monitor/getMetricGroupListData',
      payload: {
        scope: 'TENANT',
      },
    });
  }

  return (
    <PageContainer
      ghost={true}
      loading={listTenantsLoading}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'ocp-express.Detail.Monitor.PerformanceMonitoring',
              defaultMessage: '性能监控',
            })}
            spin={loading}
            onClick={() => {
              getMetricGroupList();
            }}
          />
        ),
      }}
    >
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <MonitorSearch location={location} zoneNameList={zoneNameList} serverList={serverList} />
        </Col>
        <Col span={24}>
          <Card
            className="card-without-padding card-with-grid-card"
            bordered={false}
            activeTabKey={activeKey}
            onTabChange={key => {
              history.push({
                pathname: `/tenant/${tenantId}/monitor`,
                query: {
                  ...query,
                  tab: key,
                },
              });
            }}
            tabList={metricGroupList.map(item => ({
              key: item.key,
              tab: item.name,
            }))}
          >
            <MetricChart
              metricGroupList={findBy(metricGroupList, 'key', activeKey).metricGroups || []}
              app="OB"
              tenantName={tenantData.name}
              clusterName={tenantData.clusterName}
              obClusterId={tenantData.obClusterId}
              zoneNameList={zoneNameList}
              serverList={serverList}
              {...queryData}
            />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

function mapStateToProps({ loading, monitor }) {
  return {
    loading: loading.effects['monitor/getMetricGroupListData'],
    metricGroupList: (monitor.metricGroupListData && monitor.metricGroupListData.contents) || [],
  };
}

export default connect(mapStateToProps)(Monitor);
