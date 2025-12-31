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
    id: 'OBShell.Overview.License.LicenseDetailModal.AutomaticallyGeneratedByClusterStartup',
    defaultMessage: '由集群启动时自动生成。',
  }),
  license_id: formatMessage({
    id: 'OBShell.Overview.License.LicenseDetailModal.GeneratedByOceanbase',
    defaultMessage: '由 OceanBase 生成。',
  }),
  license_code: formatMessage({
    id: 'OBShell.Overview.License.LicenseDetailModal.GeneratedByContractNumber',
    defaultMessage: '由合同号生成。',
  }),
  license_type: formatMessage({
    id: 'OBShell.Overview.License.LicenseDetailModal.CommercialBusinessEditionTrialTrial',
    defaultMessage: '· Commercial：商业版。|· Trial：试用版。',
  }),
  product_type: formatMessage({
    id: 'OBShell.Overview.License.LicenseDetailModal.SeStandsForStandAlone',
    defaultMessage: '· SE：代表单机企业版。|· DE：代表分布式企业版。',
  }),
  options: formatMessage({
    id: 'OBShell.Overview.License.LicenseDetailModal.TheFeatureOptionsIncludedIn',
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
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.ClusterId',
          defaultMessage: '集群 ID',
        }),
        value: license.cluster_ulid,
        description: LICENSE_DESCRIPTIONS.cluster_ulid.split('|'),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.AuthorizedUserName',
          defaultMessage: '授权用户名',
        }),
        value: license.end_user,
      },
      {
        name: 'License ID',
        value: license.license_id,
        description: LICENSE_DESCRIPTIONS.license_id.split('|'),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.LicenseCoding',
          defaultMessage: 'License 编码',
        }),
        value: license.license_code,
        description: LICENSE_DESCRIPTIONS.license_code.split('|'),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.LicenseType',
          defaultMessage: 'License 类型',
        }),
        value: license.license_type,
        description: LICENSE_DESCRIPTIONS.license_type.split('|'),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.ProductType',
          defaultMessage: '产品类型',
        }),
        value: license.product_type,
        description: LICENSE_DESCRIPTIONS.product_type.split('|'),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.LicenseIssuedTime',
          defaultMessage: 'License 颁发时间',
        }),
        value: formatTime(license.issuance_date),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.LicenseActivationTime',
          defaultMessage: 'License 激活时间',
        }),
        value: formatTime(license.activation_time),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.LicenseExpirationTime',
          defaultMessage: 'License 过期时间',
        }),
        value: formatTime(license.expired_time),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.LicenseOptions',
          defaultMessage: 'License 选件',
        }),
        value: isTrial ? DEFAULT_LICENSE_OPTIONS.join('、') : license?.options,
        description: LICENSE_DESCRIPTIONS.options.split('|'),
      },
      {
        name: formatMessage({
          id: 'OBShell.Overview.License.LicenseDetailModal.NumberOfAuthorizedNodes',
          defaultMessage: '授权节点数',
        }),
        value: license.node_num,
      },
    ];
  }, [license, isTrial]);

  return (
    <Modal
      title={formatMessage({
        id: 'OBShell.Overview.License.LicenseDetailModal.ViewLicense',
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
