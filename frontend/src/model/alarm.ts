import { useState } from 'react';

export default () => {
  const [clusterList, setClusterList] = useState<API.ClusterInfo[]>([]);
  const [tenantList, setTenantList] = useState<API.DbaObTenant[]>([]);

  return {
    clusterList,
    setClusterList,
    tenantList,
    setTenantList,
  };
};
