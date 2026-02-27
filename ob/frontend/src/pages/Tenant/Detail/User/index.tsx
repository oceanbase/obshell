import { history, useLocation, useParams, useSelector } from '@umijs/max';
import React, { useEffect } from 'react';
import MySQL from './MySQL';
import Oracle from './Oracle';

const Index: React.FC = () => {
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const { tenantName = '' } = useParams<{ tenantName: string }>();
  const { pathname } = useLocation();

  useEffect(() => {
    if (tenantData?.locked === 'YES' && tenantName) {
      history.push(`/cluster/tenant/${tenantName}`);
    }
  }, [tenantName, tenantData?.locked]);

  return (
    <>
      {tenantData?.mode === 'MYSQL' && <MySQL tenantName={tenantName} />}

      {tenantData?.mode === 'ORACLE' && <Oracle pathname={pathname} tenantName={tenantName} />}
    </>
  );
};

export default Index;
