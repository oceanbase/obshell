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
import { isEnglish } from '@/util';

export interface OCPLogoProps {
  onClick?: (e: React.SyntheticEvent) => void;
  height?: number;
  mode?: 'default' | 'simple';
  style?: React.CSSProperties;
  className?: string;
}

const OCPLogo: React.FC<OCPLogoProps> = ({
  mode = 'default',
  height = mode === 'default' ? 80 : 24,
  style,
  ...restProps
}) => {
  const logoUrl = isEnglish()
    ? '/assets/logo/ocp_express_logo_en.svg'
    : '/assets/logo/ocp_express_logo_zh.svg';
  const simpleLogoUrl = isEnglish()
    ? '/assets/logo/ocp_express_simple_logo_en.svg'
    : '/assets/logo/ocp_express_simple_logo_zh.svg';
  return (
    <img
      src={mode === 'default' ? logoUrl : simpleLogoUrl}
      alt="logo"
      {...restProps}
      style={{
        height,
        ...(style || {}),
      }}
    />
  );
};

export default OCPLogo;
