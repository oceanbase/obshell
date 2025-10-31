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
import MetricTopChartItem from '@/component/MetricTopChart/Item';
import type { CommonProps, MetricGroupWithChartConfig } from './Item';
import Item from './Item';

export interface MetricChartProps extends CommonProps {
  metricGroupList: MetricGroupWithChartConfig[];
  cardStyle: React.CSSProperties;
}

const MetricChart: React.FC<MetricChartProps> = ({ metricGroupList, ...restProps }) => {
  return (
    <Card bordered={false}>
      {metricGroupList.map((item, index) => {
        let padding = '0px 12px';
        if (index === 0 || index === 1) {
          padding = '12px 12px 0px 12px';
        }
        if (
          metricGroupList.length % 2 === 0 &&
          (index === metricGroupList.length - 1 || index === metricGroupList.length - 2)
        ) {
          padding = '0px 12px 12px 12px';
        }
        if (metricGroupList.length % 2 === 1 && index === metricGroupList.length - 1) {
          padding = '0px 12px 12px 12px';
        }
        return (
          <Card.Grid
            key={item.key}
            hoverable={false}
            style={{
              width: '50%',
              boxShadow: 'none',
              padding,
              ...(restProps.cardStyle || {}),
            }}
          >
            {item.withLabel ? (
              <MetricTopChartItem index={index} {...restProps} metricGroup={item} />
            ) : (
              <Item index={index} {...restProps} metricGroup={item} />
            )}
          </Card.Grid>
        );
      })}
    </Card>
  );
};

export default MetricChart;
