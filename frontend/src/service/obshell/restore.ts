// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** Get restore history GET /api/v1/restore/history */
export async function getRestoreHistory(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getRestoreHistoryParams,
  options?: { [key: string]: any }
) {
  const { tenantName: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.CdbObRestoreHistory[] }>(
    '/api/v1/restore/history',
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

/** Get original restore tenant info by ob_admin GET /api/v1/tenant/${param0}/restore/source/info */
export async function getRestoreTenantInfoByObAdmin(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getRestoreTenantInfoByObAdminParams,
  options?: { [key: string]: any }
) {
  const { tenantName: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.RestoreTenantInfo }>(
    `/api/v1/tenant/${param0}/restore/source/info`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** Restore tenant test POST /api/v1/tenant/${param0}/restore/storage/test */
export async function restoreTenantTest(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.restoreTenantTestParams,
  body: API.RestoreStorageTestParam,
  options?: { [key: string]: any }
) {
  const { tenantName: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: Record<string, any> }>(
    `/api/v1/tenant/${param0}/restore/storage/test`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
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
