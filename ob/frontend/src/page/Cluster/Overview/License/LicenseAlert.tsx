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

  return isStandalone && isTrial ? (
    <Alert
      showIcon
      type={getAlertType()}
      message={
        effectiveDays <= 0
          ? formatMessage({
              id: 'OBShell.Overview.License.LicenseAlert.OceanbaseTheStandAloneTrial',
              defaultMessage:
                'OceanBase 单机版试用 License 已过期，请申请正式 License 并正确配置。',
            })
          : formatMessage(
              {
                id: 'OBShell.Overview.License.LicenseAlert.OceanbaseTheStandAloneTrial.1',
                defaultMessage:
                  'OceanBase 单机版试用 License 即将于 {expiredTime} 过期，请申请正式 License 并正确配置。',
              },
              { expiredTime: expiredTime }
            )
      }
      style={{ marginBottom: 10 }}
    />
  ) : null;
};

export default LicenseAlert;
