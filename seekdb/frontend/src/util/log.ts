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

import FileSaver from 'file-saver';

/**
 * 下载日志
 * @param log      日志字符串
 * @param fileName 下载后生成的文件名
 * */
export function downloadLog(log?: string, fileName?: string) {
  const blob = new Blob([log || ''], { type: 'text/plain;charset=utf-8' });
  FileSaver.saveAs(blob, fileName || 'log.log');
}
