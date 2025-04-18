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
import { Progress } from '@oceanbase/design';
import type { ProgressProps } from '@oceanbase/design/es/progress';
import { useTheme } from '@oceanbase/charts';
import useStyles from './index.style';

export interface MyProgressProps extends ProgressProps {
  /* 外层容器的类型 */
  wrapperClassName?: string;
  /* 前缀 */
  prefix?: React.ReactNode;
  prefixWidth?: number;
  /* 后缀 */
  affix?: React.ReactNode;
  affixWidth?: number;
  prefixStyle?: boolean;
}

const MyProgress: React.FC<MyProgressProps> = ({
  wrapperClassName,
  strokeColor,
  strokeLinecap = 'square',
  prefix,
  prefixWidth = 40,
  affix,
  affixWidth = 40,
  prefixStyle,
  ...restProps
}: MyProgressProps) => {
  const { styles } = useStyles();
  const theme = useTheme();
  return (
    <span className={`${styles.progress} ${wrapperClassName}`}>
      {prefix && (
        <span className={prefixStyle ? null : styles.prefix} style={{ width: prefixWidth }}>
          {prefix}
        </span>
      )}

      <span className={styles.wrapper}>
        <Progress
          strokeColor={strokeColor || theme.defaultColor}
          strokeLinecap={strokeLinecap}
          {...restProps}
        />
      </span>
      {affix && (
        <span className={styles.affix} style={{ width: affixWidth }}>
          {affix}
        </span>
      )}
    </span>
  );
};

export default MyProgress;
