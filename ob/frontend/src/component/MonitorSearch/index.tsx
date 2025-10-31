import { formatMessage } from '@/util/intl';
import TenantSelect from '@/component/common/TenantSelect';
import MySelect from '@/component/MySelect';
import { DATE_TIME_FORMAT } from '@/constant/datetime';
import type { MonitorScope } from '@/page/Monitor';
import { history } from 'umi';
import React, { useEffect, useMemo, useRef } from 'react';
import { groupBy, isEqual, omitBy, toNumber } from 'lodash';
import moment from 'moment';
import { Card, Form } from '@oceanbase/design';
import { isNullValue, toBoolean } from '@oceanbase/util';
import { useGetState, useUpdate } from 'ahooks';
import AutoFresh, { FREQUENCY_TYPE } from './AutoFresh';
import styles from './index.less';
const FormItem = Form.Item;
const { Option } = MySelect;

export interface MonitorSearchQuery {
  tab?: string;
  isRealtime?: string | boolean;
  frequency?: string | 'off';
  startTime?: string;
  endTime?: string;
  clusterId?: string;
  tenantId?: string;
  zoneName?: string;
  serverIp?: string;
  dimention?: string;
}

export interface MonitorSearchQueryData {
  tab?: string;
  isRealtime?: boolean;
  frequency?: FREQUENCY_TYPE;
  startTime?: string;
  endTime?: string;
  clusterId?: number;
  tenantId?: number;
  zoneName?: string;
  serverIp?: string;
  dimention?: string;
}

export interface MonitorServer {
  ip: string;
  zoneName: string;
  hostId: number;
}

interface MonitorSearchProps {
  location?: {
    pathname?: string;
    query?: MonitorSearchQuery;
  };
  zoneNameList?: string[];
  serverList?: MonitorServer[];
  onSearch?: (queryData: MonitorSearchQueryData) => void;
  monitorScope?: MonitorScope;
}

function getQueryData(query: MonitorSearchQuery): MonitorSearchQueryData {
  const {
    tab,
    isRealtime,
    frequency,
    startTime,
    endTime,
    clusterId,
    tenantId,
    zoneName,
    serverIp,
    dimention,
  } = query;
  const queryData = {
    tab,
    isRealtime: toBoolean(isRealtime),
    frequency: frequency === 'off' ? frequency : frequency && toNumber(frequency),
    startTime,
    endTime,
    clusterId: clusterId ? toNumber(clusterId) : undefined,
    tenantId: tenantId ? toNumber(tenantId) : undefined,
    zoneName,
    serverIp,
    dimention,
  };
  return queryData;
}

const MonitorSearch: React.FC<MonitorSearchProps> = ({
  location: { query = {}, pathname } = {},
  zoneNameList,
  serverList,
  onSearch,
  monitorScope,
}) => {
  const [form] = Form.useForm();
  const { validateFields, setFieldsValue, getFieldsValue } = form;
  const update = useUpdate();
  const currentMoment = useRef(moment());
  const queryData = useMemo(() => {
    return getQueryData(query);
  }, [query]);

  const defaultFrequency = queryData.frequency || 'off';

  // 关于 React hooks 获取不到最新值的问题: https://segmentfault.com/a/1190000039054862
  // 为了保证 handleSearch 中获取的 tenantList 是最新的，使用 ahooks 的 useGetState
  // ref: https://ahooks.js.org/zh-CN/hooks/use-get-state
  const [_, setTenantList, getTenantList] = useGetState<API.Tenant[]>([]);

  const queryRange = useMemo(() => {
    return queryData.startTime && queryData.endTime
      ? [moment(queryData.startTime), moment(queryData.endTime)]
      : undefined;
  }, [queryData.startTime, queryData.endTime]);

  // 默认查看一小时的区间
  const defaultRange = queryRange || [moment().subtract(1, 'hours'), moment()];
  const defaultIsRealtime = queryData.isRealtime || false;

  useEffect(() => {
    // 实时模式下才设置range的值
    if (defaultIsRealtime) {
      setFieldsValue({
        range: defaultRange,
      });
    }
  }, [defaultIsRealtime, defaultRange]);

  const handleSearch = (mergeDefault?: boolean) => {
    validateFields().then(values => {
      const {
        range = defaultRange,
        isRealtime = defaultIsRealtime,
        frequency = defaultFrequency,
        clusterId = queryData.clusterId,
        tenantId = queryData.tenantId,
        zoneName = mergeDefault ? queryData.zoneName : undefined,
        serverIp = mergeDefault ? queryData.serverIp : undefined,
        ...restValues
      } = values;

      const tenant = getTenantList()?.find(item => item?.id === tenantId);

      const newQueryData = omitBy(
        {
          ...restValues,
          tab: queryData.tab,
          frequency,
          isRealtime,
          clusterId: isNullValue(tenant?.clusterId) ? clusterId : tenant?.clusterId,
          tenantId,
          zoneName,
          serverIp,
          // 实时模式下起止时间设置为近 1 小时
          startTime: isRealtime
            ? moment().subtract(1, 'hour').format(DATE_TIME_FORMAT)
            : range[0]?.format(DATE_TIME_FORMAT),
          endTime: isRealtime
            ? moment().format(DATE_TIME_FORMAT)
            : range[1]?.format(DATE_TIME_FORMAT),
        },
        value => isNullValue(value)
      );

      // pathname 可能为空，此时查询条件不会与 URL 进行同步
      if (pathname) {
        // 使用 history.replace 解决多次查询后，使用浏览器后退按钮不会返回前一页面
        history.replace({
          pathname,
          query: newQueryData,
        });
      }
      if (onSearch) {
        onSearch(newQueryData);
      }
    });
  };

  // 组件挂载时触发查询
  useEffect(() => {
    setTimeout(() => {
      handleSearch(true);
    }, 0);
  }, []);

  useEffect(() => {
    // 如果 query 为 {} 时，触发查询，避免 query={} 时图表数据为空
    // issue: https://aone.alipay.com/issue/38184175
    // TODO: 长期方案需要将图表中的 getData 改造为 useRequest，这样不满足请求条件时，不会用空数组兜底返回
    if (isEqual(query, {})) {
      setTimeout(() => {
        handleSearch(true);
      }, 0);
    }
  }, [query]);

  const { isRealtime = defaultIsRealtime, zoneName } = getFieldsValue();

  const realServerList = serverList?.filter(
    // 根据选中的 zone 对 server 进行筛选
    item => !zoneName || item.zoneName === zoneName
  );

  return (
    <Card className={styles.container} bodyStyle={{ padding: '16px 24px 0 24px' }} bordered={false}>
      <Form
        layout="inline"
        form={form}
        onValuesChange={() => {
          handleSearch();
        }}
      >
        <div className={styles.searchForm}>
          <AutoFresh
            onRefresh={handleSearch}
            isRealtime={isRealtime}
            queryRange={queryRange}
            frequency={defaultFrequency}
            currentMoment={currentMoment.current}
            style={{ marginRight: 25 }}
            onMenuClick={key => {
              setFieldsValue?.({
                frequency: key,
                isRealtime: key !== 'off',
              });
            }}
            onRefreshClick={() => {
              setFieldsValue?.({
                range: [moment().subtract(1, 'hours'), moment()],
              });
            }}
          />

          <div className={styles.lastForm}>
            {monitorScope === 'tenant' && (
              <FormItem
                name="tenantId"
                label={formatMessage({
                  id: 'OBShell.component.MonitorSearch.Tenant',
                  defaultMessage: '租户',
                })}
                initialValue={queryData.tenantId}
              >
                <TenantSelect
                  variant="borderless"
                  onChange={() => {
                    setFieldsValue({
                      zoneName: undefined,
                      serverIp: undefined,
                    });
                  }}
                  onSuccess={newTenantList => {
                    setTenantList(newTenantList);
                    // 当前未选中租户，则默认选中第一个租户，并重新触发查询
                    if (!queryData.clusterId && !queryData.tenantId) {
                      // 根据集群 ID 分组后的租户数据
                      const groupData = groupBy(newTenantList, item => item.clusterId);
                      const firstKey = Object.keys(groupData)?.[0];
                      const firstTenant = groupData[firstKey]?.[0];
                      setFieldsValue({
                        tenantId: firstTenant?.id,
                      });
                      setTimeout(() => {
                        handleSearch();
                      }, 0);
                    }
                  }}
                />
              </FormItem>
            )}

            <FormItem name="zoneName" label="Zone" initialValue={queryData.zoneName}>
              <MySelect
                variant="borderless"
                allowClear={true}
                showSearch={true}
                onChange={() => {
                  setFieldsValue({
                    serverIp: undefined,
                  });
                }}
                popupMatchSelectWidth={false}
                placeholder={formatMessage({
                  id: 'OBShell.component.MonitorSearch.All',
                  defaultMessage: '全部',
                })}
              >
                {zoneNameList?.map(item => (
                  <Option key={item} value={item}>
                    {item}
                  </Option>
                ))}
              </MySelect>
            </FormItem>
            <FormItem name="serverIp" initialValue={queryData.serverIp} label="OBServer">
              <MySelect
                allowClear={true}
                showSearch={true}
                variant="borderless"
                optionLabelProp="optionLabel"
                dropdownMatchSelectWidth={false}
                placeholder={formatMessage({
                  id: 'OBShell.component.MonitorSearch.All',
                  defaultMessage: '全部',
                })}
                style={{ width: '100%' }}
                dropdownClassName="select-dropdown-with-description"
              >
                {realServerList?.map(item => (
                  <Option key={item.ip} value={item.ip} optionLabel={item.ip}>
                    <span>{item.ip}</span>
                    <span>{item.zoneName}</span>
                  </Option>
                ))}
              </MySelect>
            </FormItem>
          </div>
        </div>
        <Form.Item name="isRealtime" className={styles.hiddenForm}></Form.Item>
        <Form.Item name="frequency" className={styles.hiddenForm}></Form.Item>
      </Form>
    </Card>
  );
};

MonitorSearch.getQueryData = getQueryData;

export default MonitorSearch;
