import { formatMessage } from '@/util/intl';
import React, { useRef, useState } from 'react';
import { find } from 'lodash';
import { Descriptions, Form, Modal, token } from '@oceanbase/design';
import { ExclamationCircleFilled } from '@oceanbase/icons';
import { ContentWithIcon as OBUIContentWithIcon } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import { MODAL_FORM_ITEM_LAYOUT } from '@/constant';
import { taskSuccess } from '@/util/task';
import ContentWithIcon from '@/component/ContentWithIcon';
import type { MyDrawerProps } from '@/component/MyDrawer';
import MyDrawer from '@/component/MyDrawer';
import MySelect from '@/component/MySelect';
import PackageSelect, { PackageSelectRef } from '@/component/PackageSelect';
import UploadPackageDrawer from '@/component/UploadPackageDrawer';
import { obUpgrade } from '@/service/obshell/upgrade';
import { useCluster } from '@/hook/useCluster';

const { Option } = MySelect;

export interface UpgradeDrawerProps extends MyDrawerProps {
  isStandAloneCluster?: boolean;

  clusterData: API.ClusterInfo;
  onSuccess: () => void;
  onCancel: () => void;
}

const UpgradeDrawer: React.FC<UpgradeDrawerProps> = ({
  visible,
  isStandAloneCluster,
  clusterData,
  onSuccess,
  onCancel,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;

  const { isStandalone, isDistributedBusiness } = useCluster();

  const packageSelectRef = useRef<PackageSelectRef>();

  const [uploadVisible, setUploadVisible] = useState<boolean>(false);
  // 软件包列表
  const [packageList, setPackageList] = useState<API.UpgradePkgInfo[]>([]);

  // 集群升级
  const { run: upgradeObCluster } = useRequest(obUpgrade, {
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
        // dispatch({
        //   type: 'cluster/getClusterData',
        //   payload: {
        //     id: clusterData.id,
        //   },
        // });
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
                version: currentZonePackage?.version || '',
                release: currentZonePackage?.release_distribution || '',
                freeze_server: true,
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
                ) : isDistributedBusiness ? (
                  formatMessage({
                    id: 'obshell.Overview.UpgradeDrawer.OnlyPackagesStartingWithOceanbaseAreSupported',
                    defaultMessage: 'oceanbase',
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
      </Form>
      <UploadPackageDrawer
        visible={uploadVisible}
        onCancel={() => {
          setUploadVisible(false);
        }}
        onSuccess={() => {
          if (packageSelectRef && packageSelectRef.current && packageSelectRef?.current?.refresh) {
            packageSelectRef?.current?.refresh();
          }
        }}
      />
    </MyDrawer>
  );
};

export default UpgradeDrawer;
