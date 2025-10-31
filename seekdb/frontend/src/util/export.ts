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
