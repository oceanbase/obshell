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
 * 根据备份文件的解析和计算结果，分类获取备份恢复相关的时间点数据
 * */
export function getBackupTimeData(recoverableItems: API.ObRecoverableSectionItem[]) {
  // 数据备份时间点
  const dataList = recoverableItems.filter(
    item =>
      // 数据备份: 全量
      item.obRecoverableSectionItemType === 'FULL_DATA_BACKUP' ||
      // 数据备份: 增量
      item.obRecoverableSectionItemType === 'INCREMENTAL_DATA_BACKUP'
  );
  // 日志备份时间区间
  const logList = recoverableItems.filter(
    item => item.obRecoverableSectionItemType === 'LOG_BACKUP'
  );
  // 可恢复时间区间: 对数据备份时间点和日志备份时间区间做了聚合和计算，真实可恢复
  const recoverableList = recoverableItems.filter(
    item => item.obRecoverableSectionItemType === 'RECOVERABLE'
  );
  return {
    dataList,
    logList,
    recoverableList,
  };
}
