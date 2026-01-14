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
import {
  Button,
  Col,
  Row,
  Badge,
  Space,
  Tag,
  Tooltip,
  Descriptions,
  Modal,
  message,
  Menu,
  Dropdown,
  BadgeProps,
} from '@oceanbase/design';
import { history, useDispatch, useSelector } from 'umi';
import React, { useState, useEffect } from 'react';
import { findByValue, isNullValue } from '@oceanbase/util';
import moment from 'moment';
import { PageContainer } from '@oceanbase/ui';
import { COMPACTION_STATUS_LISTV4 } from '@/constant/compaction';
import { getCompactionStatusV4 } from '@/util/cluster';
import { formatTime } from '@/util/datetime';
import { useRequest, useInterval } from 'ahooks';
import MyCard from '@/component/MyCard';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import ContentWithReload from '@/component/ContentWithReload';
import ModifyWhitelistModal from '../Component/ModifyWhitelistModal';
import {
  getCompaction,
  majorCompaction,
  clearCompactionError,
  startSeekdb,
  stopSeekdb,
  restartSeekdb,
} from '@/service/obshell/seekdb';
import InstanceConnectionString from '../Component/InstanceConnectionString';
import { EditOutlined, EllipsisOutlined, QuestionCircleOutlined } from '@oceanbase/icons';
import UpgradeAgentDrawer from '@/page/Cluster/Overview/UpgradeAgentDrawer';
import { OB_INFO_STATUS_LIST } from '@/constant/tenant';
import useUiMode from '@/hook/useUiMode';
import { getInstanceAvailableFlag } from '@/util/instance';
import useSeekdbInfo from '@/hook/usSeekdbInfo';
interface NewProps {}

const Detail: React.FC<NewProps> = ({}) => {
  const dispatch = useDispatch();
  const { seekdbInfoData: seekdbInfo } = useSeekdbInfo();
  const loading = useSelector(
    (state: DefaultRootState) => state.loading.effects['global/getSeekdbInfoData']
  );
  const { isDesktopMode } = useUiMode();
  const [modal, contextHolder] = Modal.useModal();
  const [showWhitelistModal, setShowWhitelistModal] = useState(false);
  const [upgradeVisible, setUpgradeVisible] = useState<boolean>(false);
  const [upgradeAgentVisible, setUpgradeAgentVisible] = useState<boolean>(false);

  const getSeekdbInfoData = () => {
    dispatch({
      type: 'global/getSeekdbInfoData',
      payload: {},
    });
  };

  useEffect(() => {
    getSeekdbInfoData();
  }, []);

  const seekdbInfoData: API.ObserverInfo = {
    ...(seekdbInfo || {}),
    name: seekdbInfo?.cluster_name,
  };

  const seekdbInfoStatus = seekdbInfoData?.status || '';
  const seekdbInfoStatusInfo = OB_INFO_STATUS_LIST.find(item => item.value === seekdbInfoStatus);
  const seekdbInfoStatusName = seekdbInfoStatusInfo?.label;
  const availableFlag = getInstanceAvailableFlag(seekdbInfoStatus);
  const systemdUnit = seekdbInfoData?.systemd_unit || '';

  const seekdbInfoPolling = ['STARTING', 'STOPPING', 'RESTARTING'].includes(seekdbInfoStatus);
  useInterval(
    () => {
      getSeekdbInfoData();
    },
    seekdbInfoPolling ? 3000 : undefined
  );

  // 格式化在线时长，去掉秒数的小数部分
  const formatLifeTime = (lifeTime?: string) => {
    if (!lifeTime) return '-';
    // 匹配秒数部分，将小数去掉，如 24.791556444s -> 24s
    return lifeTime.replace(/(\d+)\.\d+s/, '$1s');
  };

  // 获取租户合并详情
  const { data, refresh: getTenantCompactionRefresh } = useRequest(getCompaction, {
    defaultParams: [{}],
    ready: !!availableFlag,
  });

  const preTenantCompaction: API.TenantCompaction = data?.data || {};

  const tenantCompaction = {
    ...preTenantCompaction,
    startTime: preTenantCompaction?.start_time,
    lastFinishTime: preTenantCompaction?.last_finish_time,
    broadcastScn: preTenantCompaction?.global_broadcast_scn,
    frozenScn: preTenantCompaction?.frozen_scn,
  };

  const compactionStatus = tenantCompaction ? getCompactionStatusV4([tenantCompaction]) : null;
  const statusItem = findByValue(COMPACTION_STATUS_LISTV4, compactionStatus);
  // 最近一次合并的起止时间
  const startTime = tenantCompaction?.startTime;
  // 最近 1 次合并时间如果晚于开始时间，则说明处于新的合并中，结束时间为空
  const endTime = moment(tenantCompaction?.lastFinishTime).isAfter(
    moment(tenantCompaction?.startTime)
  )
    ? tenantCompaction?.lastFinishTime
    : undefined;

  // 合并中或等待合并调度中，进行轮询
  const compactPolling = compactionStatus === 'COMPACTING' || compactionStatus === 'WAIT_MERGE';
  useInterval(
    () => {
      getTenantCompactionRefresh();
    },
    compactPolling ? 3000 : undefined
  );

  // 发起合并
  const { run: triggerTenantCompaction } = useRequest(majorCompaction, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.Detail.Overview.MergeInitiatedSuccessfully',
            defaultMessage: '合并发起成功',
          })
        );

        getTenantCompactionRefresh();
      }
    },
  });

  // 清除租户合并异常
  const { run: clearCompactionErrorRun, loading: clearLoading } = useRequest(clearCompactionError, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.Detail.Overview.TheMergeExceptionIsClearedTheMergeWill',
            defaultMessage: '合并异常清除成功，将继续进行合并',
          })
        );

        getTenantCompactionRefresh();
      }
    },
  });

  const { run: startFn } = useRequest(startSeekdb, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        getSeekdbInfoData();
        Modal.success({
          title: formatMessage({
            id: 'ocp-v2.Cluster.Overview.TheTaskToStartThe',
            defaultMessage: '启动实例的任务提交成功',
          }),

          content: (
            <>
              {formatMessage({
                id: 'ocp-v2.Detail.Overview.ATaskHasBeenGenerated',
                defaultMessage: '已生成任务，任务 ID:',
              })}

              <a
                onClick={() => {
                  history.push(`/task/${res?.data?.id}`);
                  Modal.destroyAll();
                }}
              >
                {res?.data?.id}
              </a>
            </>
          ),
        });
      }
    },
  });

  const { run: stopFn } = useRequest(stopSeekdb, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        getSeekdbInfoData();
        Modal.success({
          title: formatMessage({
            id: 'ocp-v2.Cluster.Overview.TheTaskOfStoppingThe',
            defaultMessage: '停止实例的任务提交成功',
          }),

          content: (
            <>
              {formatMessage({
                id: 'ocp-v2.Detail.Overview.ATaskHasBeenGenerated',
                defaultMessage: '已生成任务，任务 ID:',
              })}

              <a
                onClick={() => {
                  history.push(`/task/${res?.data?.id}`);
                  Modal.destroyAll();
                  // directTo(`/task/${res?.data?.id}?backUrl=/task`);
                }}
              >
                {res?.data?.id}
              </a>
            </>
          ),
        });
      }
    },
  });

  const { run: restartFn } = useRequest(restartSeekdb, {
    manual: true,
    onSuccess: res => {
      getSeekdbInfoData();
      if (res.successful) {
        Modal.success({
          title: formatMessage({
            id: 'SeekDB.Detail.Overview.TheTaskOfRestartingThe',
            defaultMessage: '重启实例的任务提交成功',
          }),

          content: (
            <>
              {formatMessage({
                id: 'ocp-v2.Detail.Overview.ATaskHasBeenGenerated',
                defaultMessage: '已生成任务，任务 ID:',
              })}

              <a
                onClick={() => {
                  history.push(`/task/${res?.data?.id}`);
                  Modal.destroyAll();
                  // directTo(`/task/${res?.data?.id}?backUrl=/task`);
                }}
              >
                {res?.data?.id}
              </a>
            </>
          ),
        });
      }
    },
  });

  const getUnitConfig = () => {
    if (!seekdbInfoData || !seekdbInfoData?.cpu_count || !seekdbInfoData?.memory_size) {
      return '-';
    }
    const cpuCount = seekdbInfoData?.cpu_count;
    const memorySize = seekdbInfoData?.memory_size;
    // 去掉内存大小的小数部分和单位前的空格，如 12.80 GB -> 12GB
    const formattedMemorySize = memorySize.replace(/(\d+)\.\d+\s*/, '$1');
    return `${cpuCount}C ${formattedMemorySize}`;
  };

  function handleMenuClick(key: string) {
    if (key === 'upgrade') {
      setUpgradeVisible(true);
    } else if (key === 'upgradeAgent') {
      setUpgradeAgentVisible(true);
    } else if (key === 'start') {
      modal.confirm({
        title: formatMessage({
          id: 'SeekDB.Detail.Overview.AreYouSureYouWant',
          defaultMessage: '确定要启动实例吗？',
        }),
        okText: formatMessage({ id: 'SeekDB.Detail.Overview.Start', defaultMessage: '启动' }),
        width: 520,
        onOk: () => {
          startFn();
        },
      });
    } else if (key === 'stop') {
      modal.confirm({
        title: formatMessage({
          id: 'SeekDB.Detail.Overview.AreYouSureYouWant.1',
          defaultMessage: '确定要停止实例吗？',
        }),
        content: formatMessage({
          id: 'SeekDB.Detail.Overview.StoppingTheInstanceWillCause',
          defaultMessage: '停止实例将导致业务中断，请确认相关风险后再执行操作。',
        }),
        okText: formatMessage({ id: 'SeekDB.Detail.Overview.Stop', defaultMessage: '停止' }),
        okType: 'danger',
        width: 520,
        onOk: () => {
          stopFn({
            terminate: true,
          });
        },
      });
    } else if (key === 'restart') {
      modal.confirm({
        title: formatMessage({
          id: 'SeekDB.Detail.Overview.AreYouSureYouWant.2',
          defaultMessage: '确定要重启实例吗？',
        }),
        content: formatMessage({
          id: 'SeekDB.Detail.Overview.RestartingTheInstanceWillCause',
          defaultMessage: '重启实例将导致业务中断，请确认相关风险后再执行操作。',
        }),
        okText: formatMessage({ id: 'SeekDB.Detail.Overview.Restart', defaultMessage: '重启' }),
        okType: 'danger',
        width: 520,
        onOk: () => {
          restartFn({
            terminate: true,
          });
        },
      });
    }
  }

  const tooltipTitle = formatMessage(
    {
      id: 'SeekDB.Detail.Overview.InstanceSeekdbinfostatusnameThisOperationIs',
      defaultMessage: '实例{seekdbInfoStatusName}，暂不支持此操作',
    },
    { seekdbInfoStatusName: seekdbInfoStatusName }
  );

  const menu = (
    <Menu onClick={({ key }) => handleMenuClick(key)}>
      {/* <Menu.Item key="upgrade">
      <span>
       {formatMessage({
         id: 'ocp-v2.Cluster.Overview.UpgradeVersion',
         defaultMessage: '升级版本',
       })}
      </span>
      </Menu.Item> */}
      <Menu.Item key="upgradeAgent" disabled={seekdbInfoStatus !== 'AVAILABLE'}>
        <Tooltip placement="left" title={seekdbInfoStatus !== 'AVAILABLE' && tooltipTitle}>
          <span>
            {formatMessage({
              id: 'ocp-v2.Cluster.Overview.UpgradeObshell',
              defaultMessage: '升级 obshell',
            })}
          </span>
        </Tooltip>
      </Menu.Item>
    </Menu>
  );

  const renderClickableCount = (value: number | undefined, href: string) => {
    if (isNullValue(value)) {
      return '-';
    }
    if (availableFlag) {
      return (
        <a
          onClick={() => {
            history.push(href);
          }}
        >
          {value}
        </a>
      );
    }
    return value;
  };

  return (
    <PageContainer
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'SeekDB.Detail.Overview.InstanceManagement',
              defaultMessage: '实例管理',
            })}
            spin={loading}
            onClick={() => {
              getSeekdbInfoData();
              getTenantCompactionRefresh();
            }}
          />
        ),

        extra: (
          <Space>
            {!isDesktopMode &&
              ['AVAILABLE', 'UNAVAILABLE', 'STOPPING', 'RESTARTING'].includes(seekdbInfoStatus) && (
                <Tooltip
                  title={
                    (['STOPPING', 'RESTARTING'].includes(seekdbInfoStatus) && tooltipTitle) ||
                    (!!systemdUnit &&
                      formatMessage({
                        id: 'SeekDB.Detail.Overview.TheSeekdbInstanceIsRunning',
                        defaultMessage:
                          '当前 seekdb 实例运行在 systemd 管理下，如需停止服务，请执行 systemctl stop seekdb 命令。',
                      }))
                  }
                >
                  <Button
                    disabled={
                      ['STOPPING', 'RESTARTING'].includes(seekdbInfoStatus) || !!systemdUnit
                    }
                    onClick={() => handleMenuClick('stop')}
                  >
                    {formatMessage({ id: 'SeekDB.Detail.Overview.Stop', defaultMessage: '停止' })}
                  </Button>
                </Tooltip>
              )}
            {!isDesktopMode && ['STARTING', 'STOPPED'].includes(seekdbInfoStatus) && (
              <Tooltip title={['STARTING'].includes(seekdbInfoStatus) && tooltipTitle}>
                <Button
                  disabled={['STARTING'].includes(seekdbInfoStatus)}
                  onClick={() => {
                    handleMenuClick('start');
                  }}
                >
                  {formatMessage({ id: 'SeekDB.Detail.Overview.Start', defaultMessage: '启动' })}
                </Button>
              </Tooltip>
            )}
            {!isDesktopMode && (
              <Tooltip
                title={
                  ['RESTARTING', 'STOPPING', 'STARTING', 'STOPPED'].includes(seekdbInfoStatus) &&
                  tooltipTitle
                }
              >
                <Button
                  disabled={['RESTARTING', 'STOPPING', 'STARTING', 'STOPPED'].includes(
                    seekdbInfoStatus
                  )}
                  onClick={() => {
                    handleMenuClick('restart');
                  }}
                >
                  {formatMessage({ id: 'SeekDB.Detail.Overview.Restart', defaultMessage: '重启' })}
                </Button>
              </Tooltip>
            )}
            <Dropdown overlay={menu} getPopupContainer={() => document.body}>
              <Button>
                <EllipsisOutlined />
              </Button>
            </Dropdown>
            {contextHolder}
          </Space>
        ),
      }}
    >
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <MyCard
            title={
              <span style={{ marginLeft: 2 }}>
                {formatMessage({
                  id: 'ocp-express.Detail.Overview.BasicInformation',
                  defaultMessage: '基本信息',
                })}
              </span>
            }
          >
            <Descriptions column={4}>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.Status',
                  defaultMessage: '状态',
                })}
              >
                <Badge
                  status={seekdbInfoStatusInfo?.badgeStatus as BadgeProps['status']}
                  text={seekdbInfoStatusName}
                />

                {['STARTING', 'STOPPING', 'RESTARTING'].includes(seekdbInfoStatus) && (
                  <Tooltip
                    title={formatMessage({
                      id: 'SeekDB.Detail.Overview.CanEnterTheTaskCenter',
                      defaultMessage: '可进入任务中心查看',
                    })}
                  >
                    <QuestionCircleOutlined style={{ marginLeft: 8, cursor: 'pointer' }} />
                  </Tooltip>
                )}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.SeekdbVersionNumber',
                  defaultMessage: 'seekdb 版本号',
                })}
              >
                {seekdbInfoData?.version || '-'}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.Specifications',
                  defaultMessage: '规格',
                })}
              >
                {getUnitConfig()}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.HardwareArchitecture',
                  defaultMessage: '硬件架构',
                })}
              >
                {seekdbInfoData?.architecture}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.NumberOfDatabases',
                  defaultMessage: '数据库数',
                })}
              >
                {renderClickableCount(seekdbInfoData?.database_count, `/database`)}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.NumberOfUsers',
                  defaultMessage: '用户数',
                })}
              >
                {renderClickableCount(seekdbInfoData?.user_count, `/user`)}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.SqlPort',
                  defaultMessage: 'SQL 端口',
                })}
              >
                {seekdbInfoData?.port}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.ObshellPort',
                  defaultMessage: 'obshell 端口',
                })}
              >
                {seekdbInfoData?.obshell_port}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'ocp-express.Tenant.Detail.CreationTime',
                  defaultMessage: '创建时间',
                })}
              >
                {formatTime(seekdbInfoData.created_time)}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.Creator',
                  defaultMessage: '创建者',
                })}
              >
                {seekdbInfoData?.user}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.StartUpTime',
                  defaultMessage: '启动时间',
                })}
              >
                {formatTime(seekdbInfoData?.start_time)}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.OnlineDuration',
                  defaultMessage: '在线时长',
                })}
              >
                {formatLifeTime(seekdbInfoData?.life_time)}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.WorkingDirectory',
                  defaultMessage: '工作目录',
                })}
              >
                {seekdbInfoData?.base_dir || '-'}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.DataDiskPath',
                  defaultMessage: '数据盘路径',
                })}
              >
                {seekdbInfoData?.data_dir || '-'}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.LogDiskPath',
                  defaultMessage: '日志盘路径',
                })}
              >
                {seekdbInfoData?.redo_dir || '-'}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.LogDirectory',
                  defaultMessage: '日志目录',
                })}
              >
                {seekdbInfoData?.log_dir || '-'}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'SeekDB.Detail.Overview.SoftwareInstallationPath',
                  defaultMessage: '软件安装路径',
                })}
              >
                {seekdbInfoData?.bin_path || '-'}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'ocp-express.component.TenantList.ConnectionString',
                  defaultMessage: '连接字符串',
                })}
              >
                <InstanceConnectionString
                  connectionString={seekdbInfoData?.connection_string || '-'}
                />
              </Descriptions.Item>
            </Descriptions>
          </MyCard>
        </Col>
        {availableFlag && (
          <>
            <Col span={24}>
              <MyCard
                bordered={false}
                title={
                  <ContentWithQuestion
                    content={formatMessage({
                      id: 'ocp-express.Tenant.Detail.Whitelist',
                      defaultMessage: '白名单',
                    })}
                    tooltip={{
                      title: formatMessage({
                        id: 'ocp-express.Tenant.Detail.ListOfAddressesThatCan',
                        defaultMessage: '能够连接到实例的地址列表',
                      }),
                    }}
                  />
                }
              >
                <div style={{ display: 'flex', alignItems: 'center' }}>
                  {seekdbInfoData.whitelist &&
                    seekdbInfoData.whitelist.split(',').map(item => <Tag key={item}>{item}</Tag>)}
                  <Button
                    type="link"
                    icon={<EditOutlined />}
                    onClick={() => {
                      setShowWhitelistModal(true);
                    }}
                  />
                </div>
              </MyCard>
            </Col>
            <Col span={24}>
              <MyCard
                bordered={false}
                title={formatMessage({
                  id: 'ocp-express.Detail.Overview.MergeManagement',
                  defaultMessage: '合并管理',
                })}
              >
                <Descriptions>
                  <Descriptions.Item
                    label={formatMessage({
                      id: 'ocp-express.Compaction.Detail.State',
                      defaultMessage: '状态',
                    })}
                  >
                    {/**
                     * 状态对应的操作
                     * 租户合并中：  查看租户
                     * ERROR 租户合并异常： 查询租户 清除异常
                     * IDLE 租户空闲：发起合并
                     * */}
                    <Badge
                      status={statusItem.badgeStatus}
                      text={statusItem.label}
                      style={{ marginRight: 8 }}
                    />

                    {statusItem.value === 'ERROR' && (
                      <Tooltip
                        title={formatMessage({
                          id: 'ocp-express.Detail.Overview.TheFollowingTenantMergeHasAnExceptionAfter',
                          defaultMessage:
                            '合并存在异常，异常解决后，再点击「清除异常」，操作后系统会继续合并',
                        })}
                      >
                        <Button
                          type="link"
                          loading={clearLoading}
                          onClick={() => {
                            clearCompactionErrorRun();
                          }}
                        >
                          {formatMessage({
                            id: 'ocp-express.Compaction.Detail.ClearException',
                            defaultMessage: '清除异常',
                          })}
                        </Button>
                      </Tooltip>
                    )}

                    {/* 合并状态为空(说明从未发起过合并)，或者空闲状态，才能发起合并 */}
                    {statusItem.value === 'IDLE' && (
                      <a
                        onClick={() => {
                          Modal.confirm({
                            title: formatMessage({
                              id: 'ocp-express.Compaction.Detail.AreYouSureYouWant',
                              defaultMessage: '确定要发起合并吗？',
                            }),

                            onOk: () => {
                              return triggerTenantCompaction();
                            },
                          });
                        }}
                      >
                        {formatMessage({
                          id: 'ocp-express.Compaction.Detail.InitiateAMerge',
                          defaultMessage: '发起合并',
                        })}
                      </a>
                    )}
                  </Descriptions.Item>
                  <Descriptions.Item
                    label={formatMessage({
                      id: 'ocp-express.Compaction.Detail.StartTimeOfTheLastMerge',
                      defaultMessage: '最近 1 次合并开始时间',
                    })}
                  >
                    {formatTime(startTime)}
                  </Descriptions.Item>
                  <Descriptions.Item
                    label={formatMessage({
                      id: 'ocp-express.Compaction.Detail.LastMergeEndTime',
                      defaultMessage: '最近 1 次合并结束时间',
                    })}
                  >
                    {formatTime(endTime)}
                  </Descriptions.Item>
                </Descriptions>
              </MyCard>
            </Col>
          </>
        )}
      </Row>

      <ModifyWhitelistModal
        visible={showWhitelistModal}
        seekdbInfoData={seekdbInfoData}
        onCancel={() => {
          setShowWhitelistModal(false);
        }}
        onSuccess={() => {
          setShowWhitelistModal(false);
          getSeekdbInfoData();
        }}
      />

      {/* <UpgradeDrawer
           visible={upgradeVisible}
           onCancel={() => setUpgradeVisible(false)}
           onSuccess={() => {
             setUpgradeVisible(false);
           }}
          /> */}

      <UpgradeAgentDrawer
        visible={upgradeAgentVisible}
        seekdbInfoData={seekdbInfoData}
        onCancel={() => setUpgradeAgentVisible(false)}
        onSuccess={() => {
          setUpgradeAgentVisible(false);
        }}
      />
    </PageContainer>
  );
};
export default Detail;
