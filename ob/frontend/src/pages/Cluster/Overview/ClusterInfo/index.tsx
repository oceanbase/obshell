import ContentWithQuestion from '@/component/ContentWithQuestion';
import MouseTooltip from '@/component/MouseTooltip';
import MyCard from '@/component/MyCard';
import ObClusterDeployMode from '@/component/ObClusterDeployMode';
import { formatSize, isEnglish } from '@/util';
import { formatMessage } from '@/util/intl';
import { toPercent } from '@oceanbase/charts/es/util/number';
import {
  Col,
  Descriptions,
  Divider,
  Popover,
  Progress,
  Row,
  Space,
  theme,
  Tooltip,
  Typography,
} from '@oceanbase/design';
import { EllipsisOutlined } from '@oceanbase/icons';
import { isNullValue } from '@oceanbase/util';
import { useModel } from '@umijs/max';
import React from 'react';
import styles from './index.less';

const { Text } = Typography;

const Detail: React.FC = () => {
  const { token } = theme.useToken();
  const { clusterData, isSharedStorage } = useModel('cluster');
  const stats: API.BaseResourceStats = clusterData.stats || {};

  const metricList = [
    {
      key: 'cpu',
      title: 'CPU',
      description: formatMessage({
        id: 'OBShell.Overview.ClusterInfo.Allocated',
        defaultMessage: '已分配',
      }),
      percentValue: (stats.cpu_core_assigned_percent || 0).toFixed(1),
      totalValue: `${stats.cpu_core_total || 0} C`,
      assignedValue: `${stats.cpu_core_assigned || 0} C`,
      leftValue: `${(stats.cpu_core_total || 0) - (stats.cpu_core_assigned || 0)} C`,
    },
    {
      key: 'memory',
      title: formatMessage({
        id: 'ocp-express.Component.ClusterInfo.Memory',
        defaultMessage: '内存',
      }),
      description: formatMessage({
        id: 'OBShell.Overview.ClusterInfo.Allocated',
        defaultMessage: '已分配',
      }),
      percentValue: (stats.memory_assigned_percent || 0).toFixed(1),
      totalValue: stats.memory_total || 0,
      assignedValue: formatSize(stats.memory_in_bytes_assigned || 0),
      leftValue: formatSize(
        (stats.memory_in_bytes_total || 0) - (stats.memory_in_bytes_assigned || 0)
      ),
    },
    {
      key: 'disk',
      title: isSharedStorage
        ? formatMessage({
            id: 'OBShell.Overview.ClusterInfo.DataCaching',
            defaultMessage: '数据缓存',
          })
        : formatMessage({
            id: 'OBShell.Overview.ClusterInfo.DataDisk',
            defaultMessage: '数据盘',
          }),
      description: isSharedStorage
        ? formatMessage({ id: 'OBShell.Overview.ClusterInfo.Allocated', defaultMessage: '已分配' })
        : formatMessage({
            id: 'OBShell.Overview.ClusterInfo.Used',
            defaultMessage: '已使用',
          }),
      percentValue: (stats.disk_assigned_percent || 0).toFixed(1),
      totalValue: stats.disk_total || 0,
      assignedValue: formatSize(stats.disk_in_bytes_assigned || 0),
      leftValue: formatSize((stats.disk_in_bytes_total || 0) - (stats.disk_in_bytes_assigned || 0)),
    },
    ...(isSharedStorage
      ? [
          {
            key: 'shared_storage',
            title: formatMessage({
              id: 'OBShell.Overview.ClusterInfo.SharedStorage',
              defaultMessage: '共享存储',
            }),
            titleDescription: (
              <ContentWithQuestion
                content={formatMessage({
                  id: 'OBShell.Overview.ClusterInfo.SizeOfSharedStorageFootprint',
                  defaultMessage: '集群维度共享存储占用空间大小',
                })}
              />
            ),

            description: formatMessage({
              id: 'OBShell.Overview.ClusterInfo.Used',
              defaultMessage: '已使用',
            }),
            assignedValue: formatSize(stats.shared_storage_in_bytes_used || 0),
          },
        ]
      : []),
  ];

  const getStrokeColor = (percent: number, textFlag: boolean = true) => {
    if (percent >= 90) {
      return token.colorError;
    } else if (percent >= 80) {
      return token.colorWarning;
    } else {
      return textFlag ? undefined : token.colorSuccess;
    }
  };

  return (
    <MyCard
      title={formatMessage({
        id: 'OBShell.Overview.ClusterInfo.ResourceWaterLevel',
        defaultMessage: '资源水位',
      })}
      extra={
        <Popover
          placement="bottomRight"
          arrowPointAtCenter={true}
          overlayStyle={{
            maxWidth: 200,
          }}
          content={
            <Descriptions colon={false} column={1}>
              <Descriptions.Item
                label={formatMessage({
                  id: 'ocp-express.Component.ClusterInfo.ClusterName',
                  defaultMessage: '集群名称',
                })}
                className="descriptions-item-with-ellipsis"
              >
                <Text
                  ellipsis={{
                    tooltip: clusterData.cluster_name,
                  }}
                >
                  {clusterData.cluster_name}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'ocp-express.Detail.Overview.DeploymentMode',
                  defaultMessage: '部署模式',
                })}
                style={{
                  paddingBottom: 0,
                }}
              >
                {(clusterData.zones?.length || 0) > 0 ? (
                  <Tooltip title={<ObClusterDeployMode clusterData={clusterData} />}>
                    <ObClusterDeployMode clusterData={clusterData} mode="text" />
                  </Tooltip>
                ) : (
                  '-'
                )}
              </Descriptions.Item>
            </Descriptions>
          }
        >
          <EllipsisOutlined
            className="pointable"
            style={{
              fontSize: 20,
            }}
          />
        </Popover>
      }
    >
      <Row gutter={48}>
        {metricList.map((item, index) => (
          <Col
            key={item.key}
            span={isSharedStorage ? 6 : 8}
            style={
              index !== metricList.length - 1
                ? { borderRight: `1px solid ${token.colorBorderSecondary}` }
                : {}
            }
          >
            <MouseTooltip
              style={{
                maxWidth: isEnglish() ? 300 : 200,
                padding: 12,
              }}
              overlay={
                <Row gutter={[48, 12]}>
                  <Col key={item.key}>
                    <div
                      style={{
                        paddingLeft: 4,
                        fontSize: 14,
                        fontWeight: 500,
                        marginBottom: 8,
                        color: token.colorText,
                      }}
                    >
                      {item.title}
                    </div>
                    <Descriptions
                      size="small"
                      colon={false}
                      column={1}
                      labelStyle={{
                        width: isEnglish() ? 90 : 60,
                        textAlign: 'justify',
                        textAlignLast: 'justify',
                        fontSize: 12,
                      }}
                      contentStyle={{ fontSize: 12 }}
                    >
                      {item.key !== 'shared_storage' && (
                        <Descriptions.Item
                          label={formatMessage({
                            id: 'ocp-express.Component.ClusterInfo.Total',
                            defaultMessage: '总量',
                          })}
                        >
                          {item.totalValue}
                        </Descriptions.Item>
                      )}
                      <Descriptions.Item label={item.description}>
                        {item.assignedValue}
                      </Descriptions.Item>
                      {item.key !== 'shared_storage' && (
                        <Descriptions.Item
                          label={formatMessage({
                            id: 'OBShell.Overview.ClusterInfo.ConsumptionRatio',
                            defaultMessage: '消耗比',
                          })}
                        >
                          <span style={{ color: getStrokeColor(Number(item.percentValue)) }}>
                            {/* 最多保留 1 位有效小数，需要用 toPercent 处理下 */}
                            {isNullValue(item.percentValue)
                              ? '-'
                              : `${toPercent((Number(item.percentValue) || 0) / 100, 1)}%`}
                          </span>
                        </Descriptions.Item>
                      )}
                      {item.key !== 'shared_storage' && (
                        <Descriptions.Item
                          label={formatMessage({
                            id: 'ocp-express.Component.ClusterInfo.Remaining',
                            defaultMessage: '剩余',
                          })}
                        >
                          {item.leftValue}
                        </Descriptions.Item>
                      )}
                    </Descriptions>
                  </Col>
                </Row>
              }
            >
              <div className={styles.progressWrapper}>
                {item.key === 'shared_storage' ? (
                  <Progress type="circle" percent={0} showInfo={false} size={70} strokeWidth={8} />
                ) : (
                  <Progress
                    type="circle"
                    percent={Number(item.percentValue)}
                    size={70}
                    strokeWidth={8}
                    strokeColor={getStrokeColor(Number(item.percentValue))}
                  />
                )}

                <Space direction="vertical" className={styles.progressText}>
                  <div className={styles.progressTitle}>
                    {item.titleDescription ? (
                      <ContentWithQuestion
                        content={item.title}
                        tooltip={{
                          title: item.titleDescription,
                        }}
                      />
                    ) : (
                      item.title
                    )}
                  </div>
                  <div className={styles.progressDescription}>
                    <div className={styles.progressDescriptionItem}>
                      <span className={styles.progressDescriptionItemName}>{item.description}</span>
                      <span>{item.assignedValue}</span>
                    </div>
                    {!isNullValue(item.totalValue) && (
                      <>
                        <Divider style={{ margin: '4px 0' }} />
                        <div className={styles.progressDescriptionItem}>
                          <span className={styles.progressDescriptionItemName}>
                            {formatMessage({
                              id: 'OBShell.Overview.ClusterInfo.Total',
                              defaultMessage: '总量',
                            })}
                          </span>
                          <span>{item.totalValue}</span>
                        </div>
                      </>
                    )}
                  </div>
                </Space>
              </div>
            </MouseTooltip>
          </Col>
        ))}
      </Row>
    </MyCard>
  );
};

export default Detail;
