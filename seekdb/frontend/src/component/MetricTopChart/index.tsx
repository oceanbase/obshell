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
import type { CommonProps, MetricGroupWithChartConfig } from './Item';
import Item from './Item';

export interface MetricTopChartProps extends CommonProps {
  metricGroupList: MetricGroupWithChartConfig[];
  cardStyle: React.CSSProperties;
}

const MetricTopChart: React.FC<MetricTopChartProps> = ({
  metricGroupList,
  cardStyle = {},
  ...restProps
}) => {
  return (
    <Card bordered={false}>
      {metricGroupList.map(item => (
        <Card.Grid key={item.key} hoverable={false} style={{ width: '50%', ...cardStyle }}>
          <Item {...restProps} metricGroup={item} />
        </Card.Grid>
      ))}
    </Card>
  );
};

export default MetricTopChart;
