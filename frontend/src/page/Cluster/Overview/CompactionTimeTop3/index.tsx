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
import React from 'react';
import { Empty, Col, Row, theme as designTheme } from '@oceanbase/design';
import { useTheme } from '@oceanbase/charts';
import { orderBy } from 'lodash';
import { useRequest } from 'ahooks';
import MyCard from '@/component/MyCard';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import useStyles from './index.style';
import { formatDuration } from '@/util';
import { getTenantTopCompaction } from '@/service/obshell/tenant';

export interface CompactionTimeTop3Props { }

const CompactionTimeTop3: React.FC<CompactionTimeTop3Props> = () => {
  const { styles } = useStyles();
  const theme = useTheme();
  const { token } = designTheme.useToken();

  // 获取合并时间 Top3 的租户合并数据
  const { data: topCompactionListData, loading } = useRequest(() =>
    getTenantTopCompaction(
      {
        limit: '3',
      },
      {
        HIDE_ERROR_MESSAGE: true,
      }
    )
  );

  let topCompactionList = topCompactionListData?.data?.contents || [];

  // 对数据根据cost_time进行降序排序
  topCompactionList = orderBy(topCompactionList, ['cost_time'], ['desc']);

  // 数据不够，补足三列
  if (topCompactionList.length === 1) {
    topCompactionList = [...topCompactionList, {}, {}];
  } else if (topCompactionList.length === 2) {
    topCompactionList = [...topCompactionList, {}];
  }

  const Statistic = ({ value, unit }) => (
    <span
      style={{
        display: 'flex',
        flexDirection: 'row',
        alignItems: 'center',
      }}
    >
      <span
        style={{
          fontFamily: 'Avenir-Heavy',
          fontSize: 28,
        }}
      >
        {value || '-'}
      </span>
      {unit && (
        <span
          style={{
            fontSize: 12,
            color: token.colorTextTertiary,
            marginLeft: 4,
            marginTop: 12,
          }}
        >
          {unit}
        </span>
      )}
    </span>
  );

  return (
    <MyCard
      loading={loading}
      title={
        <ContentWithQuestion
          content={formatMessage({
            id: 'ocp-express.Component.CompactionTimeTop3.TenantMergeTimeTop',
            defaultMessage: '租户合并时间 Top3',
          })}
          tooltip={{
            title: (
              <div>
                <div>
                  {formatMessage({
                    id: 'ocp-express.Component.CompactionTimeTop3.SortByLastMergeTime',
                    defaultMessage: '按最近 1 次合并时间排序',
                  })}
                </div>
              </div>
            ),
          }}
        />
      }
      style={{
        height: 150,
      }}
    >
      {topCompactionList.length > 0 ? (
        <Row gutter={48}>
          {topCompactionList.map((item, index) => {
            const durationData = formatDuration(item?.cost_time, 1);

            return (
              <Col
                key={item.tenant_name}
                span={8}
                className={index !== topCompactionList.length - 1 ? styles.borderRight : ''}
              >
                <div
                  data-aspm-click="c304253.d308755"
                  data-aspm-desc="租户合并时间 Top3-跳转租户详情"
                  data-aspm-param={``}
                  data-aspm-expo
                  onClick={() => {
                    if (item.tenant_name) {
                      history.push(`/cluster/tenant/${item.tenant_name}`);
                    }
                  }}
                  className={item.tenant_name ? 'ocp-link-hover' : ''}
                  style={{
                    color: token.colorTextTertiary,
                    display: 'inline-block',
                    marginBottom: 8,
                  }}
                >
                  {item.tenant_name || '-'}
                </div>
                <Row gutter={14}>
                  <Col span={10}>
                    <Statistic value={durationData.value} unit={durationData.unitLabel} />
                  </Col>
                </Row>
              </Col>
            );
          })}
        </Row>
      ) : (
        <Empty imageStyle={{ height: 70 }} description="" />
      )}
    </MyCard>
  );
};

export default CompactionTimeTop3;
