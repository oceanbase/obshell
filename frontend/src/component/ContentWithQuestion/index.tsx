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
import { theme } from '@oceanbase/design';
import type { TooltipProps } from '@oceanbase/design/es/tooltip';
import { QuestionCircleOutlined } from '@oceanbase/icons';
import ContentWithIcon from '@/component/ContentWithIcon';

export interface ContentWithQuestionProps {
  content?: React.ReactNode;
  /* tooltip 为空，则不展示 quertion 图标和 Tooltip */
  tooltip?: TooltipProps;
  /* 是否作为 label */
  inLabel?: boolean;
  onClick?: (e: React.SyntheticEvent) => void;
  style?: React.CSSProperties;
  className?: string;
}

const ContentWithQuestion: React.FC<ContentWithQuestionProps> = ({
  content,
  tooltip,
  inLabel,
  ...restProps
}) => {
  const { token } = theme.useToken();

  return (
    <ContentWithIcon
      content={content}
      affixIcon={
        tooltip && {
          component: QuestionCircleOutlined,
          pointable: true,
          tooltip,
          style: {
            color: token.colorIcon,
            cursor: 'help',
            marginTop: '-4px',
          },
        }
      }
      {...restProps}
    />
  );
};

export default ContentWithQuestion;
