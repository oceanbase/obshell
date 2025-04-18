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
import { Button, Tooltip, Typography, Table } from '@oceanbase/design';
import React, { useState } from 'react';
import { findByValue } from '@oceanbase/util';
import { ExclamationCircleOutlined } from '@oceanbase/icons';
import { max } from 'lodash';
import { isEnglish } from '@/util';
import { formatTime } from '@/util/datetime';
import { PLAN_TYPE_LIST } from '@/constant/tenant';
import PlanDrawer from './PlanDrawer';

interface PlanTableProps {
  tenantId: string;
  startTime?: string;
  endTime?: string;
  rangeKey?: string;
  topPlans: API.PlanStatGroup[];
  topPlansLoading: boolean;
}

const PlanTable: React.FC<PlanTableProps> = ({
  tenantId,
  startTime,
  endTime,
  rangeKey,
  topPlans,
  topPlansLoading,
}) => {
  const [planGroupRecord, setPlanGroupRecord] = useState<API.PlanStatGroup | null>(null);
  const [visible, setVisible] = useState(false);

  /** 响应时间的横条图 */
  const renderLine = (record: API.PlanStatDetail, key: string) => {
    const maxValue = max(topPlans.map(plan => plan[key]));
    const rate = (record[key] || 0) / (maxValue as number);

    return (
      <div>
        <div
          style={{
            width: rate * 100,
            backgroundColor: '#5b8ff9',
            height: 10,
            display: 'inline-block',
          }}
        />

        <span style={{ marginLeft: 12 }}>{record[key]}</span>
      </div>
    );
  };

  const columns = [
    {
      title: 'Plan Hash',
      dataIndex: 'planHash',
      render: (node: number, record: API.PlanStatGroup) => {
        return (
          <div>
            <Tooltip placement="topLeft" title={node}>
              {/* Button Width 86 , marginRight 8  */}
              <Typography.Link
                style={{
                  width: record?.hitDiagnosis ? 'calc(100% - 94px)' : '100%',
                  minWidth: 100,
                  marginRight: 8,
                }}
                copyable={{ text: node }}
                ellipsis={true}
                onClick={() => {
                  setPlanGroupRecord(record);
                  setVisible(true);
                }}
              >
                {node}
              </Typography.Link>
            </Tooltip>
            {record?.hitDiagnosis && (
              <Button
                size="small"
                icon={<ExclamationCircleOutlined style={{ color: '#fa8c16' }} />}
              >
                <span style={{ fontSize: 12 }}>
                  {formatMessage({
                    id: 'ocp-express.SQLDiagnosis.Component.RecordTable.RecommendedAttention',
                    defaultMessage: '建议关注',
                  })}
                </span>
              </Button>
            )}
          </div>
        );
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.SQLDiagnosis.Component.RecordTable.ExecutionPlanType',
        defaultMessage: '执行计划类型',
      }),

      dataIndex: 'planType',
      render: (text: string) => findByValue(PLAN_TYPE_LIST, text)?.label,
    },

    {
      title: formatMessage({
        id: 'ocp-express.SQLDiagnosis.Component.RecordTable.PlanGenerationTime',
        defaultMessage: '计划生成时间',
      }),

      dataIndex: 'firstLoadTime',
      render: (text: string) => formatTime(text),
    },

    {
      title: formatMessage({
        id: 'ocp-express.SQLDiagnosis.Component.RecordTable.CpuTimeMs',
        defaultMessage: 'CPU 时间（ms）',
      }),

      dataIndex: 'avgCpuTime',
      render: (text: number, record: API.PlanStatDetail) => renderLine(record, 'avgCpuTime'),
    },
  ];

  return (
    <>
      <Table
        rowKey={record => record.planHash}
        dataSource={topPlans}
        columns={columns}
        loading={topPlansLoading}
        size={isEnglish() ? 'middle' : 'large'}
        pagination={{
          hideOnSinglePage: true,
        }}
      />

      <PlanDrawer
        tenantId={tenantId}
        PlanStatGroup={planGroupRecord}
        startTime={startTime}
        endTime={endTime}
        rangeKey={rangeKey}
        onClose={() => setVisible(false)}
        visible={visible}
      />
    </>
  );
};

export default PlanTable;
