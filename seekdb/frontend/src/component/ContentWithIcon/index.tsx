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

import { Badge, Tooltip } from '@oceanbase/design';
import React, { isValidElement } from 'react';
import classNames from 'classnames';
import Icon from '@oceanbase/icons';
import type { IconComponentProps } from '@oceanbase/icons/lib/components/Icon';
import type { BadgeProps } from '@oceanbase/design/es/badge';
import type { TooltipProps } from '@oceanbase/design/es/tooltip';
import useStyles from './index.style';

interface IconConfig extends IconComponentProps {
  badge?: BadgeProps;
  tooltip?: TooltipProps;
  pointable?: boolean;
}

type IconPosition = 'prefix' | 'affix';

export interface ContentWithIconProps {
  content?: React.ReactNode;
  prefixIcon?: IconConfig | React.ReactNode;
  affixIcon?: IconConfig | React.ReactNode;
  onClick?: (e: React.SyntheticEvent) => void;
  style?: React.CSSProperties;
  className?: string;
}

const ContentWithIcon: React.FC<ContentWithIconProps> = ({
  content,
  prefixIcon,
  affixIcon,
  className,
  ...restProps
}) => {
  const { styles } = useStyles();
  return (
    <span className={`${styles.item} ${className}`} {...restProps}>
      {prefixIcon &&
        (isValidElement(prefixIcon) ? (
          <span className={styles.prefix}>{prefixIcon}</span>
        ) : (
          getIcon('prefix', prefixIcon)
        ))}

      <span className={styles.content}>{content}</span>
      {affixIcon &&
        (isValidElement(affixIcon) ? (
          <span className={styles.affix}>{affixIcon}</span>
        ) : (
          getIcon('affix', affixIcon)
        ))}
    </span>
  );
};

function getIcon(position: IconPosition, config: IconConfig) {
  const { component, badge, tooltip, pointable = false, ...restProps } = config;
  const { styles } = useStyles();
  return (
    config && (
      <Tooltip {...tooltip} overlayStyle={{ maxWidth: 350, ...tooltip?.overlayStyle }}>
        {badge ? (
          <Badge
            {...badge}
            className={classNames(`${styles[position]}`, {
              [styles.pointable]: tooltip || pointable,
            })}
          >
            <Icon component={component} {...restProps} />
          </Badge>
        ) : (
          <span
            className={classNames(`${styles[position]}`, {
              [styles.pointable]: tooltip || pointable,
            })}
          >
            <Icon component={component} {...restProps} />
          </span>
        )}
      </Tooltip>
    )
  );
}

export default ContentWithIcon;
