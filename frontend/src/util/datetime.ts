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

import type { Moment } from 'moment';
import moment from 'moment';
import { range, toNumber } from 'lodash';
import { formatTime as formatTimeFromOBUtil, isNullValue } from '@oceanbase/util';
import { DATE_TIME_FORMAT_DISPLAY } from '@/constant/datetime';

export function formatTime(
  value: number | string | undefined,
  format: string = DATE_TIME_FORMAT_DISPLAY
) {
  return isNullValue(value) ? '-' : formatTimeFromOBUtil(value, format);
}

/**
 * 日期时间格式化: 带微秒值
 * 其中 value 是符合 RFC3339 规格格式的日期时间字符串
 * 示例: 2021-04-06T21:42:11.966709+08:00、2021-04-08T14:28:03.80652+08:00、2021-04-09T00:00:00+08:00
 * TODO: 加单测
 * */
export function formatTimeWithMicroseconds(
  value: string | undefined,
  format: string = DATE_TIME_FORMAT_DISPLAY
): string {
  // 微秒值的范围: 1 ~ 999999，长度为 1 ~ 6 位；支持 Z 零时区
  const regex = /^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}).(\d{1,6})(Z|(([+-])(\d{2}:\d{2})))$/;
  const match = regex.exec(value || '');
  // 微秒值为 0 时，后端不会返回微秒信息，此时匹配值为 null，需要使用空字符串兜底，否则 padEnd 后的字符串展示结果为 undefined
  const microseconds = (match && match[3]) || '';

  // 检查传入的格式是否已经包含微秒部分
  if (format && format.includes('S')) {
    // 如果格式包含微秒，我们需要精确控制时间格式
    // 对于 HH:mm:ss.SSSSSS 格式，我们分别格式化时、分、秒，然后添加微秒
    if (format === 'HH:mm:ss.SSSSSS') {
      const time = moment(value);
      const hours = time.format('HH');
      const minutes = time.format('mm');
      const seconds = time.format('ss');
      return `${hours}:${minutes}:${seconds}.${microseconds?.padEnd(6, '0')}`;
    }

    // 对于其他包含微秒的格式，使用通用处理
    const formatWithoutMicroseconds = format.replace(/S+/g, '');
    const baseTime = moment(value).format(formatWithoutMicroseconds);
    return `${baseTime}.${microseconds?.padEnd(6, '0')}`;
  }

  // 如果格式不包含微秒，则手动添加微秒部分
  return `${formatTimeFromOBUtil(value, format)}.${microseconds?.padEnd(6, '0')}`;
}

/**
 * 根据日期时间字符串，获取对应的微秒值
 * 其中 value 是符合 RFC3339 规格格式的日期时间字符串
 * */
export function getMicroseconds(value?: string) {
  // 微秒值的范围: 1 ~ 999999，长度为 1 ~ 6 位；支持 Z 零时区
  const regex = /^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}).(\d{1,6})(Z|(([+-])(\d{2}:\d{2})))$/;
  const match = regex.exec(value || '');
  // 微秒值为 0 时，后端不会返回微秒信息，此时匹配值为 null，需要使用空字符串兜底，否则 padEnd 后的字符串展示结果为 undefined
  const microseconds = (match && match[3]) || '';
  // padEnd 向后补 0
  return toNumber(microseconds.padEnd(6, '0'));
}

/**
 * 返回自 Unix 纪元以来的微秒数
 * 其中 value 是符合 RFC3339 规格格式的日期时间字符串
 * */
export function valueOfMicroseconds(value?: string) {
  return moment(value).unix() * 1000 * 1000 + getMicroseconds(value);
}

/**
 * 计算两个时间的微秒差值
 * 其中 value1 和 value2 是符合 RFC3339 规格格式的日期时间字符串
 * */
export function diffWithMicroseconds(value1?: string, value2?: string) {
  return valueOfMicroseconds(value1) - valueOfMicroseconds(value2);
}

export function disabledDate(current: Moment) {
  return current.isBefore(moment().format('YYYY-MM-DD'));
}

export function disabledTime(date) {
  const hours = (date && date.hours()) || 0;
  const minutes = (date && date.minutes()) || 0;
  const current = moment();
  const currentHours = current.hours();
  const currentMinutes = current.minutes();
  const currentSeconds = current.seconds();
  if (date && date.format('YYYY-MM-DD') === current.format('YYYY-MM-DD')) {
    return {
      disabledHours: () => range(0, currentHours),
      disabledMinutes: () => (hours === currentHours ? range(0, currentMinutes) : []),
      disabledSeconds: () =>
        hours === currentHours && minutes === currentMinutes ? range(0, currentSeconds) : [],
    };
  }
  return {
    disabledHours: () => [],
    disabledMinutes: () => [],
    disabledSeconds: () => [],
  };
}
