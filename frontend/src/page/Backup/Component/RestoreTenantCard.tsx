import ContentWithQuestion from '@/component/ContentWithQuestion';
import FormPrimaryZone from '@/component/FormPrimaryZone';
import ZonesFormItem from '@/component/ZonesFormItem';
import MyInput from '@/component/MyInput';
import PageLoading from '@/component/PageLoading';
import { NAME_RULE } from '@/constant';
import { formatMessage } from '@/util/intl';
import { obclusterInfo, getUnitConfigLimit } from '@/service/obshell/obcluster';
import { getMinServerCount } from '@/util/tenant';
import React, { forwardRef, useEffect } from 'react';
import { uniqueId } from 'lodash';
import { Card, Col, Form, InputNumber, Row } from '@oceanbase/design';
import type { FormInstance } from '@oceanbase/design/es/form';
import { useRequest } from 'ahooks';

export interface RestoreTenantCardProps {
  form: FormInstance;
  title?: string;
  strategyData?: API.OcpBackupSamplingInspectConfig;
  // 源集群的构建版本号
  minObBuildVersion?: string;
  useType?: string;
}

const RestoreTenantCard: React.FC<RestoreTenantCardProps> = forwardRef(
  (
    {
      form,
      title = formatMessage({
        id: 'ocp-v2.Backup.Component.RestoreTenantCard.RestoreTargetInformation',
        defaultMessage: '恢复目标信息',
      }),
      useType,
      strategyData,
    },
    ref
  ) => {
    // 获取 unit 规格的限制规则
    const { data: clusterUnitSpecLimitData } = useRequest(getUnitConfigLimit);

    const clusterUnitSpecLimit = {
      cpuLowerLimit: clusterUnitSpecLimitData?.data?.min_cpu || 0,
      memoryLowerLimit: clusterUnitSpecLimitData?.data?.min_memory || 0,
    };

    const { getFieldValue, setFieldsValue } = form;

    // 从抽检策略中获取默认的 Unit 数量
    // 由于 OB 4.0 租户副本的 Unit 数量是同构，因此取第一个 unitCount 即可
    const defaultZonesUnitCount =
      strategyData?.restoreTenantInfoParam?.zones?.map(item => item.resourcePool?.unitCount)?.[0] ||
      1;

    const defaultZones =
      strategyData?.restoreTenantInfoParam?.zones?.map(item => ({
        ...item,
        checked: true,
      })) || [];

    // 获取目标集群的数据
    const {
      data,
      runAsync: getObCluster,
      loading,
    } = useRequest(obclusterInfo, {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          // 所选集群改变时，重置 zones 和 primaryZone
          setFieldsValue({
            zones: res.data?.zones?.map(item => ({
              key: uniqueId(),
              checked: true,
              name: item.name,
              replica_type: 'FULL',
            })),
            primaryZone: undefined,
            restore_tenant_name: undefined,
          });
        }
      },
    });
    const clusterData = data?.data || {};

    const minServerCount = getMinServerCount(clusterData?.zones);

    useEffect(() => {
      getObCluster();
    }, []);

    return (
      <Card bordered={false} title={title}>
        <Form form={form} layout="vertical" colon={false} hideRequiredMark={true}>
          <Row gutter={24}>
            <Col span={8}>
              <Form.Item
                label={formatMessage({
                  id: 'ocp-v2.Backup.Component.RestoreTenantCard.TargetTenant',
                  defaultMessage: '目标租户',
                })}
                name="restore_tenant_name"
                initialValue={strategyData?.restoreTenantInfoParam?.name}
                validateFirst
                rules={[
                  {
                    required: true,
                    message: formatMessage({
                      id: 'ocp-v2.Backup.Component.RestoreTenantCard.EnterTheTenantName',
                      defaultMessage: '请输入租户名',
                    }),
                  },

                  NAME_RULE,
                ]}
              >
                <MyInput />
              </Form.Item>
            </Col>
          </Row>

          {loading ? (
            <PageLoading
              style={{
                paddingBottom: 100,
              }}
            />
          ) : (
            <>
              {clusterData.zones?.length > 0 && (
                <Row gutter={8}>
                  <Col span={12}>
                    <Form.Item
                      label={formatMessage({
                        id: 'ocp-v2.Backup.Component.RestoreTenantCard.NumberOfUnitsInEachZone',
                        defaultMessage: '各 Zone 的 Unit 数量',
                      })}
                      extra={formatMessage(
                        {
                          id: 'ocp-v2.Backup.Component.RestoreTenantCard.TheMaximumNumberOfCurrentKesheIsMinservercount',
                          defaultMessage:
                            '当前可设最大个数为 {minServerCount} (Zone 中最少 OBServer 数决定 Unit 可设最大个数)',
                        },
                        { minServerCount: minServerCount }
                      )}
                      name="zonesUnitCount"
                      initialValue={defaultZonesUnitCount}
                      rules={[
                        {
                          required: true,
                          message: formatMessage({
                            id: 'ocp-v2.component.FormZoneReplicaTable.UnitQuantityCannotBeEmpty',
                            defaultMessage: '请输入 Unit 数量',
                          }),
                        },
                      ]}
                    >
                      <InputNumber
                        min={1}
                        max={minServerCount}
                        onChange={value => {
                          const currentZones = getFieldValue('zones');
                          setFieldsValue({
                            zones: currentZones?.map(item => ({
                              ...item,
                              resourcePool: {
                                ...item?.resourcePool,
                                unitCount: value,
                              },
                            })),
                          });
                        }}
                        style={{ width: '30%' }}
                        placeholder={formatMessage({
                          id: 'ocp-v2.Tenant.New.Enter',
                          defaultMessage: '请输入',
                        })}
                      />
                    </Form.Item>
                  </Col>
                </Row>
              )}
              <ZonesFormItem
                name="zones"
                zones={clusterData?.zones}
                minServerCount={minServerCount}
                clusterUnitSpecLimit={clusterUnitSpecLimit}
              />
              <Form.Item noStyle={true} shouldUpdate={true}>
                {() => {
                  // 需要设置默认的 defaultZones，否则无法获取初始化的 zones 值，导致 primaryZone 的初始化展示出现问题
                  const currentZones = getFieldValue('zones') || defaultZones;

                  return (
                    <Form.Item
                      label={
                        <ContentWithQuestion
                          content={formatMessage({
                            id: 'ocp-v2.Backup.Component.RestoreTenantCard.ZonePriority',
                            defaultMessage: 'Zone 优先级',
                          })}
                          tooltip={{
                            title: formatMessage({
                              id: 'ocp-v2.Backup.Component.RestoreTenantCard.PriorityOfPartitionMasterReplica',
                              defaultMessage: '租户中主副本分布的优先级',
                            }),
                          }}
                        />
                      }
                      name="primaryZone"
                      initialValue={strategyData?.restoreTenantInfoParam?.primaryZone}
                      extra={formatMessage({
                        id: 'ocp-v2.Backup.Component.RestoreTenantCard.IfNoZonePriorityIs',
                        defaultMessage: '未配置任何 Zone 优先级，则为 random 优先级',
                      })}
                    >
                      <FormPrimaryZone
                        loading={loading}
                        zoneList={currentZones
                          // 仅勾选的 zone 副本可以设置优先级
                          .filter(item => item.checked)
                          .map(item => item.name)}
                      />
                    </Form.Item>
                  );
                }}
              </Form.Item>
            </>
          )}
        </Form>
      </Card>
    );
  }
);

export default RestoreTenantCard;
