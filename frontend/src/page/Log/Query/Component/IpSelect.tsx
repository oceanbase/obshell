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
import React, { useState } from 'react';
import { flatten } from 'lodash';
import { useRequest } from 'ahooks';
import * as ObClusterController from '@/service/ocp-express/ObClusterController';
import MySelect from '@/component/MySelect';
import SelectAllAndClearRender from '@/component/SelectAllAndClearRender';

interface queryHost {
  label?: string;
  value?: string;
  key?: string;
}

interface OCPObjscopeProps {
  onChange?: (value: string) => void;
  // 获取所有的 IP 节点
  getAllHostIps?: (v: queryHost[]) => void;
}

const OCPObjscope: React.FC<OCPObjscopeProps> = ({ onChange, getAllHostIps }) => {
  const [selectedValue, setSelectedValue] = useState<queryHost[] | null>();

  // OB 集群列表
  const { data: clusterInfoData } = useRequest(ObClusterController.getClusterInfo, {
    onSuccess: res => {
      if (res.successful) {
        const hostData = flatten((res?.data?.zones || []).map(item => item.servers || []));

        getAllHostIps?.(
          hostData?.map(item => ({
            value: `${item.ip}:${item.port}`,
            label: `${item.ip}:${item.port}`,
          }))
        );
      }
    },
  });

  const hostData = flatten((clusterInfoData?.data?.zones || []).map(item => item.servers || []));

  const selectHostOptions = hostData?.map(item => ({
    value: `${item.ip}:${item.port}`,
    label: `${item.ip}:${item.port}`,
  }));

  const onSelected = value => {
    setSelectedValue(value);
    if (onChange) {
      onChange(value);
    }
  };

  return (
    <MySelect
      mode="tags"
      value={selectedValue}
      allowClear={true}
      labelInValue={true}
      maxTagCount="responsive"
      onChange={onSelected}
      options={selectHostOptions}
      placeholder={formatMessage({
        id: 'ocp-express.Query.Component.IpSelect.NullIndicatesThatAllIpAddressesAreSelected',
        defaultMessage: '为空则表示选择全部 IP',
      })}
      dropdownRender={menu => {
        return (
          <SelectAllAndClearRender
            menu={menu}
            onSelectAll={() => {
              onSelected(selectHostOptions);
            }}
            onClearAll={() => {
              onSelected([]);
            }}
          />
        );
      }}
    />
  );
};

export default OCPObjscope;
