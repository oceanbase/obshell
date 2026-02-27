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
import { Typography } from '@oceanbase/design';

const { Text } = Typography;

interface InstanceConnectionStringProps {
  connectionString: string;
}
const InstanceConnectionString: React.FC<InstanceConnectionStringProps> = ({
  connectionString,
}) => {
  if (!connectionString || connectionString === '-') {
    return <>-</>;
  }

  // 确保 connectionString 是字符串类型并且有内容
  const displayText = String(connectionString).trim();

  if (!displayText) {
    return <>-</>;
  }

  return (
    <Text ellipsis={{ tooltip: displayText }} copyable={true}>
      {displayText}
    </Text>
  );
};

export default InstanceConnectionString;
