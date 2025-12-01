import ContentWithInfo from '@/component/ContentWithInfo';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import Empty from '@/component/Empty';
import { DATE_TIME_FORMAT_DISPLAY } from '@/constant/datetime';
import { formatMessage } from '@/util/intl';
import * as ObTenantController from '@/service/obshell/tenant';
import * as ObTenantDeadLockController from '@/service/obshell/tenant';
import { isZhCN } from '@/util';
import { formatTime } from '@/util/datetime';
import { getFullPath } from '@/util/global';
import React, { useEffect, useState } from 'react';
import moment from 'moment';
import { Button, Card, Col, Collapse, Form, Pagination, Row, Space, Spin } from '@oceanbase/design';
import { CaretRightOutlined, RedoOutlined } from '@oceanbase/icons';
import { directTo, sortByNumber } from '@oceanbase/util';
import { useRequest, useSetState } from 'ahooks';
import DeadlockGraph from './DeadlockGraph';
import DeadlockHistoryTable from './DeadlockHistoryTable';

const { Panel } = Collapse;

export interface DeadlockProps {
  tenantName: string;
}

const Deadlock: React.FC<DeadlockProps> = ({ tenantName }) => {
  const [checkTime, setCheckTime] = useState('');
  const [pageable, setPageable] = useSetState({
    page: 1,
    size: 10,
  });

  const { data: tenant, loading: getTenantLoading } = useRequest(ObTenantController.getTenantInfo, {
    ready: !!tenantName,
    defaultParams: [
      {
        name: tenantName,
      },
    ],
  });

  const tenantData = tenant?.data || null;
  const deadLockDetectionEnabled = tenantData?.dead_lock_detection_enabled;

  // 查询死锁历史
  const {
    data: deadLockRes,
    loading: getDeadLockHistoryListLoading,
    run: getDeadLockHistoryList,
    refresh: refreshDeadLockHistoryList,
  } = useRequest(ObTenantDeadLockController.listTenantDeadlocks, {
    manual: true,
    defaultParams: [
      {
        name: tenantName,
        page: pageable.page,
        size: pageable.size,
      },
    ],

    onSuccess: res => {
      if (res.successful) {
        setCheckTime(moment().format(DATE_TIME_FORMAT_DISPLAY));
      }
    },
  });

  useEffect(() => {
    // tenantData 拿到才进行下一步判断，避免直接进入 else
    if (tenantData && deadLockDetectionEnabled) {
      getDeadLockHistoryList({
        name: tenantName,
        page: pageable.page,
        size: pageable.size,
      });
    }
  }, [deadLockDetectionEnabled, tenantData?.obVersion, pageable.size, pageable.page, tenantName]);

  const deadLockHistoryList: API.DeadLock[] = deadLockRes?.data?.contents || [];
  const deadLockHistoryTotal: number = deadLockRes?.data?.page?.total_elements || 0;

  const loading = getTenantLoading || getDeadLockHistoryListLoading;

  if (loading) {
    return (
      <Spin
        className="mt-[15%]"
        spinning={loading}
        tip={formatMessage({
          id: 'ocp-v2.Session.Deadlock.DeadlockQueryPleaseWaitPatiently',
          defaultMessage: '死锁查询中，请耐心等待…',
        })}
      >
        <div className="h-[80vh]" />
      </Spin>
    );
  }

  return (
    <>
      {deadLockDetectionEnabled === false ? (
        <Card
          bordered={false}
          bodyStyle={{ marginBottom: 0 }}
          title={formatMessage({
            id: 'ocp-v2.Session.Deadlock.DeadlockDetails',
            defaultMessage: '死锁详情',
          })}
        >
          <Empty
            style={{ height: 'calc(100vh - 313px)' }} // 使用 Tailwind CSS 替代 className，className样式会透传到Empty里应用
            mode="pageCard"
            title={formatMessage({
              id: 'ocp-v2.Session.Deadlock.AutomaticDeadlockDetectionIsNotEnabled',
              defaultMessage: '死锁自动检测未开启',
            })}
            image={getFullPath('/assets/tenant/empty.svg')}
            description={formatMessage({
              id: 'ocp-v2.Session.Deadlock.PleaseGoToTheClusterToEnableAutomatic',
              defaultMessage: '请前往所属集群开启死锁自动检测',
            })}
          >
            <Button type="primary" onClick={() => directTo(`/overview`)}>
              {formatMessage({
                id: 'ocp-v2.Session.Deadlock.GoNow',
                defaultMessage: '立即前往',
              })}
            </Button>
          </Empty>
        </Card>
      ) : deadLockHistoryList.length === 0 ? (
        <Empty
          style={{ height: 'calc(100vh - 208px)' }} // 使用 Tailwind CSS 替代 className，className样式会透传到Empty里应用
          mode="pageCard"
          title={formatMessage({
            id: 'ocp-v2.Session.Deadlock.NoDeadlockFoundInRecent',
            defaultMessage: '近 7 天未发现死锁',
          })}
          image={getFullPath('/assets/tenant/empty.svg')}
          description={formatMessage(
            {
              id: 'ocp-v2.Session.Deadlock.LastDetectionTimeChecktime',
              defaultMessage: '最近检测时间: {checkTime}',
            },

            { checkTime: checkTime }
          )}
        >
          <Button
            type="primary"
            onClick={() => {
              refreshDeadLockHistoryList();
            }}
          >
            {formatMessage({
              id: 'ocp-v2.Session.Deadlock.Refresh',
              defaultMessage: '刷新',
            })}
          </Button>
        </Empty>
      ) : (
        <Card
          bordered={false}
          bodyStyle={{ marginBottom: 0, display: 'flex', flexDirection: 'column' }}
          title={
            <Space>
              {formatMessage({
                id: 'ocp-v2.Session.Deadlock.DeadlockDetails',
                defaultMessage: '死锁详情',
              })}

              <ContentWithInfo
                content={formatMessage(
                  {
                    id: 'ocp-v2.Session.Deadlock.TheClusterWhereTheTenant',
                    defaultMessage:
                      '此租户所在的集群已开启死锁自动检测，将自动检测并解决死锁，以下为近 7 天的死锁记录，死锁总数为 {renderDeadLockDataLength}',
                  },

                  { renderDeadLockDataLength: deadLockHistoryTotal }
                )}
              />
            </Space>
          }
          extra={
            isZhCN() && (
              <Space>
                {/* 刷新 icon */}
                <RedoOutlined
                  spin={getDeadLockHistoryListLoading}
                  onClick={() => {
                    getDeadLockHistoryList({
                      name: tenantName,
                      page: pageable.page,
                      size: pageable.size,
                    });
                  }}
                  className="ml-1 text-colorTextTertiary"
                />
              </Space>
            )
          }
        >
          <Collapse
            bordered={false}
            // 默认展开第一个死锁
            defaultActiveKey={deadLockHistoryList.map(item => item.event_id as string).slice(0, 1)}
            className="bg-colorBgContainer"
            expandIcon={({ isActive }) => <CaretRightOutlined rotate={isActive ? 90 : 0} />}
          >
            {deadLockHistoryList.map((item, index) => {
              const order = index + 1;
              return (
                <Panel
                  className={`border-none bg-colorBgLayout ${
                    index !== deadLockHistoryList.length - 1 ? 'mb-4' : ''
                  }`}
                  header={
                    <span className="font-semibold">
                      {formatMessage(
                        {
                          id: 'ocp-v2.Session.Deadlock.DeadlockOrder',
                          defaultMessage: '死锁 {order}',
                        },

                        { order: order }
                      )}
                    </span>
                  }
                  key={item?.event_id as string}
                  extra={
                    <Form layout="inline" className="text-gray-400">
                      <Form.Item
                        label={
                          <ContentWithQuestion
                            content="Event ID"
                            tooltip={{
                              placement: 'right',
                              title: formatMessage({
                                id: 'ocp-v2.Session.Deadlock.EventIdIsTheUnique',
                                defaultMessage: 'Event ID 是一次死锁事件的唯一标识',
                              }),
                            }}
                          />
                        }
                      >
                        {item?.event_id || '-'}
                      </Form.Item>
                      <Form.Item
                        label={formatMessage({
                          id: 'ocp-v2.Session.Deadlock.OccurrenceTime',
                          defaultMessage: '发生时间',
                        })}
                      >
                        {formatTime(item?.report_time)}
                      </Form.Item>
                    </Form>
                  }
                >
                  <Row gutter={[16, 16]}>
                    <Col span={9}>
                      <DeadlockGraph
                        id={item.event_id}
                        {...{
                          // 按 idx 从小到大排序，否则会影响死锁顺序
                          nodeList: item.nodes?.sort((a, b) => sortByNumber(a, b, 'idx')),
                        }}
                      />
                    </Col>
                    <Col span={15}>
                      <DeadlockHistoryTable dataSource={item?.nodes} />
                    </Col>
                  </Row>
                </Panel>
              );
            })}
          </Collapse>
          <Pagination
            current={pageable.page}
            className="self-end mt-4"
            total={deadLockHistoryTotal}
            pageSize={pageable.size}
            onChange={(page, pageSize) => {
              setPageable({
                page,
                size: pageSize,
              });
            }}
          />
        </Card>
      )}
    </>
  );
};

export default Deadlock;
