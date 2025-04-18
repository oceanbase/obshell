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
import { Input, Select } from '@oceanbase/design';
import type { InputProps } from '@oceanbase/design/es/input';
import type { SelectProps } from '@oceanbase/design/es/select';

export interface ValueType {
  key?: string;
  value?: string;
}

export interface SelectAndInputGroupProps {
  value?: ValueType;
  onChange?: (value?: ValueType) => void;
  selectProps?: SelectProps;
  inputProps?: InputProps;
}

const SelectAndInputGroup: React.FC<SelectAndInputGroupProps> = ({
  value,
  onChange,
  selectProps,
  inputProps,
}) => {
  return (
    <Input.Group compact={true}>
      <Select
        placeholder={formatMessage({
          id: 'ocp-express.src.component.MySelect.PleaseSelect',
          defaultMessage: '请选择',
        })}
        {...selectProps}
        style={{
          width: 130,
          ...selectProps?.style,
        }}
        value={value?.key}
        onChange={selectValue => {
          if (onChange) {
            onChange({
              ...value,
              key: selectValue,
              // 切换筛选维度，需要重置筛选值
              value: undefined,
            });
          }
        }}
      />
      <Input
        allowClear={true}
        placeholder={formatMessage({
          id: 'ocp-express.src.component.MyInput.PleaseEnter',
          defaultMessage: '请输入',
        })}
        {...inputProps}
        style={{
          width: 'calc(100% - 130px)',
          ...inputProps?.style,
        }}
        value={value?.value}
        onChange={e => {
          if (onChange) {
            onChange({
              ...value,
              value: e.target.value,
            });
          }
        }}
      />
    </Input.Group>
  );
};

export default SelectAndInputGroup;
