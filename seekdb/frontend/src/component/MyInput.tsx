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
import { Input } from '@oceanbase/design';
import type { InputProps } from '@oceanbase/design/es/input';

interface MyInputProps extends React.FC<InputProps> {
  Search: typeof Input.Search;
  Password: typeof Input.Password;
  TextArea: typeof Input.TextArea;
}

const MyInput: MyInputProps = props => {
  return (
    <Input
      placeholder={formatMessage({
        id: 'ocp-express.src.component.MyInput.PleaseEnter',
        defaultMessage: '请输入',
      })}
      {...props}
    />
  );
};

MyInput.Search = Input.Search;
MyInput.Password = Input.Password;
MyInput.TextArea = Input.TextArea;

export default MyInput;
