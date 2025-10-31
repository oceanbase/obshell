// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** create user create user POST /api/v1/observer/user */
export async function createUser(body: API.CreateUserParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse>('/api/v1/observer/user', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get user get user GET /api/v1/observer/user/${param0} */
export async function getUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getUserParams,
  options?: { [key: string]: any }
) {
  const { user: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.ObUser }>(`/api/v1/observer/user/${param0}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** drop user drop user DELETE /api/v1/observer/user/${param0} */
export async function dropUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.dropUserParams,
  options?: { [key: string]: any }
) {
  const { user: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/observer/user/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** modify db privilege of a user modify db privilege of a user PUT /api/v1/observer/user/${param0}/db-privileges */
export async function modifyDbPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyDbPrivilegeParams,
  body: API.ModifyUserDbPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { user: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/observer/user/${param0}/db-privileges`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** modify global privilege of a user modify global privilege of a user PUT /api/v1/observer/user/${param0}/global-privileges */
export async function modifyGlobalPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyGlobalPrivilegeParams,
  body: API.ModifyUserGlobalPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { user: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/observer/user/${param0}/global-privileges`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** lock user lock user POST /api/v1/observer/user/${param0}/lock */
export async function lockUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.lockUserParams,
  options?: { [key: string]: any }
) {
  const { user: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/observer/user/${param0}/lock`, {
    method: 'POST',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** unlock user unlock user DELETE /api/v1/observer/user/${param0}/lock */
export async function unlockUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.unlockUserParams,
  options?: { [key: string]: any }
) {
  const { user: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/observer/user/${param0}/lock`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** change user password change user password PUT /api/v1/observer/user/${param0}/password */
export async function changePassword(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.changePasswordParams,
  body: API.ChangeUserPasswordParam,
  options?: { [key: string]: any }
) {
  const { user: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/observer/user/${param0}/password`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get user stats get user stats GET /api/v1/observer/user/${param0}/stats */
export async function getStats(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getStatsParams,
  options?: { [key: string]: any }
) {
  const { user: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.ObUserStats }>(
    `/api/v1/observer/user/${param0}/stats`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** list users list users GET /api/v1/observer/users */
export async function listUsers(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ObUser[] }>('/api/v1/observer/users', {
    method: 'GET',
    ...(options || {}),
  });
}
