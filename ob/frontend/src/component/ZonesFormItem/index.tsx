import React from 'react';
import { Form, Row, Col, Checkbox, InputNumber, Tooltip, Alert } from '@oceanbase/design';
import { formatMessage } from '@/util/intl';
import { findBy } from '@oceanbase/util';
import { uniq, isNil } from 'lodash';
import { REPLICA_TYPE_LIST } from '@/constant/oceanbase';
import { validateUnitCount } from '@/util/oceanbase';
import MySelect from '@/component/MySelect';
import MyInput from '@/component/MyInput';
import { getUnitSpecLimit, getResourcesLimit } from '@/util/cluster';
import './index.less';

const { Option } = MySelect as any;

export interface ZoneItem {
  key: string;
  name: string;
  checked: boolean;
  replicaType: string;
  cpuCore: number;
  memorySize: number;
  unitCount: number;
}

export interface ZonesFormItemProps {
  name: string;
  zones: any[];
  minServerCount?: numnber;
  clusterUnitSpecLimit: {
    cpuLowerLimit: number;
    memoryLowerLimit: number;
  };
  replicaTypeTooltipConfig?: any;
  showAlert?: boolean;
  disabled?: boolean;
}

const ZonesFormItem: React.FC<ZonesFormItemProps> = ({
  name,
  zones,
  minServerCount,
  clusterUnitSpecLimit,
  replicaTypeTooltipConfig,
  showAlert = true,
  disabled = false,
}) => {
  const { getFieldValue } = Form.useFormInstance();

  return (
    <Form.Item style={{ marginBottom: 24 }}>
      <Form.List name={name}>
        {fields => {
          return (
            <>
              {fields.map((field, index: number) => {
                const currentChecked = getFieldValue([name, field.name, 'checked']);

                const currentZoneName = getFieldValue([name, field.name, 'name']);
                const zoneData = findBy(zones || [], 'name', currentZoneName);

                let idleCpuCore = 0;
                let idleMemoryInBytes = 0;
                if (zoneData?.servers?.length > 0 && zoneData?.servers[0]?.stats) {
                  const { idleCpuCoreTotal, idleMemoryInBytesTotal } = getUnitSpecLimit(
                    zoneData?.servers[0]?.stats
                  );
                  idleCpuCore = idleCpuCoreTotal || 0;
                  idleMemoryInBytes = idleMemoryInBytesTotal || 0;
                }

                const { cpuLowerLimit, memoryLowerLimit } = clusterUnitSpecLimit;

                return (
                  <div
                    key={field.key}
                    style={{
                      display: 'flex',
                    }}
                  >
                    <Form.Item
                      style={{ marginBottom: 8, width: 24 }}
                      {...field}
                      label={index === 0 && ' '}
                      name={[field.name, 'checked']}
                      rules={[
                        {
                          required: true,
                          message: formatMessage({
                            id: 'ocp-express.Tenant.New.PleaseCheckZone',
                            defaultMessage: '请勾选 Zone',
                          }),
                        },
                      ]}
                      valuePropName="checked"
                    >
                      <Checkbox disabled={disabled} />
                    </Form.Item>
                    <Row
                      gutter={16}
                      style={{
                        flex: 1,
                      }}
                    >
                      <Col span={5}>
                        <Form.Item
                          style={{ marginBottom: 8 }}
                          {...field}
                          label={
                            index === 0 &&
                            formatMessage({
                              id: 'ocp-express.Tenant.New.ZoneName',
                              defaultMessage: 'Zone 名称',
                            })
                          }
                          name={[field.name, 'name']}
                          rules={[
                            {
                              // 仅勾选行需要必填
                              required: currentChecked,
                            },
                          ]}
                        >
                          <MyInput disabled={true} />
                        </Form.Item>
                      </Col>
                      <Col span={5}>
                        <Form.Item
                          style={{ marginBottom: 8 }}
                          {...field}
                          label={
                            index === 0 &&
                            formatMessage({
                              id: 'ocp-express.Tenant.New.ReplicaType',
                              defaultMessage: '副本类型',
                            })
                          }
                          tooltip={replicaTypeTooltipConfig}
                          name={[field.name, 'replica_type']}
                          rules={[
                            {
                              // 仅勾选行需要必填
                              required: currentChecked,
                              message: formatMessage({
                                id: 'ocp-express.Tenant.New.SelectAReplicaType',
                                defaultMessage: '请选择副本类型',
                              }),
                            },
                          ]}
                          initialValue={'FULL'}
                        >
                          <MySelect disabled={!currentChecked || disabled} optionLabelProp="label">
                            {REPLICA_TYPE_LIST.map(item => (
                              <Option key={item.value} value={item.value} label={item.label}>
                                <Tooltip
                                  placement="right"
                                  title={item.description}
                                  overlayStyle={{
                                    maxWidth: 400,
                                  }}
                                >
                                  <div style={{ width: '100%' }}>{item.label}</div>
                                </Tooltip>
                              </Option>
                            ))}
                          </MySelect>
                        </Form.Item>
                      </Col>
                      <Col span={5}>
                        <Form.Item
                          style={{ marginBottom: 8 }}
                          {...field}
                          label={
                            index === 0 &&
                            formatMessage({
                              id: 'ocp-express.Tenant.New.UnitSpecifications',
                              defaultMessage: 'Unit 规格',
                            })
                          }
                          extra={
                            cpuLowerLimit && idleCpuCore && cpuLowerLimit < idleCpuCore
                              ? formatMessage(
                                  {
                                    id: 'ocp-express.Tenant.New.CurrentConfigurableRangeValueCpulowerlimitIdlecpucore',
                                    defaultMessage:
                                      '当前可配置范围值 {cpuLowerLimit}~{idleCpuCore}',
                                  },
                                  {
                                    cpuLowerLimit: cpuLowerLimit,
                                    idleCpuCore: idleCpuCore,
                                  }
                                )
                              : formatMessage({
                                  id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                                  defaultMessage: '当前可配置资源不足',
                                })
                          }
                          initialValue={
                            getResourcesLimit({
                              idleCpuCore,
                              idleMemoryInBytes,
                            })
                              ? 4
                              : null
                          }
                          name={[field.name, 'cpuCore']}
                          rules={[
                            {
                              // 仅勾选行需要必填
                              required: currentChecked,
                              message: formatMessage({
                                id: 'ocp-express.Tenant.New.EnterTheUnitSpecification',
                                defaultMessage: '请输入 unit 规格',
                              }),
                            },
                            {
                              type: 'number',
                              min: cpuLowerLimit,
                              max: idleCpuCore,
                              message:
                                cpuLowerLimit < idleCpuCore
                                  ? formatMessage(
                                      {
                                        id: 'ocp-express.Tenant.New.CurrentConfigurableRangeValueCpulowerlimitIdlecpucore',
                                        defaultMessage:
                                          '当前可配置范围值 {cpuLowerLimit}~{idleCpuCore}',
                                      },
                                      {
                                        cpuLowerLimit: cpuLowerLimit,
                                        idleCpuCore: idleCpuCore,
                                      }
                                    )
                                  : formatMessage({
                                      id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                                      defaultMessage: '当前可配置资源不足',
                                    }),
                            },
                          ]}
                        >
                          <InputNumber
                            disabled={!currentChecked || disabled}
                            addonAfter={formatMessage({
                              id: 'ocp-express.Tenant.New.Nuclear',
                              defaultMessage: '核',
                            })}
                            step={0.5}
                            style={{ width: '100%' }}
                          />
                        </Form.Item>
                      </Col>
                      <Col span={5} style={{ paddingLeft: 0 }}>
                        <Form.Item
                          style={{ marginBottom: 8 }}
                          {...field}
                          label={index === 0 && ' '}
                          name={[field.name, 'memorySize']}
                          extra={
                            memoryLowerLimit &&
                            idleMemoryInBytes &&
                            memoryLowerLimit < idleMemoryInBytes
                              ? formatMessage(
                                  {
                                    id: 'ocp-express.Tenant.New.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes',
                                    defaultMessage:
                                      '当前可配置范围值 {memoryLowerLimit}~{idleMemoryInBytes}',
                                  },
                                  {
                                    memoryLowerLimit: memoryLowerLimit,
                                    idleMemoryInBytes: idleMemoryInBytes,
                                  }
                                )
                              : formatMessage({
                                  id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                                  defaultMessage: '当前可配置资源不足',
                                })
                          }
                          initialValue={
                            getResourcesLimit({
                              idleCpuCore,
                              idleMemoryInBytes,
                            })
                              ? 8
                              : null
                          }
                          rules={[
                            {
                              // 仅勾选行需要必填
                              required: currentChecked,
                              message: formatMessage({
                                id: 'ocp-express.Tenant.New.EnterTheUnitSpecification',
                                defaultMessage: '请输入 unit 规格',
                              }),
                            },
                            {
                              type: 'number',
                              min: memoryLowerLimit || 0,
                              max: idleMemoryInBytes || 0,
                              message:
                                memoryLowerLimit < idleMemoryInBytes
                                  ? formatMessage(
                                      {
                                        id: 'ocp-express.Tenant.New.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes',
                                        defaultMessage:
                                          '当前可配置范围值 {memoryLowerLimit}~{idleMemoryInBytes}',
                                      },
                                      {
                                        memoryLowerLimit: memoryLowerLimit,
                                        idleMemoryInBytes: idleMemoryInBytes,
                                      }
                                    )
                                  : formatMessage({
                                      id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                                      defaultMessage: '当前可配置资源不足',
                                    }),
                            },
                          ]}
                        >
                          <InputNumber
                            disabled={!currentChecked || disabled}
                            style={{ width: '100%' }}
                            addonAfter="GB"
                          />
                        </Form.Item>
                      </Col>
                      <Col span={3}>
                        <Form.Item noStyle shouldUpdate={true}>
                          {() => {
                            return (
                              <Form.Item
                                style={{ marginBottom: 8 }}
                                {...field}
                                label={
                                  index === 0 &&
                                  formatMessage({
                                    id: 'ocp-express.Tenant.New.UnitQuantity',
                                    defaultMessage: 'Unit 数量',
                                  })
                                }
                                initialValue={minServerCount}
                                name={[field.name, 'unit_num']}
                                rules={[
                                  {
                                    required: true,
                                    message: formatMessage({
                                      id: 'ocp-express.component.FormZoneReplicaTable.UnitQuantityCannotBeEmpty',
                                      defaultMessage: '请输入 Unit 数量',
                                    }),
                                  },
                                  {
                                    validator: (rule, value, callback) =>
                                      validateUnitCount(
                                        rule,
                                        value,
                                        callback,
                                        findBy(zones || [], 'name', currentZoneName)
                                      ),
                                  },
                                ]}
                              >
                                <InputNumber
                                  min={1}
                                  // 设为只读模式，会去掉 InputNumber 的大部分样式
                                  readOnly={true}
                                  // 其余样式需要手动去掉
                                  className="input-number-readonly"
                                  style={{ width: '100%' }}
                                  placeholder={formatMessage({
                                    id: 'ocp-express.Tenant.New.Enter',
                                    defaultMessage: '请输入',
                                  })}
                                />
                              </Form.Item>
                            );
                          }}
                        </Form.Item>
                      </Col>
                    </Row>
                  </div>
                );
              })}
            </>
          );
        }}
      </Form.List>

      {showAlert && (
        <Form.Item noStyle shouldUpdate={true}>
          {() => {
            const currentZones = getFieldValue(name);
            const checkedZones = (currentZones || []).filter((item: any) => item.checked);
            const uniqueCheckedUnitSpecNameList = uniq(
              checkedZones
                // 去掉空值
                .filter((item: any) => !!item.unitSpecName)
                .map((item: any) => item.unitSpecName)
            );

            const uniqueCheckedUnitCountList = uniq(
              checkedZones
                // 去掉空值
                .filter((item: any) => !isNil(item.unitCount))
                .map((item: any) => item.unitCount)
            );

            // 勾选 Zone 的副本类型或副本数量不同时，展示 alert 提示
            return (
              (uniqueCheckedUnitSpecNameList.length > 1 ||
                uniqueCheckedUnitCountList.length > 1) && (
                <Alert
                  message={formatMessage({
                    id: 'ocp-express.Tenant.New.WeRecommendThatYouSetTheSameUnit',
                    defaultMessage:
                      '建议为全能型副本设置相同的 Unit 规格，不同的 Unit 规格可能造成性能或稳定性问题。',
                  })}
                  type="info"
                  showIcon={true}
                  style={{ marginTop: -8, marginBottom: 16 }}
                />
              )
            );
          }}
        </Form.Item>
      )}
    </Form.Item>
  );
};

export default ZonesFormItem;
