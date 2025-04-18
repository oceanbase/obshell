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

import { createIntl } from 'react-intl';
// import { getLocale } from 'umi';
import en_US from '@/locale/en-US';
import zh_CN from '@/locale/zh-CN';

const messages = {
  'en-US': en_US,
  'zh-CN': zh_CN,
};

export const getLocale = () => {
  const lang =
    typeof localStorage !== 'undefined'
      ? window.localStorage.getItem('umi_locale')
      : '';
  return lang || 'zh-CN';
};
export const locale = getLocale();

// const locale = getLocale();

const intl = createIntl({
  locale,
  messages: messages[locale],
});

export const { formatMessage } = intl;
export default intl;
