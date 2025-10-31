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

import request from '@/util/request';

/** 登录接口由于并非是后端显式定义的，因此 OneAPI 无法解析生成，需要手动定义 service 函数 */
export async function login(params: {
  // 用户名
  username: string;
  // 密码
  password: string;
}): Promise<any> {
  return request('/api/v1/login', {
    method: 'POST',
    requestType: 'form',
    data: params,
  });
}

/** 退出登录接口由于并非是后端显式定义的，因此 OneAPI 无法解析生成，需要手动定义 service 函数 */
export async function logout(): Promise<any> {
  return request('/api/v1/logout', {
    method: 'POST',
  });
}

// 获取用户 seesion
export async function getSession(): Promise<any> {
  return request('/api/v2/session', {
    method: 'GET',
  });
}

/** 遥测上报API */
export async function telemetryReport(params: any): Promise<any> {
  const reportUrl =
    process.env.NODE_ENV === 'production'
      ? 'https://openwebapi.oceanbase.com'
      : 'http://openwebapi.test.alipay.net';

  return fetch(`${reportUrl}/api/web/oceanbase/report`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(params),
  }).then(response => response.json());
}
