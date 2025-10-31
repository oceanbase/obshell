// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** create database create database POST /api/v1/observer/database */
export async function createDatabase(
  body: API.CreateDatabaseParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/observer/database', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get database get database GET /api/v1/observer/database/${param0} */
export async function getDatabase(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getDatabaseParams,
  options?: { [key: string]: any }
) {
  const { database: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.Database }>(
    `/api/v1/observer/database/${param0}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** update database update database PUT /api/v1/observer/database/${param0} */
export async function updateDatabase(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.updateDatabaseParams,
  body: API.ModifyDatabaseParam,
  options?: { [key: string]: any }
) {
  const { database: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/observer/database/${param0}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** delete database delete database DELETE /api/v1/observer/database/${param0} */
export async function deleteDatabase(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteDatabaseParams,
  options?: { [key: string]: any }
) {
  const { database: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/observer/database/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** list databases list databases GET /api/v1/observer/databases */
export async function listDatabases(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.Database[] }>('/api/v1/observer/databases', {
    method: 'GET',
    ...(options || {}),
  });
}
