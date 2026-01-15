import { formatMessage } from '@/util/intl';
import React from 'react';
import {
  Col,
  Descriptions,
  Row,
  Tooltip,
  Typography,
  Popover,
  theme,
  Progress,
  Space,
  Divider,
} from '@oceanbase/design';
import { EllipsisOutlined } from '@oceanbase/icons';
import { isNullValue } from '@oceanbase/util';
import { toPercent } from '@oceanbase/charts/es/util/number';
import { formatSize, isEnglish } from '@/util';
import ObClusterDeployMode from '@/component/ObClusterDeployMode';
import MyCard from '@/component/MyCard';
import MouseTooltip from '@/component/MouseTooltip';
import styles from './index.less';

const { Text } = Typography;

export interface DetailProps {
  clusterData: API.ClusterInfo;
}

const Detail: React.FC<DetailProps> = ({ clusterData }) => {
  const { token } = theme.useToken();

  const stats: API.BaseResourceStats = clusterData.stats || {};

  const metricList = [
    {
      key: 'cpu',
      title: 'CPU',
      description: '已分配',
      percentValue: (stats.cpu_core_assigned_percent || 0).toFixed(1),
      totalValue: stats.cpu_core_total || 0,
      assignedValue: stats.cpu_core_assigned || 0,
      leftValue: (stats.cpu_core_total || 0) - (stats.cpu_core_assigned || 0),
    },
    {
      key: 'memory',
      title: formatMessage({
        id: 'ocp-express.Component.ClusterInfo.Memory',
        defaultMessage: '内存',
      }),
      description: '已分配',
      percentValue: (stats.memory_assigned_percent || 0).toFixed(1),
      totalValue: stats.memory_total || 0,
      assignedValue: stats.memory_in_bytes_assigned || 0,
      leftValue: (stats.memory_in_bytes_total || 0) - (stats.memory_in_bytes_assigned || 0),
    },
    {
      key: 'disk',
      title: formatMessage({
        id: 'ocp-express.Component.ClusterInfo.Disk',
        defaultMessage: '磁盘',
      }),
      description: '已使用',
      percentValue: (stats.disk_assigned_percent || 0).toFixed(1),
      totalValue: stats.disk_total || 0,
      assignedValue: stats.disk_in_bytes_assigned || 0,
      leftValue: (stats.disk_in_bytes_total || 0) - (stats.disk_in_bytes_assigned || 0),
    },
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
      title="资源水位"
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
            span={8}
            style={index !== 2 ? { borderRight: `1px solid ${token.colorBorderSecondary}` } : {}}
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
                      <Descriptions.Item
                        label={formatMessage({
                          id: 'ocp-express.Component.ClusterInfo.Total',
                          defaultMessage: '总量',
                        })}
                      >
                        {isNullValue(item.totalValue)
                          ? '-'
                          : item.key === 'cpu'
                          ? `${item.totalValue} C`
                          : // 内存和磁盘需要进行单位换算
                            item.totalValue}
                      </Descriptions.Item>
                      <Descriptions.Item label={item.description}>
                        {isNullValue(item.assignedValue)
                          ? '-'
                          : item.key === 'cpu'
                          ? `${item.assignedValue} C`
                          : // 内存和磁盘需要进行单位换算
                            formatSize(item.assignedValue)}
                      </Descriptions.Item>
                      <Descriptions.Item label="消耗比">
                        <span style={{ color: getStrokeColor(Number(item.percentValue)) }}>
                          {/* 最多保留 1 位有效小数，需要用 toPercent 处理下 */}
                          {isNullValue(item.percentValue)
                            ? '-'
                            : `${toPercent((Number(item.percentValue) || 0) / 100, 1)}%`}
                        </span>
                      </Descriptions.Item>
                      <Descriptions.Item
                        label={formatMessage({
                          id: 'ocp-express.Component.ClusterInfo.Remaining',
                          defaultMessage: '剩余',
                        })}
                      >
                        {isNullValue(item.leftValue)
                          ? '-'
                          : item.key === 'cpu'
                          ? `${item.leftValue} C`
                          : // 内存和磁盘需要进行单位换算
                            formatSize(item.leftValue)}
                      </Descriptions.Item>
                    </Descriptions>
                  </Col>
                </Row>
              }
            >
              <div className={styles.progressWrapper}>
                <Progress
                  type="circle"
                  percent={Number(item.percentValue)}
                  size={90}
                  strokeWidth={8}
                  strokeColor={getStrokeColor(Number(item.percentValue))}
                />
                <Space direction="vertical" className={styles.progressText}>
                  <div className={styles.progressTitle}>{item.title}</div>
                  <div className={styles.progressDescription}>
                    <div className={styles.progressDescriptionItem}>
                      <span className={styles.progressDescriptionItemName}>{item.description}</span>
                      <span>
                        {isNullValue(item.assignedValue)
                          ? '-'
                          : item.key === 'cpu'
                          ? `${item.assignedValue} C`
                          : // 内存和磁盘需要进行单位换算
                            formatSize(item.assignedValue)}
                      </span>
                    </div>
                    <Divider style={{ margin: '4px 0' }} />
                    <div className={styles.progressDescriptionItem}>
                      <span className={styles.progressDescriptionItemName}>总量</span>
                      <span>
                        {isNullValue(item.totalValue)
                          ? '-'
                          : item.key === 'cpu'
                          ? `${item.totalValue} C`
                          : // 内存和磁盘需要进行单位换算
                            item.totalValue}
                      </span>
                    </div>
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
