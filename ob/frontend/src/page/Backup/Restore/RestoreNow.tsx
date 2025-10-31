import useDocumentTitle from '@/hook/useDocumentTitle';
import { formatMessage } from '@/util/intl';
import RestoreNow from '@/page/Backup/Component/RestoreNow';
import React from 'react';

export interface RestoreNowPageProps {
  location: {
    query: {
      tenantName: string;
    };
  };
}

const RestoreNowPage: React.FC<RestoreNowPageProps> = ({
  location: {
    query: { tenantName },
  },
}) => {
  useDocumentTitle(
    formatMessage({
      id: 'ocp-v2.Backup.Restore.RestoreNow.InitiateRecovery',
      defaultMessage: '发起恢复',
    })
  );
  return <RestoreNow tenantName={tenantName} />;
};

export default RestoreNowPage;
