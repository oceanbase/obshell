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
import classNames from 'classnames';
import styles from './index.less';

export interface BatchOperationBarProps {
  className?: string;
  style?: React.CSSProperties;
  size?: 'small' | 'default' | 'large';
  description?: React.ReactNode;
  selectedCount?: number;
  visible?: boolean;
  actions?: React.ReactNode[];
  onCancel?: () => void;
  selectedText?: string;
  cancelText?: string;
}

const BatchOperationBar: React.FC<BatchOperationBarProps> = ({
  description,
  className,
  style,
  actions = [],
  selectedCount = 0,
  visible,
  onCancel,
  size = 'default',
}) => {
  const realVisible = visible === undefined ? selectedCount > 0 : visible;
  return (
    <div
      style={{
        display: realVisible ? 'flex' : 'none',
        ...style,
      }}
      className={classNames(styles.container, styles[size], className)}
    >
      <div className={styles.left}>
        <div className={styles.title}>
          {formatMessage(
            {
              id: 'ocp-express.component.BatchOperationBar.SelectedcountSelected',
              defaultMessage: '已选 {selectedCount} 项',
            },
            { selectedCount }
          )}
          {onCancel && (
            <span className={styles.cancel} onClick={onCancel}>
              {formatMessage({
                id: 'ocp-express.component.BatchOperationBar.Deselect',
                defaultMessage: '取消选择',
              })}
            </span>
          )}
        </div>
        {description && <div className={styles.subtitle}>{description}</div>}
      </div>
      <div className={styles.right}>
        {actions.map((action, index) => {
          return (
            <div key={index} className={styles.actionItem}>
              {action}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default BatchOperationBar;
