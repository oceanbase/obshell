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

import React from 'react';
import { Divider } from '@oceanbase/design';
import { PlusOutlined } from '@oceanbase/icons';
import ContentWithIcon from '@/component/ContentWithIcon';

const SelectDropdownRender: React.FC<{
  menu: any;
  text: React.ReactNode;
  onClick?: () => void;
}> = ({ menu, text, ...restProps }) => {
  return (
    <div>
      {menu}
      <div>
        <Divider style={{ margin: '4px 0' }} />
        <div
          style={{
            padding: '4px 8px 8px 8px',
            cursor: 'pointer',
            // 由于 Select 在 empty 状态会将下拉菜单的文字都置灰，会影响 SelectDropdownRender，这里做样式覆盖
            color: 'rgba(0, 0, 0, 0.85)',
          }}
          {...restProps}
        >
          <ContentWithIcon
            prefixIcon={{
              component: PlusOutlined,
            }}
            content={text}
          />
        </div>
      </div>
    </div>
  );
};

export default SelectDropdownRender;
