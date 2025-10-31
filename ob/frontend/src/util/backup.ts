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

/**
 * 计算时间交集，保持微秒信息
 * @param originalTime 原始时间字符串（包含微秒）
 * @param intersectionTime 交集时间（moment对象）
 * @param isStart 是否为开始时间
 * @param dayStart 选中日期的开始时间（用于判断是否为边界时间）
 * @param dayEnd 选中日期的结束时间（用于判断是否为边界时间）
 * @returns 调整后的时间字符串
 */
function calculateIntersectionTime(
  originalTime: string,
  intersectionTime: moment.Moment,
  isStart: boolean,
  dayStart?: moment.Moment,
  dayEnd?: moment.Moment
): string {
  if (!originalTime) return intersectionTime.format();

  // 如果交集时间就是原始时间，直接返回原始时间
  const originalMoment = moment(originalTime);
  if (originalMoment.isSame(intersectionTime, 'second')) {
    return originalTime;
  }

  // 判断是否为边界时间（需要将微秒设为000000）
  const isBoundaryTime = dayStart && dayEnd && (
    (isStart && intersectionTime.isSame(dayStart, 'second')) ||
    (!isStart && intersectionTime.isSame(dayEnd, 'second'))
  );

  // 提取微秒和时区信息
  const timeMatch = originalTime.match(/\.(\d{1,6})(Z|[+-]\d{2}:\d{2})$/);
  const microseconds = isBoundaryTime ? '000000' : (timeMatch?.[1]?.padEnd(6, '0') || '000000');
  const timezone = timeMatch?.[2] || '+08:00';

  // 构建新的时间字符串
  return intersectionTime.format('YYYY-MM-DDTHH:mm:ss') + '.' + microseconds + timezone;
}

/**
 * 根据传入的时间日期和时间范围区间，匹配出符合传入日期的当日区间
 * @param timeRanges 时间范围数组，每个元素包含 start_time 和 end_time
 * @param targetDate 目标日期，可以是 moment 对象、Date 对象或日期字符串
 * @returns 匹配的时间范围数组
 */
export function getBackupTimeData(
  timeRanges: API.RestoreWindow[],
  targetDate?: moment.Moment | Date | string
) {
  // 如果没有传入目标日期，返回所有时间范围
  if (!targetDate) {
    return {
      dataList: timeRanges,
      logList: timeRanges,
      recoverableList: timeRanges,
    };
  }

  // 将目标日期转换为 moment 对象
  const targetMoment = moment.isMoment(targetDate) ? targetDate : moment(targetDate);

  // 获取目标日期的开始和结束时间（当天 00:00:00 到 23:59:59）
  const dayStart = targetMoment.clone().startOf('day');
  const dayEnd = targetMoment.clone().endOf('day');

  // 计算与目标日期的实际交集
  const matchedRanges = timeRanges
    .map(item => {
      if (!item.start_time || !item.end_time) {
        return null;
      }

      const rangeStart = moment(item.start_time);
      const rangeEnd = moment(item.end_time);

      // 检查时间范围是否与目标日期有交集
      if (rangeStart.isAfter(dayEnd) || rangeEnd.isBefore(dayStart)) {
        return null; // 无交集
      }

      // 计算交集时间范围
      const intersectionStart = rangeStart.isBefore(dayStart) ? dayStart : rangeStart;
      const intersectionEnd = rangeEnd.isAfter(dayEnd) ? dayEnd : rangeEnd;

      // 返回交集时间范围，保持原有属性
      return {
        ...item,
        // 使用交集时间作为显示时间，确保显示的是选中日期内的实际时间范围
        // 保持微秒信息，传入日期边界信息用于判断边界时间
        start_time: calculateIntersectionTime(item.start_time, intersectionStart, true, dayStart, dayEnd),
        end_time: calculateIntersectionTime(item.end_time, intersectionEnd, false, dayStart, dayEnd),
        // 添加交集信息
        intersection_start: intersectionStart.format(),
        intersection_end: intersectionEnd.format(),
        original_start: item.start_time,
        original_end: item.end_time,
      };
    })
    .filter((item): item is NonNullable<typeof item> => item !== null); // 类型安全的过滤

  // 按时间排序
  const sortedRanges = matchedRanges.sort((a, b) => {
    if (!a.start_time || !b.start_time) return 0;
    return moment(a.start_time).valueOf() - moment(b.start_time).valueOf();
  });

  return {
    dataList: sortedRanges,
    logList: sortedRanges,
    recoverableList: sortedRanges,
  };
}

/**
 * 根据传入的时间日期和时间范围区间，获取指定日期的备份时间数据
 * @param timeRanges 时间范围数组
 * @param targetDate 目标日期
 * @param dateFormat 日期格式，默认为 'YYYY-MM-DD'
 * @returns 匹配的时间范围数组
 */
export function getBackupTimeDataByDate(
  timeRanges: API.RestoreWindow[],
  targetDate: string,
  dateFormat: string = 'YYYY-MM-DD'
) {
  const targetMoment = moment(targetDate, dateFormat);

  if (!targetMoment.isValid()) {
    console.warn('Invalid target date format:', targetDate);
    return {
      recoverableList: [],
    };
  }

  return getBackupTimeData(timeRanges, targetMoment);
}

/**
 * 根据传入的时间日期和时间范围区间，获取指定日期范围内的备份时间数据
 * @param timeRanges 时间范围数组
 * @param startDate 开始日期
 * @param endDate 结束日期
 * @param dateFormat 日期格式，默认为 'YYYY-MM-DD'
 * @returns 匹配的时间范围数组
 */
export function getBackupTimeDataByDateRange(
  timeRanges: API.RestoreWindow[],
  startDate: string,
  endDate: string,
  dateFormat: string = 'YYYY-MM-DD'
) {
  const startMoment = moment(startDate, dateFormat);
  const endMoment = moment(endDate, dateFormat);

  if (!startMoment.isValid() || !endMoment.isValid()) {
    console.warn('Invalid date format:', { startDate, endDate });
    return {
      recoverableList: [],
    };
  }

  // 确保开始日期不晚于结束日期
  let finalStartMoment = startMoment;
  let finalEndMoment = endMoment;
  if (startMoment.isAfter(endMoment)) {
    finalStartMoment = endMoment;
    finalEndMoment = startMoment;
  }

  // 过滤出在指定日期范围内的时间范围
  const matchedRanges = timeRanges.filter(item => {
    if (!item.start_time || !item.end_time) {
      return false;
    }

    const rangeStart = moment(item.start_time);
    const rangeEnd = moment(item.end_time);

    // 检查时间范围是否与指定日期范围有交集
    return rangeStart.isSameOrBefore(finalEndMoment) && rangeEnd.isSameOrAfter(finalStartMoment);
  });

  // 按时间排序
  const sortedRanges = matchedRanges.sort((a, b) => {
    if (!a.start_time || !b.start_time) return 0;
    return moment(a.start_time).valueOf() - moment(b.start_time).valueOf();
  });

  return {
    recoverableList: sortedRanges,
  };
}
