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

import { formatMessage } from '@/util/intl';
import React from 'react';
import { Select } from '@oceanbase/design';
import type { SelectProps } from '@oceanbase/design/es/select';
// import { useRequest } from 'ahooks';
// import { getTenantOverView } from '@/service/obshell/tenant';

const { Option } = Select;

export interface TenantSelectProps extends SelectProps<string | number> {
  /* 值对应的属性字段 */
  valueProp?: 'obTenantId' | 'name';
  showAllOption?: boolean;
  allOption?: React.ReactNode;
  // 租户列表更新成功后的回调
  onSuccess?: (hostList: API.Tenant[]) => void;
}

const TenantSelect: React.FC<TenantSelectProps> = ({
  // 用作 value 的属性字段，默认为 obTenantId
  valueProp = 'obTenantId',
  showAllOption = false,
  allOption,
  onSuccess,
  ...restProps
}) => {
  // const { data, loading } = useRequest(getTenantOverView, {
  //   // 需要设置 defaultParams，否则请求不发起
  //   defaultParams: [{}],
  //   onSuccess: res => {
  //     if (res.successful) {
  //       if (onSuccess) {
  //         onSuccess(res?.data?.contents || []);
  //       }
  //     }
  //   },
  // });
  // const tenantList = data?.data?.contents || [];
  const tenantList = [];

  return (
    <Select
      // loading={loading}
      showSearch={true}
      optionFilterProp="children"
      dropdownMatchSelectWidth={false}
      placeholder={formatMessage({
        id: 'ocp-express.component.common.TenantSelect.SelectATenant',
        defaultMessage: '请选择租户',
      })}
      {...restProps}
    >
      {showAllOption && (
        <Option value="all">
          {formatMessage({
            id: 'ocp-express.component.common.TenantSelect.All',
            defaultMessage: '全部',
          })}
        </Option>
      )}
      {allOption}
      {tenantList.map(item => (
        <Option key={item.tenant_id} value={item[valueProp]} disabled={item.status === 'CREATING'}>
          {item.tenant_name}
        </Option>
      ))}
    </Select>
  );
};

export default TenantSelect;
