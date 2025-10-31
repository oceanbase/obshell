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
import type { EmptyProps } from '@/component/Empty';
import Empty from '@/component/Empty';

type NoAuthProps = EmptyProps;

export default ({
  image = '/assets/common/no_auth.svg',
  title = formatMessage({
    id: 'ocp-express.component.NoAuth.NoPermissionToView',
    defaultMessage: '暂无权限查看',
  }),
  description = formatMessage({
    id: 'ocp-express.component.NoAuth.PleaseContactTheAdministratorTo',
    defaultMessage: '请联系管理员开通权限',
  }),
  ...restProps
}: NoAuthProps) => <Empty image={image} title={title} description={description} {...restProps} />;
