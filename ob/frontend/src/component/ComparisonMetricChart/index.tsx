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

import React from 'react';
import { Card } from '@oceanbase/design';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import type { CommonProps, MetricGroupWithChartConfig } from './Item';
import Item from './Item';

export interface ComparisonMetricChartProps extends CommonProps {
  metricGroupList: MetricGroupWithChartConfig[];
  cardStyle?: React.CSSProperties;
  titleStyle?: React.CSSProperties;
}

const ComparisonMetricChart: React.FC<ComparisonMetricChartProps> = ({
  metricGroupList,
  titleStyle,
  ...restProps
}) => {
  return (
    <div className="card-without-padding">
      {metricGroupList.map((item, index) => {
        const title = (
          <ContentWithQuestion
            style={titleStyle}
            content={item?.name}
            tooltip={
              item?.description && {
                placement: 'right',
                // 指标组只包含一个指标，则用指标组信息代替指标信息
                title: (
                  <div>
                    {/* 用第一个指标的单位替代指标组的单位作展示 */}
                    {`${item?.description}`}
                  </div>
                ),
              }
            }
          />
        );

        return (
          <Card
            key={item.key}
            title={title}
            hoverable={false}
            style={{ width: '100%', ...(restProps.cardStyle || {}) }}
          >
            <Item index={index} {...restProps} metricGroup={item} />
          </Card>
        );
      })}
    </div>
  );
};

export default ComparisonMetricChart;
