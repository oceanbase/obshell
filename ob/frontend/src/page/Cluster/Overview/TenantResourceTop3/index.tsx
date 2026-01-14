import { formatMessage } from '@/util/intl';
import React from 'react';
import { max } from 'lodash';
import { toPercent } from '@oceanbase/charts/es/util/number';
import { Col, Descriptions, Empty, Row, theme } from '@oceanbase/design';
import MyCard from '@/component/MyCard';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import { formatSize } from '@/util';
import { isNullValue } from '@oceanbase/util';
import MyProgress from '@/component/MyProgress';
import MouseTooltip from '@/component/MouseTooltip';

export interface TenantResourceTop3Props {
  clusterData: API.ClusterInfo;
}

const TenantResourceTop3: React.FC<TenantResourceTop3Props> = ({ clusterData }) => {
  const { token } = theme.useToken();

  const tenantStats = clusterData?.tenant_stats || [];

  const tenantMetricList = [
    {
      key: 'ob_cpu_percent',
      title: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.CpuUsage',
        defaultMessage: 'CPU 消耗比',
      }),
      description: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.PercentageOfCpuUsedToTenantCpu',
        defaultMessage: '已使用 CPU 占分配 CPU 的百分比',
      }),
      usedName: '已消耗',
      chartData: tenantStats
        .map(item => ({
          ...item,
          percentValue: item.cpu_used_percent,
        }))
        .sort((a, b) => (b?.percentValue || 0) - (a?.percentValue || 0))
        .slice(0, 3),
    },

    {
      key: 'ob_memory_percent',
      title: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.MemoryUsage',
        defaultMessage: '内存消耗比',
      }),
      description: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.UsedMemoryAsAPercentageOfTenantMemory',
        defaultMessage: '分配内存总量，以及已使用内存占分配内存的百分比',
      }),
      usedName: '已消耗',
      chartData: tenantStats
        .map(item => ({
          ...item,
          percentValue: item.memory_used_percent,
        }))
        .sort((a, b) => (b?.percentValue || 0) - (a?.percentValue || 0))
        .slice(0, 3),
    },

    {
      key: 'ob_tenant_disk_usage',
      title: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.DataVolume',
        defaultMessage: '数据量',
      }),
      description: formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.TenantDataSizeAndPercentageOfClusterDisk',
        defaultMessage: '租户数据量大小，以及占集群磁盘容量的百分比',
      }),
      usedName: '已使用',
      chartData: tenantStats
        .map(item => ({
          ...item,
          usedValue: item.data_disk_usage,
          percentValue:
            ((item?.data_disk_usage || 0) / (clusterData?.stats?.disk_in_bytes_total || 0)) * 100,
        }))
        .sort((a, b) => (b?.percentValue || 0) - (a?.percentValue || 0))
        .slice(0, 3),
    },
  ];

  const tenantCount = max(tenantMetricList.map(item => item.chartData?.length)) || 3;

  const getStrokeColor = (percent: number, textFlag: boolean = true) => {
    if (percent >= 90) {
      return token.colorError;
    } else if (percent >= 80) {
      return token.colorWarning;
    } else {
      return textFlag ? undefined : token.colorTextSecondary;
    }
  };

  const getAffix = (item: any, dataItem: any) => {
    let totalValue = '';
    let usedValue = '';
    // 数据量
    if (item.key === 'ob_tenant_disk_usage') {
      totalValue = formatSize(dataItem.usedValue) as string;
      usedValue = `${
        isNullValue(dataItem.percentValue) ? '-' : toPercent(dataItem.percentValue / 100, 1)
      }%`;
    }

    // 内存消耗比
    if (item?.key === 'ob_memory_percent') {
      totalValue = formatSize(dataItem.memory_in_bytes_total) as string;
      usedValue = `${
        isNullValue(dataItem.percentValue) ? '-' : toPercent(dataItem.percentValue / 100, 1)
      }%`;
    }

    // CPU 消耗比
    if (item?.key === 'ob_cpu_percent') {
      totalValue = `${dataItem?.cpu_core_total} C`;
      usedValue = `${
        isNullValue(dataItem.percentValue) ? '-' : toPercent(dataItem.percentValue / 100, 1)
      }%`;
    }

    return (
      <>
        <span style={{ marginRight: 2 }}>总：{totalValue}</span>｜{item.usedName}：
        <span style={{ color: getStrokeColor(Number(dataItem.percentValue)) }}>{usedValue}</span>
      </>
    );
  };

  return (
    <MyCard
      title={formatMessage({
        id: 'ocp-express.Component.TenantResourceTop3.TenantResourceUsageTop',
        defaultMessage: '租户资源使用 Top3',
      })}
      headStyle={{
        marginBottom: 16,
      }}
      style={{
        height: tenantCount < 3 ? 210 - (3 - tenantCount) * 32 : 210,
      }}
      bordered={false}
    >
      <Row gutter={48}>
        {tenantMetricList.map((item, index) => {
          return (
            <Col
              key={item.key}
              span={8}
              style={
                index !== tenantMetricList.length - 1
                  ? { borderRight: `1px solid ${token.colorBorderSecondary}` }
                  : {}
              }
            >
              <MyCard
                title={
                  <div
                    style={{
                      fontSize: 14,
                      fontWeight: 500,
                      fontFamily: 'PingFangSC',
                      color: token.colorText,
                    }}
                  >
                    <ContentWithQuestion
                      content={item.title}
                      tooltip={{
                        title: item.description,
                      }}
                    />
                  </div>
                }
                headStyle={{
                  marginBottom: 16,
                }}
                className="card-without-padding"
              >
                {item.chartData.length > 0 ? (
                  <Row gutter={[0, 8]}>
                    {item.chartData.map(dataItem => (
                      <Col key={dataItem.tenant_name} span={24}>
                        <MouseTooltip
                          style={{
                            padding: 12,
                          }}
                          overlay={
                            <Row gutter={[48, 12]}>
                              <Col>
                                <Descriptions
                                  size="small"
                                  colon={false}
                                  column={1}
                                  labelStyle={{
                                    fontSize: 12,
                                    width: 80,
                                  }}
                                  contentStyle={{
                                    fontSize: 12,
                                  }}
                                >
                                  <Descriptions.Item label="租户名">
                                    {dataItem.tenant_name}
                                  </Descriptions.Item>
                                  <Descriptions.Item
                                    label={`${
                                      item.key === 'ob_cpu_percent'
                                        ? 'CPU '
                                        : item.key === 'ob_memory_percent'
                                        ? '内存'
                                        : '数据'
                                    }总量`}
                                  >
                                    {item.key === 'ob_cpu_percent'
                                      ? `${dataItem.cpu_core_total} C`
                                      : item.key === 'ob_memory_percent'
                                      ? formatSize(dataItem.memory_in_bytes_total)
                                      : formatSize(dataItem.data_disk_usage)}
                                  </Descriptions.Item>
                                  <Descriptions.Item
                                    label="已消耗占比"
                                    contentStyle={{
                                      color: getStrokeColor(Number(dataItem.percentValue)),
                                    }}
                                  >
                                    {isNullValue(dataItem.percentValue)
                                      ? '-'
                                      : toPercent((dataItem.percentValue || 0) / 100, 1)}
                                    %
                                  </Descriptions.Item>
                                </Descriptions>
                              </Col>
                            </Row>
                          }
                        >
                          <div>
                            <MyProgress
                              percent={dataItem.percentValue}
                              showInfo={false}
                              prefix={dataItem.tenant_name}
                              prefixWidth={80}
                              prefixTooltip={false}
                              strokeColor={getStrokeColor(Number(dataItem.percentValue), false)}
                              affix={getAffix(item, dataItem)}
                              affixWidth={190}
                            />
                          </div>
                        </MouseTooltip>
                      </Col>
                    ))}
                  </Row>
                ) : (
                  <Empty
                    imageStyle={{
                      height: 72,
                    }}
                    description=""
                  />
                )}
              </MyCard>
            </Col>
          );
        })}
      </Row>
    </MyCard>
  );
};

export default TenantResourceTop3;
