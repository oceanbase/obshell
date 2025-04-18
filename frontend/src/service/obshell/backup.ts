// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** Start backup for all tenants Start backup for all tenants POST /api/v1/obcluster/backup */
export async function obclusterStartBackup(
  body: API.BackupParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/obcluster/backup', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Patch backup status for all tenants Patch backup status for all tenants PATCH /api/v1/obcluster/backup */
export async function patchObclusterBackup(
  body: API.BackupStatusParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/obcluster/backup', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Set backup config for all tenants Set backup config for all tenants POST /api/v1/obcluster/backup/config */
export async function obclusterBackupConfig(
  body: API.ClusterBackupConfigParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    '/api/v1/obcluster/backup/config',
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

/** Patch backup config for all tenants Patch backup config for all tenants PATCH /api/v1/obcluster/backup/config */
export async function patchObclusterBackupConfig(
  body: API.ClusterBackupConfigParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    '/api/v1/obcluster/backup/config',
    {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      data: body,
      ...(options || {}),
    }
  );
}

/** Patch archive log status for all tenants Patch archive log status for all tenants PATCH /api/v1/obcluster/backup/log */
export async function patchObclusterArchiveLog(
  body: API.ArchiveLogStatusParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/obcluster/backup/log', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Get backup overview for all tenants Get backup overview for all tenants GET /api/v1/obcluster/backup/overview */
export async function obclusterBackupOverview(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.BackupOverview }>(
    '/api/v1/obcluster/backup/overview',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** Start backup for tenant Start backup for tenant POST /api/v1/tenant/${param0}/backup */
export async function tenantStartBackup(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantStartBackupParams,
  body: API.BackupParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    `/api/v1/tenant/${param0}/backup`,
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

/** Patch backup status for tenant Patch backup status for tenant PATCH /api/v1/tenant/${param0}/backup */
export async function patchTenantBackup(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.patchTenantBackupParams,
  body: API.BackupStatusParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/backup`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** Set backup config for tenant Set backup config for tenant POST /api/v1/tenant/${param0}/backup/config */
export async function tenantBackupConfig(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantBackupConfigParams,
  body: API.TenantBackupConfigParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    `/api/v1/tenant/${param0}/backup/config`,
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

/** Patch archive log status for tenant Patch archive log status for tenant PATCH /api/v1/tenant/${param0}/backup/log */
export async function patchTenantArchiveLog(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.patchTenantArchiveLogParams,
  body: API.ArchiveLogStatusParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/backup/log`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** Get backup overview for tenant Get backup overview for tenant GET /api/v1/tenant/${param0}/backup/overview */
export async function tenantBackupOverview(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantBackupOverviewParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.BackupOverview }>(
    `/api/v1/tenant/${param0}/backup/overview`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}
