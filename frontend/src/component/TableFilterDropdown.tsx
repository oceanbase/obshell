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
import React, { useEffect, useState } from 'react';
import { Button, Space } from '@oceanbase/design';
import type { FilterDropdownProps } from '@oceanbase/design/es/table/interface';
import MyInput from '@/component/MyInput';

export interface TableFilterDropdownProps extends FilterDropdownProps {
  /* 确定筛选后的回调函数 */
  onConfirm?: (value: React.Key) => void;
}

const TableFilterDropdown: React.FC<TableFilterDropdownProps> = ({
  setSelectedKeys,
  selectedKeys,
  confirm,
  clearFilters,
  visible,
  onConfirm,
}) => {
  const [value, setValue] = useState('');

  const confirmFilter = () => {
    confirm();
    if (onConfirm) {
      onConfirm(selectedKeys && selectedKeys[0]);
    }
  };

  useEffect(() => {
    if (!visible) {
      confirmFilter();
    }
  }, [visible]);

  return (
    <div
      style={{ padding: 8 }}
      onBlur={() => {
        confirmFilter();
      }}
    >
      <MyInput
        spm={formatMessage({
          id: 'ocp-v2.src.component.TableFilterDropdown.TableCustomSearchSearchBox',
          defaultMessage: '表格自定义搜索-搜索框',
        })}
        onChange={e => {
          setValue(e.target.value);
          setSelectedKeys(e.target.value ? [e.target.value] : []);
        }}
        value={value}
        onPressEnter={() => {
          confirmFilter();
        }}
        style={{ width: 188, marginBottom: 8, display: 'block' }}
        data-aspm-expo
        data-aspm-param={``}
      />

      <Space>
        <Button
          spm={formatMessage({
            id: 'ocp-v2.src.component.TableFilterDropdown.TableCustomSearchSearch',
            defaultMessage: '表格自定义搜索-搜索',
          })}
          type="primary"
          onClick={() => {
            confirmFilter();
          }}
          size="small"
          style={{ width: 90 }}
          data-aspm-expo
          data-aspm-param={``}
        >
          {formatMessage({ id: 'ocp-express.src.util.component.Search', defaultMessage: '搜索' })}
        </Button>
        <Button
          spm={formatMessage({
            id: 'ocp-v2.src.component.TableFilterDropdown.TableCustomSearchReset',
            defaultMessage: '表格自定义搜索-重置',
          })}
          onClick={() => {
            if (clearFilters) {
              clearFilters();
              setValue('');
            }
            confirmFilter();
          }}
          size="small"
          style={{ width: 90 }}
          data-aspm-expo
          data-aspm-param={``}
        >
          {formatMessage({ id: 'ocp-express.src.util.component.Reset', defaultMessage: '重置' })}
        </Button>
      </Space>
    </div>
  );
};

export default TableFilterDropdown;
