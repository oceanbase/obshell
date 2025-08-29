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

import moment from 'moment';
import { getLocale } from '@/util/intl';

// 动态获取当前语言的日期格式
const getLocalizedFormat = () => {
  const currentLocale = getLocale();

  // 确保 moment.js 的 locale 与当前语言保持同步
  const momentLocaleMap: Record<string, string> = {
    'en-US': 'en',
    'zh-CN': 'zh-cn',
  };
  const momentLocale = momentLocaleMap[currentLocale] || 'en';
  if (moment.locale() !== momentLocale) {
    moment.locale(momentLocale);
  }

  if (currentLocale === 'zh-CN') {
    return {
      year: 'YYYY年',
      month: 'YYYY年M月',
      date: 'YYYY年M月D日',
      dateWithoutYear: 'M月D日',
      datetime: 'YYYY年M月D日 HH:mm:ss',
      datetimeWithSSS: 'YYYY年M月D日 HH:mm:ss.SSS',
      datetimeWithoutSecond: 'YYYY年M月D日 HH:mm',
      datetimeWithoutYearAndSecond: 'M月D日 HH:mm',
    };
  } else {
    // 英文格式
    const localeData = moment.localeData();
    return {
      year: 'YYYY',
      month: 'MMM, YYYY',
      date: localeData.longDateFormat('LL') || 'MMMM D, YYYY',
      dateWithoutYear: 'MMM D',
      datetime: 'MMM D, YYYY, HH:mm:ss',
      datetimeWithSSS: 'MMM D, YYYY, HH:mm:ss.SSS',
      datetimeWithoutSecond: 'MMM D, YYYY, HH:mm',
      datetimeWithoutYearAndSecond: 'MMM D, HH:mm',
    };
  }
};

export const FOREVER_TIME = '2099-12-31T00:00:00.000Z';

/* 年 */
export const YEAR_FORMAT = 'YYYY';

// 导出动态格式获取函数
export const getYearFormatDisplay = () => getLocalizedFormat().year;
export const getMonthFormatDisplay = () => getLocalizedFormat().month;
export const getDateFormatDisplay = () => getLocalizedFormat().date;
export const getDateFormatWithoutYearDisplay = () => getLocalizedFormat().dateWithoutYear;

// 为了向后兼容，保留原来的导出，但使用当前语言的格式初始化
export const YEAR_FORMAT_DISPLAY = getLocalizedFormat().year;

/* 月 */
export const MONTH_FORMAT = 'YYYY-MM';

export const MONTH_FORMAT_DISPLAY = getLocalizedFormat().month;

/* 日期 */
export const DATE_FORMAT = 'YYYY-MM-DD';

export const DATE_FORMAT_DISPLAY = getLocalizedFormat().date;

export const DATE_FORMAT_WITHOUT_YEAR_DISPLAY = getLocalizedFormat().dateWithoutYear;

/* 日期 + 时间 */

// RFC3339 的日期时间格式
export const RFC3339_DATE_TIME_FORMAT = 'YYYY-MM-DDTHH:mm:ssZ';

// 日期时间格式
export const DATE_TIME_FORMAT = 'YYYY-MM-DD HH:mm:ss';

// 没有秒数据的日期时间格式
export const DATE_TIME_FORMAT_WITHOUT_SECOND = 'YYYY-MM-DD HH:mm';

export const DATE_TIME_FORMAT_DISPLAY = getLocalizedFormat().datetime;

export const DATE_TIME_FORMAT_WITH_SSS_DISPLAY = getLocalizedFormat().datetimeWithSSS;

export const DATE_TIME_FORMAT_WITHOUT_SECOND_DISPLAY = getLocalizedFormat().datetimeWithoutSecond;

export const DATE_TIME_FORMAT_WITHOUT_YEAR_AND_SECOND_DISPLAY =
  getLocalizedFormat().datetimeWithoutYearAndSecond;

/* 时间 */

// RFC3339 的时间格式
export const RFC3339_TIME_FORMAT = 'HH:mm:ssZ';

// 时间格式
export const TIME_FORMAT = 'HH:mm:ss';

// 带毫秒的时间格式
export const TIME_FORMAT_WITH_SSS = 'HH:mm:ss.SSS';

// 不带秒信息的时间格式
export const TIME_FORMAT_WITHOUT_SECOND = 'HH:mm';
