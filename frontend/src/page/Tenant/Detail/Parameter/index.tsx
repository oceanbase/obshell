/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

import { formatMessage } from '@/util/intl';
import { history, useSelector } from 'umi';
import React, { useState } from 'react';
import {
  Button,
  Card,
  Input,
  Select,
  Space,
  Table,
  Tooltip,
  Popconfirm,
  Form,
  Modal,
  message,
} from '@oceanbase/design';
import { isNullValue } from '@oceanbase/util';
import { PageContainer } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import {
  getTenantParameters,
  tenantSetParameters,
  getTenantVariables,
  tenantSetVariable,
} from '@/service/obshell/tenant';
import { getBooleanLabel, isEnglish } from '@/util';
import { PAGINATION_OPTION_10, BOOLEAN_LIST } from '@/constant';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import ContentWithReload from '@/component/ContentWithReload';
import Empty from '@/component/Empty';

const FormItem = Form.Item;
const { Option } = Select;

const PARAMETER = 'parameter';
const VARIABLE = 'variable';

export interface IndexProps {
  match: {
    params: {
      tenantName: string;
    };
  };
  keyword?: string;
}

const Index: React.FC<IndexProps> = ({
  match: {
    params: { tenantName },
  },
}) => {
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);

  const [form] = Form.useForm();
  const { validateFields } = form;
  const [keyword, setKeyword] = useState('');
  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState(undefined);

  const {
    data: parametersData,
    refresh: getTenantParametersRefresh,
    loading: getTenantParametersLoading,
  } = useRequest(getTenantParameters, {
    manual: false,
    defaultParams: [
      {
        name: tenantName,
      },
    ],
  });

  const {
    data: variablesData,
    refresh: getTenantVariablesRefresh,
    loading: getTenantVariablesLoading,
  } = useRequest(getTenantVariables, {
    manual: false,
    defaultParams: [
      {
        name: tenantName,
      },
    ],
  });

  const refresh = () => {
    getTenantParametersRefresh();
    getTenantVariablesRefresh();
  };

  const loading = getTenantParametersLoading || getTenantVariablesLoading;

  const parameterList = [
    ...(parametersData?.data?.contents?.map(item => ({
      ...item,
      itemType: PARAMETER,
      read_only: item?.edit_level === 'READONLY',
    })) || []),
    ...(variablesData?.data?.contents?.map(item => ({
      ...item,
      itemType: VARIABLE,
    })) || []),
  ];

  const { run: modifyTenantParameter, loading: modifyTenantParameterLoading } = useRequest(
    tenantSetParameters,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'ocp-express.src.model.tenant.ParameterValueModifiedSuccessfully',
              defaultMessage: '参数值修改成功',
            })
          );

          setVisible(false);
          setCurrentRecord(undefined);
          refresh();
        }
      },
    }
  );

  const { run: modifyTenantVariable, loading: modifyTenantVariableLoading } = useRequest(
    tenantSetVariable,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'ocp-express.src.model.tenant.ParameterValueModifiedSuccessfully',
              defaultMessage: '参数值修改成功',
            })
          );

          setVisible(false);
          setCurrentRecord(undefined);
          refresh();
        }
      },
    }
  );

  function getComponentByValueRange(record) {
    // const { type, allowedValues, minValue, maxValue } = record || {};
    // const { data_type } = record || {};
    // 枚举值类型
    // if (type === 'ENUM' && allowedValues) {
    //   return (
    //     <Select
    //       showSearch={true}
    //       placeholder={formatMessage({
    //         id: 'ocp-express.src.component.MySelect.PleaseSelect',
    //         defaultMessage: '请选择',
    //       })}
    //     >
    //       {allowedValues.split(',').map((item: string) => (
    //         <Option key={item} value={item}>
    //           {item}
    //         </Option>
    //       ))}
    //     </Select>
    //   );
    // }
    // 数字类型
    // if (type === 'INT' || type === 'NUMERIC') {
    //   // 超出 JS 能处理的最大、最小整数范围，则使用 Input，按照字符串进行处理
    //   if (
    //     minValue > Number.MAX_SAFE_INTEGER ||
    //     minValue < Number.MIN_SAFE_INTEGER ||
    //     maxValue > Number.MAX_SAFE_INTEGER ||
    //     maxValue < Number.MIN_SAFE_INTEGER
    //   ) {
    //     return (
    //       <Input
    //         placeholder={formatMessage({
    //           id: 'ocp-express.src.component.MyInput.PleaseEnter',
    //           defaultMessage: '请输入',
    //         })}
    //       />
    //     );
    //   }
    //   return (
    //     <InputNumber
    //       {...(isNullValue(minValue) ? {} : { min: minValue })}
    //       // maxValue 为空时，会导致输入的值自动被清空，无法正确输入，因此根据 maxValue 是否为空判断是否设置 max 属性
    //       {...(isNullValue(maxValue) ? {} : { max: maxValue })}
    //       // 如果是 INT 整数类型，需要控制数字精度，保证输入为整数
    //       {...(type === 'INT' ? { precision: 0 } : {})}
    //       style={{ width: '100%' }}
    //     />
    //   );
    // }
    // 字符串类型
    return (
      <Input
        placeholder={formatMessage({
          id: 'ocp-express.src.component.MyInput.PleaseEnter',
          defaultMessage: '请输入',
        })}
      />
    );
  }

  function handleSubmit() {
    validateFields().then(values => {
      const { name, itemType } = currentRecord || {};
      const { value } = values;
      if (itemType === PARAMETER) {
        modifyTenantParameter(
          {
            name: tenantName,
          },
          {
            parameters: {
              [name]: value,
            },
          }
        );
      } else if (itemType === VARIABLE) {
        modifyTenantVariable(
          {
            name: tenantName,
          },
          {
            variables: {
              [name]: value,
            },
          }
        );
      }
    });
  }

  const columns = [
    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.ParameterName',
        defaultMessage: '参数名称',
      }),

      dataIndex: 'name',
      width: '25%',
      ellipsis: true,
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          <span>{text}</span>
        </Tooltip>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.CurrentValue',
        defaultMessage: '当前值',
      }),

      ellipsis: true,
      dataIndex: 'value',
      render: (text: string) =>
        text ? (
          <Tooltip placement="topLeft" title={text}>
            <span>{text}</span>
          </Tooltip>
        ) : (
          '-'
        ),
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.Description',
        defaultMessage: '说明',
      }),

      dataIndex: 'info',
      ellipsis: true,
      render: (text: string) =>
        text ? (
          <Tooltip placement="topLeft" title={text}>
            <span>{text}</span>
          </Tooltip>
        ) : (
          '-'
        ),
    },

    {
      title: (
        <ContentWithQuestion
          content={formatMessage({
            id: 'ocp-express.Detail.Parameter.List.ReadOnly',
            defaultMessage: '只读',
          })}
          tooltip={{
            title: formatMessage({
              id: 'ocp-express.Detail.Parameter.List.CannotBeModified',
              defaultMessage: '不可修改',
            }),
          }}
        />
      ),

      dataIndex: 'read_only',
      filters: BOOLEAN_LIST.map(item => ({
        text: item.label,
        value: item.value,
      })),

      onFilter: (value: string, record: API.ClusterParameter) => record?.readonly === value,
      render: (text: boolean) => <span>{getBooleanLabel(text)}</span>,
    },
    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.Operation',
        defaultMessage: '操作',
      }),
      dataIndex: 'operation',
      render: (text, record) => {
        // 重启生效
        const needRestart = record?.edit_level === 'STATIC_EFFECTIVE';

        return (
          !record.readonly && (
            <Tooltip
              placement="topRight"
              title={
                needRestart &&
                formatMessage({
                  id: 'ocp-express.Detail.Parameter.index.ToModifyThisParameterYou',
                  defaultMessage: '修改该参数需要重启集群，请在集群参数管理中进行修改',
                })
              }
            >
              <a
                data-aspm-click="c304241.d308726"
                data-aspm-desc="租户参数列表-修改参数值"
                data-aspm-param={``}
                data-aspm-expo
                disabled={record?.needRestart}
                onClick={() => {
                  setVisible(true);
                  setCurrentRecord(record);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.Detail.Parameter.List.ModifyTheValue',
                  defaultMessage: '修改值',
                })}
              </a>
            </Tooltip>
          )
        );
      },
    },
  ];

  const formItemLayout = {
    labelCol: {
      span: isEnglish() ? 6 : 3,
    },

    wrapperCol: {
      span: isEnglish() ? 18 : 21,
    },
  };

  const currentParameterName = currentRecord && currentRecord.name;

  return tenantData.status === 'CREATING' ? (
    <Empty
      title={formatMessage({
        id: 'ocp-express.Detail.Parameter.NoData',
        defaultMessage: '暂无数据',
      })}
      description={formatMessage({
        id: 'ocp-express.Detail.Parameter.TheTenantIsBeingCreated',
        defaultMessage: '租户正在创建中，请等待租户创建完成',
      })}
    >
      <Button
        type="primary"
        onClick={() => {
          history.push(`/tenant/${tenantId}`);
        }}
      >
        {formatMessage({
          id: 'ocp-express.Detail.Parameter.AccessTheOverviewPage',
          defaultMessage: '访问总览页',
        })}
      </Button>
    </Empty>
  ) : (
    <PageContainer
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'ocp-express.Detail.Parameter.ParameterManagement',
              defaultMessage: '参数管理',
            })}
            spin={loading}
            onClick={refresh}
          />
        ),

        extra: (
          <Input.Search
            data-aspm-click="c304241.d308722"
            data-aspm-desc="租户参数列表-搜索参数"
            data-aspm-param={``}
            data-aspm-expo
            onChange={e => {
              setKeyword(e.target.value);
            }}
            placeholder={formatMessage({
              id: 'ocp-express.Detail.Parameter.List.SearchParameterName',
              defaultMessage: '搜索参数名',
            })}
            className="search-input"
            allowClear={true}
          />
        ),
      }}
    >
      <Card bordered={false} className="card-without-padding">
        <Table
          data-aspm="c304241"
          data-aspm-desc="租户参数列表"
          data-aspm-param={``}
          data-aspm-expo
          loading={loading}
          dataSource={parameterList?.filter(
            item => !keyword || (item.name && item.name.includes(keyword))
          )}
          columns={columns}
          rowKey={record => record.name}
          pagination={PAGINATION_OPTION_10}
        />

        <Modal
          title={formatMessage({
            id: 'ocp-express.page.Property.ModifyTheValue',
            defaultMessage: '修改参数值',
          })}
          visible={visible}
          destroyOnClose={true}
          footer={
            <Space>
              <Button
                data-aspm-click="c318535.d343242"
                data-aspm-desc="修改租户参数值-取消"
                data-aspm-param={``}
                data-aspm-expo
                onClick={() => {
                  setVisible(false);
                  setCurrentRecord(undefined);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.component.MyDrawer.Cancel',
                  defaultMessage: '取消',
                })}
              </Button>
              <Popconfirm
                placement="topRight"
                title={formatMessage({
                  id: 'ocp-express.Detail.Parameter.AreYouSureYouWantToModifyThe',
                  defaultMessage: '确定修改当前值吗？',
                })}
                onConfirm={() => {
                  handleSubmit();
                }}
              >
                <Button
                  data-aspm-click="c318535.d343246"
                  data-aspm-desc="修改租户参数值-提交"
                  data-aspm-param={``}
                  data-aspm-expo
                  type="primary"
                  loading={modifyTenantParameterLoading || modifyTenantVariableLoading}
                >
                  {formatMessage({
                    id: 'ocp-express.Detail.Component.AddUserDrawer.Submitted',
                    defaultMessage: '提交',
                  })}
                </Button>
              </Popconfirm>
            </Space>
          }
          onCancel={() => {
            setVisible(false);
            setCurrentRecord(undefined);
          }}
        >
          <Form form={form} preserve={false} hideRequiredMark={true} {...formItemLayout}>
            <FormItem
              label={formatMessage({
                id: 'ocp-express.page.Property.ParameterName',
                defaultMessage: '参数名',
              })}
            >
              {currentParameterName}
            </FormItem>
          </Form>
          <Form
            layout="vertical"
            form={form}
            preserve={false}
            hideRequiredMark={true}
            {...formItemLayout}
          >
            <FormItem
              label={formatMessage({
                id: 'ocp-express.component.ModifyClusterParameterDrawer.Value',
                defaultMessage: '值',
              })}
              name="value"
              initialValue={currentRecord?.value}
              style={{ marginBottom: 0 }}
            >
              {getComponentByValueRange(currentRecord)}
            </FormItem>
          </Form>
        </Modal>
      </Card>
    </PageContainer>
  );
};

export default Index;
