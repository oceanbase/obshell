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

/**
 * 以下逻辑原本应该在 src/global.ts 中进行初始化，但由于 getLocale 和 setLocale 的调用 (即 locale 插件的初始化时机)
 * 要晚于 dva 或者 useModel 等状态管理插件的初始化时机，导致执行的先后顺序不符合预期，无法获取 moment 的多语言设置，因此
 * 放到 constant/init.ts 文件中进行初始化，并在 constant/index.ts 中引入，以保证使用前已设置 moment 的多语言配置
 */
import { setLocale, getLocale } from 'umi';
import moment from 'moment';
import tracert from '@/util/tracert';

// 引入并配置 dayjs 插件，解决 DateRanger 组件中 weekday 方法缺失的问题
import dayjs from 'dayjs';
import customParseFormat from 'dayjs/plugin/customParseFormat';
import advancedFormat from 'dayjs/plugin/advancedFormat';
import weekday from 'dayjs/plugin/weekday';
import localeData from 'dayjs/plugin/localeData';
import weekOfYear from 'dayjs/plugin/weekOfYear';
import weekYear from 'dayjs/plugin/weekYear';
import isoWeek from 'dayjs/plugin/isoWeek';
import isoWeeksInYear from 'dayjs/plugin/isoWeeksInYear';
import 'dayjs/locale/zh-cn';

// 扩展 dayjs 插件
dayjs.extend(customParseFormat);
dayjs.extend(advancedFormat);
dayjs.extend(weekday);
dayjs.extend(localeData);
dayjs.extend(weekOfYear);
dayjs.extend(weekYear);
dayjs.extend(isoWeek);
dayjs.extend(isoWeeksInYear);

const validLocalMap = {
  // 英文
  'en-US': 'en-US',
  // 简体中文
  'zh-CN': 'zh-CN',
  // 兼容小写格式
  'en-us': 'en-US',
  'zh-cn': 'zh-CN',
  // 兼容单语言格式
  en: 'en-US',
  zh: 'zh-CN',
  // 兼容下划线格式
  en_US: 'en-US',
  zh_CN: 'zh-CN',
};

const locale = validLocalMap[getLocale()] || 'zh-CN';

// 只接受中划线格式
setLocale(locale);

// 自定义各个 locale 下的日期时间格式
moment.updateLocale('en', {
  longDateFormat: {
    datetime: 'MMM D, YYYY, HH:mm:ss',
    datetimeWithSSS: 'MMM D, YYYY, HH:mm:ss.SSS',
    datetimeWithoutSecond: 'MMM D, YYYY, HH:mm',
    datetimeWithoutYearAndSecond: 'MMM D, HH:mm',
    date: 'MMM D, YYYY',
    dateWithoutYear: 'MMM D',
    year: 'YYYY',
    month: 'MMM, YYYY',
  },
});

moment.updateLocale('zh-cn', {
  longDateFormat: {
    datetime: 'YYYY年M月D日 HH:mm:ss',
    datetimeWithSSS: 'YYYY年M月D日 HH:mm:ss.SSS',
    datetimeWithoutSecond: 'YYYY年M月D日 HH:mm',
    datetimeWithoutYearAndSecond: 'M月D日 HH:mm',
    date: 'YYYY年M月D日',
    dateWithoutYear: 'M月D日',
    year: 'YYYY年',
    month: 'YYYY年M月',
  },
});

// 将 Umi 的 locale 格式映射为 moment 的 locale 格式
const momentLocaleMap = {
  'en-US': 'en',
  'zh-CN': 'zh-cn',
};

moment.locale(momentLocaleMap[locale] || 'en');

// 配置 dayjs 的 locale，与 moment 保持一致
dayjs.locale(momentLocaleMap[locale] || 'en');

tracert.set({
  // 埋点 a 位
  spmAPos: 'a3647',
  bizType: 'common',
  // 开启后会监听路由变化，并触发页面访问埋点上报
  ifRouterNeedPv: true,
  // 开启曝光上报
  autoExpo: true,
  autoLogPv: true,
});
