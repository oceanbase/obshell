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
import React, { useRef } from 'react';
import { PageContainer } from '@oceanbase/ui';
import useDocumentTitle from '@/hook/useDocumentTitle';
import useReload from '@/hook/useReload';
import ContentWithReload from '@/component/ContentWithReload';
import Query from './Query';
import type { QueryLogProps } from './Query';
interface Props {
  location: QueryLogProps['location'];
}

export default (props: Props) => {
  useDocumentTitle(
    formatMessage({
      id: 'ocp-express.page.Log.LogQuery',
      defaultMessage: '日志查询',
    })
  );

  const [reloading, reload] = useReload(false);

  // 为了实现页面滚动到底部时，日志查询页能自动加载更多，因此需要获取上层父容器的 ref 用于滚动监听
  const containerRef = useRef<HTMLDivElement>(null);

  return (
    <div ref={containerRef} style={{ height: 'calc(100% - 48px)' }}>
      <PageContainer
        loading={reloading}
        ghost={true}
        header={{
          title: (
            <ContentWithReload
              content={formatMessage({
                id: 'ocp-express.page.Log.LogQuery',
                defaultMessage: '日志查询',
              })}
              spin={reloading}
              onClick={reload}
            />
          ),
        }}
      >
        <Query containerRef={containerRef} {...props} />
      </PageContainer>
    </div>
  );
};
