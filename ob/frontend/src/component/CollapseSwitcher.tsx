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
import { DownOutlined, UpOutlined } from '@oceanbase/icons';
import { noop } from 'lodash';

interface CollapseSwictherProps {
  textList?: React.ReactElement[];
  collapsed: boolean;
  onCollapse: (collapsed: boolean) => void;
}

class CollapseSwicther extends React.PureComponent<CollapseSwictherProps> {
  public state = {};

  public render() {
    const {
      textList = [
        formatMessage({
          id: 'ocp-express.component-legacy.common.CollapseSwitcher.Expand',
          defaultMessage: '展开',
        }),
        formatMessage({
          id: 'ocp-express.component-legacy.common.CollapseSwitcher.PutAway',
          defaultMessage: '收起',
        }),
      ],
      collapsed = false, // 默认展开
      onCollapse = noop,
    } = this.props;
    return (
      <a
        onClick={() => {
          onCollapse(!collapsed);
        }}
      >
        <span style={{ marginRight: 4 }}>{collapsed ? textList[0] : textList[1]}</span>
        {collapsed ? <DownOutlined /> : <UpOutlined />}
      </a>
    );
  }
}

export default CollapseSwicther;
