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
import React, { useEffect, useState } from 'react';
import { Form, InputNumber, Row, Col, Descriptions, Modal } from '@oceanbase/design';
import { isEqual, minBy, find, uniqueId } from 'lodash';
import type { ModalProps } from '@oceanbase/design/es/modal';
import { useRequest } from 'ahooks';
import { taskSuccess } from '@/util/task';
import { getUnitSpecLimit } from '@/util/cluster';
import { getMinServerCount } from '@/util/tenant';
import { tenantModifyReplicas } from '@/service/obshell/tenant';
import { unitConfigCreate } from '@/service/obshell/unit';
import { byte2GB, isNullValue } from '@oceanbase/util';

export interface BatcModifyUnitModalProps extends ModalProps {
  tenantData?: API.TenantInfo;
  tenantZones?: API.TenantZone[];
  clusterZones: API.Zone[];
  unitSpecLimit: any;
  visible?: boolean;
  onSuccess: () => void;
  onCancel: () => void;
}

const BatchModifyUnitModal: React.FC<BatcModifyUnitModalProps> = ({
  tenantData,
  tenantZones,
  clusterZones,
  unitSpecLimit,
  onSuccess,
  onCancel,
  visible,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;

  const isSingleMachine = clusterZones?.length === 1 && clusterZones[0]?.servers?.length === 1;

  const [resourcePool, setResourcePool] = useState(undefined);

  const minServerCount = getMinServerCount(clusterZones);

  useEffect(() => {
    setUnitSpecs();
  }, [visible]);

  const setUnitSpecs = () => {
    if (visible && tenantZones?.length > 0) {
      const currentResourcePool: API.ResourcePoolWithUnit = tenantZones[0]?.resourcePool;
      const currentUnitConfig: API.ObUnitConfig = currentResourcePool?.unitConfig;
      const unitConfigList: API.Zone[] = tenantZones?.filter(item =>
        isEqual(item?.resourcePool?.unitConfig, currentUnitConfig)
      );

      if (unitConfigList.length === tenantZones?.length) {
        setResourcePool({
          unitCount: currentResourcePool?.unit_num,
          cpuCore: currentUnitConfig?.max_cpu,
          memorySize: byte2GB(currentUnitConfig?.memory_size),
        });
      }
    }
  };

  const { run, loading } = useRequest(tenantModifyReplicas, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const taskId = res.data?.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'ocp-express.Detail.Component.BatchModifyUnitModal.TheTaskOfModifyingUnitInBatchesHas',
            defaultMessage: '批量修改 Unit 的任务提交成功',
          }),
        });

        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  let idleCpuCore, idleMemoryInBytes;
  if (isSingleMachine) {
    // 修改unit  计算当前 zone 内剩余资源可调整的最大值，取最小可用资源的 zone，提示规格调整的最大最小值配置
    const minIdleCpuZone = minBy(
      clusterZones?.map(
        zone => zone?.servers?.[0]?.stats && getUnitSpecLimit(zone?.servers[0]?.stats)
      ),

      'idleCpuCoreTotal'
    );

    const minIdleMemoryZone = minBy(
      clusterZones?.map(
        zone => zone?.servers?.[0]?.stats && getUnitSpecLimit(zone?.servers[0]?.stats)
      ),

      'idleMemoryInBytesTotal'
    );
    // 修改unit  当前已经分配资源最小zoen
    const minCpuCoreAssignedZone = minBy(
      clusterZones?.map(zone => zone?.servers?.[0]?.stats),

      'cpu_core_assigned'
    )?.zone;
    const minMemoryInBytesAssignedZone = minBy(
      clusterZones?.map(zone => zone?.servers?.[0]?.stats),

      'memory_in_bytes_assigned'
    )?.zone;

    const minCpuZone = find(tenantZones, item => item.name === minCpuCoreAssignedZone);
    const minMemoryZone = find(tenantZones, item => item.name === minMemoryInBytesAssignedZone);

    // 修改 unit 时，CUP可配置范围上限，当前 unit 已分配CUP + 剩余空闲CUP
    if (minIdleCpuZone && minCpuZone) {
      idleCpuCore =
        minIdleCpuZone?.idleCpuCoreTotal + minCpuZone?.resourcePool?.unitConfig?.max_cpu;
    }
    // 修改 unit 时，可配置范围上限，当前 unit 已分配内存 + 剩余空闲内存
    if (minIdleMemoryZone && minMemoryZone) {
      idleMemoryInBytes =
        minIdleMemoryZone?.idleMemoryInBytesTotal +
        minMemoryZone?.resourcePool?.unitConfig?.memorySize;
    }
  }

  const { runAsync: unitConfigCreateFn } = useRequest(unitConfigCreate, {
    manual: true,
  });

  const { cpuLowerLimit, memoryLowerLimit } = unitSpecLimit;

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-express.Detail.Component.BatchModifyUnitModal.ModifyUnit',
        defaultMessage: '修改 Unit',
      })}
      visible={visible}
      destroyOnClose={true}
      confirmLoading={loading}
      onOk={() => {
        validateFields().then(values => {
          const { cpuCore, memorySize, unitCount } = values;

          const unitName = `tenant_unit_${Date.now()}_${uniqueId()}`;

          unitConfigCreateFn({
            name: unitName,
            max_cpu: Number(cpuCore),
            memory_size: memorySize + 'GB',
          }).then(res => {
            if (res.successful) {
              run(
                {
                  name: tenantData?.tenant_name,
                },

                {
                  zone_list: tenantZones?.map(item => {
                    const { name, replicaType, resourcePool } = item;
                    return {
                      zone_name: name,
                      replica_type: replicaType,
                      unit_num: unitCount || resourcePool?.unit_num,
                      unit_config_name: unitName,
                    };
                  }),
                }
              );
            }
          });
        });
      }}
      onCancel={() => {
        onCancel();
      }}
      {...restProps}
    >
      <Descriptions column={2}>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.BatchModifyUnitModal.OperationObject',
            defaultMessage: '操作对象',
          })}
        >
          {tenantData?.tenant_name}
          {formatMessage({
            id: 'ocp-express.Detail.Component.BatchModifyUnitModal.AllZonesUnder',
            defaultMessage: '下的所有 Zone',
          })}
        </Descriptions.Item>
      </Descriptions>
      <Form form={form} preserve={false} layout="vertical" hideRequiredMark={true}>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.BatchModifyUnitModal.NumberOfUnits',
            defaultMessage: 'Unit 数量',
          })}
          name="unitCount"
          initialValue={resourcePool?.unitCount}
          extra={formatMessage(
            {
              id: 'ocp-express.Detail.Component.BatchModifyUnitModal.TheMaximumNumberOfcurrentKesheIsMinservercount',
              defaultMessage:
                '当前可设最大个数为 {minServerCount} (Zone 中最少 OBServer 数决定 Unit 可设最大个数)',
            },
            { minServerCount: minServerCount }
          )}
        >
          <InputNumber
            min={1}
            placeholder={formatMessage({
              id: 'ocp-express.Detail.Component.BatchModifyUnitModal.NullIndicatesThatTheNumberOfUnitsIs',
              defaultMessage: '为空表示不修改 Unit 数量',
            })}
            style={{ width: 168 }}
          />
        </Form.Item>
        <Row gutter={8}>
          <Col span={9} style={{ paddingRight: 8 }}>
            <Form.Item
              label={formatMessage({
                id: 'ocp-express.Detail.Component.BatchModifyUnitModal.UnitSpecification',
                defaultMessage: 'Unit 规格',
              })}
              name="cpuCore"
              initialValue={resourcePool?.cpuCore}
              extra={
                !isNullValue(cpuLowerLimit) &&
                !isNullValue(idleCpuCore) &&
                (cpuLowerLimit < idleCpuCore
                  ? formatMessage(
                      {
                        id: 'ocp-express.Detail.Component.BatchModifyUnitModal.currentConfigurableRangeValueCpulowerlimitIdlecpucore',
                        defaultMessage: '当前可配置范围值 {cpuLowerLimit}~{idleCpuCore}',
                      },
                      { cpuLowerLimit: cpuLowerLimit, idleCpuCore: idleCpuCore }
                    )
                  : formatMessage({
                      id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                      defaultMessage: '当前可配置资源不足',
                    }))
              }
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'ocp-express.Detail.Component.BatchModifyUnitModal.EnterTheUnitSpecification',
                    defaultMessage: '请输入 unit 规格',
                  }),
                },
              ]}
            >
              <InputNumber
                addonAfter={formatMessage({
                  id: 'ocp-express.Detail.Component.BatchModifyUnitModal.Nuclear',
                  defaultMessage: '核',
                })}
                step={0.5}
              />
            </Form.Item>
          </Col>
          <Col span={9}>
            <Form.Item
              label=" "
              name="memorySize"
              initialValue={resourcePool?.memorySize}
              extra={
                !isNullValue(memoryLowerLimit) &&
                !isNullValue(idleMemoryInBytes) &&
                (memoryLowerLimit < idleMemoryInBytes
                  ? formatMessage(
                      {
                        id: 'ocp-express.Detail.Component.BatchModifyUnitModal.currentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes',
                        defaultMessage: '当前可配置范围值 {memoryLowerLimit}~{idleMemoryInBytes}',
                      },
                      { memoryLowerLimit: memoryLowerLimit, idleMemoryInBytes: idleMemoryInBytes }
                    )
                  : formatMessage({
                      id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                      defaultMessage: '当前可配置资源不足',
                    }))
              }
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'ocp-express.Detail.Component.BatchModifyUnitModal.EnterTheUnitSpecification',
                    defaultMessage: '请输入 unit 规格',
                  }),
                },
              ]}
            >
              <InputNumber addonAfter="GB" />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </Modal>
  );
};

export default BatchModifyUnitModal;
