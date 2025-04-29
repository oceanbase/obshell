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
import React from 'react';
import { InputNumber } from '@oceanbase/design';
import { findByValue, findBy, byte2GB } from '@oceanbase/util';
import { REPLICA_TYPE_LIST } from '@/constant/oceanbase';
import { getUnitSpecLimit } from '@/util/cluster';
import type { FormEditTableProps } from '@/component/FormEditTable';
import FormEditTable from '@/component/FormEditTable';
import MySelect from '@/component/MySelect';
import UnitSpec from '@/component/UnitSpec';
import { format } from 'crypto-js';
import { formatSize } from '@/util';

const { Option } = MySelect;

export interface FormEditZoneReplicaTableProps<T> extends FormEditTableProps<T> {
  className?: string;
  value: T[];
  tenantData: API.TenantInfo;
  unitSpecLimit?: any;
  saveLoading?: boolean;
  dispatch?: any;
}

class FormEditZoneReplicaTable<T> extends FormEditTable<FormEditZoneReplicaTableProps<T>> {
  constructor(props: FormEditZoneReplicaTableProps) {
    super(props);
    this.state = {
      ...(super.state || {}),
    };
  }

  // 删除 zone
  public handleDeleteRecord = (record: any) => {
    const { value } = this.props;
    const newValue = value.filter(item => item.key !== record.key);
    this.handleValueChange(newValue);
  };

  public render() {
    const { saveLoading, clusterData, unitSpecLimit } = this.props;

    const columns = [
      {
        title: formatMessage({
          id: 'ocp-express.component.FormEditZoneReplicaTable.ZoneName',
          defaultMessage: 'Zone 名称',
        }),

        dataIndex: 'name',
      },

      {
        title: formatMessage({
          id: 'ocp-express.component.FormEditZoneReplicaTable.CopyType',
          defaultMessage: '副本类型',
        }),

        dataIndex: 'replicaType',
        width: '20%',
        editable: false,
        fieldProps: () => ({
          rules: [
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.component.FormEditZoneReplicaTable.SelectAReplicaType',
                defaultMessage: '请选择副本类型',
              }),
            },
          ],
        }),

        fieldComponent: () => (
          <MySelect>
            {REPLICA_TYPE_LIST.map(item => (
              <Option key={item.value} value={item.value}>
                {item.label}
              </Option>
            ))}
          </MySelect>
        ),

        render: (text: API.ReplicaType) => (
          <span>{findByValue(REPLICA_TYPE_LIST, text).label}</span>
        ),
      },
      {
        title: formatMessage({
          id: 'ocp-express.component.FormEditZoneReplicaTable.UnitSpecification',
          defaultMessage: 'Unit 规格',
        }),
        dataIndex: 'resourcePool.unitConfig',
        width: '40%',
        editable: true,
        fieldComponent: (text, record) => {
          const zoneData = findBy(clusterData?.zones || [], 'name', record.name);

          let idleCpuCore, idleMemoryInBytes;
          if (zoneData?.servers?.length > 0 && zoneData?.servers[0]?.stats) {
            const { idleCpuCoreTotal, idleMemoryInBytesTotal } = getUnitSpecLimit(
              zoneData?.servers[0]?.stats
            );

            idleCpuCore = idleCpuCoreTotal;
            idleMemoryInBytes = idleMemoryInBytesTotal;
          }

          return (
            <UnitSpec
              unitSpecLimit={unitSpecLimit}
              idleUnitSpec={{ idleCpuCore, idleMemoryInBytes }}
              defaultUnitSpec={{
                ...record?.resourcePool?.unitConfig,
              }}
            />
          );
        },

        render: (text: string, record: API.TenantZone) => {
          const { unitConfig } = (record.resourcePool as API.ResourcePool) || {};
          const { max_cpu: maxCpuCoreCount, memory_size } = (unitConfig as API.UnitConfig) || {};

          const maxMemorySizeGB = byte2GB(memory_size);

          return (
            <ul>
              <li>
                {formatMessage(
                  {
                    id: 'ocp-express.component.FormEditZoneReplicaTable.CpuCoreMincpucorecountMaxcpucorecount',
                    defaultMessage: 'CPU（核）： {maxCpuCoreCount}',
                  },

                  { maxCpuCoreCount }
                )}
              </li>
              <li>
                {formatMessage(
                  {
                    id: 'ocp-express.component.FormEditZoneReplicaTable.MemoryGbMinmemorysizegbMaxmemorysizegb',
                    defaultMessage: '内存（GB）：{maxMemorySizeGB}',
                  },

                  { maxMemorySizeGB }
                )}
              </li>
            </ul>
          );
        },
      },

      // {
      //   title: (
      //     <ContentWithQuestion
      //       content={formatMessage({
      //         id: 'ocp-express.component.FormEditZoneReplicaTable.UnitSpecification',
      //         defaultMessage: 'Unit 规格',
      //       })}
      //       tooltip={{
      //         title: unitSpecLimitRule && getUnitSpecLimitText(unitSpecLimitRule),
      //       }}
      //     />
      //   ),

      //   dataIndex: 'resourcePool.unitSpecName',
      //   width: '25%',
      //   editable: true,
      //   fieldComponent: () => (
      //     <UnitSpecSelect
      //       allowClear={true}
      //       obVersion={tenantData?.obVersion}
      //       placeholder={
      //         <Tooltip
      //           placement="topLeft"
      //           title={formatMessage({
      //             id: 'ocp-express.component.FormEditZoneReplicaTable.EmptyIndicatesThatTheUnit',
      //             defaultMessage: '为空表示不修改 Unit 规格',
      //           })}
      //         >
      //           <span>
      //             {formatMessage({
      //               id: 'ocp-express.component.FormEditZoneReplicaTable.EmptyIndicatesThatTheUnit',
      //               defaultMessage: '为空表示不修改 Unit 规格',
      //             })}
      //           </span>
      //         </Tooltip>
      //       }
      //     />
      //   ),
      //   fieldProps: (text, record) => ({
      //     rules: [
      //       {
      //         validator: (rule, value = text, callback) => {
      //           if (value) {
      //             return validatorUnitResource(
      //               rule,
      //               record.resourcePool.unitConfig,
      //               callback,
      //               findBy(tenantData.zones || [], 'name', record.name)
      //             );
      //           } else {
      //             return callback();
      //           }
      //         },
      //       },
      //     ],
      //   }),

      //   render: (text: string, record: API.TenantZone) => {
      //     const { unitConfig } = (record.resourcePool as API.ResourcePool) || {};
      //     const { maxCpuCoreCount, maxMemorySize: maxMemorySizeGB } =
      //       (unitConfig as API.UnitConfig) || {};

      //     return (
      //       <ul>
      //         <li>
      //           {formatMessage(
      //             {
      //               id: 'ocp-express.component.FormEditZoneReplicaTable.CpuCoreMincpucorecountMaxcpucorecount',
      //               defaultMessage: 'CPU（核）： {maxCpuCoreCount}',
      //             },

      //             { maxCpuCoreCount }
      //           )}
      //         </li>
      //         <li>
      //           {formatMessage(
      //             {
      //               id: 'ocp-express.component.FormEditZoneReplicaTable.MemoryGbMinmemorysizegbMaxmemorysizegb',
      //               defaultMessage: '内存（GB）：{maxMemorySizeGB}',
      //             },

      //             { maxMemorySizeGB }
      //           )}
      //         </li>
      //       </ul>
      //     );
      //   },
      // },

      {
        title: formatMessage({
          id: 'ocp-express.component.FormEditZoneReplicaTable.UnitQuantity',
          defaultMessage: 'Unit 数量',
        }),

        dataIndex: 'resourcePool.unit_num',
        editable: false,
        fieldProps: () => ({
          rules: [
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.component.FormEditZoneReplicaTable.EnterTheUnitQuantity',
                defaultMessage: '请输入 Unit 数量',
              }),
            },
          ],
        }),

        fieldComponent: () => (
          <InputNumber
            min={1}
            placeholder={formatMessage({
              id: 'ocp-express.component.FormEditZoneReplicaTable.NullIndicatesThatTheNumberOfUnitsIs',
              defaultMessage: '为空表示不修改 Unit 数量',
            })}
          />
        ),
      },
    ];

    return (
      <FormEditTable
        mode="table"
        allowSwitch={true}
        allowAdd={false}
        saveLoading={saveLoading}
        columns={columns}
        pagination={false}
        {...this.props}
      />
    );
  }
}

export default FormEditZoneReplicaTable;
