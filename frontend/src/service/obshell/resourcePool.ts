// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** drop resource pool drop resource pool which is not used by any tenant DELETE /api/v1/resource-pool/${param0} */
export async function poolDrop(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.poolDropParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/resource-pool/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** list all resource pools list all resource pools in the cluster GET /api/v1/resources-pool */
export async function poolList(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DbaObResourcePool[] }>(
    '/api/v1/resources-pool',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}
