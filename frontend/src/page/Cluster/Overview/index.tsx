/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

import { formatMessage } from '@/util/intl';
import { connect, history, useDispatch } from 'umi';
import React, { useEffect, useRef, useState } from 'react';
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
  Checkbox,
} from '@oceanbase/design';
import { flatten, reduce } from 'lodash';
import { ContentWithIcon, PageContainer } from '@oceanbase/ui';
import scrollIntoView from 'scroll-into-view';
import useReload from '@/hook/useReload';
import useDocumentTitle from '@/hook/useDocumentTitle';
import ContentWithReload from '@/component/ContentWithReload';
import MyCard from '@/component/MyCard';
import ClusterInfo from './ClusterInfo';
import CompactionTimeTop3 from './CompactionTimeTop3';
import SlowSQLTop3 from './SlowSQLTop3';
import TenantResourceTop3 from './TenantResourceTop3';
import type { ZoneListOrTopoRef } from './ZoneListOrTopo';
import ZoneListOrTopo from './ZoneListOrTopo';
import { OB_SERVER_STATUS_LIST } from '@/constant/oceanbase';
import { EllipsisOutlined } from '@oceanbase/icons';
import RestartModal from './RestartModal';
import BinlogAssociationMsg from '@/component/BinlogAssociationMsg';
import UpgradeDrawer from './UpgradeDrawer';
import { useRequest } from 'ahooks';
import { getObInfo, obStart, obStop } from '@/service/obshell/ob';
import { message } from 'antd';
import { obclusterInfo } from '@/service/obshell/obcluster';
import UpgradeAgentDrawer from './UpgradeAgentDrawer';

export interface DetailProps {
  match: {
    params: {
      clusterId: number;
    };
  };
  loading: boolean;
  clusterData: API.ClusterInfo;
}

const Detail: React.FC<DetailProps> = ({
  match: {
    params: {},
  },
  loading,
}) => {
  const { token } = theme.useToken();

  const dispatch = useDispatch();
  useDocumentTitle(
    formatMessage({
      id: 'ocp-express.Cluster.Unit.ClusterOverview',
      defaultMessage: '集群总览',
    })
  );

  const { data: obclusterInfoRes } = useRequest(obclusterInfo);
  // const { data: obInfoRes } = useRequest(getObInfo);

  const clusterData = obclusterInfoRes?.data || {};

  // console.log(obclusterInfoData, 'obclusterInfoData');

  // const clusterData = obInfoRes?.data?.obcluster_info || {};

  console.log(clusterData, 'clusterData');
  const [reloading, reload] = useReload(false);
  const zoneListOrTopoRef = useRef<ZoneListOrTopoRef>();

  // 使用空字符串兜底，避免文案拼接时出现 undefined
  const obVersion = clusterData.ob_version || '';
  const ObTenantBinlogController = {};

  // 当前 OB 集群关联的 binlog 实例
  const { data } = useRequest(ObTenantBinlogController?.listObClusterRelatedBinlogService, {
    defaultParams: [
      {
        id: clusterData.cluster_id,
      },
    ],
  });

  const obClusterRelatedBinlogService = data?.data?.contents || [];

  const getClusterData = () => {
    dispatch({
      type: 'cluster/getClusterData',
      payload: {},
    });
  };

  useEffect(() => {
    getClusterData();
  }, []);

  const scrollToZoneTable = () => {
    const zoneTable = document.getElementById('ocp-zone-table');
    if (!zoneTable) {
      zoneListOrTopoRef.current?.setType('LIST');
    }
    setTimeout(() => {
      const newZoneTable = document.getElementById('ocp-zone-table');
      if (newZoneTable) {
        zoneListOrTopoRef.current?.expandAll();
        scrollIntoView(newZoneTable, {
          align: {
            topOffset: 50,
          },
        });
      }
    }, 0);
  };

  const ResultStatusContent = ({ type }) => {
    const runningStatus = 'RUNNING';
    const normalStatus = 'NORMAL';
    const unavailableStatus = 'UNAVAILABLE';
    const data =
      type === 'tenant'
        ? clusterData?.tenants || []
        : flatten(clusterData.zones?.map(item => item.servers || []));

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
              } else if (runningCount > 0) {
                // OBServer
                zoneListOrTopoRef.current?.setStatusList([runningStatus]);
                scrollToZoneTable();
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
                } else {
                  // OBServer
                  zoneListOrTopoRef.current?.setStatusList([unavailableStatus]);
                  scrollToZoneTable();
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
                // 租户
                history.push({
                  pathname: '/tenant',
                  query: {
                    excludeStatus: [normalStatus, unavailableStatus].join(','),
                  },
                });
                if (type === 'tenant') {
                } else {
                  // OBServer
                  zoneListOrTopoRef.current?.setStatusList(
                    OB_SERVER_STATUS_LIST.map(item => item.value as API.ObServerStatus).filter(
                      item => ![runningStatus, unavailableStatus].includes(item)
                    )
                  );
                  scrollToZoneTable();
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
  };

  // 组装大盘类型数据
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

  const { run: startFn } = useRequest(obStart, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        Modal.success({
          title: '启动集群的任务提交成功',

          content: (
            <>
              {formatMessage({
                id: 'ocp-v2.Detail.Overview.ATaskHasBeenGenerated',
                defaultMessage: '已生成任务，任务 ID:',
              })}

              <a
                onClick={() => {
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
          title: '停止集群的任务提交成功',

          content: (
            <>
              {formatMessage({
                id: 'ocp-v2.Detail.Overview.ATaskHasBeenGenerated',
                defaultMessage: '已生成任务，任务 ID:',
              })}

              <a
                onClick={() => {
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

  function handleMenuClick(key: string) {
    if (key === 'upgrade') {
      setUpgradeVisible(true);
    } else if (key === 'upgradeAgent') {
      setUpgradeAgentVisible(true);
    } else if (key === 'start') {
      Modal.confirm({
        title: `确定要启动 OB 集群 ${clusterData.cluster_name} 吗？`,

        okText: '启动',
        onOk: () => {
          startFn({ scope: { type: 'GLOBAL' } });
          // dispatch({
          //   type: 'cluster/startCluster',
          //   payload: {
          //     id: clusterData.id,
          //   },
          // });
        },
      });
    } else if (key === 'stop') {
      let freezeServerValue = true;
      Modal.confirm({
        title: formatMessage(
          {
            id: 'ocp-v2.Cluster.Detail.AreYouSureYouWant.1',
            defaultMessage: '确定要停止 OB 集群 {clusterDataName} 吗？',
          },

          { clusterDataName: clusterData.cluster_name }
        ),

        content: (
          <>
            <BinlogAssociationMsg
              clusterType="OceanBase"
              type="text"
              binlogService={obClusterRelatedBinlogService}
            />

            <div>
              {formatMessage({
                id: 'ocp-v2.Cluster.Detail.StoppingTheClusterWillCause',
                defaultMessage: '停止集群会导致集群中所有的服务被终止，请谨慎操作',
              })}
            </div>
            {/* <ContentWithIcon
              iconType="question"
              content={formatMessage({
                id: 'ocp-v2.Detail.Overview.PerformADumpOperationBeforeStoppingTheProcess',
                defaultMessage: '停止进程前执行转储操作',
              })}
              tooltip={{
                title: formatMessage({
                  id: 'ocp-v2.Detail.Overview.PerformingThisActionWillProlongTheResponseTime',
                  defaultMessage:
                    '执行本动作会延长停止进程的响应时间，但可以显著缩短 OBServer 恢复时间。',
                }),
              }}
            /> */}

            {/* <Checkbox
              style={{ marginLeft: '8px' }}
              defaultChecked={freezeServerValue}
              onChange={v => {
                freezeServerValue = v.target.checked;
              }}
            /> */}
          </>
        ),

        okText: formatMessage({ id: 'ocp-v2.Cluster.Detail.Stop', defaultMessage: '停止' }),
        okButtonProps: {
          danger: true,
          ghost: true,
        },

        onOk: () => {
          stopFn({
            force: true,
            scope: {
              type: 'GLOBAL',
            },
          });
        },
      });
    }
  }

  const menu = (
    <Menu onClick={({ key }) => handleMenuClick(key)}>
      <Menu.Item key="upgrade">
        <span>升级版本</span>
      </Menu.Item>
      <Menu.Item key="upgradeAgent">
        <span>升级 Agent</span>
      </Menu.Item>
      {/*
      {(clusterData.status === 'RUNNING' ||
        clusterData.status === 'UNAVAILABLE' ||
        clusterData.syncStatus === 'DISABLED_WITH_READ_ONLY') && (
        <Menu.Item key="restart">
          <span>启动集群</span>
        </Menu.Item>
      )}

      {(clusterData.status === 'RUNNING' || clusterData.status === 'UNAVAILABLE') && (
        <Menu.Item key="stop">
          <span>停止集群</span>
        </Menu.Item>
      )} */}
    </Menu>
  );

  return (
    <PageContainer
      loading={reloading}
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
                id: 'ocp-express.Cluster.Overview.ClusterOverview',
                defaultMessage: '集群总览',
              })}
            </span>
            <ContentWithReload
              content={
                <Tag
                  color="geekblue"
                  style={{
                    fontWeight: 'normal',
                    lineHeight: '24px',
                    position: 'relative',
                    top: -4,
                  }}
                >
                  {formatMessage(
                    {
                      id: 'ocp-express.Cluster.Overview.Obversion',
                      defaultMessage: '{obVersion} 版本',
                    },
                    { obVersion: obVersion }
                  )}
                </Tag>
              }
              spin={loading || reloading}
              onClick={() => {
                getClusterData();
                reload();
              }}
            />
          </div>
        ),

        extra: (
          <Space>
            {/* TODO: version1 先屏蔽 */}
            {/* <Button
              data-aspm-click="c304254.d308756"
              data-aspm-desc="集群详情-Unit 分布跳转"
              data-aspm-param={``}
              data-aspm-expo
              onClick={() => {
                history.push('/overview/unit');
              }}
            >
              {formatMessage({
                id: 'ocp-express.Cluster.Overview.UnitDistribution',
                defaultMessage: 'Unit 分布',
              })}
            </Button> */}
            <Button
              data-aspm-click="c304254.d308757"
              data-aspm-desc="集群详情-参数管理跳转"
              data-aspm-param={``}
              data-aspm-expo
              onClick={() => {
                history.push('/overview/parameter');
              }}
            >
              {formatMessage({
                id: 'ocp-express.Cluster.Overview.ParameterManagement',
                defaultMessage: '参数管理',
              })}
            </Button>
            <Button
              onClick={() => {
                handleMenuClick('start');
              }}
            >
              <span>启动集群</span>
            </Button>

            <Button
              onClick={() => {
                handleMenuClick('stop');
              }}
            >
              <span>停止集群</span>
            </Button>

            <Dropdown overlay={menu} getPopupContainer={() => document.body}>
              <Button>
                <EllipsisOutlined />
              </Button>
            </Dropdown>
          </Space>
        ),
      }}
    >
      <Row gutter={[16, 16]}>
        <Col span={12}>
          <ClusterInfo clusterData={clusterData} />
        </Col>
        {overviewStatusType.map(item => {
          return (
            <Col key={item.key} span={6}>
              <MyCard
                title={
                  <div style={{ backgroundImage: item.img }}>
                    <span>{item.title}</span>
                  </div>
                }
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
        <Col span={12}>
          <CompactionTimeTop3 />
        </Col>
        <Col span={12}>
          <SlowSQLTop3 />
        </Col>
        <Col span={24}>
          <TenantResourceTop3 clusterData={clusterData} />
        </Col>
        <Col span={24}>
          <ZoneListOrTopo ref={zoneListOrTopoRef} clusterData={clusterData} />
        </Col>
      </Row>

      <UpgradeDrawer
        obClusterRelatedBinlogService={obClusterRelatedBinlogService}
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
    </PageContainer>
  );
};

export default Detail;
