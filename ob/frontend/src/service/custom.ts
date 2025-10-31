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

export async function getSqlTraceTemplate() {
  return request<Blob>('/templates/sqlTraceReportTemplate.html', {
    method: 'GET',
    responseType: 'blob',
  });
}

export async function getInspectionTemplate() {
  return request<Blob>('/templates/inspectionReportTemplate.html', {
    method: 'GET',
    responseType: 'blob',
  });
}

// 由于OneApi生成的service函数，FormData中不含参数
// 接管OBProxy 导入负载均衡
export async function parseLbInstanceConfigFile(
  params: {
    // path
    id?: number;
  },
  files?: File[],
  options?: Record<string, any>
) {
  const { id: param0 } = params;
  const formData = new FormData();
  if (files) {
    formData.append('file', files || '');
  }
  return request<API.ComOceanbaseOcpCoreResponseIterableResponse_LBInstance_>(
    `/api/v2/obproxy/clusters/${param0}/lb/instance/parseFile`,
    {
      method: 'POST',
      params: { ...params },
      data: formData,
      ...(options || {}),
    }
  );
}

// 由于OneApi生成的service函数，FormData中不含参数
// 凭据管理导入预检查
export async function importCredentialPreCheck(
  params: {
    // query
    secret?: string;
  },
  files?: File[],
  options?: Record<string, any>
) {
  const formData = new FormData();
  if (files) {
    formData.append('file', files || '');
  }
  if (params.secret) {
    formData.append('secret', params.secret || '');
  }
  return request<API.SuccessResponse_CredentialValidateResult_>(
    '/api/v2/profiles/me/credentials/importPreCheck',
    {
      method: 'POST',
      params: {
        // ...params,
      },
      data: formData,
      ...(options || {}),
    }
  );
}

// 凭据管理导入
export async function importCredential(
  params: {
    // query
    secret?: string;
    numbers?: number[];
  },
  files?: File[],
  options?: Record<string, any>
) {
  const formData = new FormData();
  if (files) {
    formData.append('file', files || '');
  }
  if (params.secret) {
    formData.append('secret', params.secret || '');
  }
  if (params.numbers) {
    formData.append('numbers', params.numbers || []);
  }
  return request<API.SuccessResponse_CredentialValidateResult_>(
    '/api/v2/profiles/me/credentials/import',
    {
      method: 'POST',
      params: {
        // ...params,
      },
      data: formData,
      ...(options || {}),
    }
  );
}

// 获取用户 seesion
export async function getSession(): Promise<any> {
  return request('/api/v2/session', {
    method: 'GET',
  });
}
// obproxy 下线
export async function offlineObproxies(
  params: {
    // path
    obproxyIds?: number[];
    id?: number[];
  },
  options?: Record<string, any>
) {
  const { obproxyIds: param0, id: param1 } = params;

  return request<API.ComOceanbaseOcpCoreResponseNoDataResponse>(
    `/api/v2/obproxy/clusters/${param1}/obproxies/${param0}/offline`,
    {
      method: 'POST',
      params: { ...params },
      ...(options || {}),
    }
  );
}
// obproxy 上线
export async function onlineObproxies(
  params: {
    // path
    obproxyIds?: number[];
    id?: number[];
  },
  options?: Record<string, any>
) {
  const { obproxyIds: param0, id: param1 } = params;
  return request<API.ComOceanbaseOcpCoreResponseNoDataResponse>(
    `/api/v2/obproxy/clusters/${param1}/obproxies/${param0}/online`,
    {
      method: 'POST',
      params: { ...params },
      ...(options || {}),
    }
  );
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
