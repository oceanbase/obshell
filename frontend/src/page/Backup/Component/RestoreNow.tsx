import ContentWithQuestion from '@/component/ContentWithQuestion';
import MySelect from '@/component/MySelect';
import Result from '@/component/Result';
import { DATE_FORMAT_DISPLAY, DATE_TIME_FORMAT, TIME_FORMAT } from '@/constant/datetime';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { formatMessage } from '@/util/intl';
import { getTenantOverView } from '@/service/obshell/tenant';
import { obclusterInfo } from '@/service/obshell/obcluster';
import { unitConfigCreate } from '@/service/obshell/unit';
import { getTenantBackupConfig } from '@/service/obshell/backup';
import * as ObTenantBackupController from '@/service/obshell/restore';
import {
  breadcrumbItemRender,
  getLatestRecoverableRangeTime,
  getRecoverableRangeTime,
} from '@/util/component';
import { getBackupTimeData } from '@/util/backup';
import { getMicroseconds } from '@/util/datetime';
import { history } from 'umi';
import React, { useRef, useState, useEffect } from 'react';
import { find, uniqueId, uniqWith } from 'lodash';
import type { Moment } from 'moment';
import moment from 'moment';
import type { FormComponentProps } from '@ant-design/compatible/lib/form';
import type { WrappedFormUtils } from '@ant-design/compatible/lib/form/Form';
import {
  Button,
  Card,
  Col,
  DatePicker,
  Form,
  Input,
  message,
  Popconfirm,
  Descriptions,
  Popover,
  Row,
  Space,
  Spin,
  theme,
  TimePicker,
} from '@oceanbase/design';
import type { Route } from '@oceanbase/design/es/breadcrumb/Breadcrumb';
import { ExclamationCircleFilled, InfoCircleFilled } from '@oceanbase/icons';
import { PageContainer } from '@oceanbase/ui';
import { isNullValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import RestoreTenantCard from './RestoreTenantCard';
import StorageConfigCard from './StorageConfigCard';

const { Option } = MySelect;

const TimeScopeType = {
  label: 'LATEST_SIX_MONTHS' as API.ComOceanbaseOcpBackupEnumsObQueryTimeScopeType,
  days: 180,
};

export interface RestoreNowProps extends FormComponentProps {
  form?: WrappedFormUtils;
  tenantName?: string; // 源租户
  useType?: string;
  // 源集群的构建版本号
  minObBuildVersion?: string;
  isCopy?: boolean; // 因为具备集群、租户信息，复制等同于从租户视图发起恢复任务
  taskId?: string;
  restoreDetail?: API.ComOceanbaseOcpBackupModelRestoreHistoryTask;
}

const RestoreNow: React.FC<RestoreNowProps> = ({
  useType,
  tenantName,
  minObBuildVersion,
  isCopy,
}) => {
  const isGlobalRestore = isNullValue(tenantName);

  const restoreTenantRef = useRef();

  const { token } = theme.useToken();
  const [form] = Form.useForm();

  const title = formatMessage({
    id: 'ocp-v2.Backup.Component.RestoreNow.InitiateRecovery',
    defaultMessage: '发起恢复',
  });

  useDocumentTitle(title);

  const { getFieldsValue, getFieldValue, setFieldsValue, validateFields } = form;
  const [visible, setVisible] = useState(false);
  const [taskId, setTaskId] = useState(false);

  // 是否测试过
  const [checkedStatus, setCheckedStatus] = useState<string>('unchecked');

  const [recoveryCheckResult, setRecoveryCheckResult] = useState(undefined);
  const [showObAdminTip, setShowObAdminTip] = useState<boolean>(false);

  const { restoreDate, restoreTime, restoreMicroseconds } = getFieldsValue();

  const [currentRestoreDate, setCurrentRestoreDate] = useState(restoreDate);
  const [currentRestoreTime, setCurrentRestoreTime] = useState(restoreTime);
  const [currentRestoreMicroseconds, setCurrentRestoreMicroseconds] = useState(restoreTime);

  // 对恢复的微秒值进行格式化
  const formatRestoreMicroseconds = (value: number | undefined) => {
    if (value === undefined || value === null) {
      return '000000';
    }
    return value.toString().padStart(6, '0');
  };

  const { data: clusterInfo } = useRequest(obclusterInfo, {
    defaultParams: [{}],
  });

  const clusterData = clusterInfo?.data || {};

  const { data: tenantBackupConfigData, run } = useRequest(getTenantBackupConfig, {
    manual: true,
    defaultParams: [
      {
        name: tenantName,
      },
    ],

    onSuccess: res => {
      if (res?.successful) {
        const { archive_base_uri, data_base_uri } = res?.data;
        getRestoreSourceTenantInfo({
          archive_log_uri: archive_base_uri,
          data_backup_uri: data_base_uri,
        }).then(res => {
          if (res?.successful && res?.data) {
            setCheckedStatus('passed');
            setShowObAdminTip(false);
          } else if (res?.error && res?.error?.errCode === 'Environment.WithoutObAdmin') {
            setShowObAdminTip(true);
            setCheckedStatus('passed');
          } else {
            setCheckedStatus('failed');
            setShowObAdminTip(false);
            setRecoveryCheckResult(res.error);
          }
        });
      }
    },
  });

  const tenantBackupConfigInfo = tenantBackupConfigData?.data || {};

  useEffect(() => {
    if (tenantName) {
      run({
        name: tenantName,
      });
    }
  }, [tenantName]);

  const { data } = useRequest(getTenantOverView, {
    defaultParams: [{}],
  });
  const tenantList = data?.data?.contents || [];

  const onParseBackupInfoDataForDateSuccess = res => {
    if (res.successful && res?.data) {
      const { tenant_name, restore_windows } = res.data;
      // 获取最新时间点（保持原始精度）
      if (restore_windows?.restore_windows?.length > 0) {
        // 遍历找到真正的最新时间点，同时保持原始字符串的微秒精度
        let maxEndTimeItem = restore_windows.restore_windows[0];
        let maxEndTimeMoment = moment(maxEndTimeItem.end_time);

        restore_windows.restore_windows.forEach(item => {
          const itemMoment = moment(item.end_time);
          if (itemMoment.isAfter(maxEndTimeMoment)) {
            maxEndTimeItem = item;
            maxEndTimeMoment = itemMoment;
          }
        });

        // 从原始字符串中提取微秒值，保持精度
        const originalEndTimeStr = maxEndTimeItem.end_time;

        // 尝试多种微秒格式匹配
        let microseconds = 0;
        let microsecondsMatch = originalEndTimeStr.match(/\.(\d{6})$/);

        if (microsecondsMatch) {
          // 标准6位微秒格式
          microseconds = parseInt(microsecondsMatch[1], 10);
        } else {
          // 尝试其他可能的微秒格式
          microsecondsMatch = originalEndTimeStr.match(/\.(\d{1,6})/);
          if (microsecondsMatch) {
            const microStr = microsecondsMatch[1];
            // 补齐到6位
            microseconds = parseInt(microStr.padEnd(6, '0'), 10);
          }
        }

        setFieldsValue({
          restoreDate: maxEndTimeMoment,
          restoreTime: maxEndTimeMoment,
          restoreMicroseconds: microseconds,
        });

        // 同时更新当前状态
        setCurrentRestoreDate(maxEndTimeMoment);
        setCurrentRestoreTime(maxEndTimeMoment);
        setCurrentRestoreMicroseconds(microseconds);

        // 为了保证在 handleRestoreEndTimeChange 获取最新的数据，这里使用异步执行
        setTimeout(() => {
          handleRestoreEndTimeChange(maxEndTimeMoment, maxEndTimeMoment, microseconds);
        }, 0);
      }
      setFieldsValue({
        sourceTenantName: tenant_name,
      });
    }
  };

  // 获取租户在选中日期的备份信息
  const {
    data: tenantParseBackupInfoDataForDate,
    runAsync: getRestoreSourceTenantInfo,
    loading: getRestoreSourceTenantInfoLoading,
  } = useRequest(ObTenantBackupController.getRestoreSourceTenantInfo, {
    manual: true,
    onSuccess: onParseBackupInfoDataForDateSuccess,
  });

  const tenantParseInfoForDate = tenantParseBackupInfoDataForDate?.data;

  // 租户视图发起恢复
  const { runAsync: tenantRestore, loading: tenantRestoreLoading } = useRequest(
    ObTenantBackupController.tenantRestore,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'ocp-v2.Backup.Component.RestoreNow.TheRecoveryTaskIsInitiated',
              defaultMessage: '恢复任务发起成功',
            })
          );

          setVisible(true);
          setTaskId(res.data?.id);
        }
      },
    }
  );

  const handleCheck = () => {
    validateFields(['archive_log_uri', 'data_backup_uri']).then(values => {
      const { archive_log_uri, data_backup_uri } = values;

      getRestoreSourceTenantInfo({
        archive_log_uri: 'file://' + archive_log_uri,
        data_backup_uri: 'file://' + data_backup_uri,
      }).then(res => {
        if (res?.successful && res?.data) {
          setCheckedStatus('passed');
          setShowObAdminTip(false);
        } else if (res?.error && res?.error?.errCode === 'Environment.WithoutObAdmin') {
          setShowObAdminTip(true);
          setCheckedStatus('passed');
        } else {
          setCheckedStatus('failed');
          setShowObAdminTip(false);
          setRecoveryCheckResult(res.error);
        }
      });
    });
  };

  const handleSubmit = () => {
    if (checkedStatus === 'unchecked') {
      return message.error(
        formatMessage({
          id: 'OBShell.Backup.Component.RestoreNow.TestTheStorageDirectoryFirst',
          defaultMessage: '请先测试存储目录',
        })
      );
    }
    validateFields().then(values => {
      const {
        sourceTenantName,
        restoreDate: restoreDateValue,
        restoreTime: restoreTimeValue,
        restoreMicroseconds: restoreMicrosecondsValue,
        restore_tenant_name,
        primaryZone,
        zones,
        archive_log_uri,
        data_backup_uri,
        ...restValues
      } = values;
      const primary_zone = primaryZone;
      const zoneUnitMap = {};
      const checkZones = zones.filter(item => !!item.checked);

      uniqWith(checkZones, (a, b) => {
        return a.cpuCore === b.cpuCore && a.memorySize === b.memorySize;
      }).map(({ cpuCore, memorySize }) => {
        zoneUnitMap['' + cpuCore + memorySize] = {
          name: `tenant_unit_${Date.now()}_${uniqueId()}`,
          cpuCore: Number(cpuCore),
          memorySize: memorySize + 'GB',
        };
      });

      const zoneList = checkZones?.map(zone => {
        return {
          name: zone.name,
          replica_type: zone.replica_type,
          unit_config_name: zoneUnitMap['' + zone.cpuCore + zone.memorySize]?.name,
          unit_num: zone.unit_num,
        };
      });

      // 需要精确到微秒级
      const timestamp = restoreDateValue
        ? !isNullValue(restoreMicrosecondsValue) &&
        `${restoreDateValue.format('YYYY-MM-DD')}T${(restoreTimeValue || 0).format(
          'HH:mm:ss'
        )}.${formatRestoreMicroseconds(restoreMicrosecondsValue || 0)}${restoreTimeValue.format(
          'Z'
        )}`
        : null;

      const { archive_base_uri, data_base_uri } = tenantBackupConfigInfo;
      const restoreParam = {
        ...restValues,
        archive_log_uri: isGlobalRestore ? 'file://' + archive_log_uri : archive_base_uri,
        data_backup_uri: isGlobalRestore ? 'file://' + data_backup_uri : data_base_uri,
        timestamp,
        restore_tenant_name,
        primary_zone,
        // 仅提交勾选的 Zone 副本
        zone_list: zoneList,
      };

      Promise.all(
        Object.entries(zoneUnitMap).map(([key, item]) => {
          return unitConfigCreate({
            max_cpu: item.cpuCore,
            memory_size: item.memorySize,
            name: item.name,
          });
        })
      ).then(res => {
        if (!res.some(item => item.successful === false)) {
          tenantRestore(restoreParam);
        }
      });
    });
  };

  // 格式化后的所选恢复日期
  const selectedRestoreDate =
    currentRestoreDate && (currentRestoreDate as Moment).format(DATE_FORMAT_DISPLAY);

  // 格式化后的所选恢复时刻
  const selectedRestoreDateTime =
    currentRestoreDate && currentRestoreTime && !isNullValue(currentRestoreMicroseconds)
      ? `${moment(currentRestoreDate).format(DATE_FORMAT_DISPLAY)} ${moment(
        currentRestoreTime
      ).format(TIME_FORMAT)}.${formatRestoreMicroseconds(currentRestoreMicroseconds)}`
      : '-';

  // 源租户在不同视图下的展示 label
  const backupTenantLabel = tenantParseInfoForDate?.tenant_name;
  // 租户在最近七天的备份信息
  const recoverableItems = tenantParseInfoForDate?.restore_windows?.restore_windows || [];
  // 租户在最近七天的可恢复时间列表
  const recoverableList = recoverableItems;

  // 租户在选中日期的备份信息
  const backupTenantForDate = tenantParseInfoForDate;

  // 租户在选中日期的可恢复时间列表
  const { recoverableList: recoverableListForDate } = getBackupTimeData(
    backupTenantForDate?.restore_windows?.restore_windows || [],
    currentRestoreDate
  );

  const routes: Route[] = [
    {
      path: isGlobalRestore ? `/tenant` : `/cluster/tenant/${tenantName}/backup`,
      breadcrumbName: isGlobalRestore
        ? formatMessage({
          id: 'OBShell.component.MonitorSearch.Tenant',
          defaultMessage: '租户',
        })
        : formatMessage({
          id: 'OBShell.Backup.Component.RestoreNow.BackupRecovery',
          defaultMessage: '备份恢复',
        }),
    },

    { breadcrumbName: title },
  ];

  // 恢复时刻改变的回调，当修改恢复时间和微秒值，以及选中日期的备份信息返回时进行调用
  // issue: https://aone.alipay.com/issue/101439637
  const handleRestoreEndTimeChange = (
    newRestoreDate: Moment | null,
    newRestoreTime: Moment | null,
    newRestoreMicroseconds?: number
  ) => {
    // 恢复日期、时间和微秒值都不为空，才执行可恢复时间点的校验逻辑
    if (newRestoreDate && newRestoreTime && !isNullValue(newRestoreMicroseconds)) {
      const restoreEndTime = moment(
        `${(newRestoreDate as Moment).format('YYYY-MM-DD')} ${newRestoreTime.format('HH:mm:ss')}`
      );

      const recoverableItem = find(
        // 需要将最近七天和选中日期的可恢复时间 item 合并
        [...recoverableList, ...recoverableListForDate],
        item => {
          // 开始时间的微秒值
          const startMicroseconds = getMicroseconds(item?.start_time);

          // 结束时间的微秒值
          const endMicroseconds = getMicroseconds(item?.end_time);

          // 去掉微秒值的开始时间
          const startTime = moment(moment(item?.start_time).format(DATE_TIME_FORMAT));

          // 去掉微秒值的结束时间
          const endTime = moment(moment(item?.end_time).format(DATE_TIME_FORMAT));

          // 未匹配的条件:
          if (
            // 秒数小于开始时间
            restoreEndTime.isBefore(startTime) ||
            // 秒数等于开始时间，但微秒值小于开始时间的微秒值
            (restoreEndTime.isSame(startTime) &&
              (newRestoreMicroseconds as number) < startMicroseconds) ||
            // 秒数大于结束时间
            restoreEndTime.isAfter(endTime) ||
            // 秒数等于开始时间，但微秒值大于结束时间的微秒值
            (restoreEndTime.isSame(endTime) && (newRestoreMicroseconds as number) > endMicroseconds)
          ) {
            return false;
          }
          return true;
        }
      );

      if (recoverableItem) {
        setCurrentRestoreDate(newRestoreDate);
        setCurrentRestoreTime(newRestoreTime);
        setCurrentRestoreMicroseconds(newRestoreMicroseconds);
      }
    }
  };

  return visible ? (
    <Result
      title={formatMessage({
        id: 'ocp-v2.Backup.Component.RestoreNow.TheRecoveryTaskIsInitiated',
        defaultMessage: '恢复任务发起成功',
      })}
      subTitle={formatMessage({
        id: 'ocp-v2.Backup.Component.RestoreNow.ItTakesSomeTimeTo',
        defaultMessage: '恢复任务需要一定的时间，可前往查看任务详情',
      })}
      extra={
        <Space>
          <Button
            type="primary"
            onClick={() => {
              history.push(`/task/${taskId}`);
            }}
          >
            {formatMessage({
              id: 'ocp-v2.Backup.Component.RestoreNow.ViewTaskDetails',
              defaultMessage: '查看任务详情',
            })}
          </Button>
          <Button
            onClick={() => {
              history.goBack();
            }}
          >
            {formatMessage({
              id: 'ocp-v2.Backup.Component.RestoreNow.Return',
              defaultMessage: '返回',
            })}
          </Button>
        </Space>
      }
    />
  ) : (
    <PageContainer
      ghost={true}
      header={{
        title,
        breadcrumb: { routes, itemRender: breadcrumbItemRender },
        onBack: () => history.goBack(),
      }}
      footer={[
        <Popconfirm
          placement="topRight"
          title={formatMessage({
            id: 'ocp-v2.Backup.Component.RestoreNow.AreYouSureYouWantToInitiateA',
            defaultMessage: '确定发起恢复吗？',
          })}
          onConfirm={() => {
            handleSubmit();
          }}
          okText={formatMessage({
            id: 'ocp-v2.component.Compaction.DumpPolicyDrawer.Ok',
            defaultMessage: '确定',
          })}
          cancelText={formatMessage({
            id: 'ocp-v2.component.Compaction.DumpPolicyDrawer.Cancel',
            defaultMessage: '取消',
          })}
        >
          <Button loading={tenantRestoreLoading} type="primary">
            {formatMessage({
              id: 'ocp-v2.Backup.Component.RestoreNow.InitiateRecovery',
              defaultMessage: '发起恢复',
            })}
          </Button>
        </Popconfirm>,
        isGlobalRestore && (
          <Button
            key="check"
            loading={tenantRestoreLoading}
            onClick={() => {
              handleCheck();
            }}
          >
            {formatMessage({
              id: 'OBShell.Backup.Component.RestoreNow.Test',
              defaultMessage: '测试',
            })}
          </Button>
        ),
        <Button
          key="Cancel"
          onClick={() => {
            history.goBack();
          }}
        >
          {formatMessage({
            id: 'ocp-v2.Backup.Component.RestoreNow.Cancel',
            defaultMessage: '取消',
          })}
        </Button>,
      ]}
    >
      <Row gutter={[0, 16]}>
        {isGlobalRestore && (
          <Col span={24}>
            <Form
              form={form}
              layout="vertical"
              colon={false}
              hideRequiredMark={true}
              onValuesChange={() => {
                setCheckedStatus('unchecked');
              }}
            >
              <StorageConfigCard
                form={form}
                useType={'restore'}
                obVersion={minObBuildVersion}
                checkedStatus={checkedStatus}
                loading={getRestoreSourceTenantInfoLoading}
                recoveryCheckResult={recoveryCheckResult}
              />
            </Form>
          </Col>
        )}

        <Col span={24}>
          <Spin spinning={getRestoreSourceTenantInfoLoading}>
            <Card
              bordered={false}
              title={formatMessage({
                id: 'ocp-v2.Backup.Component.RestoreNow.RestoreSourceInformation',
                defaultMessage: '恢复源信息',
              })}
            >
              <Form form={form} layout="vertical" colon={false} hideRequiredMark={true}>
                <Row gutter={24}>
                  <Col span={8}>
                    {isGlobalRestore ? (
                      <Form.Item noStyle={true} shouldUpdate={true}>
                        {() => {
                          return (
                            <Form.Item
                              label={formatMessage({
                                id: 'ocp-v2.Backup.Component.RestoreNow.SourceTenant',
                                defaultMessage: '源租户',
                              })}
                              name="sourceTenantName"
                              validateFirst
                              rules={[
                                {
                                  required: showObAdminTip ? false : true,
                                  message: formatMessage({
                                    id: 'ocp-v2.Backup.Component.RestoreNow.SelectASourceTenant',
                                    defaultMessage: '请选择源租户',
                                  }),
                                },
                              ]}
                            >
                              {/* 由于恢复数据可能从存储目录解析而来，因此租户名和 obTenantId 均可能重复，但两者的组合是唯一的 */}
                              <MySelect
                                loading={getRestoreSourceTenantInfoLoading}
                                disabled={!isCopy && isGlobalRestore}
                                showSearch={true}
                                optionFilterProp="label"
                                optionLabelProp="label"
                                dropdownMatchSelectWidth={false}
                                onChange={() => {
                                  setFieldsValue({
                                    restoreDate: undefined,
                                    restoreTime: undefined,
                                    restoreMicroseconds: undefined,
                                  });
                                }}
                              >
                                {tenantList.map(item => {
                                  const value = item.tenant_name;
                                  return (
                                    <Option key={value} value={value} label={value}>
                                      <span>
                                        <Space>
                                          <span>{value}</span>
                                        </Space>
                                      </span>
                                    </Option>
                                  );
                                })}
                              </MySelect>
                            </Form.Item>
                          );
                        }}
                      </Form.Item>
                    ) : (
                      <Descriptions column={4} style={{ marginBottom: 24 }}>
                        <Descriptions.Item
                          label={formatMessage({
                            id: 'ocp-v2.Backup.Component.RestoreNow.SourceTenant',
                            defaultMessage: '源租户',
                          })}
                        >
                          {clusterData?.cluster_name
                            ? `${clusterData.cluster_name}/${tenantName}`
                            : tenantName}
                        </Descriptions.Item>
                      </Descriptions>
                    )}
                  </Col>
                </Row>
                <Row gutter={8}>
                  <Col span={6}>
                    <Form.Item
                      label={formatMessage({
                        id: 'ocp-v2.Backup.Component.RestoreNow.RecoveryDate',
                        defaultMessage: '恢复日期',
                      })}
                      style={{ marginBottom: 0 }}
                      name="restoreDate"
                      rules={[
                        {
                          required: showObAdminTip ? false : true,
                          message: formatMessage({
                            id: 'ocp-v2.Backup.Component.RestoreNow.SelectARecoveryDate',
                            defaultMessage: '请选择恢复日期',
                          }),
                        },
                      ]}
                    >
                      <DatePicker
                        allowClear={false}
                        format={DATE_FORMAT_DISPLAY}
                        onChange={value => {
                          setCurrentRestoreDate(value);
                          handleRestoreEndTimeChange(
                            value,
                            getFieldValue('restoreTime'),
                            restoreMicroseconds
                          );
                        }}
                        style={{ width: '100%' }}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={5}>
                    <Form.Item noStyle shouldUpdate>
                      {() => {
                        const currentRestoreDate = getFieldValue('restoreDate');
                        return (
                          <Form.Item
                            label={formatMessage({
                              id: 'ocp-v2.Backup.Component.RestoreNow.TimeEveryMinute',
                              defaultMessage: '时分秒',
                            })}
                            style={{ marginBottom: 0 }}
                            name="restoreTime"
                            rules={[
                              {
                                required: showObAdminTip ? false : true,
                                message: formatMessage({
                                  id: 'ocp-v2.Backup.Component.RestoreNow.SelectARecoveryTime',
                                  defaultMessage: '请选择恢复时间点',
                                }),
                              },
                            ]}
                          >
                            <TimePicker
                              allowClear={false}
                              format={TIME_FORMAT}
                              showNow={false}
                              disabled={isNullValue(currentRestoreDate)}
                              onChange={value => {
                                setCurrentRestoreTime(value);
                                handleRestoreEndTimeChange(
                                  currentRestoreDate,
                                  value,
                                  restoreMicroseconds
                                );
                              }}
                              style={{ width: '100%' }}
                            />
                          </Form.Item>
                        );
                      }}
                    </Form.Item>
                  </Col>
                  <Col span={4}>
                    <Form.Item noStyle shouldUpdate>
                      {() => {
                        const currentRestoreTime = getFieldValue('restoreTime');
                        return (
                          <Form.Item
                            label={
                              <ContentWithQuestion
                                content={formatMessage({
                                  id: 'ocp-v2.Backup.Component.RestoreNow.Microseconds',
                                  defaultMessage: '微秒',
                                })}
                                tooltip={{
                                  title: formatMessage({
                                    id: 'ocp-v2.Backup.Component.RestoreNow.YouCanEnterUpTo',
                                    defaultMessage: '微秒最多可输入 6 位数字，即 0 ~ 999999',
                                  }),
                                }}
                              />
                            }
                            style={{ marginBottom: 0 }}
                            name="restoreMicroseconds"
                            initialValue={0}
                            rules={[
                              {
                                required: showObAdminTip ? false : true,
                                message: formatMessage({
                                  id: 'ocp-v2.Backup.Component.RestoreNow.EnterMicroseconds',
                                  defaultMessage: '请输入微秒',
                                }),
                              },
                              {
                                pattern: /^\d{1,6}$/,
                                message: formatMessage({
                                  id: 'ocp-v2.Backup.Component.RestoreNow.MicrosecondsFormat',
                                  defaultMessage: '微秒只能输入1-6位数字',
                                }),
                              },
                            ]}
                          >
                            <Input
                              type="text"
                              disabled={isNullValue(currentRestoreTime)}
                              onChange={e => {
                                const value = e.target.value;
                                // 只允许输入数字，最多6位
                                if (/^\d{0,6}$/.test(value)) {
                                  const numValue = value === '' ? undefined : parseInt(value, 10);
                                  setCurrentRestoreMicroseconds(numValue);
                                  handleRestoreEndTimeChange(
                                    getFieldValue('restoreDate'),
                                    getFieldValue('restoreTime'),
                                    numValue
                                  );
                                }
                              }}
                              onFocus={e => {
                                const currentValue = getFieldValue('restoreMicroseconds');
                                if (currentValue !== undefined && currentValue !== null) {
                                  e.target.value = currentValue.toString();
                                }
                              }}
                              onBlur={e => {
                                const value = e.target.value;
                                // 失焦时补齐前导零到6位
                                if (value && value.length < 6) {
                                  const paddedValue = value.padStart(6, '0');
                                  e.target.value = paddedValue;
                                  const numValue = parseInt(paddedValue, 10);
                                  setCurrentRestoreMicroseconds(numValue);
                                  handleRestoreEndTimeChange(
                                    getFieldValue('restoreDate'),
                                    getFieldValue('restoreTime'),
                                    numValue
                                  );
                                }
                              }}
                              placeholder="000000"
                              maxLength={6}
                              style={{
                                width: '100%',
                                // 未选择时间时，对微秒置灰处理
                                ...(restoreTime ? {} : { color: token.colorTextQuaternary }),
                              }}
                            />
                          </Form.Item>
                        );
                      }}
                    </Form.Item>
                  </Col>
                  <Col span={9}>
                    <Form.Item label=" " style={{ marginBottom: 0 }}>
                      <span style={{ color: token.colorTextTertiary }}>
                        {formatMessage(
                          {
                            id: 'ocp-v2.Backup.Component.RestoreNow.SelectedRecoveryTimeSelectedrestoredatetime',
                            defaultMessage: '所选恢复时刻：{selectedRestoreDateTime}',
                          },

                          { selectedRestoreDateTime }
                        )}
                      </span>
                    </Form.Item>
                  </Col>
                </Row>
                {tenantParseInfoForDate?.tenant_name || showObAdminTip ? (
                  <div style={{ color: token.colorTextTertiary }}>
                    <div>
                      <Space>
                        <InfoCircleFilled style={{ color: token.colorPrimary }} />
                        <span>
                          {showObAdminTip ? (
                            <>
                              {formatMessage({
                                id: 'ocp-v2.Backup.Component.RestoreNow.PleaseInstallObAdmin',
                                defaultMessage:
                                  '若需获取可恢复时间范围，请安装工具集成包：oceanbase-ce-utils, 以便系统解析备份集相关信息',
                              })}
                              <div>
                                {formatMessage({
                                  id: 'OBShell.Backup.Component.RestoreNow.IfTheRecoveryTimeIs',
                                  defaultMessage: '若未指定恢复时间，默认恢复至最新位点',
                                })}
                              </div>
                            </>
                          ) : recoverableList.length > 0 ? (
                            <>
                              {formatMessage(
                                {
                                  id: 'ocp-v2.Backup.Component.RestoreNow.BackuptenantlabelTheRecoverableTimeIn',
                                  defaultMessage:
                                    '{backupTenantLabel} 最近 {backupDays} 天的可恢复时间：',
                                },

                                {
                                  backupTenantLabel,
                                  backupDays: TimeScopeType.days,
                                }
                              )}
                              {getLatestRecoverableRangeTime(recoverableList)}
                            </>
                          ) : (
                            formatMessage(
                              {
                                id: 'ocp-v2.Backup.Component.RestoreNow.BackuptenantlabelNoRecoverableTimeIn',
                                defaultMessage:
                                  '{backupTenantLabel} 最近 {backupDays} 天无可恢复时间',
                              },

                              {
                                backupTenantLabel,
                                backupDays: TimeScopeType.days,
                              }
                            )
                          )}
                        </span>
                        {recoverableList.length > 1 && (
                          <Popover
                            placement="topRight"
                            // .ant-popover 设置了全局的最大宽度为 300px，如果要定制 Popover 宽度，则需要关闭 maxWidth
                            overlayStyle={{ maxWidth: 'none', width: 620 }}
                            content={getRecoverableRangeTime(recoverableList)}
                          >
                            {/* 左侧存在 tag，tag 本身就带有 margin-right，加上 Space 的 margin，这里需要设置 -margin */}
                            <a style={{ marginLeft: -8 }}>
                              {formatMessage({
                                id: 'ocp-v2.Backup.Component.RestoreNow.MoreTime',
                                defaultMessage: '更多时间',
                              })}
                            </a>
                          </Popover>
                        )}
                      </Space>
                    </div>
                    <div>
                      {tenantParseInfoForDate?.tenant_name && getFieldValue('restoreDate') && (
                        <>
                          {recoverableListForDate.length > 0 ? ( // 选中日期存在可恢复时间区间
                            <Space>
                              <InfoCircleFilled style={{ color: token.colorPrimary }} />
                              <span>
                                {formatMessage(
                                  {
                                    id: 'ocp-v2.Backup.Component.RestoreNow.TheRecoverableTimeOfBackuptenantlabel',
                                    defaultMessage:
                                      '{backupTenantLabel} 在 {selectedRestoreDate} 的可恢复时间：',
                                  },

                                  {
                                    backupTenantLabel,
                                    selectedRestoreDate,
                                  }
                                )}
                                {getLatestRecoverableRangeTime(
                                  recoverableListForDate,
                                  'HH:mm:ss.SSSSSS' // 使用包含微秒的格式
                                )}
                              </span>
                              {recoverableListForDate.length > 1 && (
                                <Popover
                                  // .ant-popover 设置了全局的最大宽度为 300px，如果要定制 Popover 宽度，则需要关闭 maxWidth
                                  overlayStyle={{ maxWidth: 'none', width: 400 }}
                                  content={getRecoverableRangeTime(
                                    recoverableListForDate,
                                    // 在选中日期内的可恢复时间，只展示时间即可，不需要展示日期
                                    'HH:mm:ss.SSSSSS' // 使用包含微秒的格式
                                  )}
                                >
                                  {/* 左侧存在 tag，tag 本身就带有 margin-right，加上 Space 的 margin，这里需要设置 -margin */}
                                  <a style={{ marginLeft: -8 }}>
                                    {formatMessage({
                                      id: 'ocp-v2.Backup.Component.RestoreNow.MoreTime',
                                      defaultMessage: '更多时间',
                                    })}
                                  </a>
                                </Popover>
                              )}
                            </Space>
                          ) : (
                            // 选中日期不存在可恢复时间区间
                            <Space style={{ color: token.colorError }}>
                              <ExclamationCircleFilled style={{ color: token.colorError }} />
                              {formatMessage(
                                {
                                  id: 'ocp-v2.Backup.Component.RestoreNow.BackuptenantlabelCannotBeRecoveredAt',
                                  defaultMessage:
                                    '{backupTenantLabel} 在 {selectedRestoreDate}无可恢复时间',
                                },

                                {
                                  backupTenantLabel,
                                  selectedRestoreDate,
                                }
                              )}
                            </Space>
                          )}
                        </>
                      )}
                    </div>
                  </div>
                ) : null}
              </Form>
            </Card>
          </Spin>
        </Col>
        <Col span={24}>
          <RestoreTenantCard
            ref={restoreTenantRef}
            form={form}
            useType={useType}
            backupCluster={clusterData}
            minObBuildVersion={minObBuildVersion}
          />
        </Col>
      </Row>
    </PageContainer>
  );
};

export default RestoreNow;
