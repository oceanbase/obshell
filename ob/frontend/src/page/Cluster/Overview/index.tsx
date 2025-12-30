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
import { history } from 'umi';
import React, { useRef, useState } from 'react';
import {
  Space,
  Tag,
  Col,
  Row,
  Badge,
  Button,
  theme,
  Dropdown,
  Menu,
  Modal,
} from '@oceanbase/design';
import { flatten, isEmpty, reduce } from 'lodash';
import { PageContainer } from '@oceanbase/ui';
import useReload from '@/hook/useReload';
import useDocumentTitle from '@/hook/useDocumentTitle';
import ContentWithReload from '@/component/ContentWithReload';
import MyCard from '@/component/MyCard';
import ClusterInfo from './ClusterInfo';
import CompactionTimeTop3 from './CompactionTimeTop3';
import TenantResourceTop3 from './TenantResourceTop3';
import type { ZoneListOrTopoRef } from './ZoneListOrTopo';
import ZoneListOrTopo from './ZoneListOrTopo';
import { EllipsisOutlined } from '@oceanbase/icons';
import UpgradeDrawer from './UpgradeDrawer';
import { useRafInterval, useRequest } from 'ahooks';
import { obStart, obStop } from '@/service/obshell/ob';
import { obclusterInfo, obclusterSetParameters } from '@/service/obshell/obcluster';
import UpgradeAgentDrawer from './UpgradeAgentDrawer';
import moment from 'moment';
import { getAgentMainDags } from '@/service/obshell/task';
import { message } from 'antd';
import useUiMode from '@/hook/useUiMode';
import LicenseDetailModal from './License/LicenseDetailModal';
import LicenseAlert from './License/LicenseAlert';

export const getServerStatus = (server: API.Observer) => {
  let status = 'OTHER';
  if (!server) return status;

  if (
    server.inner_status === 'ACTIVE' &&
    // 未开始时 start_time 的时间戳等于 0（1970年01月01日00时00分00），所以判断时间戳大于 0
    moment(server.start_time).valueOf() > 0
  ) {
    status = 'RUNNING';
  }
  if (server.inner_status === 'INACTIVE') {
    status = 'UNAVAILABLE';
  }

  return status;
};

export interface DetailProps {}

const Detail: React.FC<DetailProps> = ({}) => {
  const { token } = theme.useToken();
  const [clusterStatus, setClusterStatus] = useState<'normal' | 'abnormal'>('normal');

  const isAbnormal = clusterStatus === 'abnormal';
  const { isDesktopMode } = useUiMode();

  useDocumentTitle(
    formatMessage({ id: 'ocp-v2.Cluster.Overview.ClusterManagement', defaultMessage: '集群管理' })
  );

  const {
    data: obclusterInfoRes,
    refresh,
    loading: obclusterInfoLoading,
  } = useRequest(
    () =>
      obclusterInfo({
        HIDE_ERROR_MESSAGE: true,
      }),
    {
      onSuccess: res => {
        if (res.successful) {
          if (isAbnormal) {
            setClusterStatus('normal');
            reload();
          }
        } else {
          setClusterStatus('abnormal');
        }
      },
    }
  );

  useRafInterval(
    () => {
      refresh();
    },
    isAbnormal ? 3000 : undefined
  );

  const clusterData = obclusterInfoRes?.data || {};

  const [reloading, reload] = useReload(false);
  const zoneListOrTopoRef = useRef<ZoneListOrTopoRef>();

  // 使用空字符串兜底，避免文案拼接时出现 undefined
  const obVersion = clusterData.ob_version || '';
  const deadLockDetectionEnabled = clusterData.dead_lock_detection_enabled;

  const realServerList = flatten(clusterData.zones?.map(item => item.servers || [])).map(item => {
    const status = getServerStatus(item);

    return { ...item, status };
  });

  const ResultStatusContent = ({ type }) => {
    const runningStatus = 'RUNNING';
    const normalStatus = 'NORMAL';
    const unavailableStatus = 'UNAVAILABLE';
    const data = type === 'tenant' ? clusterData?.tenants || [] : realServerList;

    const normalCount = data?.filter(item => item?.status === normalStatus)?.length || 0;
    const unavailableCount = data?.filter(item => item?.status === unavailableStatus)?.length || 0;
    const runningCount = data?.filter(item => item?.status === runningStatus)?.length || 0;

    const otherCount =
      reduce(
        data
          ?.filter(item => ![normalStatus, unavailableStatus, runningStatus].includes(item.status))
          .map(e => e.count),
        (sum, n) => sum + n
      ) || 0;

    return (
      <div style={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-around' }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
          }}
        >
          <Badge
            status="success"
            text={formatMessage({
              id: 'ocp-express.Cluster.Overview.Running',
              defaultMessage: '运行中',
            })}
            style={{
              marginRight: 24,
              color: token.colorTextTertiary,
            }}
          />

          <a
            onClick={() => {
              // 租户
              if (normalCount > 0) {
                history.push({
                  pathname: '/tenant',
                  query: {
                    status: normalStatus,
                  },
                });
              }
            }}
            style={
              (type === 'tenant' ? normalCount : runningCount) === 0
                ? { color: token.colorTextTertiary, cursor: 'default' }
                : {}
            }
          >
            {type === 'tenant' ? normalCount : runningCount}
          </a>
        </div>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
          }}
        >
          <Badge
            status="error"
            text={formatMessage({
              id: 'ocp-express.Cluster.Overview.Unavailable',
              defaultMessage: '不可用',
            })}
            style={{
              marginRight: 24,
              color: token.colorTextTertiary,
            }}
          />

          <a
            onClick={() => {
              if (unavailableCount > 0) {
                if (type === 'tenant') {
                  // 租户
                  history.push({
                    pathname: '/tenant',
                    query: {
                      status: unavailableStatus,
                    },
                  });
                }
              }
            }}
            style={
              unavailableCount === 0
                ? { color: token.colorTextTertiary, cursor: 'default' }
                : { color: '#ff4b4b' }
            }
          >
            {unavailableCount}
          </a>
        </div>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
          }}
        >
          <Badge
            status="default"
            text={formatMessage({
              id: 'ocp-express.Cluster.Overview.Other',
              defaultMessage: '其他',
            })}
            style={{
              marginRight: 24,
              color: token.colorTextTertiary,
            }}
          />

          <a
            onClick={() => {
              if (otherCount > 0) {
                if (type === 'tenant') {
                  // 租户
                  history.push({
                    pathname: '/tenant',
                    query: {
                      excludeStatus: [normalStatus, unavailableStatus].join(','),
                    },
                  });
                }
              }
            }}
            style={otherCount === 0 ? { color: token.colorTextTertiary, cursor: 'default' } : {}}
          >
            {otherCount}
          </a>
        </div>
      </div>
    );
  }; // 组装大盘类型数据
  const overviewStatusType = [
    {
      key: 'observer',
      title: formatMessage({
        id: 'ocp-express.Cluster.Overview.ObserverStatistics',
        defaultMessage: 'OBServer 统计',
      }),
      img: `/assets/server/running.svg`,
      content: <ResultStatusContent type={'OBServer'} />,
      totalCount: flatten(clusterData.zones?.map(item => item.servers || []))?.length,
    },
    {
      key: 'tenant',
      title: formatMessage({
        id: 'ocp-express.Cluster.Overview.TenantStatistics',
        defaultMessage: '租户统计',
      }),
      img: `/assets/overview/tenant.svg`,
      content: <ResultStatusContent type={'tenant'} />,
      totalCount: (clusterData?.tenants || [])?.length,
    },
  ];

  const [upgradeVisible, setUpgradeVisible] = useState<boolean>(false);
  const [upgradeAgentVisible, setUpgradeAgentVisible] = useState<boolean>(false);
  const [licenseDetailOpen, setLicenseDetailOpen] = useState(false);

  const { run: startFn } = useRequest(obStart, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        Modal.success({
          title: formatMessage({
            id: 'ocp-v2.Cluster.Overview.TheTaskToStartThe',
            defaultMessage: '启动集群的任务提交成功',
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

  const { run: stopFn } = useRequest(obStop, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        Modal.success({
          title: formatMessage({
            id: 'ocp-v2.Cluster.Overview.TheTaskOfStoppingThe',
            defaultMessage: '停止集群的任务提交成功',
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

  const { run: updateClusterParameter } = useRequest(obclusterSetParameters, {
    manual: true,
    onSuccess: (res, params) => {
      if (res?.successful) {
        refresh();
        if (params[0]?.params[0]?.value === '30ms') {
          message.success(
            formatMessage({
              id: 'OBShell.Cluster.Overview.DeadlockAutoDetectionIsOn',
              defaultMessage: '死锁自动检测已开启',
            })
          );
        } else {
          message.success(
            formatMessage({
              id: 'OBShell.Cluster.Overview.DeadlockAutoDetectionIsOff',
              defaultMessage: '死锁自动检测已关闭',
            })
          );
        }
      }
    },
  });

  function handleMenuClick(key: string) {
    if (key === 'upgrade') {
      setUpgradeVisible(true);
    } else if (key === 'upgradeAgent') {
      setUpgradeAgentVisible(true);
    } else if (key === 'start' || key === 'stop') {
      getAgentMainDags().then(res => {
        const dagList = res.data?.contents || [];

        if (dagList.some(item => item.state !== 'FAILED')) {
          message.warning(
            formatMessage({
              id: 'ocp-v2.Cluster.Overview.ThereIsAnOngoingCluster',
              defaultMessage: '当前有正在进行的集群启停任务，此时禁止继续启停',
            })
          );
          return;
        }

        let content = '';
        let forcePassDag = { id: [] };
        const list = dagList.filter(item => item.state === 'FAILED');
        if (list.length > 0) {
          const listString = list
            .map(item => {
              return `[${item.name}:${item.id}]`;
            })
            .join('，');

          content = formatMessage(
            {
              id: 'ocp-v2.Cluster.Overview.WillAutomaticallySkipTasksListstring',
              defaultMessage: '将会自动跳过任务 {listString}',
            },
            { listString: listString }
          );
          forcePassDag = { id: list.map(item => item.id) };
        }

        if (key === 'start') {
          const realClusterName = clusterData.cluster_name || '';
          Modal.confirm({
            title: formatMessage(
              {
                id: 'ocp-v2.Cluster.Overview.AreYouSureYouWant',
                defaultMessage: '确定要启动 OB 集群 {realClusterName} 吗？',
              },
              { realClusterName: realClusterName }
            ),
            content,
            okText: formatMessage({ id: 'ocp-v2.Cluster.Overview.Start', defaultMessage: '启动' }),
            onOk: () => {
              startFn({ scope: { type: 'GLOBAL' }, forcePassDag });
            },
          });
        } else {
          Modal.confirm({
            title: formatMessage(
              {
                id: 'ocp-express.Cluster.Detail.AreYouSureYouWant.1',
                defaultMessage: '确定要停止 OB 集群 {clusterDataName} 吗？',
              },

              { clusterDataName: clusterData.cluster_name || '' }
            ),

            content: (
              <>
                <div>
                  {formatMessage({
                    id: 'ocp-express.Cluster.Detail.StoppingTheClusterWillCause',
                    defaultMessage: '停止集群会导致集群中所有的服务被终止，请谨慎操作',
                  })}
                </div>
                <div>{content}</div>
              </>
            ),

            okText: formatMessage({
              id: 'ocp-express.Cluster.Detail.Stop',
              defaultMessage: '停止',
            }),
            okButtonProps: {
              danger: true,
              ghost: true,
            },

            onOk: () => {
              stopFn({
                terminate: true,
                scope: {
                  type: 'GLOBAL',
                },
                forcePassDag,
              });
            },
          });
        }
      });
    } else if (key === 'openAutoDeadlockDetection') {
      Modal.confirm({
        title: formatMessage({
          id: 'OBShell.Cluster.Overview.AreYouSureYouWant',
          defaultMessage: '确认要开启死锁自动检测吗？',
        }),

        content: formatMessage({
          id: 'OBShell.Cluster.Overview.ThisAutomaticallyDetectsAndResolves',
          defaultMessage:
            '本操作会自动检测并解决死锁，并保存 7 天的死锁记录。本操作为集群级别行为，其他租户也将同步开启，本功能对集群有 2% 左右的性能损耗，请谨慎操作。',
        }),

        onOk: () => {
          return updateClusterParameter({
            params: [{ name: '_lcl_op_interval', value: '30ms', scope: 'CLUSTER' }],
          });
        },
      });
    } else if (key === 'closeAutoDeadlockDetection') {
      Modal.confirm({
        title: formatMessage({
          id: 'OBShell.Cluster.Overview.AreYouSureYouWant.1',
          defaultMessage: '确认要关闭死锁自动检测吗？',
        }),
        content: formatMessage({
          id: 'OBShell.Cluster.Overview.AfterAutomaticDeadlockDetectionIs',
          defaultMessage: '关闭死锁自动检测后，不再自动检测和解决死锁。',
        }),

        onOk: () => {
          return updateClusterParameter({
            params: [{ name: '_lcl_op_interval', value: '0ms', scope: 'CLUSTER' }],
          });
        },
      });
    } else if (key === 'license') {
      setLicenseDetailOpen(true);
    }
  }

  // 计算是否为单机版
  const isStandalone = clusterData?.is_standalone || false;

  const menu = (
    <Menu onClick={({ key }) => handleMenuClick(key)}>
      <Menu.Item key="upgrade">
        <span>
          {formatMessage({
            id: 'ocp-v2.Cluster.Overview.UpgradeVersion',
            defaultMessage: '升级版本',
          })}
        </span>
      </Menu.Item>
      <Menu.Item key="upgradeAgent">
        <span>
          {formatMessage({
            id: 'ocp-v2.Cluster.Overview.UpgradeObshell',
            defaultMessage: '升级 obshell',
          })}
        </span>
      </Menu.Item>

      {deadLockDetectionEnabled === true && (
        <Menu.Item key="closeAutoDeadlockDetection">
          <span>
            {formatMessage({
              id: 'OBShell.Cluster.Overview.TurnOffAutomaticDeadlockDetection',
              defaultMessage: '关闭死锁自动检测',
            })}
          </span>
        </Menu.Item>
      )}
      {deadLockDetectionEnabled === false && (
        <Menu.Item key="openAutoDeadlockDetection">
          <span>
            {formatMessage({
              id: 'ocp-v2.Detail.Overview.EnableAutomaticDeadlockDetection',
              defaultMessage: '开启死锁自动检测',
            })}
          </span>
        </Menu.Item>
      )}
      {isStandalone && (
        <Menu.Item key="license">
          <span>
            {formatMessage({
              id: 'ocp-v2.Detail.Overview.ViewLicense',
              defaultMessage: '查看 License',
            })}
          </span>
        </Menu.Item>
      )}
    </Menu>
  );

  // 是否为单节点集群
  const isStandAloneCluster =
    clusterData?.zones?.length === 1 && clusterData?.zones?.[0]?.servers?.length === 1;

  return (
    <PageContainer
      loading={reloading || (!isAbnormal && obclusterInfoLoading)}
      ghost={true}
      header={{
        title: (
          <div>
            <span
              style={{
                marginRight: 12,
              }}
            >
              {formatMessage({
                id: 'ocp-v2.Cluster.Overview.ClusterManagement',
                defaultMessage: '集群管理',
              })}
            </span>
            <ContentWithReload
              content={
                <Tag
                  color={isAbnormal ? 'warning' : 'geekblue'}
                  style={{
                    fontWeight: 'normal',
                    lineHeight: '24px',
                    position: 'relative',
                    top: -4,
                  }}
                >
                  {isAbnormal
                    ? formatMessage({
                        id: 'ocp-v2.Cluster.Overview.ClusterStatusIsAbnormal',
                        defaultMessage: '集群状态异常',
                      })
                    : formatMessage(
                        {
                          id: 'ocp-express.Cluster.Overview.Obversion',
                          defaultMessage: '{obVersion} 版本',
                        },
                        { obVersion: obVersion }
                      )}
                </Tag>
              }
              spin={reloading}
              onClick={() => {
                refresh();
                reload();
              }}
            />
          </div>
        ),

        extra: (
          <Space>
            <Button
              onClick={() => {
                history.push('/overview/parameter');
              }}
            >
              {formatMessage({
                id: 'ocp-express.Cluster.Overview.ParameterManagement',
                defaultMessage: '参数管理',
              })}
            </Button>

            {!isDesktopMode && clusterData.status !== 'AVAILABLE' && (
              <Button
                onClick={() => {
                  handleMenuClick('start');
                }}
              >
                <span>
                  {formatMessage({
                    id: 'ocp-v2.Cluster.Overview.StartTheCluster',
                    defaultMessage: '启动集群',
                  })}
                </span>
              </Button>
            )}

            {!isDesktopMode && clusterData.status === 'AVAILABLE' && (
              <Button
                onClick={() => {
                  handleMenuClick('stop');
                }}
              >
                <span>
                  {formatMessage({
                    id: 'ocp-v2.Cluster.Overview.StopTheCluster',
                    defaultMessage: '停止集群',
                  })}
                </span>
              </Button>
            )}

            <Dropdown overlay={menu} getPopupContainer={() => document.body}>
              <Button>
                <EllipsisOutlined />
              </Button>
            </Dropdown>
          </Space>
        ),
      }}
    >
      <LicenseAlert clusterData={clusterData} />
      <Row gutter={[16, 16]}>
        <Col span={14}>
          <ClusterInfo clusterData={clusterData} />
        </Col>
        {overviewStatusType.map(item => {
          return (
            <Col key={item.key} span={5}>
              <MyCard
                title={
                  <div style={{ backgroundImage: item.img }}>
                    <span>{item.title}</span>
                  </div>
                }
                style={{ height: '100%' }}
                headStyle={{
                  marginBottom: 16,
                }}
                bodyStyle={{
                  padding: '16px 24px',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <div
                    style={{
                      fontSize: '28px',
                      fontFamily: 'Avenir-Heavy',
                      lineHeight: '66px',
                    }}
                  >
                    {item.totalCount}
                  </div>
                  {item.content}
                </div>
              </MyCard>
            </Col>
          );
        })}
        <Col span={24}>
          <CompactionTimeTop3 />
        </Col>
        <Col span={24}>
          <TenantResourceTop3 clusterData={clusterData} />
        </Col>
        <Col span={24}>
          <ZoneListOrTopo
            ref={zoneListOrTopoRef}
            clusterData={clusterData}
            serverList={realServerList}
          />
        </Col>
      </Row>

      <UpgradeDrawer
        // 单节点集群和空集群都只显示停服升级
        isStandAloneCluster={isStandAloneCluster || isEmpty(clusterData)}
        visible={upgradeVisible}
        clusterData={clusterData}
        onCancel={() => setUpgradeVisible(false)}
        onSuccess={() => {
          setUpgradeVisible(false);
        }}
      />

      <UpgradeAgentDrawer
        visible={upgradeAgentVisible}
        clusterData={clusterData}
        onCancel={() => setUpgradeAgentVisible(false)}
        onSuccess={() => {
          setUpgradeAgentVisible(false);
        }}
      />

      <LicenseDetailModal
        clusterData={clusterData}
        open={licenseDetailOpen}
        onCancel={() => setLicenseDetailOpen(false)}
      />
    </PageContainer>
  );
};

export default Detail;
