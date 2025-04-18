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

const localeData = moment.localeData();

export const FOREVER_TIME = '2099-12-31T00:00:00.000Z';

/* 年 */
export const YEAR_FORMAT = 'YYYY';

export const YEAR_FORMAT_DISPLAY = localeData.longDateFormat('year');

/* 月 */
export const MONTH_FORMAT = 'YYYY-MM';

export const MONTH_FORMAT_DISPLAY = localeData.longDateFormat('month');

/* 日期 */
export const DATE_FORMAT = 'YYYY-MM-DD';

export const DATE_FORMAT_DISPLAY = localeData.longDateFormat('date');

export const DATE_FORMAT_WITHOUT_YEAR_DISPLAY = localeData.longDateFormat('dateWithoutYear');

/* 日期 + 时间 */

// RFC3339 的日期时间格式
export const RFC3339_DATE_TIME_FORMAT = 'YYYY-MM-DDTHH:mm:ssZ';

// 日期时间格式
export const DATE_TIME_FORMAT = 'YYYY-MM-DD HH:mm:ss';

// 没有秒数据的日期时间格式
export const DATE_TIME_FORMAT_WITHOUT_SECOND = 'YYYY-MM-DD HH:mm';

export const DATE_TIME_FORMAT_DISPLAY = localeData.longDateFormat('datetime');

export const DATE_TIME_FORMAT_WITH_SSS_DISPLAY = localeData.longDateFormat('datetimeWithSSS');

export const DATE_TIME_FORMAT_WITHOUT_SECOND_DISPLAY =
  localeData.longDateFormat('datetimeWithoutSecond');

export const DATE_TIME_FORMAT_WITHOUT_YEAR_AND_SECOND_DISPLAY = localeData.longDateFormat(
  'datetimeWithoutYearAndSecond'
);

/* 时间 */

// RFC3339 的时间格式
export const RFC3339_TIME_FORMAT = 'HH:mm:ssZ';

// 时间格式
export const TIME_FORMAT = 'HH:mm:ss';

// 带毫秒的时间格式
export const TIME_FORMAT_WITH_SSS = 'HH:mm:ss.SSS';

// 不带秒信息的时间格式
export const TIME_FORMAT_WITHOUT_SECOND = 'HH:mm';
