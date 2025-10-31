// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** Get original restore tenant info by ob_admin POST /api/v1/restore/source-tenant-info */
export async function getRestoreSourceTenantInfo(
  body: API.RestoreStorageParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.RestoreTenantInfo }>(
    '/api/v1/restore/source-tenant-info',
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      data: body,
      ...(options || {}),
    }
  );
}

/** List restore tasks GET /api/v1/restore/tasks */
export async function listRestoreTasks(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.listRestoreTasksParams,
  options?: { [key: string]: any }
) {
  const { tenantName: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.PaginatedRestoreTaskResponse }>(
    '/api/v1/restore/tasks',
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** Get restore windows GET /api/v1/restore/windows */
export async function getRestoreWindows(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.RestoreWindowsParam }>(
    '/api/v1/restore/windows',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** Get restore task id DELETE /api/v1/tenant/${param0}/restore */
export async function cancelRestoreTask(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.cancelRestoreTaskParams,
  options?: { [key: string]: any }
) {
  const { tenantName: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    `/api/v1/tenant/${param0}/restore`,
    {
      method: 'DELETE',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** Get restore overview GET /api/v1/tenant/${param0}/restore/overview */
export async function getRestoreOverview(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getRestoreOverviewParams,
  options?: { [key: string]: any }
) {
  const { tenantName: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.RestoreOverview }>(
    `/api/v1/tenant/${param0}/restore/overview`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** Restore tenant POST /api/v1/tenant/restore */
export async function tenantRestore(body: API.RestoreParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/tenant/restore', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
