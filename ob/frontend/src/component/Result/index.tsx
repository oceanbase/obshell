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

import PageCard from '@/component/PageCard';
import { formatMessage } from '@/util/intl';
import { Button, Result, Space } from '@oceanbase/design';
import type { ResultProps } from '@oceanbase/design/es/result';
import { PageContainer } from '@oceanbase/ui';
import { history } from '@umijs/max';
import React from 'react';
import styles from './index.less';

export interface SuccessProps extends ResultProps {
  iconUrl?: string;
  taskId?: number;
  children?: React.ReactNode;
  style?: React.CSSProperties;
  className?: string;
}

const Success: React.FC<SuccessProps> = ({
  iconUrl = '/assets/icon/success.svg',
  taskId,
  children,
  className,
  style,
  ...restProps
}) => {
  return (
    <PageContainer className={`${styles.container} ${className}`} style={style}>
      <PageCard
        style={{
          height: 'calc(100vh - 72px)',
        }}
      >
        <Result
          icon={<img src={iconUrl} alt="" className={styles.icon} />}
          extra={
            <Space>
              <Button
                onClick={() => {
                  history.goBack();
                }}
              >
                {formatMessage({
                  id: 'ocp-express.component.Result.Return',
                  defaultMessage: '返回',
                })}
              </Button>
              {taskId && (
                <Button
                  type="primary"
                  onClick={() => {
                    history.push(`/task/${taskId}`);
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.component.Result.ViewTaskDetails',
                    defaultMessage: '查看任务详情',
                  })}
                </Button>
              )}
            </Space>
          }
          {...restProps}
        >
          {children}
        </Result>
      </PageCard>
    </PageContainer>
  );
};

export default Success;
