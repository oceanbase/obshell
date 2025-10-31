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

/* eslint-disable */
// 该文件由 OneAPI 自动生成，请勿手动修改！
import request from '@/util/request';

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/sessions/closeQuery */
export async function closeTenantQuery(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.CloseSessionParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/sessions/closeQuery`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/sessions/close */
export async function closeTenantSession(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.CloseSessionParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/sessions/close`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/sessions/stats */
export async function getSessionStats(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_SessionStats_>(`/api/v1/ob/tenants/${param0}/sessions/stats`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/sessions/${param1} */
export async function getTenantSession(
  params: {
    // path
    tenantId?: number;
    sessionId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, sessionId: param1 } = params;
  return request<API.SuccessResponse_TenantSession_>(
    `/api/v1/ob/tenants/${param0}/sessions/${param1}`,
    {
      method: 'GET',
      params: { ...params },
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/sessions */
export async function listTenantSessions(
  params: {
    // query
    pageable?: API.Pageable;
    dbUser?: string;
    dbName?: string;
    clientIp?: string;
    activeOnly?: boolean;
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, ...queryParams } = params;
  return request<API.PaginatedResponse_TenantSession_>(`/api/v1/ob/tenants/${param0}/sessions`, {
    method: 'GET',
    params: {
      ...queryParams,
      pageable: undefined,
      ...queryParams['pageable'],
    },
    ...(options || {}),
  });
}
