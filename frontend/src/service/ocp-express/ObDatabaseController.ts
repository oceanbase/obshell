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

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/databases */
export async function createDatabase(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.CreateDatabaseParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_Database_>(`/api/v1/ob/tenants/${param0}/databases`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 DELETE /api/v1/ob/tenants/${param0}/databases/${param1} */
export async function deleteDatabase(
  params: {
    // path
    tenantId?: number;
    dbName?: string;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, dbName: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/databases/${param1}`, {
    method: 'DELETE',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/databases */
export async function listDatabases(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.IterableResponse_Database_>(`/api/v1/ob/tenants/${param0}/databases`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 PUT /api/v1/ob/tenants/${param0}/databases/${param1} */
export async function modifyDatabase(
  params: {
    // path
    tenantId?: number;
    dbName?: string;
  },
  body?: API.ModifyDatabaseParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, dbName: param1 } = params;
  return request<API.SuccessResponse_Database_>(
    `/api/v1/ob/tenants/${param0}/databases/${param1}`,
    {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}
