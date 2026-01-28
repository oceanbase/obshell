import { formatMessage } from '@/util/intl';
import { history, useDispatch } from 'umi';
import {
  Button,
  Col,
  Form,
  Row,
  Badge,
  Space,
  Tag,
  Tooltip,
  Typography,
  Descriptions,
  Modal,
  message,
  Switch,
} from '@oceanbase/design';
import { QuestionCircleOutlined } from '@oceanbase/icons';
import React, { useEffect, useState } from 'react';
import { uniqueId, find } from 'lodash';
import { byte2GB, findByValue } from '@oceanbase/util';
import moment from 'moment';
import { PageContainer } from '@oceanbase/ui';
import { COMPACTION_STATUS_LISTV4 } from '@/constant/compaction';
import { TENANT_MODE_LIST } from '@/constant/tenant';
import { getCompactionStatusV4 } from '@/util/cluster';
import { taskSuccess } from '@/util/task';
import { formatTime } from '@/util/datetime';
import { useRequest, useInterval } from 'ahooks';
import MyCard from '@/component/MyCard';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import ContentWithReload from '@/component/ContentWithReload';
import BatchOperationBar from '@/component/BatchOperationBar';
import FormEditZoneReplicaTable from '@/component/FormEditZoneReplicaTable';
import RenderConnectionString from '@/component/RenderConnectionString';
import BatchModifyUnitModal from '../Component/BatchModifyUnitModal';
import AddReplicaModal from '../Component/AddReplicaModal';
import DeleteTenantModal from '../Component/DeleteTenantModal';
import DeleteReplicaModal from '../Component/DeleteReplicaModal';
import ModifyWhitelistModal from '../Component/ModifyWhitelistModal';
import ModifyPrimaryZoneDrawer from '../Component/ModifyPrimaryZoneDrawer';

import OBProxyAndConnectionStringModal from '../Component/OBProxyAndConnectionStringModal';
import {
  getTenantInfo,
  tenantModifyReplicas,
  getTenantCompaction,
  tenantMajorCompaction,
  tenantClearCompactionError,
  tenantLock,
  tenantUnlock,
} from '@/service/obshell/tenant';

import { unitConfigCreate } from '@/service/obshell/unit';
import { obclusterInfo, getUnitConfigLimit } from '@/service/obshell/obcluster';
import { getZonesFromTenant } from '@/util/tenant';
const { Text } = Typography;

interface NewProps {
  match: {
    params: {
      tenantName: string;
    };
  };
}

const Detail: React.FC<NewProps> = ({
  match: {
    params: { tenantName },
  },
}) => {
  const dispatch = useDispatch();

  const [form] = Form.useForm();
  const { setFieldsValue } = form;

  const [showAddReplicaModal, setShowAddReplicaModal] = useState(false);
  const [showDeleteReplicaModal, setShowDeleteReplicaModal] = useState(false);
  const [showBatchModifyUnitModal, setShowBatchModifyUnitModal] = useState(false);

  const [showDeleteTenantModal, setShowDeleteTenantModal] = useState(false);
  const [showWhitelistModal, setShowWhitelistModal] = useState(false);
  const [showModifyPrimaryZoneDrawer, setShowModifyPrimaryZoneDrawer] = useState(false);
  const [connectionStringModalVisible, setConnectionStringModalVisible] = useState(false);

  const [currentTenantZone, setCurrentTenantZone] = useState(null);
  const [selectedRows, setSelectedRows] = useState([]);
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);

  const { data: clusterInfo } = useRequest(obclusterInfo, {
    defaultParams: [{}],
  });

  const clusterData = clusterInfo?.data || {};
  const clusterZones = clusterData?.zones || [];

  const {
    data: tenantInfo,
    run: getTenantData,
    loading,
    refresh,
  } = useRequest(getTenantInfo, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const currentTenantData = res?.data || {};
        const _zones = getZonesFromTenant(currentTenantData);

        setFieldsValue({
          zones: (_zones || []).map(item => ({
            key: uniqueId(),
            name: item.name,
            replicaType: item.replicaType,
            resourcePool: item.resourcePool,
          })),
        });
      }
    },
  });

  const tenantData: API.TenantInfo = {
    ...(tenantInfo?.data || {}),
    name: tenantInfo?.data?.tenant_name,
  };

  const zones = getZonesFromTenant(tenantData);

  // 获取 unit 规格的限制规则
  const { data: clusterUnitSpecLimitData } = useRequest(getUnitConfigLimit);

  const clusterUnitSpecLimit = {
    cpuLowerLimit: clusterUnitSpecLimitData?.data?.min_cpu || 0,
    memoryLowerLimit: clusterUnitSpecLimitData?.data?.min_memory || 0,
  };

  useEffect(() => {
    if (tenantName) {
      getTenantData({ name: tenantName });
    }
  }, [tenantName]);

  // 修改副本
  const { runAsync: modifyReplica, loading: modifyUnitLoading } = useRequest(tenantModifyReplicas, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const taskId = res?.data?.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'ocp-express.Detail.Overview.TheTaskOfModifyingTheReplicaWasSubmitted',
            defaultMessage: '修改副本的任务提交成功',
          }),
        });
        setFieldsValue({
          // 由于表格表单的最新值和展示态的数据格式不一样，因此修改成功后还需要重新设置副本的值
          zones: (zones || []).map(item => ({
            key: uniqueId(),
            name: item.name,
            replicaType: item.replicaType,
            resourcePool: item.resourcePool,
          })),
        });
      }
    },
  });

  const updateTenant = () => {
    dispatch({
      type: 'tenant/getTenantData',
      payload: {
        name: tenantName,
      },
    });
  };

  const { run: lockTenant } = useRequest(tenantLock, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.TenantLockedSuccessfully',
            defaultMessage: '租户锁定成功',
          })
        );
        refresh();
        updateTenant();
      }
    },
  });

  const { run: unlockTenant } = useRequest(tenantUnlock, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.tenant.TenantUnlockedSuccessfully',
            defaultMessage: '租户解锁成功',
          })
        );
        refresh();
        updateTenant();
      }
    },
  });

  // 获取租户合并详情
  const { data, refresh: getTenantCompactionRefresh } = useRequest(getTenantCompaction, {
    defaultParams: [
      {
        name: tenantName,
      },
    ],
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
  const polling = compactionStatus === 'COMPACTING' || compactionStatus === 'WAIT_MERGE';
  useInterval(
    () => {
      getTenantCompactionRefresh();
    },
    polling ? 3000 : undefined
  );

  // 发起合并
  const { run: triggerTenantCompaction } = useRequest(tenantMajorCompaction, {
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
  const { run: clearCompactionError, loading: clearLoading } = useRequest(
    tenantClearCompactionError,
    {
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
    }
  );

  const { runAsync: unitConfigCreateFn } = useRequest(unitConfigCreate, {
    manual: true,
  });

  const handleModifyReplica = async (value: API.TenantZone, onSuccess: () => void) => {
    const { name, replicaType, resourcePool } = value || {};
    const currentModifyTenantZone = find(zones, item => item.name === name);

    const maxMemorySizeGB = byte2GB(currentModifyTenantZone?.resourcePool?.unitConfig?.memory_size);

    let unitName: string = currentModifyTenantZone?.resourcePool?.unitConfig?.name;
    if (
      resourcePool &&
      (resourcePool.unitConfig?.cpuCore !==
        currentModifyTenantZone?.resourcePool?.unitConfig?.max_cpu ||
        resourcePool.unitConfig?.memorySize !== maxMemorySizeGB)
    ) {
      unitName = `tenant_unit_${Date.now()}_${uniqueId()}`;
      const res = await unitConfigCreateFn({
        name: unitName,
        max_cpu: Number(resourcePool.unitConfig?.cpuCore),
        memory_size: `${resourcePool.unitConfig?.memorySize}GB`,
      });
      if (!res.successful) {
        return;
      }
    } else if (replicaType === currentModifyTenantZone?.replicaType) {
      return message.info(
        formatMessage({
          id: 'ocp-express.Detail.Overview.ZoneInformationHasNotChanged',
          defaultMessage: 'Zone 的信息没有变更',
        })
      );
    }
    modifyReplica(
      {
        name: tenantData.tenant_name!,
      },
      {
        zone_list: [
          {
            zone_name: name!,
            replica_type: replicaType,
            unit_num: resourcePool?.unitCount || currentModifyTenantZone?.resourcePool?.unit_num,
            unit_config_name: unitName,
          },
        ],
      }
    ).then(res => {
      if (res.successful) {
        if (onSuccess) {
          refresh();
          onSuccess();
        }
      }
    });
  };

  const handleDeleteReplica = (record: API.TenantZone) => {
    setShowDeleteReplicaModal(true);
    setCurrentTenantZone(record);
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: (newSelectedRowKeys: string[], newSelectedRows: API.TenantZone[]) => {
      setSelectedRowKeys(newSelectedRowKeys);
      setSelectedRows(newSelectedRows);
    },
  };

  return (
    <PageContainer
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'ocp-express.Detail.Overview.Overview',
              defaultMessage: '总览',
            })}
            spin={loading}
            onClick={() => {
              refresh();
              getTenantCompactionRefresh();
            }}
          />
        ),

        extra: (
          <Space>
            <Tooltip
              placement="topRight"
              title={
                tenantData.tenant_name === 'sys' &&
                formatMessage({
                  id: 'ocp-express.Detail.Overview.TheSysTenantCannotBe',
                  defaultMessage: 'sys 租户无法删除',
                })
              }
            >
              <Button
                disabled={tenantData.tenant_name === 'sys'}
                onClick={() => {
                  setShowDeleteTenantModal(true);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.Tenant.Detail.DeleteATenant',
                  defaultMessage: '删除租户',
                })}
              </Button>
            </Tooltip>

            <Button
              type="primary"
              onClick={() => {
                setShowAddReplicaModal(true);
              }}
            >
              {formatMessage({
                id: 'ocp-express.Tenant.Detail.NewCopy',
                defaultMessage: '新增副本',
              })}
            </Button>
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
            <div>
              <Descriptions column={4}>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'OBShell.Detail.Overview.TenantName',
                    defaultMessage: '租户名称',
                  })}
                >
                  <Typography.Text copyable>{tenantData?.tenant_name}</Typography.Text>
                </Descriptions.Item>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Tenant.Detail.TenantMode',
                    defaultMessage: '租户模式',
                  })}
                >
                  {findByValue(TENANT_MODE_LIST, tenantData.mode).label}
                </Descriptions.Item>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Detail.Overview.ObVersion',
                    defaultMessage: 'OB 版本号',
                  })}
                >
                  {clusterData?.ob_version}
                </Descriptions.Item>

                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Tenant.Detail.CharacterSet',
                    defaultMessage: '字符集',
                  })}
                >
                  {tenantData.charset || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="Collation">
                  {tenantData.collation || '-'}
                </Descriptions.Item>
                <Descriptions.Item
                  label={
                    <ContentWithQuestion
                      content={formatMessage({
                        id: 'OBShell.Detail.Overview.DeploymentStepByStep',
                        defaultMessage: '部署分步',
                      })}
                      tooltip={{
                        title: formatMessage({
                          id: 'OBShell.Detail.Overview.TypesOfTenantReplicasIn',
                          defaultMessage: '租户在各个 Zone 上的副本类型',
                        }),
                      }}
                    />
                  }
                >
                  {tenantData.locality || '-'}
                </Descriptions.Item>
                <Descriptions.Item
                  span={2}
                  label={formatMessage({
                    id: 'ocp-express.component.TenantList.ConnectionString',
                    defaultMessage: '连接字符串',
                  })}
                >
                  <RenderConnectionString
                    connectionStrings={tenantData.connection_strings || []}
                    maxWidth={500}
                    callBack={() => {
                      setConnectionStringModalVisible(true);
                    }}
                  />
                </Descriptions.Item>
                {tenantData.mode === 'MYSQL' && (
                  <Descriptions.Item
                    label={
                      <Space size={4}>
                        {formatMessage({
                          id: 'OBShell.Detail.Overview.TableNamesAreCaseSensitive',
                          defaultMessage: '表名大小写敏感',
                        })}

                        <Tooltip
                          title={
                            <div>
                              <div>
                                {formatMessage({
                                  id: 'OBShell.Detail.Overview.UseCreateTableOrCreate',
                                  defaultMessage:
                                    '· 0：使用 CREATE TABLE 或 CREATE DATABASE\n                                语句指定的大小写字母在硬盘上保存表名和数据库名。名称比较对大小写敏感。',
                                })}
                              </div>
                              <div>
                                {formatMessage({
                                  id: 'OBShell.Detail.Overview.TableNamesAreSavedIn',
                                  defaultMessage:
                                    '· 1：表名在硬盘上以小写保存，名称比较对大小写不敏感。',
                                })}
                              </div>
                              <div>
                                {formatMessage({
                                  id: 'OBShell.Detail.Overview.TableNameAndDatabaseName',
                                  defaultMessage:
                                    '· 2：表名和数据库名在硬盘上使用 CREATE TABLE 或 CREATE DATABASE\n                                语句指定的大小写字母进行保存，但名称比较对大小写不敏感。',
                                })}
                              </div>
                            </div>
                          }
                          placement="topLeft"
                        >
                          <QuestionCircleOutlined
                            style={{
                              color: '#999',
                              cursor: 'pointer',
                              fontSize: '12px',
                            }}
                          />
                        </Tooltip>
                      </Space>
                    }
                  >
                    {tenantData.lower_case_table_names || '-'}
                  </Descriptions.Item>
                )}
                <Descriptions.Item
                  label={formatMessage({
                    id: 'OBShell.Detail.Overview.TimeZone',
                    defaultMessage: '时区',
                  })}
                >
                  GMT{tenantData.time_zone || '-'}
                </Descriptions.Item>

                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Detail.Overview.LockedState',
                    defaultMessage: '锁定状态',
                  })}
                >
                  <Switch
                    size="small"
                    disabled={tenantData?.tenant_name === 'sys'}
                    checked={tenantData?.locked === 'YES'}
                    onChange={() => {
                      Modal.confirm({
                        title: `确定要${tenantData?.locked === 'YES' ? '解锁' : '锁定'}租户 ${
                          tenantData.tenant_name
                        } 吗？`,

                        content:
                          tenantData?.locked === 'YES'
                            ? formatMessage({
                                id: 'OBShell.Detail.Overview.UnlockingTheTenantWillRestore',
                                defaultMessage: '解锁租户将恢复用户对租户的访问',
                              })
                            : formatMessage({
                                id: 'OBShell.Detail.Overview.LockingATenantWillCause',
                                defaultMessage: '锁定租户将导致用户无法访问该租户，请谨慎操作',
                              }),

                        okText:
                          tenantData?.locked === 'YES'
                            ? formatMessage({
                                id: 'OBShell.Detail.Overview.Unlock',
                                defaultMessage: '解锁',
                              })
                            : formatMessage({
                                id: 'OBShell.Detail.Overview.Lock',
                                defaultMessage: '锁定',
                              }),
                        okButtonProps: {
                          danger: true,
                          ghost: true,
                        },
                        onOk: () => {
                          if (tenantData?.locked === 'NO') {
                            lockTenant({
                              name: tenantData?.tenant_name!,
                            });
                          } else {
                            unlockTenant({
                              name: tenantData?.tenant_name!,
                            });
                          }
                        },
                      });
                    }}
                  />
                </Descriptions.Item>
                <Descriptions.Item
                  label={formatMessage({
                    id: 'ocp-express.Tenant.Detail.CreationTime',
                    defaultMessage: '创建时间',
                  })}
                >
                  {formatTime(tenantData.created_time)}
                </Descriptions.Item>
                <Descriptions.Item
                  className="descriptions-item-with-ellipsis"
                  label={formatMessage({
                    id: 'ocp-express.Detail.Overview.Note',
                    defaultMessage: '备注',
                  })}
                >
                  <Tooltip placement="bottomLeft" title={tenantData?.comment}>
                    <Text style={{ width: '95%', minWidth: 180 }} ellipsis={true}>
                      {tenantData?.comment || '-'}
                    </Text>
                  </Tooltip>
                </Descriptions.Item>
              </Descriptions>
            </div>
          </MyCard>
        </Col>
        <Col span={24}>
          <MyCard
            title={formatMessage({
              id: 'ocp-express.Tenant.Detail.CopyDetails',
              defaultMessage: '副本详情',
            })}
            extra={
              <Button
                onClick={() => {
                  setShowBatchModifyUnitModal(true);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.Detail.Overview.ModifyUnit',
                  defaultMessage: '修改 Unit',
                })}
              </Button>
            }
          >
            <Form form={form} preserve={false} layout="vertical" hideRequiredMark={true}>
              <Form.Item name="zones" initialValue={[]}>
                <FormEditZoneReplicaTable
                  onSave={handleModifyReplica}
                  onDelete={handleDeleteReplica}
                  saveLoading={modifyUnitLoading}
                  showDeletePopconfirm={false}
                  tenantData={tenantData}
                  zones={zones}
                  clusterData={clusterData}
                  tenantName={tenantName}
                  rowSelection={rowSelection}
                  rowKey={record => record.name}
                  unitSpecLimit={clusterUnitSpecLimit}
                />
              </Form.Item>
            </Form>
            {selectedRowKeys && selectedRowKeys.length > 0 && (
              <BatchOperationBar
                size="small"
                selectedCount={selectedRowKeys && selectedRowKeys.length}
                onCancel={() => {
                  setSelectedRowKeys([]);
                }}
                actions={[
                  <Button
                    key="batch-delete"
                    danger={true}
                    ghost={true}
                    onClick={() => {
                      setShowDeleteReplicaModal(true);
                    }}
                  >
                    {formatMessage({
                      id: 'ocp-express.Tenant.Detail.BatchDelete',
                      defaultMessage: '批量删除',
                    })}
                  </Button>,
                ]}
                style={{ marginBottom: 16 }}
              />
            )}
          </MyCard>
        </Col>
        <Col span={24}>
          <MyCard
            bordered={false}
            title={
              <ContentWithQuestion
                content={formatMessage({
                  id: 'ocp-express.Tenant.Detail.ZonePriority',
                  defaultMessage: 'Zone 优先级',
                })}
                tooltip={{
                  title: formatMessage({
                    id: 'ocp-express.Tenant.Detail.PriorityOfTheDistributionOf',
                    defaultMessage: '租户中主副本分布的优先级',
                  }),
                }}
              />
            }
            extra={
              <Button
                onClick={() => {
                  setShowModifyPrimaryZoneDrawer(true);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.Tenant.Detail.Modify',
                  defaultMessage: '修改',
                })}
              </Button>
            }
          >
            {tenantData?.primary_zone &&
              tenantData?.primary_zone
                ?.split(';')
                .map((item: string) => <Tag key={item}>{item}</Tag>)}
          </MyCard>
        </Col>
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
                    defaultMessage: '能够连接到租户的地址列表',
                  }),
                }}
              />
            }
            extra={
              <Button
                onClick={() => {
                  setShowWhitelistModal(true);
                }}
              >
                {formatMessage({
                  id: 'ocp-express.Tenant.Detail.Modify',
                  defaultMessage: '修改',
                })}
              </Button>
            }
          >
            <div className="multiple-tag-wrapper">
              {tenantData.whitelist &&
                tenantData.whitelist.split(',').map(item => <Tag key={item}>{item}</Tag>)}
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
                        '以下租户合并存在异常，异常解决后，再点击「清除异常」，操作后系统会继续合并',
                    })}
                  >
                    <Button
                      type="link"
                      loading={clearLoading}
                      onClick={() => {
                        clearCompactionError({
                          name: tenantName,
                        });
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
                          return triggerTenantCompaction({
                            name: tenantName,
                          });
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
      </Row>

      <AddReplicaModal
        visible={showAddReplicaModal}
        tenantName={tenantName}
        tenantData={tenantData}
        tenantZones={zones}
        clusterZones={clusterZones}
        unitSpecLimit={clusterUnitSpecLimit}
        onCancel={() => {
          setShowAddReplicaModal(false);
        }}
        onSuccess={() => {
          setShowAddReplicaModal(false);
          refresh();
        }}
      />

      <DeleteReplicaModal
        visible={showDeleteReplicaModal}
        tenantData={tenantData}
        tenantZones={currentTenantZone ? [currentTenantZone] : selectedRows}
        onCancel={() => {
          setShowDeleteReplicaModal(false);
          setCurrentTenantZone(null);
        }}
        onSuccess={() => {
          setShowDeleteReplicaModal(false);
          setCurrentTenantZone(null);
          refresh();
        }}
      />

      <BatchModifyUnitModal
        visible={showBatchModifyUnitModal}
        tenantData={tenantData}
        clusterZones={clusterZones}
        tenantZones={zones}
        unitSpecLimit={clusterUnitSpecLimit}
        onCancel={() => {
          setShowBatchModifyUnitModal(false);
        }}
        onSuccess={() => {
          setShowBatchModifyUnitModal(false);
          setSelectedRows([]);
          setSelectedRowKeys([]);
          refresh();
        }}
      />

      <DeleteTenantModal
        tenantData={tenantData}
        visible={showDeleteTenantModal}
        onCancel={() => {
          setShowDeleteTenantModal(false);
        }}
        onSuccess={() => {
          setShowDeleteTenantModal(false);
          // 租户接口第一时间请求的状态不是最新的，加个 delay 兜底
          setTimeout(() => {
            history.push('/tenant');
          }, 700);
        }}
      />

      <ModifyWhitelistModal
        visible={showWhitelistModal}
        tenantData={tenantData}
        onCancel={() => {
          setShowWhitelistModal(false);
        }}
        onSuccess={() => {
          setShowWhitelistModal(false);
          refresh();
        }}
      />

      <ModifyPrimaryZoneDrawer
        zones={zones}
        visible={showModifyPrimaryZoneDrawer}
        tenantData={tenantData}
        onClose={() => {
          setShowModifyPrimaryZoneDrawer(false);
        }}
        onSuccess={() => {
          setShowModifyPrimaryZoneDrawer(false);
          refresh();
        }}
      />

      <OBProxyAndConnectionStringModal
        width={900}
        visible={connectionStringModalVisible}
        obproxyAndConnectionStrings={tenantData?.obproxyAndConnectionStrings || []}
        onCancel={() => {
          setConnectionStringModalVisible(false);
        }}
        onSuccess={() => {
          setConnectionStringModalVisible(false);
          refresh();
        }}
      />
    </PageContainer>
  );
};
export default Detail;
