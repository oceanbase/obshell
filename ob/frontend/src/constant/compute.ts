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

export const AUTHORIZE_TYPE_LIST = [
  {
    label: formatMessage({
      id: 'OBShell.src.constant.compute.PasswordVerification',
      defaultMessage: '密码验证',
    }),
    value: 'PASSWORD',
  },
];

export const USET_TYPE_LIST = [
  {
    label: formatMessage({
      id: 'OBShell.src.constant.compute.RootUser',
      defaultMessage: 'root 用户',
    }),
    value: 'root',
  },

  {
    label: formatMessage({
      id: 'OBShell.src.constant.compute.OrdinaryUser',
      defaultMessage: '普通用户',
    }),
    value: 'normal',
  },
];
