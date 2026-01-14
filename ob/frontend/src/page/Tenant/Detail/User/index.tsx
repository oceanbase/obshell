import { useSelector, history } from 'umi';
import React, { useEffect } from 'react';
import MySQL from './MySQL';
import Oracle from './Oracle';

export interface IndexProps {
  match: {
    params: {
      tenantName: string;
    };
  };
  location: {
    pathname: string;
  };
}

const Index: React.FC<IndexProps> = ({
  match: {
    params: { tenantName },
  },
  location: { pathname },
}) => {
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);

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
