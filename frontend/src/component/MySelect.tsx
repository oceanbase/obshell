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

const { Option, OptGroup } = Select;

const MySelect: React.FC<SelectProps<any>> = ({ children, ...restProps }) => (
  <Select
    placeholder={formatMessage({
      id: 'ocp-express.src.component.MySelect.PleaseSelect',
      defaultMessage: '请选择',
    })}
    {...restProps}
  >
    {children}
  </Select>
);

MySelect.Option = Option;
MySelect.OptGroup = OptGroup;

export default MySelect;
