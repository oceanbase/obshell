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
 * 中文和英文、数字之间使用空格分隔，并去除中文之间的空格，以保证文本展示符合内容规范
 * @param{string} text 原始文本
 * @return{string} 格式化后的文本
 */
export function formatTextWithSpace(text?: string) {
  /* 中文和英文、数字之间使用空格分隔 */
  // 中文在前，英文和数字在后
  const regex1 = /([\u4e00-\u9fa5]+)([a-zA-Z0-9]+)/g;
  // 中文在后，英文和数字在前
  const regex2 = /([a-zA-Z0-9]+)([\u4e00-\u9fa5]+)/g;
  let str = text;
  str = str?.replace(regex1, '$1 $2');
  str = str?.replace(regex2, '$1 $2');
  /* 去除中文之间的空格 (即空白字符，包括空格、制表符、换页符等)，需要使用断言 */
  const regex3 = /([\u4e00-\u9fa5]+)\s+(?=[\u4e00-\u9fa5]+)/g;
  str = str?.replace(regex3, '$1');
  return str;
}
