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

import AobException from '@/component/AobException';
import { formatMessage } from '@/util/intl';
import { history } from '@umijs/max';
import React from 'react';

export default () => {
  return (
    <AobException
      title="404"
      desc={formatMessage({
        id: 'ocp-express.page.Error.404.SorryThePageYouVisited',
        defaultMessage: '抱歉，你访问的页面不存在',
      })}
      img="/assets/common/404.svg"
      backText={formatMessage({
        id: 'ocp-express.page.Error.404.ReturnToHomePage',
        defaultMessage: '返回首页',
      })}
      onBack={() => {
        history.push(`/`);
      }}
      style={{
        paddingTop: 50,
        height: '100%',
      }}
    />
  );
};
