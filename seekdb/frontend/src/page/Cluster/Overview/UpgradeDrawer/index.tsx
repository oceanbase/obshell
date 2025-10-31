import { formatMessage } from '@/util/intl';
import { useDispatch, useSelector } from 'umi';
import React, { useRef, useState, useEffect } from 'react';
import { find } from 'lodash';
import { Button, Descriptions, Form, Modal, Space, token } from '@oceanbase/design';
import { ExclamationCircleFilled } from '@oceanbase/icons';
import { ContentWithIcon as OBUIContentWithIcon } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import { MODAL_FORM_ITEM_LAYOUT } from '@/constant';
import { taskSuccess } from '@/util/task';
import ContentWithIcon from '@/component/ContentWithIcon';
import type { MyDrawerProps } from '@/component/MyDrawer';
import MyDrawer from '@/component/MyDrawer';
import MySelect from '@/component/MySelect';
import PackageSelect from '@/component/PackageSelect';
import UploadPackageDrawer from '@/component/UploadPackageDrawer';
import styles from './index.less';
import { obUpgrade } from '@/service/obshell/upgrade';
import { useCluster } from '@/hook/useCluster';

const { Option } = MySelect;
export interface UpgradeDrawerProps extends MyDrawerProps {
  isStandAloneCluster?: boolean;
  // 集群的架构列表
  architectureList?: string[];
  clusterData: API.ClusterInfo;
  onSuccess: () => void;
  onCancel: () => void;
}

export type ExtendedUpgradeNode = API.UpgradeNode & {
  isCurrent?: boolean;
  noRpms?: API.UpgradeNode['rpms'];
};

export interface ExtendedOcpCluster extends API.OcpClusterView {
  upgradePath?: ExtendedUpgradeNode[];
}

const UpgradeDrawer: React.FC<UpgradeDrawerProps> = ({
  visible,
  isStandAloneCluster,
  architectureList,
  clusterData,
  onSuccess,
  onCancel,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { getFieldsValue, validateFields } = form;

  const { isStandalone } = useCluster();

  const dispatch = useDispatch();
  const packageSelectRef = useRef<PackageSelectRef>();

  const { ocpMultiClusterInfo: { multiClusterEnabled, currentOcpInfo } = {}, ocpMultiClusterList } =
    useSelector((state: DefaultRootState) => state.global);

  const [pathVisible, setPathVisible] = useState<boolean>(false);
  const [uploadVisible, setUploadVisible] = useState<boolean>(false);
  // 软件包列表
  const [packageList, setPackageList] = useState<API.UpgradePkgInfo[]>([]);

  const initialData = clusterData?.zones?.map((item, index) => ({
    name: item.name,
    sequenceNumber: index + 1,
    key: index + 1,
  }));

  const [zoneOrderList, setZoneOrderList] = useState(initialData);

  useEffect(() => {
    if (visible && multiClusterEnabled && ocpMultiClusterList.length < 1) {
      //  获取 OCP 列表
      dispatch({
        type: 'global/getOcpMultiAllOcpClusters',
      });
    }
  }, [visible, multiClusterEnabled]);

  useEffect(() => {
    setZoneOrderList(initialData);
  }, [clusterData?.zones]);

  // 集群升级
  const { run: upgradeObCluster, loading: upgradeLoading } = useRequest(obUpgrade, {
    // 手动调用
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const taskId = res.data?.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'ocp-v2.Overview.UpgradeDrawer.VersionUpgradeTaskSubmittedSuccessfully',
            defaultMessage: '版本升级的任务提交成功',
          }),
        });

        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  return (
    <MyDrawer
      width={620}
      title={formatMessage({
        id: 'ocp-v2.Overview.UpgradeDrawer.UpgradeVersion',
        defaultMessage: '升级版本',
      })}
      visible={visible}
      destroyOnClose={true}
      onOk={() => {
        validateFields().then(values => {
          const { upgradeMode, rpmFileName } = values;
          const currentZonePackage = find(packageList, rpm => rpm.pkg_id === rpmFileName.value);

          Modal.confirm({
            title: formatMessage({
              id: 'ocp-v2.Overview.UpgradeDrawer.AreYouSureYouWant',
              defaultMessage: '确定要升级吗？',
            }),
            content: (
              <ul>
                {/* 停服升级 */}
                {upgradeMode === 'stopService' && (
                  <li>
                    {formatMessage({
                      id: 'ocp-v2.Overview.UpgradeDrawer.StopTheServiceUpgradeWill',
                      defaultMessage:
                        '停服升级会中断业务对数据库的访问，对业务造成不可用的影响，请谨慎操作',
                    })}
                  </li>
                )}
              </ul>
            ),

            onOk: () => {
              return upgradeObCluster({
                mode: upgradeMode,
                version: currentZonePackage?.version,
                release: currentZonePackage?.release_distribution,
              });
            },
          });
        });
      }}
      onCancel={() => onCancel()}
      okText={formatMessage({
        id: 'ocp-v2.Overview.UpgradeDrawer.Upgrade',
        defaultMessage: '升级',
      })}
      // confirmLoading={pathLoading || checkBinlogCompatibilityLoading}
      {...restProps}
    >
      <Descriptions column={1} style={{ marginBottom: 12 }}>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-v2.Overview.UpgradeDrawer.Cluster',
            defaultMessage: '集群',
          })}
        >
          {clusterData.cluster_name}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-v2.Overview.UpgradeDrawer.OceanbaseVersion',
            defaultMessage: 'OceanBase 版本',
          })}
        >
          {clusterData.ob_version}
        </Descriptions.Item>
      </Descriptions>

      <Form form={form} preserve={false} layout="horizontal" hideRequiredMark={true}>
        {/* <Form.Item
            label={
              <OBUIContentWithIcon
                iconType="question"
                content={'停止进程前执行转储操作'}
                tooltip={{
                  title: '执行本动作会延长停止进程的响应时间，但可以显著缩短 OBServer 恢复时间。',
                }}
              />
            }
            initialValue={true}
            valuePropName="checked"
            name="freezeServer"
            style={{ marginTop: '8px' }}
           >
            <Checkbox />
           </Form.Item> */}
        <Form.Item shouldUpdate={true} noStyle={true}>
          {({ getFieldValue }) => (
            <Form.Item
              {...MODAL_FORM_ITEM_LAYOUT}
              label={
                <OBUIContentWithIcon
                  iconType="question"
                  content={formatMessage({
                    id: 'ocp-v2.Overview.UpgradeDrawer.UpgradeMethod',
                    defaultMessage: '升级方式',
                  })}
                  tooltip={{
                    placement: 'right',
                    title: (
                      <div>
                        <div>
                          {formatMessage({
                            id: 'ocp-v2.Overview.UpgradeDrawer.RollingUpgradeNoInterruptionOf',
                            defaultMessage: '滚动升级：不会中断业务对数据库的访问',
                          })}
                        </div>
                        <div>
                          {formatMessage({
                            id: 'ocp-v2.Overview.UpgradeDrawer.StopServiceUpgradeItWill',
                            defaultMessage:
                              '停服升级：会中断业务对数据库的访问，对业务造成不可用的影响，请谨慎操作',
                          })}
                        </div>
                      </div>
                    ),
                  }}
                />
              }
              name="upgradeMode"
              initialValue={isStandAloneCluster ? 'stopService' : 'rolling'}
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'ocp-v2.Overview.UpgradeDrawer.PleaseSelectAnUpgradeMethod',
                    defaultMessage: '请选择升级方式',
                  }),
                },
              ]}
              extra={
                getFieldValue('upgradeMode') === 'stopService' && (
                  <ContentWithIcon
                    prefixIcon={{
                      component: ExclamationCircleFilled,
                      style: {
                        color: token.colorWarning,
                        marginTop: 4,
                      },
                    }}
                    content={formatMessage({
                      id: 'ocp-v2.Overview.UpgradeDrawer.ServiceSuspensionAndUpgradeWill',
                      defaultMessage: '停服升级会中断业务对数据库的访问，对业务造成不可用的影响',
                    })}
                    style={{
                      color: token.colorWarning,
                      // 与 icon 的 marginTop 配合，可以实现 icon 和文本的上对齐
                      alignItems: 'flex-start',
                    }}
                  />
                )
              }
              style={{ marginTop: '-8px' }}
            >
              <MySelect>
                {!isStandAloneCluster && (
                  <Option value="rolling">
                    {formatMessage({
                      id: 'ocp-v2.Overview.UpgradeDrawer.RollingUpgradeRecommend',
                      defaultMessage: '滚动升级（推荐）',
                    })}
                  </Option>
                )}

                <Option value="stopService">
                  {formatMessage({
                    id: 'ocp-v2.Overview.UpgradeDrawer.StopServiceUpgrade',
                    defaultMessage: '停服升级',
                  })}
                </Option>
              </MySelect>
            </Form.Item>
          )}
        </Form.Item>
        <Form.Item
          {...MODAL_FORM_ITEM_LAYOUT}
          label={
            <OBUIContentWithIcon
              iconType="question"
              content={formatMessage({
                id: 'ocp-v2.Overview.UpgradeDrawer.UpgradeVersion',
                defaultMessage: '升级版本',
              })}
              tooltip={{
                title: formatMessage({
                  id: 'ocp-v2.Overview.UpgradeDrawer.TheOptionalVersionIsBased',
                  defaultMessage: '可选的版本是根据当前已上传 OB 软件包的版本集合',
                }),
              }}
            />
          }
          name="rpmFileName"
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-v2.Overview.UpgradeDrawer.PleaseSelectOceanbaseVersion',
                defaultMessage: '请选择 OceanBase 版本',
              }),
            },
          ]}
          extra={
            <div>
              <div>
                {formatMessage({
                  id: 'ocp-v2.Overview.UpgradeDrawer.UploadRequiredDuringUpgrade',
                  defaultMessage: '升级时需要上传',
                })}{' '}
                {isStandalone ? (
                  formatMessage({
                    id: 'obshell.Overview.UpgradeDrawer.OnlyPackagesStartingWithOceanbaseStandaloneAreSupported',
                    defaultMessage: 'oceanbase-standalone 包',
                  })
                ) : (
                  <>
                    <a
                      href="https://mirrors.aliyun.com/oceanbase/community/stable/el/7/x86_64/"
                      target="_blank"
                    >
                      {formatMessage({
                        id: 'ocp-v2.Overview.UpgradeDrawer.OceanbaseCeAndOceanbaseCe',
                        defaultMessage: 'oceanbase-ce 和 oceanbase-ce-libs',
                      })}
                    </a>{' '}
                    {formatMessage({
                      id: 'ocp-v2.Overview.UpgradeDrawer.TwoBags',
                      defaultMessage: '两个包',
                    })}
                  </>
                )}
              </div>
              {architectureList?.length === 2 && (
                <div>
                  {formatMessage(
                    {
                      id: 'ocp-v2.Overview.UpgradeDrawer.OnlyPackagesHigherThanThe',
                      defaultMessage: '仅可选择高于当前版本 {clusterDataObVersion} 的软件包',
                    },
                    { clusterDataObVersion: clusterData.ob_version }
                  )}
                </div>
              )}

              {architectureList?.length === 1 && (
                <div>
                  {formatMessage({
                    id: 'ocp-v2.Overview.UpgradeDrawer.OnlyTheCurrentClusterHardware',
                    defaultMessage: '仅可选择匹配当前集群硬件架构',
                  })}
                  {architectureList?.join()}
                  {formatMessage({
                    id: 'ocp-v2.Overview.UpgradeDrawer.ThePackage',
                    defaultMessage: '的软件包',
                  })}
                </div>
              )}
            </div>
          }
        >
          <PackageSelect
            labelInValue={true}
            useType="UPGRADE_CLUSTER"
            isHeterogeneousUpgrade={true}
            type={['OB_SERVER_INSTALL_PACKAGE']}
            currentObVersion={clusterData?.ob_version}
            ref={packageSelectRef}
            onSuccess={newPackageList => {
              setPackageList(newPackageList);
            }}
          />
        </Form.Item>

        {/* <Form.Item shouldUpdate={true} noStyle={true}>
            {() => {
              const { rpmFileName } = getFieldsValue();
              const currentZonePackage = find(
                packageList,
                rpm =>
                  `${rpm?.name}-${rpm?.version}-${rpm?.buildNumber}-${rpm?.operatingSystem}` ===
                  rpmFileName?.label
              );
              return (
                <>
                  {zoneSwitch && (
                    <>
                      <Alert
                        showIcon
                        type="info"
                        style={{ marginBottom: '24px' }}
                        message={
                          <>
                            <div style={{ fontWeight: 500 }}>
                              { '升级注意事项'}
                            </div>
                            {isGte4_1(currentZonePackage?.version) ? (
                              <>
                                <div>
                                  {formatMessage({
                                    id: 'ocp-v2.Component.UpgradeDrawer.IfThePrimaryAndSecondaryTenantsAreIn',
                                    defaultMessage:
                                      '1.如果主备租户在不同集群，建议您升级前将主租户切换至另一个集群，先升级备租户所在集群，再升级主租户所在集群',
                                  })}
                                </div>
                                <div>
                                  {formatMessage({
                                    id: 'ocp-v2.Component.UpgradeDrawer.AfterYouChooseToAdjustTheUpgradeOrder',
                                    defaultMessage:
                                      '2. 当您选择调整 Zone 的升级顺序后，您需要在任务流程：Stop Zone\n                      之前人工确认是否继续执行升级操作，同时建议您尽快执行完升级操作，避免出现异常行为造成集群升级问题，敬请知悉；',
                                  })}
                                </div>
                              </>
                            ) : (
                              <>
                                <div>
                                  {formatMessage({
                                    id: 'ocp-v2.Component.UpgradeDrawer.OnlyThePrimaryClusterZoneCanBeSet',
                                    defaultMessage:
                                      '1. 仅支持设置主集群 Zone\n                      升级顺序，若存在备集群，则先升级备集群，且备集群按默认顺序进行升级且中间不会停止。注意：在备集群升级完成至主集群后，需要暂停，由用户确认后再进行升级动作；',
                                  })}
                                </div>
                                <div>
                                  {formatMessage({
                                    id: 'ocp-v2.Component.UpgradeDrawer.AfterYouChooseToAdjustTheUpgradeOrder',
                                    defaultMessage:
                                      '2. 当您选择调整 Zone 的升级顺序后，您需要在任务流程：Stop Zone\n                      之前人工确认是否继续执行升级操作，同时建议您尽快执行完升级操作，避免出现异常行为造成集群升级问题，敬请知悉；',
                                  })}
                                </div>
                              </>
                            )}
                          </>
                        }
                      />
                       <Form.Item
                        name="primaryZone"
                        label={ 'Zone 升级顺序'}
                        {...MODAL_FORM_ITEM_LAYOUT}
                        wrapperCol={{ span: 24 }}
                      >
                        <DraggableTable
                          zoneOrderList={zoneOrderList}
                          setZoneOrderList={setZoneOrderList}
                        />
                      </Form.Item>
                    </>
                  )}
                </>
              );
            }}
           </Form.Item> */}
      </Form>
      <MyDrawer
        className={styles.upgradePathContainer}
        width={560}
        title={formatMessage(
          {
            id: 'ocp-v2.Overview.UpgradeDrawer.ClusterdatanameUpgradePathConfirmation',
            defaultMessage: '{clusterDataName} 升级路径确认',
          },
          { clusterDataName: clusterData.name }
        )}
        visible={pathVisible}
        onCancel={() => {
          setPathVisible(false);
          // 需要同时关闭上传文件的弹窗，否则下次展示升级路径弹窗时，上传文件弹窗也会一同出现
          setUploadVisible(false);
        }}
        footer={
          <div
            style={{
              width: 512,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Space>
              <Button
                onClick={() => {
                  setPathVisible(false);
                }}
              >
                {formatMessage({
                  id: 'ocp-v2.Overview.UpgradeDrawer.Cancel',
                  defaultMessage: '取消',
                })}
              </Button>
              <Button
                loading={upgradeLoading}
                type="primary"
                onClick={() => {
                  validateFields().then(values => {
                    const { upgradeMode, rpmFileName, freezeServer, zone } = values;
                    const currentZonePackage = find(
                      packageList,
                      rpm =>
                        `${rpm?.name}-${rpm?.version}-${rpm?.buildNumber}-${rpm?.operatingSystem}` ===
                        rpmFileName?.label
                    );
                    const destVersion = `${currentZonePackage.version}-${currentZonePackage.buildNumber}`;
                    const params = {
                      upgradeMode,
                      operatingSystem: currentZonePackage?.operatingSystem,
                      destVersion,
                      freezeServer,
                      ...(zone
                        ? {
                            zoneOrder: zoneOrderList?.map(item => item.name),
                          }
                        : {}),
                    };

                    // mode: string;
                    // release: string;
                    // version: string;
                    // upgradeObCluster({
                    //   id: clusterData.id,
                    // });
                  });
                }}
              >
                {formatMessage({
                  id: 'ocp-v2.Overview.UpgradeDrawer.Determine',
                  defaultMessage: '确定',
                })}
              </Button>
            </Space>
          </div>
        }
      ></MyDrawer>
      <UploadPackageDrawer
        visible={uploadVisible}
        onCancel={() => {
          setUploadVisible(false);
        }}
        onSuccess={() => {
          const { rpmFileName } = getFieldsValue();
          if (packageSelectRef && packageSelectRef.current && packageSelectRef?.current?.refresh) {
            packageSelectRef?.current?.refresh();
          }
        }}
      />
    </MyDrawer>
  );
};

export default UpgradeDrawer;
