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
import React, { createElement } from 'react';
import { Button } from '@oceanbase/design';
import useStyles from './index.style';

interface AobExceptionProps {
  title?: React.ReactNode;
  desc?: React.ReactNode;
  img?: string;
  actions?: React.ReactNode;
  style?: React.CSSProperties;
  className?: string;
  linkElement?: string;
  backText?: string;
  redirect?: string;
  onBack?: () => void;
}

const AobException: React.FC<AobExceptionProps> = ({
  className,
  backText = formatMessage({
    id: 'ocp-express.component.AobException.ReturnToHomePage',
    defaultMessage: '返回首页',
  }),
  title,
  desc,
  img,
  linkElement = 'a',
  actions,
  redirect = '/',
  onBack,
  ...rest
}) => {
  const { styles } = useStyles();
  return (
    <div className={`${styles.container} ${className}`} {...rest}>
      <div className={styles.imgWrapper}>
        <div className={styles.img} style={{ backgroundImage: `url(${img})` }} />
      </div>
      <div className={styles.content}>
        <h1>{title}</h1>
        <div className={styles.desc}>{desc}</div>
        <div className={styles.actions}>
          {actions ||
            (onBack ? (
              <Button
                data-aspm-click="ca48180.da30493"
                data-aspm-desc="异常页-返回首页"
                data-aspm-param={``}
                data-aspm-expo
                type="primary"
                onClick={onBack}
              >
                {backText}
              </Button>
            ) : (
              createElement(
                linkElement,
                {
                  to: redirect,
                  href: redirect,
                },
                <Button
                  data-aspm-click="ca48180.da30493"
                  data-aspm-desc="异常页-返回首页"
                  data-aspm-param={``}
                  data-aspm-expo
                  type="primary"
                >
                  {backText}
                </Button>
              )
            ))}
        </div>
      </div>
    </div>
  );
};

export default AobException;
