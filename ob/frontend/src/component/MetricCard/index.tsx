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

import React, { useState } from 'react';
import { Empty, Spin } from '@oceanbase/design';
import { Modal } from '@oceanbase/design';
import { FullscreenOutlined } from '@oceanbase/icons';
import type { ChartProps } from '@/component/Chart';
import Chart from '@/component/Chart';
import MyCard from '@/component/MyCard';

export interface MetricCardProps {
  loading?: boolean;
  title: React.ReactNode;
  extra?: React.ReactNode;
  range?: React.ReactNode;
  typeButton?: React.ReactNode;
  chartConfig: ChartProps;
}

const MetricCard: React.FC<MetricCardProps> = ({
  loading = false,
  title,
  extra,
  range,
  typeButton,
  chartConfig = { data: [] },
  ...restProps
}) => {
  const [visible, setVisible] = useState(false);
  const { data = [] } = chartConfig || {};
  const realChartConfig = {
    ...chartConfig,
    ...(chartConfig.type === 'Line'
      ? {
          interactions: [
            {
              type: 'brush-x',
            },
          ],
        }
      : {}),
  };

  return (
    <MyCard
      loading={loading}
      title={title}
      extra={
        <span style={{ display: 'flex', alignItems: 'center' }}>
          {extra}
          <FullscreenOutlined onClick={() => setVisible(true)} style={{ cursor: 'pointer' }} />
        </span>
      }
      {...restProps}
    >
      {data?.length > 0 ? (
        <>
          {typeButton}
          <Chart height={186} {...realChartConfig} />
        </>
      ) : (
        <>
          {typeButton}
          <Empty style={{ height: 160 }} imageStyle={{ marginTop: 46 }} />
        </>
      )}
      <Modal
        width={960}
        title={title}
        visible={visible}
        destroyOnClose={true}
        footer={false}
        onCancel={() => setVisible(false)}
      >
        <div>
          {range}
          <Spin spinning={loading}>
            {data?.length > 0 ? (
              <Chart height={300} {...realChartConfig} />
            ) : (
              <Empty style={{ height: 240 }} imageStyle={{ marginTop: 60 }} />
            )}
          </Spin>
        </div>
      </Modal>
    </MyCard>
  );
};

export default MetricCard;
