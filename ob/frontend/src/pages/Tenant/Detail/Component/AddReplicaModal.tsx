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

import BackGroundContainer from '@/component/BackGroundContainer';
import MySelect from '@/component/MySelect';
import { REPLICA_TYPE_LIST } from '@/constant/oceanbase';
import { tenantAddReplicas } from '@/service/obshell/tenant';
import { unitConfigCreate } from '@/service/obshell/unit';
import { getUnitSpecLimit } from '@/util/cluster';
import { formatMessage } from '@/util/intl';
import { taskSuccess } from '@/util/task';
import { TenantZone } from '@/util/tenant';
import { Col, Descriptions, Form, InputNumber, Modal, Row } from '@oceanbase/design';
import { byte2GB, findBy } from '@oceanbase/util';
import { useModel } from '@umijs/max';
import { useRequest } from 'ahooks';
import { isEqual, uniqueId } from 'lodash';
import React, { useEffect, useState } from 'react';

const { Option } = MySelect;

export interface AddReplicaModalProps {
  visible: boolean;
  tenantData: API.TenantInfo;
  tenantZones: TenantZone[];
  clusterZones: API.Zone[];
  unitSpecLimit: any;
  onSuccess: () => void;
  onCancel: () => void;
}

interface ResourcePool {
  cpuCore?: number;
  memorySize?: number;
  unitNum?: number;
  dataDiskSize?: number;
}

const AddReplicaModal: React.FC<AddReplicaModalProps> = ({
  visible,
  tenantData,
  tenantZones,
  clusterZones,
  unitSpecLimit,
  onSuccess,
  ...restProps
}) => {
  const { isSharedStorage } = useModel('cluster');
  const [form] = Form.useForm();
  const { validateFields } = form;
  const [resourcePool, setResourcePool] = useState<ResourcePool>({});

  const tenantZoneNames = (tenantZones || []).map(item => item.name);
  // 已经设置了副本的 zone，不可重复设置
  const unsetZones = clusterZones.filter(item => !tenantZoneNames.includes(item.name!));

  const sharedStorageColSpan = isSharedStorage ? 8 : 12;

  useEffect(() => {
    if (visible && tenantZones?.length > 0) {
      const currentResourcePool: API.ResourcePoolWithUnit = tenantZones[0]?.resourcePool;
      const currentUnitConfig: API.ObUnitConfig = currentResourcePool?.unit_config || {};
      const unitConfigList: API.Zone[] = tenantZones?.filter(item =>
        isEqual(item?.resourcePool?.unit_config, currentUnitConfig)
      );
      if (unitConfigList.length === tenantZones?.length) {
        setResourcePool({
          cpuCore: currentUnitConfig?.max_cpu,
          memorySize: byte2GB(currentUnitConfig?.memory_size || 0),
          unitNum: currentResourcePool?.unit_num,
          dataDiskSize: isSharedStorage
            ? byte2GB(currentUnitConfig?.data_disk_size || 0)
            : undefined,
        });
      }
    }
  }, [visible]);

  const { run: addReplica, loading: addReplicaLoading } = useRequest(tenantAddReplicas, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const taskId = res.data?.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'ocp-express.src.model.tenant.TheTaskForAddingA',
            defaultMessage: '新增副本的任务提交成功',
          }),
        });

        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  const { runAsync: unitConfigCreateFn } = useRequest(unitConfigCreate, {
    manual: true,
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const { zoneName, replicaType, cpuCore, memorySize, dataDiskSize } = values;

      const unitName = `tenant_unit_${Date.now()}_${uniqueId()}`;

      unitConfigCreateFn({
        name: unitName,
        max_cpu: Number(cpuCore),
        memory_size: `${memorySize}GB`,
        data_disk_size: isSharedStorage ? `${dataDiskSize}GB` : undefined,
      }).then(res => {
        if (res.successful) {
          addReplica(
            {
              name: tenantData.tenant_name!,
            },

            {
              zone_list: [
                {
                  name: zoneName,
                  replica_type: replicaType,
                  unit_num: resourcePool?.unitNum || 0,
                  unit_config_name: unitName,
                },
              ],
            }
          );
        }
      });
    });
  };

  const currentZoneName = Form.useWatch('zoneName', form);

  let idleCpuCore, idleMemoryInBytes;
  if (unsetZones?.length > 0) {
    const zoneData = findBy(unsetZones, 'name', currentZoneName);
    const isSingleServer = zoneData?.servers?.length === 1;
    if (isSingleServer) {
      const { idleCpuCoreTotal, idleMemoryInBytesTotal } = getUnitSpecLimit(
        zoneData?.servers?.[0]?.stats!
      );

      idleCpuCore = idleCpuCoreTotal;
      idleMemoryInBytes = idleMemoryInBytesTotal;
    }
  }

  const { cpuLowerLimit, memoryLowerLimit } = unitSpecLimit;

  return (
    <Modal
      width={600}
      visible={visible}
      title={formatMessage({
        id: 'ocp-express.Detail.Component.AddReplicaModal.NewCopy',
        defaultMessage: '新增副本',
      })}
      destroyOnClose={true}
      confirmLoading={addReplicaLoading}
      onOk={handleSubmit}
      {...restProps}
    >
      <Descriptions column={1} style={{ marginBottom: 16 }}>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.AddReplicaModal.Tenant',
            defaultMessage: '租户',
          })}
        >
          {tenantData?.tenant_name}
        </Descriptions.Item>
      </Descriptions>
      <Form form={form} preserve={false} layout="vertical" hideRequiredMark={true}>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.AddReplicaModal.TargetZone',
            defaultMessage: '目标 Zone',
          })}
          name="zoneName"
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.Detail.Component.AddReplicaModal.SelectAZone',
                defaultMessage: '请选择 Zone',
              }),
            },
          ]}
          extra={formatMessage({
            id: 'ocp-v2.Detail.Component.AddReplicaModal.OnlyOneReplicaCanBe',
            defaultMessage: '每个 Zone 只能设置一个副本',
          })}
        >
          <MySelect showSearch={true}>
            {unsetZones.map(item => (
              <Option key={item.name} value={item.name}>
                {item.name}
              </Option>
            ))}
          </MySelect>
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.AddReplicaModal.CopyType',
            defaultMessage: '副本类型',
          })}
          name="replicaType"
          initialValue="FULL"
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.Detail.Component.AddReplicaModal.PleaseSelectTheCopyType',
                defaultMessage: '请选择副本类型',
              }),
            },
          ]}
        >
          <MySelect>
            {REPLICA_TYPE_LIST?.map(item => (
              <Option key={item.value} value={item.value}>
                {item.label}
              </Option>
            ))}
          </MySelect>
        </Form.Item>
        <div style={{ marginBottom: 8 }}>
          {formatMessage({
            id: 'ocp-express.Detail.Component.AddReplicaModal.UnitSpecification',
            defaultMessage: 'Unit 规格',
          })}
        </div>
        <BackGroundContainer style={{ marginBottom: 16 }}>
          <Row gutter={8}>
            <Col span={sharedStorageColSpan}>
              <Form.Item
                label="CPU"
                name="cpuCore"
                initialValue={resourcePool?.cpuCore}
                style={{ marginBottom: 0 }}
                extra={
                  cpuLowerLimit && idleCpuCore && cpuLowerLimit < idleCpuCore
                    ? `[${cpuLowerLimit}，${idleCpuCore}]`
                    : formatMessage({
                        id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                        defaultMessage: '当前可配置资源不足',
                      })
                }
                rules={[
                  {
                    required: true,
                    message: formatMessage({
                      id: 'ocp-express.Detail.Component.AddReplicaModal.EnterTheUnitSpecification',
                      defaultMessage: '请输入 unit 规格',
                    }),
                  },
                ]}
              >
                <InputNumber
                  addonAfter={formatMessage({
                    id: 'ocp-express.Detail.Component.AddReplicaModal.Nuclear',
                    defaultMessage: '核',
                  })}
                  step={0.5}
                  min={cpuLowerLimit || 0.5}
                  max={idleCpuCore}
                />
              </Form.Item>
            </Col>
            <Col span={sharedStorageColSpan}>
              <Form.Item
                label={formatMessage({
                  id: 'OBShell.Detail.Component.AddReplicaModal.Memory',
                  defaultMessage: '内存',
                })}
                name="memorySize"
                initialValue={resourcePool?.memorySize}
                style={{ marginBottom: 0 }}
                extra={
                  memoryLowerLimit && idleMemoryInBytes && memoryLowerLimit < idleMemoryInBytes
                    ? `[${memoryLowerLimit}，${idleMemoryInBytes}]`
                    : formatMessage({
                        id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                        defaultMessage: '当前可配置资源不足',
                      })
                }
                rules={[
                  {
                    required: true,
                    message: formatMessage({
                      id: 'ocp-express.Detail.Component.AddReplicaModal.EnterTheUnitSpecification',
                      defaultMessage: '请输入 unit 规格',
                    }),
                  },
                ]}
              >
                <InputNumber addonAfter="GB" min={memoryLowerLimit} max={idleMemoryInBytes} />
              </Form.Item>
            </Col>
            {isSharedStorage && (
              <Col span={8}>
                <Form.Item
                  label={formatMessage({
                    id: 'OBShell.Detail.Component.AddReplicaModal.DataCaching',
                    defaultMessage: '数据缓存',
                  })}
                  name="dataDiskSize"
                  initialValue={resourcePool?.dataDiskSize}
                  style={{ marginBottom: 0 }}
                  rules={[
                    {
                      required: true,
                      message: formatMessage({
                        id: 'ocp-express.Detail.Component.AddReplicaModal.EnterTheUnitSpecification',
                        defaultMessage: '请输入 unit 规格',
                      }),
                    },
                  ]}
                >
                  <InputNumber addonAfter="GB" min={0} />
                </Form.Item>
              </Col>
            )}
          </Row>
        </BackGroundContainer>
      </Form>
      <Descriptions column={1}>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.Detail.Component.AddReplicaModal.NumberOfUnits',
            defaultMessage: 'Unit 数量',
          })}
        >
          {resourcePool?.unitNum}
        </Descriptions.Item>
      </Descriptions>
    </Modal>
  );
};

export default AddReplicaModal;
