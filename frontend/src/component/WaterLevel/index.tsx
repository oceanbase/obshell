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
import useStyles from './index.style';
import MouseTooltip from '@/component/MouseTooltip';

export interface WaterLevelProps extends ProgressProps {
  title: React.ReactNode;
  description: React.ReactNode;
  tooltip: React.ReactNode;
}

const WaterLevel: React.FC<WaterLevelProps> = ({ title, description, tooltip, ...restProps }) => {
  const { styles } = useStyles();
  return (
    <MouseTooltip
      overlay={tooltip}
      style={{
        color: 'rgba(0, 0, 0, 0.65)',
        fontSize: 14,
      }}
    >
      <Progress
        type="circle"
        width={90}
        strokeLinecap="square"
        format={() => (
          <div style={{ width: '80%', margin: '0 auto' }}>
            <div className={styles.title}>{title}</div>
            <div className={styles.description}>{description}</div>
          </div>
        )}
        {...restProps}
      />
    </MouseTooltip>
  );
};

export default WaterLevel;
