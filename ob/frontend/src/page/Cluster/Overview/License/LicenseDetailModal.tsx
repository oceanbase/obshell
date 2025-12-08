import { formatMessage } from '@/util/intl';
import { isTrialLicense } from '@/util/cluster';
import { formatTime } from '@/util/datetime';
import React, { useMemo } from 'react';
import { Descriptions, Modal } from '@oceanbase/design';
import { ContentWithQuestion } from '@oceanbase/ui';

interface LicenseDetailModalProps {
  clusterData: API.ClusterInfo;
  open: boolean;
  onCancel: () => void;
}

const LICENSE_DESCRIPTIONS = {
  cluster_ulid: formatMessage({
    id: 'ocp-v2.Component.License.LicenseDetailModal.AutomaticallyGeneratedByClusterStartup',
    defaultMessage: '由集群启动时自动生成。',
  }),
  license_id: formatMessage({
    id: 'ocp-v2.Component.License.LicenseDetailModal.GeneratedByOceanbase',
    defaultMessage: '由 OceanBase 生成。',
  }),
  license_code: formatMessage({
    id: 'ocp-v2.Component.License.LicenseDetailModal.GeneratedByContractNumber',
    defaultMessage: '由合同号生成。',
  }),
  license_type: formatMessage({
    id: 'ocp-v2.Component.License.LicenseDetailModal.CommercialBusinessEditionTrialTrial',
    defaultMessage: '· Commercial：商业版。|· Trial：试用版。',
  }),
  product_type: formatMessage({
    id: 'ocp-v2.Component.License.LicenseDetailModal.SeStandsForStandAlone',
    defaultMessage: '· SE：代表单机企业版。|· DE：代表分布式企业版。',
  }),
  options: formatMessage({
    id: 'ocp-v2.Component.License.LicenseDetailModal.TheFeatureOptionsIncludedIn',
    defaultMessage:
      'License 中所包含的功能选件。|· EE：代表 OceanBase 单机企业版基础功能。|· PS：代表 OceanBase 单机集群间复制选件。|· MT：代表 OceanBase 单机多租户选件。|· AP：代表 OceanBase 单机 OLAP 选件。',
  }),
};

const DEFAULT_LICENSE_OPTIONS = ['EE', 'PS', 'MT', 'AP'];

const LicenseDetailModal: React.FC<LicenseDetailModalProps> = ({ clusterData, open, onCancel }) => {
  const { license = {} as API.ObLicense } = clusterData || {};
  const isTrial = isTrialLicense(license);

  const formatLicenseValue = useMemo(() => {
    return [
      {
        name: '集群 ID',
        value: license.cluster_ulid,
        description: LICENSE_DESCRIPTIONS.cluster_ulid.split('|'),
      },
      {
        name: '授权用户名',
        value: license.end_user,
      },
      {
        name: 'License ID',
        value: license.license_id,
        description: LICENSE_DESCRIPTIONS.license_id.split('|'),
      },
      {
        name: 'License 编码',
        value: license.license_code,
        description: LICENSE_DESCRIPTIONS.license_code.split('|'),
      },
      {
        name: 'License 类型',
        value: license.license_type,
        description: LICENSE_DESCRIPTIONS.license_type.split('|'),
      },
      {
        name: '产品类型',
        value: license.product_type,
        description: LICENSE_DESCRIPTIONS.product_type.split('|'),
      },
      {
        name: 'License 颁发时间',
        value: formatTime(license.issuance_date),
      },
      {
        name: 'License 激活时间',
        value: formatTime(license.activation_time),
      },
      {
        name: 'License 过期时间',
        value: formatTime(license.expired_time),
      },
      {
        name: 'License 选件',
        value: isTrial ? DEFAULT_LICENSE_OPTIONS.join('、') : license.options?.join('、'),
        description: LICENSE_DESCRIPTIONS.options.split('|'),
      },
      {
        name: '授权节点数',
        value: license.node_num,
      },
    ];
  }, [license]);

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-v2.Component.License.LicenseDetailModal.ViewLicense',
        defaultMessage: '查看 License',
      })}
      footer={null}
      open={open}
      onCancel={onCancel}
      width={424}
    >
      <Descriptions column={1}>
        {formatLicenseValue.map(item => {
          const { description, value, name } = item;
          return (
            <Descriptions.Item
              key={name}
              label={
                description?.length ? (
                  <ContentWithQuestion
                    content={item.name}
                    tooltip={{
                      title: description?.map((des, index) => <div key={index}>{des}</div>),
                      overlayStyle: { maxWidth: 400 },
                    }}
                  />
                ) : (
                  name
                )
              }
            >
              {value}
            </Descriptions.Item>
          );
        })}
      </Descriptions>
    </Modal>
  );
};

export default LicenseDetailModal;
