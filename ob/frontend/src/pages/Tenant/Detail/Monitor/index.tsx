import MonitorDetail from '@/component/MonitorDetail';
import { formatMessage } from '@/util/intl';
import { useDeepCompareEffect } from 'ahooks';
import React, { useEffect, useState } from 'react';

import ContentWithReload from '@/component/ContentWithReload';
import { getFilterData } from '@/component/MonitorDetail/helper';
import useReload from '@/hook/useReload';
import { MonitorScope } from '@/pages/Monitor';
import MonitorConfig from '@/pages/Monitor/MonitorConfig';
import { getSystemExternalPrometheus } from '@/service/obshell/system';
import { PageContainer } from '@ant-design/pro-components';
import { useSelector } from '@umijs/max';
import { useRequest } from 'ahooks';
import { isEmpty } from 'lodash';

export interface MonitorProps {
  monitorScope?: MonitorScope; // 租户详情页的 monitorScope 为空, 一级菜单性能监控页面 monitorScope 为 tenant
}

const Monitor: React.FC<MonitorProps> = ({ monitorScope }) => {
  const [reloading, reload] = useReload(false);
  const { clusterData } = useSelector((state: DefaultRootState) => state.cluster);
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const [filterLabel, setFilterLabel] = useState<Monitor.LabelType[]>([
    {
      key: 'ob_cluster_name',
      value: clusterData.cluster_name || '',
    },
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
  useEffect(() => {
    if (!monitorScope && clusterData.cluster_name && tenantData?.tenant_name) {
      setFilterLabel([
        {
          key: 'ob_cluster_name',
          value: clusterData.cluster_name,
        },
        {
          key: 'tenant_name',
          value: tenantData.tenant_name,
        },
      ]);
    }
  }, [clusterData.cluster_name, tenantData?.tenant_name, monitorScope]);

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
          key: 'ob_cluster_name',
          value: clusterData.cluster_name || '',
        },
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
