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

const tracert = {
  // 设置 Tracert 属性
  set(params: Record<string, any>) {
    if (window.Tracert) {
      window.Tracert.call('set', params);
    }
  },
  // 将对象解析为符合埋点参数格式的字符串
  stringify(params: Record<string, any>) {
    if (window.Tracert) {
      return window.Tracert.call('stringify', params);
    }
    return '';
  },
};

export default tracert;
