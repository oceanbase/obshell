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
import { history } from 'umi';
import {
  Alert,
  Button,
  Checkbox,
  Col,
  Form,
  InputNumber,
  Row,
  Space,
  Spin,
  Card,
  Switch,
  Table,
  Tooltip,
  Typography,
  token,
} from '@oceanbase/design';
import { InfoCircleFilled } from '@oceanbase/icons';
import type { Route } from '@oceanbase/design/es/breadcrumb/Breadcrumb';
import React, { useEffect, useState } from 'react';
import { uniq, uniqueId, uniqWith } from 'lodash';
import { findBy, isNullValue } from '@oceanbase/util';
import { PageContainer } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import * as ObshellTenantController from '@/service/obshell/tenant';
import { NAME_RULE, PASSWORD_REGEX } from '@/constant';
import { REPLICA_TYPE_LIST } from '@/constant/oceanbase';
import { TENANT_MODE_LIST, LOWER_CASE_TABLE_NAMES } from '@/constant/tenant';
import { getTextLengthRule, validatePassword } from '@/util';
import { getUnitSpecLimit, getResourcesLimit } from '@/util/cluster';
import { getMinServerCount } from '@/util/tenant';
import { validateUnitCount } from '@/util/oceanbase';
import { breadcrumbItemRender } from '@/util/component';
import useDocumentTitle from '@/hook/useDocumentTitle';
import MyInput from '@/component/MyInput';
import MyCard from '@/component/MyCard';
import MySelect from '@/component/MySelect';
import Password from '@/component/Password';
import WhitelistInput from '@/component/WhitelistInput';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import FormPrimaryZone from '@/component/FormPrimaryZone';
import SetParameterEditableProTable from '@/component/ParameterTemplate/SetParameterEditableProTable';
import { getObclusterCharsets } from '@/service/obshell/ob';
import { tenantCreate } from '@/service/obshell/tenant';
import { unitConfigCreate } from '@/service/obshell/unit';
import { getUnitConfigLimit } from '@/service/obshell/obcluster';
import { message } from 'antd';
import { obclusterInfo } from '@/service/obshell/obcluster';
import { TIMEZONE_LIST, filterTimezonesByMode } from '@/constant/timezone';
import styles from './index.less';

const { Option } = MySelect;
const { Text } = Typography;

interface NewProps {}

const New: React.FC<NewProps> = ({}) => {
  const [form] = Form.useForm();
  const { getFieldValue, validateFields, setFieldsValue } = form;

  const [passed, setPassed] = useState(true);
  const [currentMode, setCurrentMode] = useState('MYSQL');
  const [collations, setCollations] = useState<API.Collation[]>([]);

  const [currentlowerCaseTableNames, setCurrentlowerCaseTableNames] = useState(1);

  const [scenarioDescription, setScenarioDescription] = useState();

  // 参数设置开关
  const [advanceSettingSwitch, setAdvanceSettingSwitch] = useState(false);
  const [compatibleOracleAlret, setCompatibleOracleAlret] = useState(false);

  // 推荐
  useDocumentTitle(
    formatMessage({
      id: 'ocp-express.Tenant.New.CreateATenant',
      defaultMessage: '新建租户',
    })
  );

  const {
    data: obclusterInfoRes,
    loading: clusterDataLoading,
    runAsync: getClusterData,
  } = useRequest(obclusterInfo, {
    manual: false,
  });

  const zones = obclusterInfoRes?.data?.zones || [];

  const clusterData = obclusterInfoRes?.data || {};
  const minServerCount = getMinServerCount(zones);

  // 获取 unit 规格的限制规则
  const { data: clusterUnitSpecLimitData } = useRequest(getUnitConfigLimit);

  const clusterUnitSpecLimit = {
    cpuLowerLimit: clusterUnitSpecLimitData?.data?.min_cpu || 0,
    memoryLowerLimit: clusterUnitSpecLimitData?.data?.min_memory || 0,
  };

  const { data: charsetListData, runAsync: listCharsets } = useRequest(getObclusterCharsets, {
    manual: false,
  });

  const charsetList = charsetListData?.data?.contents || [];

  // 查询集群租户负载类型
  const { run: getLoadTypeData, data: loadTypeParameterTemplate } = useRequest(
    ObshellTenantController.listSupportParameterTemplates,
    {
      manual: true,
      onSuccess: res => {
        const currentScenario = (res.data?.contents || []).find(item => item.scenario === 'HTAP');
        setScenarioDescription(currentScenario.description);
      },
    }
  );

  const loadTypeParameterList = loadTypeParameterTemplate?.data?.contents || [];

  useEffect(() => {
    getClusterData({}).then(() => {
      // 获取字符集列表后再设置 collation 列表，因为两者有先后依赖关系
      listCharsets({
        tenant_mode: 'MYSQL',
      }).then(charsetRes => {
        setFieldsValue({
          charset: 'utf8mb4',
        });
        handleCharsetChange('utf8mb4', charsetRes.data?.contents || []);
      });
      getLoadTypeData();
    });
  }, []);

  useEffect(() => {
    setFieldsValue({
      zones: (zones || []).map(item => ({
        checked: true,
        name: item.name,
        replicaType: 'FULL',
        unitCount: 1,
      })),

      mode: 'MYSQL',
      charset: 'utf8mb4',
      primaryZone: '',
    });
  }, [clusterData]);

  // 创建租户更改为异步
  const { run: addTenant, loading: createTenantLoading } = useRequest(tenantCreate, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-v2.Tenant.New.TenantCreationSucceeded',
            defaultMessage: '创建租户成功',
          })
        );
        const taskId = res?.data?.id;
        history.push(`/tenant/result/${taskId}`);
      }
    },
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const {
        zones,
        zonesUnitCount,
        primaryZone,
        rootPassword,
        lowerCaseTableNames,
        timeZone,
        parameters = [],
        ...restValues
      } = values;
      const realParameters = {};
      const zoneUnitMap = {};
      const checkZones = zones.filter(item => !!item.checked);

      parameters.forEach(parameter => {
        realParameters[parameter.name] = parameter.value;
      });

      uniqWith(checkZones, (a, b) => {
        return a.cpuCore === b.cpuCore && a.memorySize === b.memorySize;
      }).map(({ cpuCore, memorySize }) => {
        zoneUnitMap['' + cpuCore + memorySize] = {
          name: `tenant_unit_${Date.now()}_${uniqueId()}`,
          cpuCore: Number(cpuCore),
          memorySize: memorySize + 'GB',
        };
      });

      const zoneList = checkZones?.map(zone => {
        return {
          name: zone.name,
          replica_type: zone.replicaType,
          unit_config_name: zoneUnitMap['' + zone.cpuCore + zone.memorySize]?.name,
          unit_num: zone.unitCount,
        };
      });

      const variables = new Map([
        ['lower_case_table_names', lowerCaseTableNames],
        ['time_zone', timeZone],
      ]);

      Promise.all(
        Object.entries(zoneUnitMap).map(([key, item]) => {
          return unitConfigCreate({
            max_cpu: item.cpuCore,
            memory_size: item.memorySize,
            name: item.name,
          });
        })
      ).then(res => {
        if (!res.some(item => item.successful === false)) {
          addTenant({
            ...restValues,
            zone_list: zoneList,
            import_script: true,
            root_password: rootPassword,
            primary_zone: primaryZone,
            variables: Object.fromEntries(variables),
            parameters: realParameters,
          });
        }
      });
    });
  };

  const handleCharsetChange = (charset: string, newCharsetList: API.Charset[]) => {
    const newCollations = findBy(newCharsetList, 'name', charset)?.collations || [];
    const defaultCollation = findBy(newCollations, 'is_default', true);
    setCollations(newCollations);
    setFieldsValue({ collation: defaultCollation.name });
  };

  const replicaTypeTooltipConfig = {
    color: '#fff',
    overlayStyle: {
      maxWidth: 400,
    },
    title: formatMessage({
      id: 'ocp-v2.Tenant.New.FullFunctionalAndReadOnly',
      defaultMessage: '当前支持全功能型副本和只读型副本',
    }),
  };

  const routes: Route[] = [
    {
      path: '/tenant',
      breadcrumbName: formatMessage({
        id: 'ocp-express.Tenant.New.Tenant',
        defaultMessage: '租户',
      }),
    },

    {
      breadcrumbName: formatMessage({
        id: 'ocp-express.Tenant.New.CreateATenant',
        defaultMessage: '新建租户',
      }),
    },
  ];

  const columns = [
    {
      title: ' ',
      dataIndex: 'checked',
      width: 24,
    },
    {
      title: formatMessage({
        id: 'ocp-express.Tenant.New.ZoneName',
        defaultMessage: 'Zone 名称',
      }),
      dataIndex: 'name',
    },

    {
      title: (
        <ContentWithQuestion
          content={formatMessage({
            id: 'ocp-express.Tenant.New.ReplicaType',
            defaultMessage: '副本类型',
          })}
          tooltip={replicaTypeTooltipConfig}
        />
      ),

      dataIndex: 'replicaType',
    },

    {
      title: formatMessage({
        id: 'ocp-express.Tenant.New.UnitSpecifications',
        defaultMessage: 'Unit 规格',
      }),

      dataIndex: 'unitSpecName',
    },

    {
      title: formatMessage({
        id: 'ocp-express.Tenant.New.UnitQuantity',
        defaultMessage: 'Unit 数量',
      }),
      dataIndex: 'unitCount',
    },
  ];

  return (
    <PageContainer
      className={styles.container}
      ghost={true}
      header={{
        title: formatMessage({
          id: 'ocp-express.Tenant.New.NewTenant',
          defaultMessage: '新建租户',
        }),
        breadcrumb: { routes, itemRender: breadcrumbItemRender },
        onBack: () => {
          history.push('/tenant');
        },
      }}
      footer={[
        <Button
          key="cancel"
          onClick={() => {
            history.push('/tenant');
          }}
        >
          {formatMessage({
            id: 'ocp-express.Tenant.New.Cancel',
            defaultMessage: '取消',
          })}
        </Button>,
        <Button key="submit" type="primary" loading={createTenantLoading} onClick={handleSubmit}>
          {formatMessage({
            id: 'ocp-express.Tenant.New.Submitted',
            defaultMessage: '提交',
          })}
        </Button>,
      ]}
    >
      <Form
        form={form}
        layout="vertical"
        colon={false}
        hideRequiredMark={true}
        requiredMark="optional"
      >
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <MyCard
              title={formatMessage({
                id: 'ocp-express.Tenant.New.BasicInformation',
                defaultMessage: '基本信息',
              })}
              bordered={false}
            >
              <Row gutter={24}>
                <Col span={8}>
                  <Form.Item
                    label={formatMessage({
                      id: 'ocp-express.Tenant.New.TenantMode',
                      defaultMessage: '租户模式',
                    })}
                    name="mode"
                    initialValue="MYSQL"
                    rules={[
                      {
                        required: true,
                        message: formatMessage({
                          id: 'ocp-express.Tenant.New.SelectTenantMode',
                          defaultMessage: '请选择租户模式',
                        }),
                      },
                    ]}
                  >
                    <MySelect
                      onChange={value => {
                        setFieldsValue({
                          charset: 'utf8mb4',
                        });

                        setCurrentMode(value);
                        listCharsets({
                          tenant_mode: value,
                        }).then(charsetRes => {
                          handleCharsetChange('utf8mb4', charsetRes.data?.contents || []);
                        });
                      }}
                    >
                      {TENANT_MODE_LIST.filter(item =>
                        clusterData?.is_community_edition ? item.value !== 'ORACLE' : true
                      ).map(item => (
                        <Option key={item.value} value={item.value}>
                          {item.label}
                        </Option>
                      ))}
                    </MySelect>
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    label={formatMessage({
                      id: 'ocp-express.Tenant.New.TenantName',
                      defaultMessage: '租户名称',
                    })}
                    name="name"
                    rules={[
                      {
                        required: true,
                        message: formatMessage({
                          id: 'ocp-express.Tenant.New.EnterTheTenantName',
                          defaultMessage: '请输入租户名称',
                        }),
                      },

                      NAME_RULE,
                    ]}
                  >
                    <MyInput
                      placeholder={formatMessage({
                        id: 'ocp-express.Tenant.New.PleaseEnter',
                        defaultMessage: '请输入',
                      })}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={24}>
                <Col span={8}>
                  <Form.Item
                    label={formatMessage({
                      id: 'ocp-express.Tenant.New.AdministratorInitialPassword',
                      defaultMessage: '管理员初始密码',
                    })}
                    name="rootPassword"
                    rules={[
                      {
                        required: true,
                        message: formatMessage({
                          id: 'ocp-express.Tenant.New.PleaseEnterOrRandomlyGenerate',
                          defaultMessage: '请输入或随机生成管理员初始密码',
                        }),
                      },

                      {
                        validator: validatePassword(passed),
                      },
                    ]}
                  >
                    <Password generatePasswordRegex={PASSWORD_REGEX} onValidate={setPassed} />
                  </Form.Item>
                </Col>

                <Col span={8}>
                  <Form.Item
                    label={formatMessage({
                      id: 'ocp-express.Tenant.New.CharacterSet',
                      defaultMessage: '字符集',
                    })}
                    name="charset"
                    rules={[
                      {
                        required: true,
                        message: formatMessage({
                          id: 'ocp-express.Tenant.New.SelectACharacterSet',
                          defaultMessage: '请选择字符集',
                        }),
                      },
                    ]}
                  >
                    <MySelect
                      onChange={value => {
                        handleCharsetChange(value, charsetList);
                      }}
                      showSearch={true}
                      optionLabelProp="label"
                      dropdownClassName="select-dropdown-with-description"
                    >
                      {charsetList.map(item => (
                        <Option key={item.name} value={item.name} label={item.name}>
                          <span>{item.name}</span>
                        </Option>
                      ))}
                    </MySelect>
                  </Form.Item>
                </Col>
                {/* 如果是 oracle 模式，不展示 collation */}
                {currentMode !== 'ORACLE' && (
                  <Col span={8}>
                    <Form.Item label="Collation" name="collation">
                      <MySelect>
                        {collations?.map(item => (
                          <Option key={item.name} value={item.name} label={item.name}>
                            <span>{item.name}</span>
                          </Option>
                        ))}
                      </MySelect>
                    </Form.Item>
                  </Col>
                )}
              </Row>
              <Row gutter={24}>
                {currentMode === 'MYSQL' && (
                  <Col span={8}>
                    <Form.Item
                      label={formatMessage({
                        id: 'OBShell.Tenant.New.TableNamesAreCaseSensitive',
                        defaultMessage: '表名大小写敏感',
                      })}
                      name="lowerCaseTableNames"
                      initialValue={1}
                      rules={[
                        {
                          required: true,
                        },
                      ]}
                      tooltip={
                        <div>
                          <div>
                            {formatMessage({
                              id: 'OBShell.Tenant.New.DescriptionOfCaseSensitiveParameters',
                              defaultMessage: '大小写敏感参数（lower_case_table_names）说明：',
                            })}
                          </div>
                          <div>
                            {formatMessage({
                              id: 'OBShell.Tenant.New.UseCreateTableOrCreate',
                              defaultMessage:
                                '· 0：使用 CREATE TABLE 或 CREATE DATABASE\n                            语句指定的大小写字母在硬盘上保存表名和数据库名。名称比较对大小写敏感。',
                            })}
                          </div>
                          <div>
                            {formatMessage({
                              id: 'OBShell.Tenant.New.TableNamesAreSavedIn',
                              defaultMessage:
                                '· 1：表名在硬盘上以小写保存，名称比较对大小写不敏感。',
                            })}
                          </div>
                          <div>
                            {formatMessage({
                              id: 'OBShell.Tenant.New.TableNameAndDatabaseName',
                              defaultMessage:
                                '· 2：表名和数据库名在硬盘上使用 CREATE TABLE 或 CREATE DATABASE\n                            语句指定的大小写字母进行保存，但名称比较对大小写不敏感。',
                            })}
                          </div>
                        </div>
                      }
                      extra={
                        LOWER_CASE_TABLE_NAMES?.find(
                          item => item.value === currentlowerCaseTableNames
                        )?.description
                      }
                    >
                      <MySelect
                        onChange={val => {
                          setCurrentlowerCaseTableNames(val);
                        }}
                      >
                        {LOWER_CASE_TABLE_NAMES?.map(item => (
                          <Option key={item.value} value={item.value} label={item.label}>
                            <span>{item.label}</span>
                          </Option>
                        ))}
                      </MySelect>
                    </Form.Item>
                  </Col>
                )}
                <Col span={8}>
                  <Form.Item noStyle shouldUpdate>
                    {() => {
                      // 根据租户模式过滤时区列表
                      const REAL_TIMEZONE_LIST = filterTimezonesByMode(
                        TIMEZONE_LIST,
                        currentMode as 'MYSQL' | 'ORACLE'
                      );

                      return (
                        <Form.Item
                          label={formatMessage({
                            id: 'OBShell.Tenant.New.TimeZone',
                            defaultMessage: '时区',
                          })}
                          name="timeZone"
                          initialValue="+08:00"
                          rules={[
                            {
                              required: true,
                            },
                          ]}
                        >
                          <MySelect
                            showSearch={true}
                            optionLabelProp="label"
                            placeholder={formatMessage({
                              id: 'ocp-express.Tenant.New.SelectTimezone',
                              defaultMessage: '请选择时区',
                            })}
                          >
                            {REAL_TIMEZONE_LIST.map(item => (
                              <Option
                                key={item.value}
                                value={item.value}
                                label={`${item.gmt} ${item.label}`}
                              >
                                <div>
                                  <div>
                                    {item.gmt} {item.label}
                                  </div>
                                  <div style={{ fontSize: '12px', color: '#999' }}>
                                    {item.description}
                                  </div>
                                </div>
                              </Option>
                            ))}
                          </MySelect>
                        </Form.Item>
                      );
                    }}
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    label={formatMessage({
                      id: 'ocp-express.Tenant.New.Remarks',
                      defaultMessage: '备注',
                    })}
                    name="comment"
                    rules={[getTextLengthRule(0, 200)]}
                  >
                    <MyInput.TextArea
                      rows={3}
                      showCount={true}
                      maxLength={200}
                      placeholder={formatMessage({
                        id: 'ocp-express.Tenant.New.PleaseEnter',
                        defaultMessage: '请输入',
                      })}
                    />
                  </Form.Item>
                </Col>
              </Row>
              {loadTypeParameterList.length > 0 && (
                <Row>
                  <Col span={24}>
                    <Form.Item
                      label={
                        <Space>
                          <div style={{ width: 56 }}>
                            {formatMessage({
                              id: 'OBShell.Tenant.New.LoadType',
                              defaultMessage: '负载类型',
                            })}
                          </div>
                          <InfoCircleFilled style={{ color: '#006aff' }} />
                          <Text
                            style={{
                              width: 'calc(100vw - 450px)',
                              fontSize: '14px',
                              fontWeight: 400,
                              color: token.colorTextTertiary,
                            }}
                            ellipsis={{
                              tooltip: {
                                overlayInnerStyle: {
                                  background: '#fff',
                                  width: 484,
                                  color: token.colorTextTertiary,
                                },
                                color: '#fff',
                              },
                            }}
                          >
                            {formatMessage({
                              id: 'OBShell.Tenant.New.LoadTypeMainlyAffectsSql',
                              defaultMessage:
                                '负载类型主要影响 SQL\n                                类大查询判断时间（参数：large_query_threshold），对 OLTP 类型负载的\n                                RT 可能存在较大影响，请谨慎选择',
                            })}
                          </Text>
                        </Space>
                      }
                      name={'scenario'}
                      initialValue={'HTAP'}
                      extra={scenarioDescription}
                      style={{ marginBottom: 0 }}
                      rules={[
                        {
                          required: true,
                        },
                      ]}
                    >
                      <MySelect
                        style={{
                          width: '33%',
                        }}
                        onChange={val => {
                          const currentScenario = loadTypeParameterList.find(
                            item => item.scenario === val
                          );
                          setScenarioDescription(currentScenario?.description);
                        }}
                      >
                        {loadTypeParameterList?.map(item => (
                          <Option key={item.scenario} value={item.scenario}>
                            {item.scenario}
                          </Option>
                        ))}
                      </MySelect>
                    </Form.Item>
                  </Col>
                </Row>
              )}
            </MyCard>
          </Col>
          <Col span={24}>
            <MyCard
              title={formatMessage({
                id: 'ocp-express.Tenant.New.ZoneInformation',
                defaultMessage: 'Zone 信息',
              })}
              bordered={false}
            >
              <Spin spinning={false}>
                {zones?.length > 0 && (
                  <Row gutter={8}>
                    <Col span={12}>
                      <Form.Item
                        label={formatMessage({
                          id: 'ocp-express.Tenant.New.NumberOfUnitsInEachZone',
                          defaultMessage: '各 Zone 的 Unit 数量',
                        })}
                        extra={formatMessage(
                          {
                            id: 'ocp-express.Tenant.New.TheMaximumNumberOfCurrentKesheIsMinservercount',
                            defaultMessage:
                              '当前可设最大个数为 {minServerCount} (Zone 中最少 OBServer 数决定 Unit 可设最大个数)',
                          },
                          { minServerCount: minServerCount }
                        )}
                        name="zonesUnitCount"
                        initialValue={1}
                        rules={[
                          {
                            // 仅勾选行需要必填
                            required: true,
                            message: formatMessage({
                              id: 'ocp-express.component.FormZoneReplicaTable.UnitQuantityCannotBeEmpty',
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
                                unitCount: value,
                              })),
                            });
                          }}
                          style={{ width: '30%' }}
                          placeholder={formatMessage({
                            id: 'ocp-express.Tenant.New.Enter',
                            defaultMessage: '请输入',
                          })}
                        />
                      </Form.Item>
                    </Col>
                  </Row>
                )}

                {obclusterInfoRes?.data ? (
                  <Form.Item style={{ marginBottom: 24 }}>
                    <Form.List
                      name="zones"
                      rules={[
                        {
                          required: true,
                          message: formatMessage({
                            id: 'ocp-express.Tenant.New.PleaseSetACopy',
                            defaultMessage: '请设置副本',
                          }),
                        },
                      ]}
                    >
                      {(fields, {}) => {
                        return (
                          <>
                            {fields.map((field, index: number) => {
                              const currentChecked = getFieldValue([
                                'zones',
                                field.name,
                                'checked',
                              ]);

                              const currentZoneName = getFieldValue(['zones', field.name, 'name']);
                              const zoneData = findBy(zones || [], 'name', currentZoneName);
                              const isSingleServer = zoneData?.servers?.length === 1;

                              let idleCpuCore = 0;
                              let idleMemoryInBytes = 0;
                              if (isSingleServer) {
                                const { idleCpuCoreTotal, idleMemoryInBytesTotal } =
                                  getUnitSpecLimit(zoneData?.servers[0]?.stats);
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
                                      },
                                    ]}
                                    valuePropName="checked"
                                  >
                                    <Checkbox />
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
                                        name={[field.name, 'replicaType']}
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
                                      >
                                        <MySelect
                                          disabled={!currentChecked}
                                          optionLabelProp="label"
                                        >
                                          {REPLICA_TYPE_LIST.map(item => (
                                            <Option
                                              key={item.value}
                                              value={item.value}
                                              label={item.label}
                                            >
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
                                          isSingleServer &&
                                          (cpuLowerLimit < idleCpuCore!
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
                                              }))
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
                                          ...(isSingleServer
                                            ? [
                                                {
                                                  type: 'number' as const,
                                                  min: cpuLowerLimit,
                                                  max: idleCpuCore,
                                                  message:
                                                    cpuLowerLimit < idleCpuCore!
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
                                              ]
                                            : []),
                                        ]}
                                      >
                                        <InputNumber
                                          disabled={!currentChecked}
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
                                          isSingleServer &&
                                          (memoryLowerLimit < idleMemoryInBytes!
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
                                              }))
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
                                          ...(isSingleServer
                                            ? [
                                                {
                                                  type: 'number' as const,
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
                                              ]
                                            : []),
                                        ]}
                                      >
                                        <InputNumber
                                          disabled={!currentChecked}
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
                                              initialValue={1}
                                              name={[field.name, 'unitCount']}
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
                  </Form.Item>
                ) : (
                  <Table className={styles.table} dataSource={[]} columns={columns} />
                )}

                <Form.Item noStyle shouldUpdate={true}>
                  {() => {
                    const currentZones = getFieldValue('zones');
                    const checkedZones = (currentZones || []).filter(item => item.checked);
                    const uniqueCheckedUnitSpecNameList = uniq(
                      checkedZones
                        // 去掉空值
                        .filter(item => !!item.unitSpecName)
                        .map(item => item.unitSpecName)
                    );

                    const uniqueCheckedUnitCountList = uniq(
                      checkedZones
                        // 去掉空值
                        .filter(item => !isNullValue(item.unitCount))
                        .map(item => item.unitCount)
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
              </Spin>
              <Form.Item noStyle shouldUpdate={true}>
                {() => {
                  const currentZones = getFieldValue('zones');
                  return (
                    <Form.Item
                      style={{ marginBottom: 0 }}
                      label={formatMessage({
                        id: 'ocp-express.Tenant.New.ZonePrioritySorting',
                        defaultMessage: 'Zone 优先级排序',
                      })}
                      tooltip={{
                        title: formatMessage({
                          id: 'ocp-express.Tenant.New.PriorityOfTheDistributionOf',
                          defaultMessage: '租户中主副本分布的优先级',
                        }),
                      }}
                      // extra={formatMessage({
                      //   id: 'ocp-express.Tenant.New.IfThisParameterIsEmpty',
                      //   defaultMessage: '如果为空，则默认继承 sys 租户的 Zone 优先级',
                      // })}
                      name="primaryZone"
                      rules={[
                        {
                          required: true,
                          message: formatMessage({
                            id: 'ocp-v2.Tenant.New.PleaseSelectZonePrioritySort',
                            defaultMessage: '请选择 Zone 优先级排序',
                          }),
                        },
                      ]}
                    >
                      <FormPrimaryZone
                        loading={clusterDataLoading}
                        zoneList={(currentZones || [])
                          // 仅勾选的 zone 副本可以设置优先级
                          .filter(item => item.checked)
                          .map(item => item.name)}
                      />
                    </Form.Item>
                  );
                }}
              </Form.Item>
            </MyCard>
          </Col>
          <Col span={24}>
            <MyCard
              title={formatMessage({
                id: 'ocp-express.Tenant.New.SecuritySettings',
                defaultMessage: '安全设置',
              })}
              bordered={false}
            >
              <Row>
                <Col span={10}>
                  <Form.Item
                    style={{ marginBottom: 0 }}
                    name="whitelist"
                    rules={[
                      {
                        validator: WhitelistInput.validate,
                      },
                    ]}
                  >
                    <WhitelistInput layout="vertical" />
                  </Form.Item>
                </Col>
              </Row>
            </MyCard>
          </Col>
          <Col span={24}>
            <Card
              className={advanceSettingSwitch ? '' : 'card-without-padding'}
              title={
                <Space>
                  {formatMessage({
                    id: 'ocp-express.Cluster.New.ParameterSettings',
                    defaultMessage: '参数设置',
                  })}

                  <Switch
                    checked={advanceSettingSwitch}
                    onChange={checked => {
                      setAdvanceSettingSwitch(checked);
                    }}
                  />
                </Space>
              }
              bordered={false}
            >
              {advanceSettingSwitch && (
                <>
                  {compatibleOracleAlret && currentMode === 'MYSQL' && (
                    <Alert
                      message={formatMessage({
                        id: 'ocp-express.Tenant.New.DeleteTheParametersMarkedWith',
                        defaultMessage:
                          '请删除红色边框标记的参数，这些参数仅适用于 Oracle 模式的租户，当前租户为MySQL',
                      })}
                      type="error"
                      showIcon={true}
                      style={{ marginBottom: 16 }}
                    />
                  )}

                  <Form.Item
                    className={styles.parameters}
                    label={
                      <div
                        style={{
                          width: '100%',
                          display: 'flex',
                          justifyContent: 'space-between',
                        }}
                      >
                        <span>
                          {formatMessage({
                            id: 'ocp-express.Tenant.New.ParameterModification',
                            defaultMessage: '参数修改',
                          })}
                        </span>
                        {/* 一期暂不开放此功能 */}
                        {/* <ParameterTemplateDropdown
                      type="tenant"
                      tenantMode={getFieldValue('mode')}
                      parameters={getFieldValue('parameters')}
                      onSetParameter={value => {
                      setFieldsValue({
                      parameters: value,
                      });
                      setInitParameters(value);
                      }}
                      /> */}
                      </div>
                    }
                    name="parameters"
                    rules={[
                      {
                        required: true,
                        message: formatMessage({
                          id: 'ocp-express.Tenant.New.SetStartupParameters',
                          defaultMessage: '请设置租户参数',
                        }),
                      },
                    ]}
                  >
                    <SetParameterEditableProTable
                      type="tenant"
                      drawerType="NEW_TENANT"
                      addButtonText={formatMessage({
                        id: 'ocp-express.Tenant.New.AddStartupParameters',
                        defaultMessage: '添加租户参数',
                      })}
                      // initialParameters={initParameters}
                      tenantMode={currentMode}
                      setCompatibleOracleAlret={val => {
                        setCompatibleOracleAlret(val);
                      }}
                    />
                  </Form.Item>
                </>
              )}
            </Card>
          </Col>
        </Row>
      </Form>
    </PageContainer>
  );
};
export default New;
