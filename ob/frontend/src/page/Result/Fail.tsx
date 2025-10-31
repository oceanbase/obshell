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
import { history } from 'umi';
import React from 'react';
import { Button, Result } from '@oceanbase/design';
import Icon from '@oceanbase/icons';
import { PageContainer } from '@oceanbase/ui';
import PageCard from '@/component/PageCard';
import { ReactComponent as FailSvg } from '@/asset/fail.svg';

export interface TaskFailProps {
  style?: React.CSSProperties;
  className?: string;
}

const TaskFail: React.FC<TaskFailProps> = () => (
  <PageContainer>
    <PageCard
      style={{
        height: 'calc(100vh - 72px)',
      }}
    >
      <Result
        icon={<Icon component={FailSvg} />}
        title={formatMessage({
          id: 'ocp-express.page.Result.Fail.ClusterCreationFailed',
          defaultMessage: '创建集群失败',
        })}
        subTitle={formatMessage({
          id: 'ocp-express.page.Result.Fail.YouCanTryToRe',
          defaultMessage: '你可以尝试重新创建或返回集群概览',
        })}
        extra={
          <div>
            <Button
              type="primary"
              onClick={() => {
                history.push('/cluster/new');
              }}
              style={{ marginRight: 8 }}
            >
              {formatMessage({
                id: 'ocp-express.page.Result.Fail.ReCreateACluster',
                defaultMessage: '重新创建集群',
              })}
            </Button>
            <Button
              onClick={() => {
                history.push('/cluster');
              }}
            >
              {formatMessage({
                id: 'ocp-express.page.Result.Fail.ReturnToClusterOverview',
                defaultMessage: '返回集群概览',
              })}
            </Button>
          </div>
        }
      />
    </PageCard>
  </PageContainer>
);

export default TaskFail;
