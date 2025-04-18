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

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/addReplica */
export async function addReplica(
  params: {
    // path
    tenantId?: number;
  },
  body?: Array<API.AddReplicaParam>,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_TaskInstance_>(`/api/v1/ob/tenants/${param0}/addReplica`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/changePassword */
export async function changePasswordForTenant(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.TenantChangePasswordParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/changePassword`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/checkTenantPassword */
export async function checkTenantPassword(
  body?: API.CreatePasswordInVaultParam,
  options?: { [key: string]: any },
) {
  return request<API.SuccessResponse_CheckTenantPasswordResult_>(
    '/api/v1/ob/tenants/checkTenantPassword',
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/createOrReplacePassword */
export async function createOrReplacePassword(
  body?: API.CreatePasswordInVaultParam,
  options?: { [key: string]: any },
) {
  return request<API.NoDataResponse>('/api/v1/ob/tenants/createOrReplacePassword', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants */
export async function createTenant(body?: API.CreateTenantParam, options?: { [key: string]: any }) {
  return request<API.SuccessResponse_TaskInstance_>('/api/v1/ob/tenants', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/deleteReplica */
export async function deleteReplica(
  params: {
    // path
    tenantId?: number;
  },
  body?: Array<API.DeleteReplicaParam>,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_TaskInstance_>(`/api/v1/ob/tenants/${param0}/deleteReplica`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 DELETE /api/v1/ob/tenants/${param0} */
export async function deleteTenant(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}`, {
    method: 'DELETE',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0} */
export async function getTenant(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_TenantInfo_>(`/api/v1/ob/tenants/${param0}`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants */
export async function listTenants(
  params: {
    // query
    pageable?: API.Pageable;
    name?: string;
    mode?: Array<API.TenantMode>;
    locked?: boolean;
    readonly?: boolean;
    status?: Array<API.TenantStatus>;
  },
  options?: { [key: string]: any },
) {
  return request<API.PaginatedResponse_TenantInfo_>('/api/v1/ob/tenants', {
    method: 'GET',
    params: {
      ...params,
      pageable: undefined,
      ...params['pageable'],
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/lock */
export async function lockTenant(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/lock`, {
    method: 'POST',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/modifyPrimaryZone */
export async function modifyPrimaryZone(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.ModifyPrimaryZoneParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/modifyPrimaryZone`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/modifyReplica */
export async function modifyReplica(
  params: {
    // path
    tenantId?: number;
  },
  body?: Array<API.ModifyReplicaParam>,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_TaskInstance_>(`/api/v1/ob/tenants/${param0}/modifyReplica`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/modifyDescription */
export async function modifyTenantDescription(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.ModifyTenantDescriptionParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/modifyDescription`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/modifyWhitelist */
export async function modifyWhitelist(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.ModifyWhitelistParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/modifyWhitelist`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/preCheck */
export async function tenantPreCheck(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_TenantPreCheckResult_>(
    `/api/v1/ob/tenants/${param0}/preCheck`,
    {
      method: 'GET',
      params: { ...params },
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/unlock */
export async function unlockTenant(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/unlock`, {
    method: 'POST',
    params: { ...params },
    ...(options || {}),
  });
}
