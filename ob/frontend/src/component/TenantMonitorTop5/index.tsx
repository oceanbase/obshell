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
import { connect } from 'umi';
import React, { useState, useEffect } from 'react';
import { Card, Dropdown, Menu } from '@oceanbase/design';
import { ClockCircleOutlined, DownOutlined } from '@oceanbase/icons';
import moment from 'moment';
import { findBy } from '@oceanbase/util';
import { RFC3339_DATE_TIME_FORMAT } from '@/constant/datetime';
import ContentWithIcon from '@/component/ContentWithIcon';
import MetricTopChart from '@/component/MetricTopChart';

export interface TenantMonitorTop5Props {
  dispatch: any;
  clusterName?: string;
  metricGroupList: API.MetricClass[];
}

const TenantMonitorTop5: React.FC<TenantMonitorTop5Props> = ({
  dispatch,
  clusterName,
  metricGroupList,
}) => {
  const [menuKey, setMenuKey] = useState('hour');
  const menuList = [
    {
      value: 'hour',
      label: formatMessage({
        id: 'ocp-express.component.TenantMonitorTop5.LastHour',
        defaultMessage: '最近一小时',
      }),
      range: [moment().subtract(1, 'hours'), moment()],
    },

    {
      value: 'day',
      label: formatMessage({
        id: 'ocp-express.component.TenantMonitorTop5.LastDay',
        defaultMessage: '最近一天',
      }),
      range: [moment().subtract(1, 'days'), moment()],
    },

    {
      value: 'week',
      label: formatMessage({
        id: 'ocp-express.component.TenantMonitorTop5.LastWeek',
        defaultMessage: '最近一周',
      }),
      range: [moment().subtract(1, 'weeks'), moment()],
    },
  ];

  const menuItem = findBy(menuList, 'value', menuKey);
  const menu = (
    <Menu
      onClick={({ key }) => {
        setMenuKey(key);
      }}
    >
      {menuList.map(item => (
        <Menu.Item key={item.value}>{item.label}</Menu.Item>
      ))}
    </Menu>
  );

  useEffect(() => {
    dispatch({
      type: 'monitor/getMetricGroupListData',
      payload: {
        scope: 'TENANT',
        type: 'TOP',
      },
    });
  }, []);

  return (
    <Card
      className="card-without-padding"
      title={formatMessage({
        id: 'ocp-express.component.TenantMonitorTop5.TopTenantMonitoring',
        defaultMessage: '租户监控 Top5',
      })}
      bordered={false}
      extra={
        <Dropdown placement="bottomLeft" overlay={menu}>
          <ContentWithIcon
            className="pointable"
            prefixIcon={{
              component: ClockCircleOutlined,
            }}
            affixIcon={{
              component: DownOutlined,
            }}
            content={menuItem.label}
          />
        </Dropdown>
      }
    >
      <MetricTopChart
        clusterName={clusterName}
        startTime={moment(menuItem.range && menuItem.range[0]).format(RFC3339_DATE_TIME_FORMAT)}
        endTime={moment(menuItem.range && menuItem.range[1]).format(RFC3339_DATE_TIME_FORMAT)}
        // 获取 top5 数据
        limit={5}
        // top 数据的最终聚合维度
        scope="tenant_name"
        metricGroupList={findBy(metricGroupList, 'key', 'metric_tenant_top').metricGroups || []}
        menuKey={menuKey}
      />
    </Card>
  );
};

function mapStateToProps({ monitor }) {
  return {
    metricGroupList: (monitor.metricGroupListData && monitor.metricGroupListData.contents) || [],
  };
}

export default connect(mapStateToProps)(TenantMonitorTop5);
