import ContentWithQuestion from '@/component/ContentWithQuestion';
import Empty from '@/component/Empty';
import {
  BACKUP_SCHEDULE_MODE_LIST,
  DATA_BACKUP_MODE_LIST,
  MONTH_OPTIONS,
  S3_BACKUP_TYPE_LIST,
  STORAGE_TYPE_LIST,
  WEEK_OPTIONS,
} from '@/constant/backup';
import { RFC3339_TIME_FORMAT, TIME_FORMAT } from '@/constant/datetime';
import { formatMessage } from '@/util/intl';
import { isGte4 } from '@/util/package';
import React from 'react';
import { toNumber } from 'lodash';
import moment from 'moment';
import {
  Card,
  Col,
  Descriptions,
  Divider,
  Form,
  Row,
  Space,
  Switch,
  Table,
  token,
} from '@oceanbase/design';
import type { CardProps } from '@oceanbase/design/es/card';
import { findByValue, isNullValue } from '@oceanbase/util';
import styles from './index.less';

export interface BackupStrategyCardProps extends CardProps {
  backupStrategy?: API.OcpBackupStrategy;
  obVersion?: string;
}

const BackupStrategyCard: React.FC<BackupStrategyCardProps> = ({
  backupStrategy,
  obVersion,
  ...restProps
}) => {
  const { backupDataRetentionDays } = backupStrategy?.obBackupScheduleConfig || {};
  const {
    dataBackupTimeoutMinutes,
    logBackupDelaySeconds,
    noSuccessfulDataBackupAlarmDays,
    libObLogExpireDays,
  } = backupStrategy?.obBackupThresholdConfig || {};

  const customBackupRangeList =
    backupStrategy?.obBackupServiceStorageConfig?.obBackupStorageBaseInfo?.customBackupRangeList ||
    [];
  const dataBackupTypeInfo = customBackupRangeList.find(item => item.backupType === 'DATA');
  const logBackupTypeInfo = customBackupRangeList.find(item => item.backupType === 'LOG');

  const dataStatusItem = findByValue(S3_BACKUP_TYPE_LIST, dataBackupTypeInfo?.nodeType);
  const logStatusItem = findByValue(S3_BACKUP_TYPE_LIST, logBackupTypeInfo?.nodeType);

  const formItemLayout = {
    labelCol: {
      span: 24,
    },

    wrapperCol: {
      span: 24,
    },
  };

  const columns = [
    {
      title: findByValue(
        BACKUP_SCHEDULE_MODE_LIST,
        backupStrategy?.obBackupScheduleConfig?.backupScheduleMode
      ).label,
      dataIndex: 'backupScheduleMode',
    },
  ];

  const dataSourceItem = {
    // 对应第一列的 dataIndex
    backupScheduleMode: formatMessage({
      id: 'ocp-v2.Component.AddBackupStrategy.DataBackupMethod',
      defaultMessage: '数据备份方式',
    }),
  };

  const { dayAndDataBackupModelMap = {} } = backupStrategy?.obBackupScheduleConfig || {};
  Object.keys(dayAndDataBackupModelMap).forEach(key => {
    const options =
      backupStrategy?.obBackupScheduleConfig?.backupScheduleMode === 'WEEK'
        ? WEEK_OPTIONS
        : MONTH_OPTIONS;
    columns.push({
      // key 为字符串，而 options 中的值为数字，因此筛选时需要将 key 转换成数字
      title: findByValue(options, toNumber(key)).label,
      dataIndex: key,
      render: (text: API.DataBackupMode) => findByValue(DATA_BACKUP_MODE_LIST, text).label,
    });

    dataSourceItem[key] = dayAndDataBackupModelMap[key];
  });

  return (
    <Card bordered={false} {...restProps}>
      {/* 集群级备份策略可能为空 */}
      {isNullValue(backupStrategy?.id) ? (
        <Empty
          mode="pageCard"
          description={formatMessage({
            id: 'ocp-v2.Component.BackupStrategyCard.ClusterLevelBackupPolicyNotConfigured',
            defaultMessage: '未配置集群级备份策略',
          })}
        />
      ) : (
        <Form
          layout="horizontal"
          colon={true}
          {...formItemLayout}
          className={`form-with-small-margin ${styles.container}`}
        >
          <Row gutter={[0, 16]}>
            <Col span={24}>
              <Row gutter={[0, 16]}>
                <Col span={24}>
                  <Card className="ocp-card-without-border">
                    <div
                      style={{
                        fontSize: 16,
                        fontWeight: 500,
                        marginBottom: '16px',
                      }}
                    >
                      {formatMessage({
                        id: 'ocp-v2.Component.BackupStrategyCard.BasicConfiguration',
                        defaultMessage: '基础配置',
                      })}
                    </div>
                    <Form.Item
                      label={formatMessage({
                        id: 'ocp-v2.Component.BackupStrategyDetail.BackupStorage',
                        defaultMessage: '备份存储',
                      })}
                    >
                      <div className="ocp-sub-detail-card">
                        <Descriptions column={3}>
                          <Descriptions.Item
                            label={formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyDetail.Service',
                              defaultMessage: '服务名',
                            })}
                          >
                            {backupStrategy?.obBackupServiceStorageConfig?.serviceName || 'OB'}
                          </Descriptions.Item>
                          <Descriptions.Item
                            label={formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyDetail.StorageConfigurationName',
                              defaultMessage: '存储配置名',
                            })}
                          >
                            {backupStrategy?.obBackupServiceStorageConfig?.configName}
                          </Descriptions.Item>
                          <Descriptions.Item
                            label={formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyDetail.StorageType',
                              defaultMessage: '存储类型',
                            })}
                          >
                            {
                              findByValue(
                                STORAGE_TYPE_LIST,
                                backupStrategy?.obBackupServiceStorageConfig
                                  ?.obBackupStorageBaseInfo?.backupStorageType
                              ).label
                            }
                          </Descriptions.Item>
                          <Descriptions.Item
                            label={formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyDetail.StorageDirectory',
                              defaultMessage: '存储目录',
                            })}
                            className="descriptions-item-with-ellipsis"
                          >
                            {
                              backupStrategy?.obBackupServiceStorageConfig?.obBackupStorageBaseInfo
                                ?.storageUrl
                            }
                          </Descriptions.Item>
                          {/* 仅 File 类型支持容量报警阈值 */}
                          {/* issue: https://aone.alipay.com/issue/100596595 */}
                          {backupStrategy?.obBackupServiceStorageConfig?.obBackupStorageBaseInfo
                            ?.backupStorageType === 'BACKUP_STORAGE_FILE' && (
                            <Descriptions.Item
                              label={formatMessage({
                                id: 'ocp-v2.Component.BackupStrategyDetail.CapacityWarningThresholdForBackup',
                                defaultMessage: '备份存储的容量预警阈值',
                              })}
                            >
                              {`${backupStrategy?.obBackupServiceStorageConfig?.alarmStoragePercentageThreshold}%`}
                            </Descriptions.Item>
                          )}

                          {backupStrategy?.obBackupServiceStorageConfig?.obBackupStorageBaseInfo
                            .backupStorageType === 'BACKUP_STORAGE_OSS' && (
                            <>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.AccessDomainName',
                                  defaultMessage: '访问域名',
                                })}
                                className="descriptions-item-with-ellipsis"
                              >
                                {
                                  backupStrategy?.obBackupServiceStorageConfig
                                    ?.obBackupStorageBaseInfo?.ossAccessKey?.endpoint
                                }
                              </Descriptions.Item>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.User',
                                  defaultMessage: '访问用户',
                                })}
                              >
                                {
                                  backupStrategy?.obBackupServiceStorageConfig
                                    ?.obBackupStorageBaseInfo?.ossAccessKey?.accessKeyId
                                }
                              </Descriptions.Item>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.Accesskey',
                                  defaultMessage: '访问秘钥',
                                })}
                              >
                                ******
                              </Descriptions.Item>
                            </>
                          )}
                          {backupStrategy?.obBackupServiceStorageConfig?.obBackupStorageBaseInfo
                            .backupStorageType === 'BACKUP_STORAGE_S3' && (
                            <>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.AccessDomainName',
                                  defaultMessage: '访问域名',
                                })}
                                className="descriptions-item-with-ellipsis"
                              >
                                {
                                  backupStrategy?.obBackupServiceStorageConfig
                                    ?.obBackupStorageBaseInfo?.s3AccessKey?.endpoint
                                }
                              </Descriptions.Item>
                              <Descriptions.Item label="AK">
                                {
                                  backupStrategy?.obBackupServiceStorageConfig
                                    ?.obBackupStorageBaseInfo?.s3AccessKey?.accessKeyId
                                }
                              </Descriptions.Item>
                              <Descriptions.Item label="SK">******</Descriptions.Item>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.Region',
                                  defaultMessage: '地域',
                                })}
                              >
                                {
                                  backupStrategy?.obBackupServiceStorageConfig
                                    ?.obBackupStorageBaseInfo?.s3AccessKey?.region
                                }
                              </Descriptions.Item>
                            </>
                          )}
                          {backupStrategy?.obBackupServiceStorageConfig?.obBackupStorageBaseInfo
                            .backupStorageType === 'BACKUP_STORAGE_COS' && (
                            <>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.AccessDomainName',
                                  defaultMessage: '访问域名',
                                })}
                              >
                                {
                                  backupStrategy?.obBackupServiceStorageConfig
                                    ?.obBackupStorageBaseInfo?.cosAccessKey?.endpoint
                                }
                              </Descriptions.Item>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.ResourceIdentifierAppid',
                                  defaultMessage: '资源标识（APPID）',
                                })}
                              >
                                {
                                  backupStrategy?.obBackupServiceStorageConfig
                                    ?.obBackupStorageBaseInfo?.cosAccessKey?.appId
                                }
                              </Descriptions.Item>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.TheIdOfTheProject',
                                  defaultMessage: '项目身份识别 ID',
                                })}
                                className="descriptions-item-with-ellipsis"
                              >
                                {
                                  backupStrategy?.obBackupServiceStorageConfig
                                    ?.obBackupStorageBaseInfo?.cosAccessKey?.accessKeyId
                                }
                              </Descriptions.Item>
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.ProjectIdentityKey',
                                  defaultMessage: '项目身份秘钥',
                                })}
                              >
                                ******
                              </Descriptions.Item>
                            </>
                          )}
                        </Descriptions>
                      </div>
                    </Form.Item>
                    {backupStrategy?.backupDimension === 'TENANT' && (
                      <Form.Item
                        label={formatMessage({
                          id: 'ocp-v2.Component.BackupStrategyCard.BackupScope',
                          defaultMessage: '备份范围',
                        })}
                      >
                        {customBackupRangeList?.length > 0 ? (
                          <div className="ocp-sub-detail-card">
                            <Descriptions
                              column={3}
                              className={styles.descriptionTitle}
                              title={formatMessage({
                                id: 'ocp-v2.Component.BackupStrategyCard.DataBackup',
                                defaultMessage: '数据备份',
                              })}
                            >
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyCard.BackupType',
                                  defaultMessage: '备份类型',
                                })}
                                style={{
                                  marginBottom: '16px',
                                }}
                              >
                                {dataStatusItem.label}
                                {formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyCard.Priority',
                                  defaultMessage: '（优先级：',
                                })}
                                {dataBackupTypeInfo?.nodeList?.join(';')}）
                              </Descriptions.Item>
                            </Descriptions>
                            <Descriptions
                              column={3}
                              title={formatMessage({
                                id: 'ocp-v2.Component.BackupStrategyCard.LogBackup',
                                defaultMessage: '日志备份',
                              })}
                              className={styles.descriptionTitle}
                              style={{
                                marginBottom: '0',
                              }}
                            >
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyCard.BackupType',
                                  defaultMessage: '备份类型',
                                })}
                              >
                                {logStatusItem.label}
                                {formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyCard.Priority',
                                  defaultMessage: '（优先级：',
                                })}
                                {logBackupTypeInfo?.nodeList?.join(';')}）
                              </Descriptions.Item>
                            </Descriptions>
                          </div>
                        ) : (
                          formatMessage({
                            id: 'ocp-v2.Component.BackupStrategyCard.Closed',
                            defaultMessage: '已关闭',
                          })
                        )}
                      </Form.Item>
                    )}
                    <Form layout="horizontal">
                      <Form.Item
                        label={formatMessage({
                          id: 'ocp-v2.Component.BackupStrategyDetail.ExpiredBackupCleanupSchedulingConfiguration',
                          defaultMessage: '过期备份清理调度配置',
                        })}
                      >
                        {backupStrategy?.obBackupScheduleConfig?.cleanEnabled
                          ? formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyCard.Opened',
                              defaultMessage: '已开启',
                            })
                          : formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyCard.Closed',
                              defaultMessage: '已关闭',
                            })}
                      </Form.Item>
                    </Form>
                    {backupStrategy?.obBackupScheduleConfig?.cleanEnabled && (
                      <Form.Item>
                        <div className="ocp-sub-detail-card">
                          <Descriptions>
                            {/* 物理备份的清理调度时间是在 OB 内核实现，不支持 OCP 白屏设置，因此只有逻辑备份才展示清理调度时间 */}
                            {/* issue: https://aone.alipay.com/issue/31207988 */}
                            {backupStrategy.backupMode === 'LOGICAL_BACKUP' && (
                              <Descriptions.Item
                                label={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.SchedulingTime',
                                  defaultMessage: '调度时间',
                                })}
                              >
                                {moment(
                                  backupStrategy?.obBackupScheduleConfig?.cleanScheduleTime,
                                  RFC3339_TIME_FORMAT
                                ).format(TIME_FORMAT)}
                              </Descriptions.Item>
                            )}

                            <Descriptions.Item
                              label={formatMessage({
                                id: 'ocp-v2.Component.BackupStrategyDetail.LastRetentionDaysOfBackup',
                                defaultMessage: '备份文件最近保留天数',
                              })}
                            >
                              {
                                backupStrategy?.obSecondaryBackupConfig?.scheduleConfig
                                  ?.backupDataRetentionDays
                              }
                              {formatMessage({
                                id: 'OBShell.Component.BackupStrategyCard.Day',
                                defaultMessage: '天',
                              })}
                            </Descriptions.Item>
                          </Descriptions>
                        </div>
                      </Form.Item>
                    )}

                    <Form.Item
                      label={formatMessage({
                        id: 'ocp-v2.Component.BackupStrategyCard.AlarmThreshold',
                        defaultMessage: '报警阈值',
                      })}
                    >
                      <div className="ocp-sub-detail-card">
                        <Descriptions>
                          <Descriptions.Item
                            label={
                              <ContentWithQuestion
                                content={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.AlarmThresholdForDataBackup',
                                  defaultMessage: '数据备份超时报警阈值',
                                })}
                                tooltip={{
                                  title: formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.AnAlarmIsTriggeredWhen',
                                    defaultMessage:
                                      '当数据备份的时间（不包括备份调度的时间）超过设定阈值的时候会触发告警，建议参考之前的数据备份任务的起止时间的耗时进行调整，数据备份任务的时长可能会因数据大小以及备份介质的性能等因素有所不同',
                                  }),

                                  placement: 'top',
                                }}
                              />
                            }
                          >
                            {formatMessage(
                              {
                                id: 'ocp-v2.Component.BackupStrategyDetail.DatabackuptimeoutminutesMinutes',
                                defaultMessage: '{dataBackupTimeoutMinutes} 分钟',
                              },

                              { dataBackupTimeoutMinutes }
                            )}
                          </Descriptions.Item>
                          <Descriptions.Item
                            label={
                              <ContentWithQuestion
                                content={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.LogBackupLatencyAlarmThreshold',
                                  defaultMessage: '日志备份延时报警阈值',
                                })}
                                tooltip={{
                                  title: formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.AnAlertIsTriggeredWhen',
                                    defaultMessage:
                                      '当日志备份的延时时间超过设定阈值时会触发告警，延迟时间会因为日志大小以及备份介质的性能等因素有所不同',
                                  }),

                                  placement: 'top',
                                }}
                              />
                            }
                          >
                            {formatMessage(
                              {
                                id: 'ocp-v2.Component.BackupStrategyDetail.LogbackupdelaysecondsSeconds',
                                defaultMessage: '{logBackupDelaySeconds} 秒',
                              },

                              { logBackupDelaySeconds }
                            )}
                          </Descriptions.Item>
                          <Descriptions.Item
                            label={
                              <ContentWithQuestion
                                content={formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.TheNumberOfDaysThat',
                                  defaultMessage: '无成功数据备份预警天数',
                                })}
                                tooltip={{
                                  title: formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.IfNoDataBackupIs',
                                    defaultMessage:
                                      '如果在设置的报警天数内无成功的数据备份，则会发送告警，设置的天数不要小于备份调度周期，请根据设置的备份调度周期进行设置',
                                  }),

                                  placement: 'top',
                                }}
                              />
                            }
                          >
                            {formatMessage(
                              {
                                id: 'ocp-v2.Component.BackupStrategyDetail.NosuccessfuldatabackupalarmdaysDays',
                                defaultMessage: '{noSuccessfulDataBackupAlarmDays} 天',
                              },

                              { noSuccessfulDataBackupAlarmDays }
                            )}
                          </Descriptions.Item>
                          {/* 只有逻辑备份支持 liboblog 保留天数这一配置，物理备份不支持 */}
                          {backupStrategy?.backupMode === 'LOGICAL_BACKUP' && (
                            <Descriptions.Item
                              label={
                                <ContentWithQuestion
                                  content={formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.LiboblogRetentionDays',
                                    defaultMessage: 'liboblog 的保留天数',
                                  })}
                                  tooltip={{
                                    title: formatMessage({
                                      id: 'ocp-v2.Component.BackupStrategyDetail.TheNumberOfDaysThat.1',
                                      defaultMessage:
                                        '日志备份组件产生的日志（系统运行日志，非备份文件）的保留天数，系统将自动清理超期的日志',
                                    }),

                                    placement: 'top',
                                  }}
                                />
                              }
                            >
                              {formatMessage(
                                {
                                  id: 'ocp-v2.Component.BackupStrategyDetail.LiboblogexpiredaysDays',
                                  defaultMessage: '{libObLogExpireDays} 天',
                                },

                                { libObLogExpireDays }
                              )}
                            </Descriptions.Item>
                          )}
                        </Descriptions>
                      </div>
                    </Form.Item>
                  </Card>
                  <Divider style={{ margin: '8px 0px' }} />
                </Col>
                <Col span={24}>
                  <Card className="ocp-card-without-border">
                    <div
                      style={{
                        fontSize: 16,
                        fontWeight: 500,
                        marginBottom: '16px',
                      }}
                    >
                      {formatMessage({
                        id: 'ocp-v2.Component.BackupStrategyCard.SchedulingConfiguration',
                        defaultMessage: '调度配置',
                      })}
                    </div>
                    <Form layout="horizontal">
                      <Form.Item
                        colon={false}
                        label={formatMessage(
                          {
                            id: 'ocp-v2.Component.BackupStrategyDetail.SchedulingCycleFindbyvaluebackupschedule',
                            defaultMessage: '调度周期（{findByValueBACKUPSCHEDULE}）',
                          },

                          {
                            findByValueBACKUPSCHEDULE: findByValue(
                              BACKUP_SCHEDULE_MODE_LIST,
                              backupStrategy?.obBackupScheduleConfig?.backupScheduleMode
                            ).label,
                          }
                        )}
                        labelCol={{
                          span: 24,
                        }}
                        wrapperCol={{
                          span: 24,
                        }}
                        style={{ marginBottom: 8 }}
                      >
                        <Table
                          bordered={true}
                          dataSource={[dataSourceItem]}
                          columns={columns}
                          pagination={false}
                          // 尽管 backupScheduleMode 是一个固定值，但由于只有一行数据，也能保证 rowKey 是唯一的
                          rowKey={record => record.backupScheduleMode}
                          className={styles.scheduleCycleTable}
                        />
                      </Form.Item>
                    </Form>
                    <Descriptions column={3}>
                      <Descriptions.Item
                        label={formatMessage({
                          id: 'ocp-v2.Component.BackupStrategyDetail.SchedulingTime',
                          defaultMessage: '调度时间',
                        })}
                      >
                        {moment(
                          backupStrategy?.obBackupScheduleConfig?.backupScheduleTime,
                          RFC3339_TIME_FORMAT
                        ).format(TIME_FORMAT)}
                      </Descriptions.Item>
                      <Descriptions.Item
                        label={formatMessage({
                          id: 'ocp-v2.Component.BackupStrategyDetail.LogBackup',
                          defaultMessage: '日志备份',
                        })}
                      >
                        {backupStrategy?.obBackupScheduleConfig?.startLogBackup
                          ? formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyCard.Opened',
                              defaultMessage: '已开启',
                            })
                          : formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyCard.Closed',
                              defaultMessage: '已关闭',
                            })}
                      </Descriptions.Item>
                      {/* 仅物理备份支持集群自动合并 */}
                      {backupStrategy?.backupMode === 'PHYSICAL_BACKUP' && (
                        <Descriptions.Item
                          label={
                            <ContentWithQuestion
                              content={formatMessage({
                                id: 'ocp-v2.Component.BackupStrategyDetail.AutoMerge',
                                defaultMessage: '自动合并',
                              })}
                              tooltip={{
                                placement: 'right',
                                overlayStyle: {
                                  maxWidth: 'none',
                                  width: 590,
                                },

                                title: (
                                  <div>
                                    <div>
                                      {formatMessage({
                                        id: 'ocp-v2.Component.BackupStrategyDetail.IfAutomaticMergeIsEnabled',
                                        defaultMessage:
                                          '若自动合并开启，在备份调度的过程中，当备份依赖合并时将自动发起一次合并。',
                                      })}
                                    </div>
                                    <div>
                                      {formatMessage({
                                        id: 'ocp-v2.Component.BackupStrategyDetail.IfAutomaticMergeIsDisabled',
                                        defaultMessage:
                                          '若自动合并关闭，在备份调度的过程中，始终不会自动发起合并，可能导致备份调度失败。',
                                      })}
                                    </div>
                                  </div>
                                ),
                              }}
                            />
                          }
                        >
                          <span>
                            {backupStrategy?.autoCompactEnabled
                              ? formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyCard.Opened',
                                  defaultMessage: '已开启',
                                })
                              : formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyCard.Closed',
                                  defaultMessage: '已关闭',
                                })}
                          </span>
                          <span style={{ color: token.colorTextTertiary, marginLeft: 8 }}>
                            {backupStrategy?.autoCompactEnabled
                              ? formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.WhenThisFeatureIsEnabled',
                                  defaultMessage:
                                    '开启时，在备份调度的过程中，当备份依赖合并时将自动发起一次合并',
                                })
                              : formatMessage({
                                  id: 'ocp-v2.Component.BackupStrategyDetail.WhenTheBackupSchedulingIs',
                                  defaultMessage:
                                    '关闭时，在备份调度的过程中，始终不会自动发起合并，可能导致备份调度失败',
                                })}
                          </span>
                        </Descriptions.Item>
                      )}
                    </Descriptions>
                  </Card>
                </Col>
              </Row>
            </Col>
            {!isGte4(obVersion) && (
              <Col span={24}>
                <Card
                  type="inner"
                  title={
                    <Space
                      style={{
                        fontSize: 16,
                      }}
                    >
                      <ContentWithQuestion
                        content={formatMessage({
                          id: 'ocp-v2.Component.BackupStrategyDetail.SecondaryBackup',
                          defaultMessage: '二次备份',
                        })}
                        tooltip={{
                          title: formatMessage({
                            id: 'ocp-v2.Component.BackupStrategyDetail.YouCanBackUpData',
                            defaultMessage:
                              '将备份数据文件和日志文件备份至另一个目录，可以实现备份文件',
                          }),
                        }}
                      />

                      <Switch
                        size="small"
                        disabled={true}
                        checked={backupStrategy?.secondaryBackupEnabled}
                      />

                      {/* <span
                   style={{
                     color: token.colorTextTertiary,
                     fontSize: 14,
                     fontWeight: 'normal',
                   }}
                  >
                   {supportSecondaryBackupTip}
                  </span> */}
                    </Space>
                  }
                  bodyStyle={
                    backupStrategy?.secondaryBackupEnabled
                      ? {
                          paddingTop: 8,
                        }
                      : {
                          display: 'none',
                        }
                  }
                >
                  {backupStrategy?.secondaryBackupEnabled && (
                    <Form
                      layout="horizontal"
                      {...formItemLayout}
                      className="form-with-small-margin"
                    >
                      <Space direction="vertical" size={8}>
                        <Card className="ocp-card-without-border">
                          <div
                            style={{
                              fontSize: 16,
                              fontWeight: 500,
                            }}
                          >
                            {formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyDetail.StorageConfiguration',
                              defaultMessage: '存储配置',
                            })}
                          </div>
                          <Form.Item
                            label={formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyCard.SecondaryBackupStorage',
                              defaultMessage: '二次备份存储',
                            })}
                          >
                            <div className="ocp-sub-detail-card">
                              <Descriptions column={3}>
                                <Descriptions.Item
                                  label={formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.StorageConfigurationName',
                                    defaultMessage: '存储配置名',
                                  })}
                                >
                                  {
                                    backupStrategy?.obSecondaryBackupConfig?.storageConfig
                                      ?.configName
                                  }
                                </Descriptions.Item>
                                <Descriptions.Item
                                  label={formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.StorageDirectory',
                                    defaultMessage: '存储目录',
                                  })}
                                >
                                  {
                                    backupStrategy?.obSecondaryBackupConfig?.storageConfig
                                      ?.obBackupStorageBaseInfo.storageUrl
                                  }
                                </Descriptions.Item>
                                <Descriptions.Item
                                  label={formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.CapacityAlertThresholdOfBackup',
                                    defaultMessage: '备份存储的容量告警阈值',
                                  })}
                                >
                                  {`${backupStrategy?.obSecondaryBackupConfig?.storageConfig?.alarmStoragePercentageThreshold}%`}
                                </Descriptions.Item>
                              </Descriptions>
                            </div>
                          </Form.Item>
                        </Card>
                        <Divider style={{ margin: '0px 0px 8px 0px' }} />
                        <Card className="ocp-card-without-border">
                          <div
                            style={{
                              fontSize: 16,
                              fontWeight: 500,
                            }}
                          >
                            {formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyDetail.SchedulingConfiguration',
                              defaultMessage: '调度配置',
                            })}
                          </div>
                          <Form.Item
                            label={formatMessage({
                              id: 'ocp-v2.Component.BackupStrategyCard.SecondaryBackupScheduling',
                              defaultMessage: '二次备份调度',
                            })}
                          >
                            <div className="ocp-sub-detail-card">
                              <Descriptions column={1}>
                                <Descriptions.Item
                                  label={formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.SecondaryDataBackup',
                                    defaultMessage: '数据二次备份',
                                  })}
                                >
                                  {formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.AfterADataBackupTask',
                                    defaultMessage:
                                      '在数据备份任务成功后会自动发起数据二次备份任务',
                                  })}
                                </Descriptions.Item>
                                <Descriptions.Item
                                  label={formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.SecondaryLogBackup',
                                    defaultMessage: '日志二次备份',
                                  })}
                                >
                                  {/* 小尺寸的 Switch 组件直接作为 Descriptions.Item 的直接子元素，无法垂直居中，需要使用 span 元素在外面包裹一层 */}
                                  {/* 因为小尺寸的 Switch 高度和行高都写死成了 16px，无法继承 Descriptions.Item 的行高，但 span 元素可以默认继承 */}
                                  <span>
                                    <Switch
                                      size="small"
                                      disabled={true}
                                      checked={
                                        backupStrategy?.obSecondaryBackupConfig?.scheduleConfig
                                          ?.logBackupScheduleEnabled
                                      }
                                    />
                                  </span>
                                </Descriptions.Item>
                                {backupStrategy?.obSecondaryBackupConfig?.scheduleConfig
                                  ?.logBackupScheduleEnabled && (
                                  <Descriptions.Item
                                    label={formatMessage({
                                      id: 'ocp-v2.Component.BackupStrategyDetail.SecondaryLogBackupCycle',
                                      defaultMessage: '日志二次备份周期',
                                    })}
                                  >
                                    {formatMessage({
                                      id: 'ocp-v2.Component.BackupStrategyDetail.OnceADay',
                                      defaultMessage: '1 天一次',
                                    })}
                                  </Descriptions.Item>
                                )}
                              </Descriptions>
                            </div>
                          </Form.Item>
                          <Form.Item
                            label={
                              <div>
                                <span style={{ marginRight: 8 }}>
                                  {formatMessage({
                                    id: 'ocp-v2.Component.BackupStrategyDetail.ExpiredBackupCleanupSchedulingConfiguration',
                                    defaultMessage: '过期备份清理调度配置',
                                  })}
                                </span>
                                <Switch
                                  size="small"
                                  disabled={true}
                                  checked={
                                    backupStrategy?.obSecondaryBackupConfig?.scheduleConfig
                                      ?.cleanEnabled
                                  }
                                />
                              </div>
                            }
                          >
                            <Card className="ocp-card-without-border">
                              {backupStrategy?.obSecondaryBackupConfig?.scheduleConfig
                                ?.cleanEnabled && (
                                <div className="ocp-sub-detail-card">
                                  <Descriptions>
                                    <Descriptions.Item
                                      label={formatMessage({
                                        id: 'ocp-v2.Component.BackupStrategyDetail.LastRetentionDaysOfBackup',
                                        defaultMessage: '备份文件最近保留天数',
                                      })}
                                    >
                                      {
                                        backupStrategy?.obSecondaryBackupConfig?.scheduleConfig
                                          ?.backupDataRetentionDays
                                      }
                                      {formatMessage({
                                        id: 'OBShell.Component.BackupStrategyCard.Day',
                                        defaultMessage: '天',
                                      })}
                                    </Descriptions.Item>
                                  </Descriptions>
                                </div>
                              )}
                            </Card>
                          </Form.Item>
                        </Card>
                      </Space>
                    </Form>
                  )}
                </Card>
              </Col>
            )}
          </Row>
        </Form>
      )}
    </Card>
  );
};

export default BackupStrategyCard;
