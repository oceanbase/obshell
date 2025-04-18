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
import { useSelector, useDispatch } from 'umi';
import React, { useEffect } from 'react';
import { Select } from '@oceanbase/design';
import type { SelectProps } from '@oceanbase/design/es/select';

const { Option } = Select;

export interface IUserSelectProps extends SelectProps<number | string> {
  valueProp?: 'id' | 'username';
}

const UserSelect: React.FC<IUserSelectProps> = ({ valueProp = 'id', ...restProps }) => {
  const { userListData } = useSelector((state: DefaultRootState) => state.iam);
  const userList = userListData.contents || [];

  const dispatch = useDispatch();

  useEffect(() => {
    dispatch({
      type: 'iam/getUserListData',
      payload: {},
    });
  }, []);

  return (
    <Select
      showSearch={true}
      optionFilterProp="children"
      placeholder={formatMessage({
        id: 'ocp-express.component.common.UserSelect.Select',
        defaultMessage: '请选择',
      })}
      {...restProps}
    >
      {userList.map(item => (
        <Option key={item.id} value={item[valueProp]}>
          {item.username}
        </Option>
      ))}
    </Select>
  );
};

export default UserSelect;
