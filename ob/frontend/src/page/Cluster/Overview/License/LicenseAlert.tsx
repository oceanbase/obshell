import { formatMessage } from '@/util/intl';
import { getDaysUntilExpiry, getLicenseExpiredTime, isTrialLicense } from '@/util/cluster';
import React from 'react';
import { Alert } from '@oceanbase/design';

interface LicenseAlertProps {
  clusterData: API.ClusterInfo;
}

const LicenseAlert: React.FC<LicenseAlertProps> = ({ clusterData }) => {
  const { license = {} as API.ObLicense } = clusterData || {};
  // 计算是否为单机版
  const isStandalone = clusterData?.is_standalone || false;

  const isTrial = isTrialLicense(license);
  const expiredTime = getLicenseExpiredTime(license);

  const effectiveDays = getDaysUntilExpiry(expiredTime);
  const getAlertType = () => {
    if (effectiveDays <= 0) {
      return 'error';
    } else if (effectiveDays <= 30) {
      return 'warning';
    } else {
      return 'info';
    }
  };

  const trialLabel = isTrial
    ? formatMessage({ id: 'ocp-v2.Component.License.LicenseAlert.Trial', defaultMessage: '试用' })
    : '';

  return isStandalone ? (
    <Alert
      showIcon
      type={getAlertType()}
      message={
        <>
          {effectiveDays <= 0
            ? formatMessage(
                {
                  id: 'ocp-v2.Component.License.LicenseAlert.OceanbaseTheLicenseOfThe',
                  defaultMessage:
                    'OceanBase 单机版{trialLabel} License 已过期，请申请正式 License 并正确配置。',
                },
                { trialLabel: trialLabel }
              )
            : formatMessage(
                {
                  id: 'ocp-v2.Component.License.LicenseAlert.TheStandAloneVersionTriallabel',
                  defaultMessage:
                    'OceanBase 单机版{trialLabel} License 即将于 {expiredTime} 过期，请申请正式 License 并正确配置。',
                },
                { trialLabel: trialLabel, expiredTime: expiredTime }
              )}
        </>
      }
      style={{ marginBottom: 10 }}
    />
  ) : null;
};

export default LicenseAlert;
