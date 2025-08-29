import { formatMessage } from '@/util/intl';
/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { useSelector, history } from 'umi';
import React, { useEffect } from 'react';
import MySQL from './MySQL';
import Oracle from './Oracle';

export interface IndexProps {
  match: {
    params: {
      clusterId: number;
      tenantId: number;
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
      history.push(`/tenant/${tenantName}`);
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
