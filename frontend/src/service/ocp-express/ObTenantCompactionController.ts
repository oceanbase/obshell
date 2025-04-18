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

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/clearCompactionError */
export async function clearCompactionError(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/clearCompactionError`, {
    method: 'POST',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/compaction */
export async function getTenantCompaction(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_TenantCompaction_>(`/api/v1/ob/tenants/${param0}/compaction`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/topCompactions */
export async function topCompactions(
  params: {
    // query
    top?: number;
    times?: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_TenantCompactionHistory_>(
    '/api/v1/ob/tenants/topCompactions',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/triggerCompaction */
export async function triggerTenantCompaction(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/triggerCompaction`, {
    method: 'POST',
    params: { ...params },
    ...(options || {}),
  });
}
