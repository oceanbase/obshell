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

import ContentWithQuestion from '@/component/ContentWithQuestion';
import ContentWithReload from '@/component/ContentWithReload';
import Empty from '@/component/Empty';
import { BOOLEAN_LIST, PAGINATION_OPTION_10 } from '@/constant';
import {
  getTenantParameters,
  getTenantVariables,
  tenantSetParameters,
  tenantSetVariable,
} from '@/service/obshell/tenant';
import { getBooleanLabel, isEnglish } from '@/util';
import { formatMessage } from '@/util/intl';
import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Space,
  Table,
  Tooltip,
  message,
} from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { history, useParams, useSelector } from '@umijs/max';
import { useRequest } from 'ahooks';
import React, { useState } from 'react';

const FormItem = Form.Item;

const PARAMETER = 'parameter';
const VARIABLE = 'variable';

const Index: React.FC = () => {
  const { tenantName = '' } = useParams<{ tenantName: string }>();
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

  function getComponentByValueRange() {
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
          history.push(`/cluster/tenant/${tenantName}`);
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
