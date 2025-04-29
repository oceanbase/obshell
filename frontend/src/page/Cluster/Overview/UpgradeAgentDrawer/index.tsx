import { formatMessage } from '@/util/intl';
import { useDispatch, useSelector } from 'umi';
import React, { useRef, useState, useEffect } from 'react';
import { compact, filter, find, flatten, groupBy, includes, some, sum, uniq, uniqBy } from 'lodash';
import {
  Alert,
  Button,
  Descriptions,
  Form,
  Modal,
  Space,
  Spin,
  Steps,
  Switch,
  Tabs,
  Tag,
  token,
  Tooltip,
  Typography,
} from '@oceanbase/design';
import {
  ExclamationCircleFilled,
  InfoCircleFilled,
  QuestionCircleOutlined,
} from '@oceanbase/icons';
import { directTo, isNullValue, joinComponent } from '@oceanbase/util';
import { ContentWithIcon as OBUIContentWithIcon } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import { MODAL_FORM_ITEM_LAYOUT } from '@/constant';
import BinlogAssociationMsg from '@/component/BinlogAssociationMsg';
// import * as ObClusterController from '@/service/ocp-all-in-one/ObClusterController';
import { buildVersionCompare, versionCompare } from '@/util/package';
import { taskSuccess } from '@/util/task';
import ContentWithIcon from '@/component/ContentWithIcon';
import ContentWithReload from '@/component/ContentWithReload';
import type { MyDrawerProps } from '@/component/MyDrawer';
import MyDrawer from '@/component/MyDrawer';
import MySelect from '@/component/MySelect';
// import type { PackageSelectRef } from '@/component/PackageSelect';
import PackageSelect from '@/component/PackageSelect';
import UploadPackageDrawer from '@/component/UploadPackageDrawer';
// import DraggableTable from '../DraggableTable';
import styles from './index.less';
import { agentUpgrade } from '@/service/obshell/upgrade';
import { getObInfo } from '@/service/obshell/ob';
import { getAgentInfo, getStatus } from '@/service/obshell/v1';

const { Step } = Steps;
const { TabPane } = Tabs;
const { Option } = MySelect;
const { Paragraph } = Typography;
const ObClusterController = {};
export interface UpgradeDrawerProps extends MyDrawerProps {
  isStandAloneCluster?: boolean;
  // 集群的架构列表
  architectureList?: string[];
  clusterData: API.ClusterConfig;
  obClusterRelatedBinlogService?: API.TenantBinlogService;
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

const UpgradeAgentDrawer: React.FC<UpgradeDrawerProps> = ({
  visible,
  obClusterRelatedBinlogService,
  isStandAloneCluster,
  architectureList,
  clusterData,
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
        // dispatch({
        //   type: 'cluster/getClusterData',
        //   payload: {
        //     id: clusterData.id,
        //   },
        // });
      }
    },
  });
  const { data } = useRequest(getAgentInfo);
  const agentData = data?.data || {};

  return (
    <MyDrawer
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
      cancelText={formatMessage({
        id: 'ocp-express.component.FormEditTable.Cancel',
        defaultMessage: '取消',
      })}
      confirmLoading={loading}
      {...restProps}
    >
      <Form form={form} preserve={false} hideRequiredMark={true} className="form-with-small-margin">
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
          {uniq(
            flatten(clusterData?.zones?.map(item => item.servers || [])).map(
              item => item?.architecture
            ) || []
          ).join('、')}
        </Form.Item>
        <Form.Item
          label={formatMessage({
            id: 'ocp-v2.Overview.UpgradeAgentDrawer.ObshellAgentVersion',
            defaultMessage: 'OBShell Agent 版本',
          })}
          name="fileName"
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-v2.Overview.UpgradeAgentDrawer.PleaseSelectTheObshellAgent',
                defaultMessage: '请选择 OBShell Agent 版本',
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
    </MyDrawer>
  );
};

export default UpgradeAgentDrawer;
