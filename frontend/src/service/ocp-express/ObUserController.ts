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

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/changePassword */
export async function changeDbUserPassword(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.ModifyDbUserPasswordParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/changePassword`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles */
export async function createDbRole(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.CreateDbRoleParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_DbRole_>(`/api/v1/ob/tenants/${param0}/roles`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users */
export async function createDbUser(
  params: {
    // path
    tenantId?: number;
  },
  body?: API.CreateDbUserParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.SuccessResponse_DbUser_>(`/api/v1/ob/tenants/${param0}/users`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 DELETE /api/v1/ob/tenants/${param0}/roles/${param1} */
export async function deleteDbRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/roles/${param1}`, {
    method: 'DELETE',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 DELETE /api/v1/ob/tenants/${param0}/users/${param1} */
export async function deleteDbUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/users/${param1}`, {
    method: 'DELETE',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/roles/${param1} */
export async function getDbRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.SuccessResponse_DbRole_>(`/api/v1/ob/tenants/${param0}/roles/${param1}`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/users/${param1} */
export async function getDbUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.SuccessResponse_DbUser_>(`/api/v1/ob/tenants/${param0}/users/${param1}`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/grantDbPrivilege */
export async function grantDbPrivilege(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.GrantDbPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/grantDbPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/grantGlobalPrivilege */
export async function grantGlobalPrivilege(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.GrantGlobalPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/grantGlobalPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/grantGlobalPrivilege */
export async function grantGlobalPrivilegeToRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.GrantGlobalPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/roles/${param1}/grantGlobalPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/grantObjectPrivilege */
export async function grantObjectPrivilegeToRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.GrantObjectPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/roles/${param1}/grantObjectPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/grantObjectPrivilege */
export async function grantObjectPrivilegeToUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.GrantObjectPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/grantObjectPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/grantRole */
export async function grantRoleToRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.GrantRoleParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/roles/${param1}/grantRole`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/grantRole */
export async function grantRoleToUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.GrantRoleParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/users/${param1}/grantRole`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/objects */
export async function listDbObjects(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.IterableResponse_DbObject_>(`/api/v1/ob/tenants/${param0}/objects`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/roles */
export async function listDbRoles(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.IterableResponse_DbRole_>(`/api/v1/ob/tenants/${param0}/roles`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/tenants/${param0}/users */
export async function listDbUsers(
  params: {
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0 } = params;
  return request<API.IterableResponse_DbUser_>(`/api/v1/ob/tenants/${param0}/users`, {
    method: 'GET',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/lock */
export async function lockDbUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/users/${param1}/lock`, {
    method: 'POST',
    params: { ...params },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/modifyDbPrivilege */
export async function modifyDbPrivilege(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.ModifyDbPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/modifyDbPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/modifyGlobalPrivilege */
export async function modifyGlobalPrivilege(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.ModifyGlobalPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/modifyGlobalPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/modifyGlobalPrivilege */
export async function modifyGlobalPrivilegeFromRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.ModifyGlobalPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/roles/${param1}/modifyGlobalPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/modifyObjectPrivilege */
export async function modifyObjectPrivilegeFromRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.ModifyObjectPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/roles/${param1}/modifyObjectPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/modifyObjectPrivilege */
export async function modifyObjectPrivilegeFromUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.ModifyObjectPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/modifyObjectPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/modifyRole */
export async function modifyRoleFromRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.ModifyRoleParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/roles/${param1}/modifyRole`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/modifyRole */
export async function modifyRoleFromUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.ModifyRoleParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/users/${param1}/modifyRole`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/revokeDbPrivilege */
export async function revokeDbPrivilege(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.RevokeDbPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/revokeDbPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/revokeGlobalPrivilege */
export async function revokeGlobalPrivilege(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.RevokeGlobalPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/revokeGlobalPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/revokeGlobalPrivilege */
export async function revokeGlobalPrivilegeFromRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.RevokeGlobalPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/roles/${param1}/revokeGlobalPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/revokeObjectPrivilege */
export async function revokeObjectPrivilegeFromRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.RevokeObjectPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/roles/${param1}/revokeObjectPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/revokeObjectPrivilege */
export async function revokeObjectPrivilegeFromUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.RevokeObjectPrivilegeParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(
    `/api/v1/ob/tenants/${param0}/users/${param1}/revokeObjectPrivilege`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/roles/${param1}/revokeRole */
export async function revokeRoleFromRole(
  params: {
    // path
    tenantId?: number;
    roleName?: string;
  },
  body?: API.RevokeRoleParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, roleName: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/roles/${param1}/revokeRole`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/revokeRole */
export async function revokeRoleFromUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  body?: API.RevokeRoleParam,
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/users/${param1}/revokeRole`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...params },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/tenants/${param0}/users/${param1}/unlock */
export async function unlockDbUser(
  params: {
    // path
    tenantId?: number;
    username?: string;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, username: param1 } = params;
  return request<API.NoDataResponse>(`/api/v1/ob/tenants/${param0}/users/${param1}/unlock`, {
    method: 'POST',
    params: { ...params },
    ...(options || {}),
  });
}
