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

import { getLocale } from '@umijs/max';

// 在应用挂载点后将 locale 信息以 script 标签的形式插入到报告中，以便下载的报告能以默认语言进行展示
export const insertLocaleScript = (html: string) => {
  return html.replace(
    '<div id="root"></div>',
    `<div id="root"></div>
    <script>
      window.__OCP_REPORT_LOCALE = "${getLocale() || 'zh-CN'}";
    </script>`
  );
};

// 下载文件和HTML
export const download = (content: string, fileName?: string) => {
  const blob = new Blob([content]);
  const blobUrl = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.download = fileName;
  a.href = blobUrl;
  a.click();
  a.remove();
  window.URL.revokeObjectURL(blobUrl);
};
