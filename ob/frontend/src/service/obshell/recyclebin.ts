// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** purge recyclebin tenant purge tenant in recyclebin DELETE /api/v1/recyclebin/bin/${param0} */
export async function recyclebinTenantPurge(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.recyclebinTenantPurgeParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/recyclebin/bin/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** flashback tenant from recyclebin flashback tenant from recyclebin POST /api/v1/recyclebin/tenant/${param0} */
export async function recyclebinFlashbackTenant(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.recyclebinFlashbackTenantParams,
  body: API.FlashBackTenantParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/recyclebin/tenant/${param0}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** list all tenants in recyclebin list all tenants in recyclebin GET /api/v1/recyclebin/tenants */
export async function recyclebinTenantList(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DbaRecyclebin[] }>(
    '/api/v1/recyclebin/tenants',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}
