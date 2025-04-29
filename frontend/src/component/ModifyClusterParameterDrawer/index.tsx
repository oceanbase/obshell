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

import MyDrawer from '@/component/MyDrawer';
import MyInput from '@/component/MyInput';
import MySelect from '@/component/MySelect';
import SelectAllAndClearRender from '@/component/SelectAllAndClearRender';
import {
  getDetailComponentByParameterValue,
  getSimpleComponentByClusterParameterValue,
} from '@/util/component';
import { formatMessage } from '@/util/intl';
import { useSelector, useDispatch } from 'umi';
import {
  Alert,
  Button,
  Col,
  Form,
  Row,
  Table,
  Tree,
  Typography,
  Descriptions,
  message,
  DrawerProps,
} from '@oceanbase/design';
import React, { useEffect, useState } from 'react';
import { flatten, groupBy, isEqual, unionWith, uniq, uniqBy } from 'lodash';
import { DeleteOutlined, PlusOutlined } from '@oceanbase/icons';
import { useRequest } from 'ahooks';
import { getTenantOverView } from '@/service/obshell/tenant';
import { obclusterSetParameters } from '@/service/obshell/obcluster';
import styles from './index.less';

const { Option, OptGroup } = MySelect;
const { Text } = Typography;

export interface ModifyClusterParameterDrawerProps extends DrawerProps {
  parameter?: API.ClusterParameter;
  onSuccess?: () => void;
}

interface obTenantParametersType {
  value?: number | string | boolean;
  tenantName?: string;
  tenantId?: number;
  isNewValue?: boolean;
}

interface TreeNode {
  key?: string;
  title?: string;
  children?: TreeNode[];
}

const ModifyClusterParameterDrawer: React.FC<ModifyClusterParameterDrawerProps> = ({
  parameter = {},
  onSuccess,
  ...restProps
}) => {
  // const { styles } = useStyles();
  const rangeList = [
    {
      name: formatMessage({
        id: 'ocp-express.component.ModifyClusterParameterDrawer.Cluster',
        defaultMessage: '集群',
      }),
      value: 'cluster',
    },

    {
      name: 'Zone',
      value: 'Zone',
    },

    {
      name: 'OBServer',
      value: 'OBServer',
    },
  ];

  const [form] = Form.useForm();
  const { getFieldValue, setFieldsValue, validateFields } = form;

  // 确认抽屉是否可见
  const [confirmVisible, setConfirmVisible] = useState(false);
  // 查看参数值抽屉是否可见
  const [checkVisible, setCheckVisible] = useState(false);
  // 选择全部租户
  const [allTenantSelectedStatus, setAllTenantSelectedStatus] = useState('allowed');
  // 判断是否有当前参数对应多个生效对象
  let hasSameScope = false;

  const { clusterData } = useSelector((state: DefaultRootState) => state.cluster);
  const dispatch = useDispatch();

  const getClusterData = () => {
    dispatch({
      type: 'cluster/getClusterData',
      payload: {},
    });
  };

  useEffect(() => {
    if (!clusterData?.name) {
      getClusterData();
    }
  }, []);

  const removeSelectDisabled = () => {
    setAllTenantSelectedStatus('allowed');
  };
  useEffect(() => {
    if (!restProps.open) {
      removeSelectDisabled();
    } else {
      // Form List 无法通过 preserve false 来清除
      setFieldsValue({ parameter: [{ key: 0 }] });
    }
  }, [restProps.open]);

  useEffect(() => {
    if (parameter?.scope === 'TENANT') {
      getTenantList({});
    }
  }, [parameter]);

  const { run: updateClusterParameter, loading } = useRequest(obclusterSetParameters, {
    manual: true,
    onSuccess: res => {
      if (res?.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.component.ModifyClusterParameterDrawer.ParameterValueModified',
            defaultMessage: '参数值修改成功',
          })
        );
        setConfirmVisible(false);
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  const handleModifyValue = () => {
    validateFields().then(values => {
      const { parameter: modifyParameter } = values;
      let otherParam = {};
      // 修改租户类型参数 整理接口数据
      if (parameter?.scope === 'TENANT') {
        otherParam = modifyParameter?.map(item => {
          if (item?.target?.filter(o => o.value === 'all')?.length !== 0) {
            const tenantNameList = uniq(
              parameter.tenant_value?.map(item => item.tenant_name) || []
            );
            return {
              name: parameter?.name,
              value: item.value,
              scope: parameter?.scope,
              // all_user_tenant 不包含 sys 租户，获取所有 tenantName 进行修改
              // all_user_tenant: true,
              tenants: tenantNameList,
            };
          } else {
            return {
              name: parameter?.name,
              value: item.value,
              scope: parameter?.scope,
              tenants: item.target.map(o => o.label),
            };
          }
        });
      } else {
        // 修改集群类型参数 整理接口数据
        otherParam = modifyParameter?.map(item => {
          // 含有 Zone 级别
          if (item.applyTo === 'Zone') {
            return {
              name: parameter?.name,
              value: item.value,
              scope: parameter?.scope,
              zones: item.target.map(o => o.label),
            };
          } else if (item?.applyTo === 'OBServer') {
            // 含有 OBServer 级别
            return {
              name: parameter?.name,
              value: item.value,
              scope: parameter?.scope,
              servers: item.target.map(o => o.label),
            };
          } else {
            // 集群级别
            return {
              name: parameter?.name,
              value: item.value,
              scope: parameter?.scope,
            };
          }
        });
      }

      updateClusterParameter({ params: otherParam });
    });
  };
  const serverValues = parameter?.ob_parameters;
  const zoneList = Object.entries(groupBy(serverValues, item => item.zone)).map(([key, value]) => {
    return {
      name: key,
      serverList: value.map(item => {
        return {
          ip: item.svr_ip,
          id: item.zone,
          port: item.svr_port,
        };
      }),
    };
  });

  // // 对集群下的 Zone 格式整理
  // const zoneList = (clusterData?.zones || []).map(item => ({
  //   label: item.name,
  //   value: item.name,
  //   serverList: (item.servers || []).map(server => ({
  //     ip: server.ip,
  //     id: server.id,
  //     port: server.port,
  //   })),
  // }));

  // 获取当前集群下的租户
  const { data, run: getTenantList } = useRequest(getTenantOverView, {
    manual: true,
    defaultParams: [
      {
        mode: 'MYSQL',
      },
    ],
  });

  const tenantList = data?.data?.contents || [];
  const editParameter = getFieldValue(['parameter']);

  if (parameter?.scope === 'CLUSTER') {
    const serverList: string[] = [];
    if (editParameter?.length > 0 && editParameter[0] !== undefined) {
      editParameter.map(item => {
        if (item?.applyTo === 'cluster') {
          zoneList.map(zone => zone.serverList.map(server => serverList.push(server.ip)));
        } else if (item?.applyTo === 'Zone') {
          zoneList.map(zone => {
            item?.target?.map(target => {
              if (target.label === zone.label) {
                zone.serverList.map(server => serverList.push(server.ip));
              }
            });
          });
        } else if (item?.applyTo === 'OBServer') {
          zoneList.map(zone => {
            item?.target?.map(target => {
              zone.serverList.map(server => {
                if (target.label?.split(':')[0] === server.ip) {
                  serverList.push(server.ip);
                }
              });
            });
          });
        }
      });
    }
    hasSameScope = uniq(serverList).length < serverList.length;
  }

  // 将整理数据为适合 Tree 渲染的数据
  const getParameterValue = () => {
    if (editParameter?.length > 0 && editParameter[0] !== undefined) {
      // 修改租户类型参数 数据整理
      if (parameter.scope === 'TENANT') {
        const obTenantParameters: obTenantParametersType[] = [];
        parameter?.tenant_value?.map(tenantValue => {
          editParameter.map(item => {
            // 如果选择了 全部租户
            if (item?.target?.filter(o => o.value === 'all')?.length !== 0) {
              obTenantParameters.push({
                tenantName: tenantValue.tenant_name,
                value: item.value,
                tenantId: tenantValue.tenant_id,
                isNewValue: true,
              });
            } else {
              item?.target?.map(target => {
                if (tenantValue.tenant_name === target.label) {
                  obTenantParameters.push({
                    tenantName: target.label,
                    value: item.value,
                    tenantId: tenantValue.tenant_id,
                    isNewValue: true,
                  });
                }
              });
            }
          });
        });

        // 对租户数据进行去重
        return uniqBy(
          unionWith(
            obTenantParameters,
            parameter?.tenant_value?.map(item => {
              return {
                tenantName: item.tenant_name,
                value: item.value,
                tenantId: item.tenant_id,
              };
            }),
            isEqual
          ),
          'tenantName'
        );
      } else {
        const serverValues = parameter?.ob_parameters;
        const zones = Object.entries(groupBy(serverValues, item => item.zone)).map(
          ([key, value]) => {
            return {
              name: key,
              servers: value,
            };
          }
        );

        // 修改参数值，取原参数值生成 treeData
        let parametersTreeData: TreeNode[] = [];
        // 收集所修改参数值 生效范围的 server
        parametersTreeData = [
          {
            title: 'cluster',
            key: '1',
            children: zones?.map((zone, index) => ({
              title: zone.name,
              key: `1-${index}`,
              children: zone?.servers?.map((server, key) => ({
                title:
                  !!server?.value && server?.value !== server?.svr_ip
                    ? `${server?.svr_ip}:${server?.svr_port}：${server?.value}`
                    : `${server?.svr_ip}`,
                key: `1-${index}-${key}`,
                // 用来判断 newValue
                originValue: server.value,
              })),
            })),
          },
        ];

        editParameter?.map(item => {
          // 修改参数值，生效范围包含集群
          if (item?.applyTo === 'cluster') {
            if (Array.isArray(parametersTreeData[0]?.children)) {
              parametersTreeData[0].children = parametersTreeData[0].children.map(child => {
                return {
                  ...child,
                  children: child.children?.map(server => {
                    return {
                      ...server,
                      title: `${server?.title?.split('：')[0]}：${item?.value}`,
                      isNewValue: server.originValue !== item.value,
                    };
                  }),
                };
              });
            }
          } else if (item?.applyTo === 'Zone') {
            item?.target?.forEach(targetZone => {
              parametersTreeData = parametersTreeData.map(param => ({
                title: 'cluster',
                key: '1',
                children: param?.children?.map((zone, index) => {
                  if (targetZone.label === zone?.title) {
                    return {
                      key: zone.key,
                      title: zone.title,
                      children: zone?.children?.map((server, key) => {
                        return {
                          title: `${server?.title?.split('：')[0]}：${item?.value}`,
                          key: `1-${index}-${key}`,
                          isNewValue: server?.originValue !== item?.value,
                        };
                      }),
                    };
                  } else {
                    return zone;
                  }
                }),
              }));
            });
          } else if (item?.applyTo === 'OBServer') {
            item?.target?.map(targetZone => {
              parametersTreeData = parametersTreeData.map(param => ({
                title: 'cluster',
                key: '1',
                children: param?.children?.map(zone => ({
                  title: zone.title,
                  key: zone.key,
                  children: zone?.children?.map(server => {
                    if (
                      targetZone?.label?.split('：')[0] === server?.title?.split('：')[0] ||
                      `${targetZone?.label?.split('：')[0]}:${targetZone?.port}` ===
                        server?.title?.split('：')[0]
                    ) {
                      return {
                        key: server.key,
                        title: `${server?.title?.split('：')[0]}：${item?.value}`,
                        isNewValue: server?.title?.split('：')[1] !== item?.value,
                      };
                    } else {
                      return server;
                    }
                  }),
                })),
              }));
            });
          }
        });
        return parametersTreeData;
      }
    }
  };

  const confirmColumns = [
    {
      title: formatMessage({
        id: 'ocp-express.component.ModifyClusterParameterDrawer.ValueBeforeChange',
        defaultMessage: '值-变更前',
      }),

      width: '50%',
      dataIndex: 'oldValue',
      render: () => getDetailComponentByParameterValue(parameter),
    },

    {
      title: formatMessage({
        id: 'ocp-express.component.ModifyClusterParameterDrawer.ValueAfterChange',
        defaultMessage: '值-变更后',
      }),

      width: '50%',
      dataIndex: 'newValue',
      render: text => {
        if (parameter?.scope === 'TENANT') {
          return text
            ?.sort((a, b) => {
              const nameA = a?.tenantName?.toUpperCase();
              const nameB = b?.tenantName?.toUpperCase();
              if (nameA < nameB) {
                return -1;
              }
              if (nameA > nameB) {
                return 1;
              }
              return 0;
            })
            ?.map(item => (
              <div className={item?.isNewValue ? styles.newValue : styles.tenantName}>
                {`${item?.tenantName}: ${item?.value}`}
              </div>
            ));
        }
        return (
          <Tree
            defaultExpandAll={true}
            treeData={text}
            titleRender={node => {
              // 改动的值标记颜色
              if (node?.isNewValue) {
                return <div className={styles.treeNewValue}>{node.title}</div>;
              }
              return <div>{node.title}</div>;
            }}
          />
        );
      },
    },
  ];

  const getSelectAllOptions = (applyTo: string) => {
    if (applyTo === 'OBServer') {
      const selectAllOptions = [];
      zoneList?.map(zone => {
        zone?.serverList.map(server =>
          selectAllOptions.push({
            value: server.id,
            label: `${server.ip}:${server.port}`,
            port: server.port,
          })
        );
      });
      return selectAllOptions;
    } else {
      return zoneList.map(zone => ({ value: zone.name, label: zone.name }));
    }
  };

  const filterTenantList = (tenantData, text, tenants) => {
    const selectedtenantList: number[] = flatten(
      (tenants && tenants?.map(item => item?.target?.map(o => o?.value) || [])) || []
    );

    return tenantData.filter(
      item =>
        !selectedtenantList
          .filter(tenant => !(text || []).includes(tenant))
          .includes(item.tenant_id)
    );
  };

  return (
    <MyDrawer
      className={styles.drawer}
      width={966}
      title={formatMessage({
        id: 'ocp-express.component.ModifyClusterParameterDrawer.ModifyParameterValues',
        defaultMessage: '修改参数值',
      })}
      onOk={() => {
        validateFields().then(() => {
          setConfirmVisible(true);
        });
      }}
      destroyOnClose
      {...restProps}
    >
      {parameter?.scope === 'OB_CLUSTER_PARAMETER' && (
        <Alert
          message={formatMessage({
            id: 'ocp-express.component.ModifyClusterParameterDrawer.RelationshipBetweenValueAndEffective',
            defaultMessage: '值和生效范围的关系',
          })}
          description={
            <>
              <div>
                {formatMessage({
                  id: 'ocp-express.component.ModifyClusterParameterDrawer.WhenTheEffectiveRangeIs',
                  defaultMessage:
                    '·生效范围为集群时，表示集群当前 OBServer 和此集群后续添加的 OBServer 都使用此值',
                })}
              </div>
              <div>
                {formatMessage({
                  id: 'ocp-express.component.ModifyClusterParameterDrawer.WhenTheEffectiveRangeIs.1',
                  defaultMessage:
                    '·生效范围为 Zone 时，表示所选择 Zone 的当前 OBServer 和此 Zone 后续添加的 OBServer\n                都使用此值',
                })}
              </div>
              <div>
                {formatMessage({
                  id: 'ocp-express.component.ModifyClusterParameterDrawer.WhenTheEffectiveRangeIs.2',
                  defaultMessage: '·生效范围为 OBServer 时，表示所选择的 OBServer 使用此值',
                })}
              </div>
            </>
          }
          type="info"
          showIcon={true}
          style={{ marginBottom: 24 }}
        />
      )}

      <Descriptions column={3}>
        <Descriptions.Item
          span={2}
          label={formatMessage({
            id: 'ocp-express.component.ModifyClusterParameterDrawer.Parameter',
            defaultMessage: '参数名',
          })}
        >
          {parameter.name}
        </Descriptions.Item>
        <Descriptions.Item
          span={1}
          label={formatMessage({
            id: 'ocp-express.component.ModifyClusterParameterDrawer.CurrentValue',
            defaultMessage: '当前值',
          })}
          className="descriptions-item-with-ellipsis"
        >
          {parameter.values && (
            <Text ellipsis={{ tooltip: true, placement: 'topRight ' }}>
              {getSimpleComponentByClusterParameterValue(parameter, parameter?.scope, () => {
                setCheckVisible(true);
              })}
            </Text>
          )}
        </Descriptions.Item>
      </Descriptions>
      <Form form={form} layout="vertical" requiredMark={false} preserve={false}>
        <Form.List name="parameter" initialValue={[{ key: 0 }]}>
          {(fields, { add, remove: removeItem }) => {
            return (
              <>
                {fields.map((field, index) => (
                  <div key={field.key}>
                    <Row gutter={[8, 4]}>
                      <Col span={9}>
                        <Form.Item
                          {...field}
                          label={
                            index === 0 &&
                            formatMessage({
                              id: 'ocp-express.component.ModifyClusterParameterDrawer.Value',
                              defaultMessage: '值',
                            })
                          }
                          name={[field.name, 'value']}
                          fieldKey={[field.fieldKey, 'value']}
                          rules={[
                            {
                              required: true,
                              message: formatMessage({
                                id: 'ocp-express.component.ModifyClusterParameterDrawer.EnterAValue',
                                defaultMessage: '请输入值',
                              }),
                            },
                          ]}
                        >
                          <MyInput
                            placeholder={formatMessage({
                              id: 'ocp-express.component.ModifyClusterParameterDrawer.Value',
                              defaultMessage: '值',
                            })}
                            style={{ width: '100%' }}
                            allowClear={true}
                          />
                        </Form.Item>
                      </Col>
                      {parameter?.scope === 'TENANT' ? (
                        <Col span={14}>
                          <Form.Item noStyle shouldUpdate={true}>
                            {() => {
                              const currentParameter = getFieldValue(['parameter']);
                              const params = currentParameter[index];
                              return (
                                <Form.Item
                                  {...field}
                                  label={
                                    index === 0 &&
                                    formatMessage({
                                      id: 'ocp-express.component.ModifyClusterParameterDrawer.EffectiveObject',
                                      defaultMessage: '生效对象',
                                    })
                                  }
                                  name={[field.name, 'target']}
                                  fieldKey={[field.fieldKey, 'target']}
                                  rules={[
                                    {
                                      required: true,
                                      message: formatMessage({
                                        id: 'ocp-express.component.ModifyClusterParameterDrawer.EnterAnEffectiveObject',
                                        defaultMessage: '请输入生效对象',
                                      }),
                                    },
                                  ]}
                                >
                                  <MySelect
                                    mode="tags"
                                    allowClear={true}
                                    labelInValue={true}
                                    placeholder={formatMessage({
                                      id: 'ocp-express.component.ModifyClusterParameterDrawer.SelectAnActiveObjectFirst',
                                      defaultMessage: '请先选择生效对象',
                                    })}
                                    onChange={val => {
                                      const values = val?.map(item => item.value);
                                      if (values.length === 0) {
                                        setAllTenantSelectedStatus('allowed');
                                      } else if (values[0] === 'all') {
                                        setAllTenantSelectedStatus('selected');
                                      } else {
                                        setAllTenantSelectedStatus('disabled');
                                      }
                                    }}
                                  >
                                    <Option
                                      value="all"
                                      key="all"
                                      label={formatMessage({
                                        id: 'ocp-express.component.ModifyClusterParameterDrawer.AllTenants',
                                        defaultMessage: '全部租户',
                                      })}
                                      disabled={allTenantSelectedStatus === 'disabled'}
                                    >
                                      {formatMessage({
                                        id: 'ocp-express.component.ModifyClusterParameterDrawer.AllTenants',
                                        defaultMessage: '全部租户',
                                      })}
                                    </Option>
                                    {
                                      // 租户类型参数，租户只可赋值一次
                                      filterTenantList(
                                        tenantList,
                                        params?.target?.map(item => item?.value),
                                        currentParameter
                                      )
                                        ?.sort((a, b) => {
                                          const nameA = a?.name?.toUpperCase();
                                          const nameB = b?.name?.toUpperCase();
                                          if (nameA < nameB) {
                                            return -1;
                                          }
                                          if (nameA > nameB) {
                                            return 1;
                                          }
                                          return 0;
                                        })
                                        ?.map(item => {
                                          return (
                                            <Option
                                              value={item.tenant_id}
                                              key={item.tenant_id}
                                              label={item.tenant_name}
                                              disabled={allTenantSelectedStatus === 'selected'}
                                            >
                                              {item.tenant_name}
                                            </Option>
                                          );
                                        })
                                    }
                                  </MySelect>
                                </Form.Item>
                              );
                            }}
                          </Form.Item>
                        </Col>
                      ) : (
                        <>
                          <Col span={4}>
                            <Form.Item
                              {...field}
                              label={
                                index === 0 &&
                                formatMessage({
                                  id: 'ocp-express.component.ModifyClusterParameterDrawer.EffectiveRange',
                                  defaultMessage: '生效范围',
                                })
                              }
                              name={[field.name, 'applyTo']}
                              fieldKey={[field.fieldKey, 'applyTo']}
                              rules={[
                                {
                                  required: true,
                                  message: formatMessage({
                                    id: 'ocp-express.component.ModifyClusterParameterDrawer.EnterAValidRange',
                                    defaultMessage: '请输入生效范围',
                                  }),
                                },
                              ]}
                            >
                              <MySelect
                                allowClear={true}
                                onChange={val => {
                                  const currentParameter = getFieldValue(['parameter']);
                                  setFieldsValue({
                                    parameter: currentParameter?.map((param, key) =>
                                      index === key
                                        ? {
                                            applyTo: val,
                                            value: param.value,
                                            target: [],
                                          }
                                        : param
                                    ),
                                  });
                                }}
                              >
                                {rangeList.map(item => {
                                  return (
                                    <Option value={item.value} key={item.value}>
                                      {item.name}
                                    </Option>
                                  );
                                })}
                              </MySelect>
                            </Form.Item>
                          </Col>
                          <Col span={10}>
                            <Form.Item noStyle shouldUpdate={true}>
                              {() => {
                                const currentParameter = getFieldValue(['parameter']);
                                const params = getFieldValue(['parameter'])[index];
                                return (
                                  <Form.Item
                                    {...field}
                                    label={
                                      index === 0 &&
                                      formatMessage({
                                        id: 'ocp-express.component.ModifyClusterParameterDrawer.EffectiveObject',
                                        defaultMessage: '生效对象',
                                      })
                                    }
                                    name={[field.name, 'target']}
                                    fieldKey={[field.fieldKey, 'target']}
                                    rules={[
                                      {
                                        required: params?.applyTo !== 'cluster',
                                        message: formatMessage({
                                          id: 'ocp-express.component.ModifyClusterParameterDrawer.EnterAnEffectiveObject',
                                          defaultMessage: '请输入生效对象',
                                        }),
                                      },
                                    ]}
                                  >
                                    <MySelect
                                      mode="multiple"
                                      allowClear={true}
                                      labelInValue={true}
                                      maxTagCount={params?.applyTo === 'OBServer' ? 2 : 3}
                                      disabled={!params?.applyTo || params?.applyTo === 'cluster'}
                                      placeholder={
                                        params?.applyTo !== 'cluster' &&
                                        formatMessage({
                                          id: 'ocp-express.component.ModifyClusterParameterDrawer.SelectAValidRange',
                                          defaultMessage: '请先选择生效范围',
                                        })
                                      }
                                      dropdownRender={menu => (
                                        <SelectAllAndClearRender
                                          menu={menu}
                                          onSelectAll={() => {
                                            setFieldsValue({
                                              parameter: currentParameter?.map((param, key) => {
                                                return index === key
                                                  ? {
                                                      applyTo: param?.applyTo,
                                                      value: param.value,
                                                      target: getSelectAllOptions(param?.applyTo),
                                                    }
                                                  : param;
                                              }),
                                            });
                                          }}
                                          onClearAll={() => {
                                            setFieldsValue({
                                              parameter: currentParameter?.map((param, key) =>
                                                index === key
                                                  ? {
                                                      applyTo: param?.applyTo,
                                                      value: param.value,
                                                      target: [],
                                                    }
                                                  : param
                                              ),
                                            });
                                          }}
                                        />
                                      )}
                                    >
                                      {params?.applyTo === 'Zone' &&
                                        zoneList.map(item => {
                                          return (
                                            <Option value={item.name} key={item.name}>
                                              {item.name}
                                            </Option>
                                          );
                                        })}
                                      {params?.applyTo === 'OBServer' &&
                                        zoneList?.map(zone => {
                                          return (
                                            <OptGroup label={zone.name}>
                                              {zone?.serverList.map(server => (
                                                <Option key={server.id} value={server.id}>
                                                  {`${server.ip}:${server.port}`}
                                                </Option>
                                              ))}
                                            </OptGroup>
                                          );
                                        })}
                                    </MySelect>
                                  </Form.Item>
                                );
                              }}
                            </Form.Item>
                          </Col>
                        </>
                      )}

                      <Col span={1}>
                        <Form.Item noStyle shouldUpdate={true}>
                          {() => {
                            const currentParameter = getFieldValue(['parameter']);
                            return currentParameter?.length > 1 ? (
                              <Form.Item label={index === 0 && ' '} style={{ marginBottom: 8 }}>
                                <DeleteOutlined
                                  style={{ color: 'rgba(0,0,0, .45)' }}
                                  onClick={() => {
                                    removeItem(field.name);
                                  }}
                                />
                              </Form.Item>
                            ) : null;
                          }}
                        </Form.Item>
                      </Col>
                    </Row>
                  </div>
                ))}

                <Row gutter={8}>
                  <Col span={23}>
                    <Form.Item style={{ marginBottom: 8 }}>
                      <Button
                        type="dashed"
                        // 修改租户级别参数，选择全部租户后，不可再新增一行
                        disabled={
                          parameter?.scope === 'OB_TENANT_PARAMETER' &&
                          allTenantSelectedStatus === 'selected'
                        }
                        onClick={() => {
                          validateFields().then(() => {
                            add();
                          });
                        }}
                        style={{ width: '100%' }}
                      >
                        <PlusOutlined />
                        {formatMessage({
                          id: 'ocp-express.component.ModifyClusterParameterDrawer.AddValue',
                          defaultMessage: '添加值',
                        })}
                      </Button>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            );
          }}
        </Form.List>
      </Form>
      <MyDrawer
        className={styles.confirmDrawer}
        width={688}
        title={formatMessage({
          id: 'ocp-express.component.ModifyClusterParameterDrawer.ConfirmValueModificationInformation',
          defaultMessage: '确认值修改信息',
        })}
        visible={confirmVisible}
        confirmLoading={loading}
        onCancel={() => {
          setConfirmVisible(false);
        }}
        onOk={handleModifyValue}
      >
        {parameter.edit_level === 'STATIC_EFFECTIVE' && (
          <Alert
            type="warning"
            message={formatMessage({
              id: 'ocp-express.component.ModifyClusterParameterDrawer.NoteThatTheCurrentParameter',
              defaultMessage: '请注意，当前参数修改需要重启 OBServer 才能生效。',
            })}
            showIcon={true}
            className={styles.restartAlert}
          />
        )}

        {hasSameScope && (
          <Alert
            type="warning"
            message={formatMessage({
              id: 'ocp-express.component.ModifyClusterParameterDrawer.CurrentlyMultipleValuesGouXuan',
              defaultMessage: '当前存在多个值勾选同一生效范围，系统将按最新修改项自动分配结果。',
            })}
            showIcon={true}
            className={styles.alert}
          />
        )}

        <Table
          bordered={true}
          columns={confirmColumns}
          dataSource={[
            {
              key: '1',
              // oldValue: valueToServers,
              newValue: getParameterValue(),
            },
          ]}
          rowKey={record => record.key}
          pagination={false}
          className="table-without-hover-style"
        />
      </MyDrawer>
      <MyDrawer
        className={styles.confirmDrawer}
        width={480}
        title={formatMessage(
          {
            id: 'ocp-express.component.ModifyClusterParameterDrawer.ViewParameternameParameterValues',
            defaultMessage: '查看 {parameterName} 参数值',
          },
          { parameterName: parameter.name }
        )}
        visible={checkVisible}
        footer={false}
        onCancel={() => {
          setCheckVisible(false);
        }}
      >
        {getDetailComponentByParameterValue(parameter)}
      </MyDrawer>
    </MyDrawer>
  );
};

export default ModifyClusterParameterDrawer;
