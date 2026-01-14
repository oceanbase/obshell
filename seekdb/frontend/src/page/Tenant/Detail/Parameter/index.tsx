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
import React, { useMemo, useState } from 'react';
import {
  Button,
  Card,
  Input,
  Space,
  Table,
  Tooltip,
  Popconfirm,
  Form,
  Modal,
  message,
} from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import { getParameters, setParameters, getVariables, setVariables } from '@/service/obshell/seekdb';
import { isEnglish } from '@/util';
import { PAGINATION_OPTION_10 } from '@/constant';
import ContentWithReload from '@/component/ContentWithReload';

const FormItem = Form.Item;

const PARAMETER = 'parameter';
const VARIABLE = 'variable';

export interface IndexProps {}

const Index: React.FC<IndexProps> = ({}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;
  const [keyword, setKeyword] = useState('');
  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState(undefined);

  const {
    data: parametersData,
    refresh: getParametersRefresh,
    loading: getParametersLoading,
  } = useRequest(getParameters, {
    manual: false,
    defaultParams: [{}],
  });

  const {
    data: variablesData,
    refresh: getVariablesRefresh,
    loading: getVariablesLoading,
  } = useRequest(getVariables, {
    manual: false,
    defaultParams: [{}],
  });

  const refresh = () => {
    getParametersRefresh();
    getVariablesRefresh();
  };

  const loading = getParametersLoading || getVariablesLoading;

  const parameterList = useMemo(() => {
    return [
      ...(parametersData?.data?.contents?.map(item => ({
        ...item,
        itemType: PARAMETER,
        read_only: item?.edit_level === 'READONLY',
      })) || []),
      ...(variablesData?.data?.contents?.map(item => ({
        ...item,
        itemType: VARIABLE,
      })) || []),
    ].sort((a, b) => {
      const aStartsWithUnderscore = a.name?.startsWith('_');
      const bStartsWithUnderscore = b.name?.startsWith('_');

      // 如果一个以_开头，另一个不以_开头，将_开头的放后面
      if (aStartsWithUnderscore && !bStartsWithUnderscore) return 1;
      if (!aStartsWithUnderscore && bStartsWithUnderscore) return -1;

      // 否则按字母排序
      return (a.name || '').localeCompare(b.name || '');
    });
  }, [parametersData?.data?.contents, variablesData?.data?.contents]);

  const { run: modifyTenantParameter, loading: modifyTenantParameterLoading } = useRequest(
    setParameters,
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
    setVariables,
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
        modifyTenantParameter({
          parameters: {
            [name]: value,
          },
        });
      } else if (itemType === VARIABLE) {
        modifyTenantVariable({
          variables: {
            [name]: value,
          },
        });
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

  return (
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
