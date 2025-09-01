import { formatMessage } from '@/util/intl';
import MonitorDetail from '@/component/MonitorDetail';
import { useDeepCompareEffect } from 'ahooks';
import React, { useState } from 'react';

import { useSelector } from 'umi';
import { MonitorScope } from '@/page/Monitor';
import { isEmpty } from 'lodash';
import { getSystemExternalPrometheus } from '@/service/obshell/system';
import { useRequest } from 'ahooks';
import ContentWithReload from '@/component/ContentWithReload';
import { PageContainer } from '@ant-design/pro-components';
import useReload from '@/hook/useReload';
import MonitorConfig from '@/page/Monitor/MonitorConfig';
import { getFilterData } from '@/component/MonitorDetail/helper';

export interface MonitorProps {
  monitorScope?: MonitorScope; // 租户详情页的 monitorScope 为空, 一级菜单性能监控页面 monitorScope 为 tenant
}

const Monitor: React.FC<MonitorProps> = ({ monitorScope }) => {
  const [reloading, reload] = useReload(false);
  const { clusterData } = useSelector((state: DefaultRootState) => state.cluster);
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const [filterLabel, setFilterLabel] = useState<Monitor.LabelType[]>([
    {
      key: 'tenant_name',
      value: tenantData?.tenant_name || '',
    },
  ]);

  const {
    data: prometheusData,
    loading: prometheusLoading,
    refresh: refreshPrometheusData,
  } = useRequest(getSystemExternalPrometheus, {
    ready: !monitorScope, // 如果租户详情页，则请求 Prometheus 数据
    defaultParams: [{ HIDE_ERROR_MESSAGE: true }],
  });

  // 如果是一级菜单性能监控页面，则不请求 Prometheus 数据
  const isPrometheusConfigured =
    !!monitorScope ||
    (prometheusData?.successful && prometheusData.status === 200 && !!prometheusData?.data);

  // 当 tenantData.tenant_name 变化时，更新 filterLabel（针对租户详情页）
  useDeepCompareEffect(() => {
    if (tenantData?.tenant_name && !monitorScope) {
      setFilterLabel([
        {
          key: 'tenant_name',
          value: tenantData.tenant_name,
        },
      ]);
    }
  }, [tenantData?.tenant_name, monitorScope]);
  const [filterData, setFilterData] = useState<Monitor.FilterDataType>({
    tenantList: monitorScope ? [] : undefined,
    date: '',
  });

  useDeepCompareEffect(() => {
    if (isEmpty(clusterData.tenants)) return;
    const tenantOptions = clusterData?.tenants?.map((item: API.DbaObTenant) => {
      return {
        label: item.tenant_name!,
        value: item.tenant_name!,
      };
    });

    setFilterData({
      tenantList: monitorScope ? tenantOptions : undefined,
      ...getFilterData(clusterData as API.ClusterInfo),
    });

    if (!monitorScope) return;

    // 当存在 monitorScope 且 tenantList 有数据时，默认选中第一条租户
    if (tenantOptions && tenantOptions.length > 0) {
      const firstTenant = tenantOptions[0];
      setFilterLabel([
        {
          key: 'tenant_name',
          value: firstTenant.value,
        },
      ]);
    }
  }, [clusterData, monitorScope]);

  const renderMonitor = () => {
    return !isPrometheusConfigured && !prometheusLoading ? (
      <MonitorConfig
        targetPath={`/cluster/tenant/${tenantData?.tenant_name}/monitor`}
        onConfigSuccess={() => {
          // 配置成功后刷新 Prometheus 状态
          refreshPrometheusData();
        }}
      />
    ) : (
      <MonitorDetail
        filterData={filterData}
        setFilterData={setFilterData}
        filterLabel={filterLabel}
        setFilterLabel={setFilterLabel}
        queryScope="OBTENANT"
        groupLabels={['tenant_name']}
        useFor="tenant"
      />
    );
  };

  return monitorScope ? (
    renderMonitor()
  ) : (
    <PageContainer
      loading={reloading || prometheusLoading}
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'OBShell.Detail.Monitor.PerformanceMonitoring',
              defaultMessage: '性能监控',
            })}
            spin={reloading || prometheusLoading}
            onClick={() => {
              reload();
            }}
          />
        ),
      }}
    >
      {renderMonitor()}
    </PageContainer>
  );
};

export default Monitor;
