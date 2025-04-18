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
import { SyncOutlined } from '@oceanbase/icons';
import ContentWithIcon from '@/component/ContentWithIcon';

export interface ContentWithReloadProps {
  content?: React.ReactNode;
  spin?: boolean;
  onClick?: (e: React.UIEvent) => void;
  style?: React.CSSProperties;
  className?: string;
}

const ContentWithReload: React.FC<ContentWithReloadProps> = ({
  content,
  spin = false,
  onClick,
  ...restProps
}) => {
  return (
    <ContentWithIcon
      content={content}
      affixIcon={{
        component: SyncOutlined,
        spin,
        pointable: true,
        style: {
          fontSize: 14,
          marginTop: -4,
          marginRight: 4,
        },
        onClick,
        tooltip: {
          title: formatMessage({
            id: 'ocp-express.src.component.ContentWithReload.Refresh',
            defaultMessage: '刷新',
          }),
        },
      }}
      {...restProps}
    />
  );
};

export default ContentWithReload;
