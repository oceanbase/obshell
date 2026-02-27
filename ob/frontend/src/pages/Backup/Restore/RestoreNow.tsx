import useDocumentTitle from '@/hook/useDocumentTitle';
import { useQueryParam } from '@/hook/useQueryParams';
import RestoreNow from '@/pages/Backup/Component/RestoreNow';
import { formatMessage } from '@/util/intl';
import React from 'react';

const RestoreNowPage: React.FC = () => {
  useDocumentTitle(
    formatMessage({
      id: 'ocp-v2.Backup.Restore.RestoreNow.InitiateRecovery',
      defaultMessage: '发起恢复',
    })
  );
  const tenantName = useQueryParam('tenantName');
  return <RestoreNow tenantName={tenantName} />;
};

export default RestoreNowPage;
