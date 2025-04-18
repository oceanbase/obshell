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
import { Tooltip, Typography, Space } from '@oceanbase/design';
import { FolderOpenOutlined } from '@oceanbase/icons';

const { Text } = Typography;

interface RenderConnectionStringProps {
  maxWidth?: number;
  connectionStrings: API.ObproxyAndConnectionString[];
  callBack?: () => void;
}

const RenderConnectionString: React.FC<RenderConnectionStringProps> = ({
  maxWidth = 180,
  connectionStrings,
  callBack,
}) => {
  if (connectionStrings?.length && connectionStrings?.length > 0) {
    const connectionString = connectionStrings[0]?.connection_string;
    return (
      <Tooltip placement="topLeft" title={connectionString}>
        <Space>
          <Text style={{ maxWidth }} ellipsis={true} copyable={true}>
            {connectionString}
          </Text>
          {connectionStrings?.length > 1 && (
            <Tooltip
              title={formatMessage({
                id: 'ocp-express.Detail.Component.RenderConnectionString.ViewMore',
                defaultMessage: '查看更多',
              })}
            >
              <a
                onClick={() => {
                  if (callBack) {
                    callBack();
                  }
                }}
              >
                <FolderOpenOutlined />
              </a>
            </Tooltip>
          )}
        </Space>
      </Tooltip>
    );
  }
  return <>-</>;
};

export default RenderConnectionString;
