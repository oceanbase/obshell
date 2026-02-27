import type { MyDrawerProps } from '@/component/MyDrawer';
import PackageSelect from '@/component/PackageSelect';
import { agentUpgrade } from '@/service/obshell/upgrade';
import { getAgentInfo } from '@/service/obshell/v1';
import { formatMessage } from '@/util/intl';
import { taskSuccess } from '@/util/task';
import { Form, Modal } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import React, { useState } from 'react';
export interface UpgradeDrawerProps extends MyDrawerProps {
  seekdbInfoData: API.ObserverInfo;
  onSuccess: () => void;
  onCancel: () => void;
}

const UpgradeAgentDrawer: React.FC<UpgradeDrawerProps> = ({
  visible,
  seekdbInfoData,
  onSuccess,
  onCancel,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;
  const [packageList, setPackageList] = useState<API.UpgradePkgInfo[]>([]);

  const { runAsync, loading } = useRequest(agentUpgrade, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const taskId = res.data?.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'ocp-v2.Overview.UpgradeAgentDrawer.VersionUpgradeTaskSubmittedSuccessfully',
            defaultMessage: '版本升级的任务提交成功',
          }),
        });

        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });
  const { data } = useRequest(getAgentInfo);
  const agentData = data?.data || {};

  return (
    <Modal
      width={620}
      title={formatMessage({
        id: 'ocp-v2.Overview.UpgradeAgentDrawer.AgentUpgradeVersion',
        defaultMessage: 'Agent 升级版本',
      })}
      visible={visible}
      destroyOnClose={true}
      onOk={() => {
        validateFields().then(({ fileName }) => {
          const pkg = packageList.find(item => item.pkg_id === fileName.value);
          Modal.confirm({
            title: formatMessage({
              id: 'ocp-v2.Overview.UpgradeAgentDrawer.AreYouSureYouWant',
              defaultMessage: '确定要升级吗？',
            }),
            onOk: () => {
              if (pkg) {
                runAsync({
                  version: pkg?.version,
                  release: pkg.release_distribution,
                });
              }
            },
          });
        });
      }}
      onCancel={() => onCancel()}
      okText={formatMessage({
        id: 'ocp-v2.Overview.UpgradeAgentDrawer.Upgrade',
        defaultMessage: '升级',
      })}
      confirmLoading={loading}
      {...restProps}
    >
      <Form form={form} preserve={false} hideRequiredMark={true} className="form-with-small-margin">
        <div style={{ borderRadius: 6, background: '#f8fafe', padding: 16, marginBottom: 24 }}>
          <Form.Item
            label={formatMessage({
              id: 'ocp-v2.Overview.UpgradeAgentDrawer.InstalledVersion',
              defaultMessage: '已安装版本',
            })}
          >
            {agentData?.version}
          </Form.Item>
          <Form.Item
            label={formatMessage({
              id: 'ocp-v2.Overview.UpgradeAgentDrawer.HardwareArchitecture',
              defaultMessage: '硬件架构',
            })}
          >
            {seekdbInfoData?.architecture}
          </Form.Item>
          <Form.Item
            label={formatMessage({
              id: 'seekdb.Overview.UpgradeAgentDrawer.ObshellPort',
              defaultMessage: 'obshell 端口',
            })}
            style={{ marginBottom: 0 }}
          >
            {seekdbInfoData?.obshell_port}
          </Form.Item>
        </div>
        <Form.Item
          label={formatMessage({
            id: 'ocp-v2.Overview.UpgradeAgentDrawer.ObshellAgentVersion',
            defaultMessage: 'obshell Agent 版本',
          })}
          name="fileName"
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-v2.Overview.UpgradeAgentDrawer.PleaseSelectTheObshellAgent',
                defaultMessage: '请选择 obshell Agent 版本',
              }),
            },
          ]}
          style={{ marginBottom: 0 }}
        >
          <PackageSelect
            useType="UPGRADE_AGENT"
            labelInValue={true}
            style={{ width: '100%' }}
            isHeterogeneousUpgrade={true}
            currentBuildVersionForUpgrade={agentData?.version}
            onSuccess={newPackageList => {
              setPackageList(newPackageList);
            }}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default UpgradeAgentDrawer;
