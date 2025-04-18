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
import { history, useSelector } from 'umi';
import React, { useState } from 'react';
import { PageContainer } from '@oceanbase/ui';
import { uniq } from 'lodash';
import { Card, Table, Tooltip } from '@oceanbase/design';
import type { TablePaginationConfig } from '@oceanbase/design/es/table';
import { getBooleanLabel, isEnglish } from '@/util';
import {
  breadcrumbItemRender,
  getDetailComponentByParameterValue,
  getSimpleComponentByClusterParameterValue,
} from '@/util/component';
import { useRequest } from 'ahooks';
import ModifyClusterParameterDrawer from '@/component/ModifyClusterParameterDrawer';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import ContentWithReload from '@/component/ContentWithReload';
import MyDrawer from '@/component/MyDrawer';
import MyInput from '@/component/MyInput';
import { BOOLEAN_LIST } from '@/constant';
import { obclusterParameters } from '@/service/obshell/obcluster';

export interface ListProps {}

const List: React.FC<ListProps> = ({}) => {
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: 1,
    pageSize: 10,
  });

  // 列表搜索关键字
  const [keyword, setKeyword] = useState('');

  // 修改集群级参数值的抽屉是否可见
  const [visible, setVisible] = useState(false);
  // 查看参数值的抽屉是否可见
  const [valueVisible, setValueVisible] = useState(false);
  // 当前编辑的参数项
  const [currentRecord, setCurrentRecord] = useState<API.ClusterParameter | null>(null);

  const { data, loading, refresh } = useRequest(obclusterParameters, {
    defaultParams: [],

    onSuccess: res => {
      if (res.successful) {
        setPagination({
          ...pagination,
          current: 1,
        });
      }
    },
  });

  const parameterList = data?.data?.contents || [];
  const sectionList = uniq(parameterList.map(item => item.section).filter(item => item));

  const columns = [
    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.ParameterName',
        defaultMessage: '参数名称',
      }),

      dataIndex: 'name',
      ellipsis: true,
      width: '25%',
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          <span>{text}</span>
        </Tooltip>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.Category',
        defaultMessage: '类别',
      }),
      dataIndex: 'section',
      filters: sectionList.map(item => ({
        text: item,
        value: item,
      })),

      onFilter: (value: string, record: API.ClusterParameter) => record.section === value,
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.CurrentValue',
        defaultMessage: '当前值',
      }),

      dataIndex: 'currentValue',
      ellipsis: true,
      filters: [
        {
          text: formatMessage({
            id: 'ocp-express.Detail.Parameter.List.ClusterScope',
            defaultMessage: '集群范围',
          }),

          value: true,
        },

        {
          text: formatMessage({
            id: 'ocp-express.Detail.Parameter.List.TenantScope',
            defaultMessage: '租户范围',
          }),
          value: 'tenant',
        },

        {
          text: formatMessage({
            id: 'ocp-express.Detail.Parameter.List.CustomRange',
            defaultMessage: '自定义范围',
          }),

          value: false,
        },
      ],

      onFilter: (value, record: API.ClusterParameter) => {
        if (value === 'tenant') {
          return record.scope === 'TENANT';
        } else {
          return record.scope && record.is_single_value === value;
        }
      },
      render: (text: API.ClusterParameterValue, record: API.ClusterParameter) => (
        <Tooltip
          placement="topLeft"
          title={getSimpleComponentByClusterParameterValue(record, record?.scope, () => {}, 'text')}
        >
          <span>
            {getSimpleComponentByClusterParameterValue(record, record?.scope, () => {
              setValueVisible(true);
              setCurrentRecord(record);
            })}
          </span>
        </Tooltip>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.Description',
        defaultMessage: '说明',
      }),

      dataIndex: 'info',
      ellipsis: true,
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          <span>{text || '-'}</span>
        </Tooltip>
      ),
    },

    {
      title: (
        <ContentWithQuestion
          content={formatMessage({
            id: 'ocp-express.Detail.Parameter.List.WhetherTheRestartTakesEffect',
            defaultMessage: '是否重启生效',
          })}
          tooltip={{
            title: formatMessage({
              id: 'ocp-express.Detail.Parameter.List.WhetherTheParameterTakesEffect',
              defaultMessage: '参数是否在 OBServer 重启之后才能生效',
            }),
          }}
        />
      ),

      filters: BOOLEAN_LIST.map(item => ({
        text: item.label,
        value: item.value,
      })),

      onFilter: (value: boolean, record: API.ClusterParameter) => {
        const isReload = record?.edit_level === 'STATIC_EFFECTIVE';
        return value ? isReload : !isReload;
      },
      width: isEnglish() ? 240 : undefined,
      dataIndex: 'edit_level',
      render: text => <span>{getBooleanLabel(text === 'STATIC_EFFECTIVE')}</span>,
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

      dataIndex: 'readonly',
      width: isEnglish() ? 140 : 100,
      filters: BOOLEAN_LIST.map(item => ({
        text: item.label,
        value: item.value,
      })),

      onFilter: (value: string, record: API.ClusterParameter) => {
        const isReadonly = record?.edit_level === 'READONLY';
        return value ? isReadonly : !isReadonly;
      },
      render: (text, record: API.ClusterParameter) => (
        <span>{getBooleanLabel(record?.edit_level === 'READONLY')}</span>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-express.Detail.Parameter.List.Operation',
        defaultMessage: '操作',
      }),
      width: isEnglish() ? 130 : 100,
      dataIndex: 'operation',
      render: (text: string, record: API.ClusterParameter) => {
        if (!record.readonly) {
          return (
            <a
              data-aspm-click="c304242.d308727"
              data-aspm-desc="集群参数列表-修改参数值"
              data-aspm-param={``}
              data-aspm-expo
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
          );
        }
      },
    },
  ];

  const currentParameterName = currentRecord && currentRecord.name;

  const routes = [
    {
      path: '/overview',
      breadcrumbName: formatMessage({
        id: 'ocp-express.Cluster.Parameter.ClusterOverview',
        defaultMessage: '集群总览',
      }),
    },

    {
      breadcrumbName: formatMessage({
        id: 'ocp-express.Detail.Parameter.ParameterManagement',
        defaultMessage: '参数管理',
      }),
    },
  ];

  return (
    <PageContainer
      ghost={true}
      header={{
        breadcrumb: { routes, itemRender: breadcrumbItemRender },
        title: (
          <ContentWithReload
            spin={loading}
            content={formatMessage({
              id: 'ocp-express.Detail.Parameter.ParameterManagement',
              defaultMessage: '参数管理',
            })}
            onClick={() => {
              refresh();
            }}
          />
        ),

        onBack: () => {
          history.push('/overview');
        },
        extra: (
          <MyInput.Search
            data-aspm-click="c304242.d308724"
            data-aspm-desc="集群参数列表-搜索参数"
            data-aspm-param={``}
            data-aspm-expo
            onChange={e => {
              setKeyword(e.target.value);
              setPagination({
                current: 1,
                pageSize: 10,
              });
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
          data-aspm="c304242"
          data-aspm-desc="集群参数列表"
          data-aspm-param={``}
          data-aspm-expo
          loading={loading}
          dataSource={parameterList
            .sort((a, b) => {
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
            .filter(item => !keyword || (item && item.name.includes(keyword)))}
          columns={columns}
          rowKey={record => record.name}
          onChange={newPagination => {
            setPagination(newPagination);
          }}
          pagination={pagination}
        />

        <MyDrawer
          width={480}
          visible={valueVisible}
          title={formatMessage(
            {
              id: 'ocp-express.Detail.Parameter.List.ViewTheValueOfCurrentparametername',
              defaultMessage: '查看 {currentParameterName} 的值',
            },

            { currentParameterName }
          )}
          onCancel={() => {
            setValueVisible(false);
            setCurrentRecord(null);
          }}
          footer={false}
          bodyStyle={{ fontSize: 14, lineHeight: '22px' }}
        >
          {getDetailComponentByParameterValue(currentRecord)}
        </MyDrawer>
        <ModifyClusterParameterDrawer
          parameter={currentRecord || {}}
          visible={visible}
          onCancel={() => {
            setVisible(false);
            setCurrentRecord(null);
          }}
          onSuccess={() => {
            setVisible(false);
            setCurrentRecord(null);
            refresh();
          }}
        />
      </Card>
    </PageContainer>
  );
};

export default List;
