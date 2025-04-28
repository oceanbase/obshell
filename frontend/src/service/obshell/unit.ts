// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** create resource unit config create resource unit config POST /api/v1/unit/config */
export async function unitConfigCreate(
  body: API.CreateResourceUnitConfigParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/unit/config', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get resource unit config get resource unit config GET /api/v1/unit/config/${param0} */
export async function unitConfigGet(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.unitConfigGetParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DbaObUnitConfig }>(
    `/api/v1/unit/config/${param0}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** drop resource unit config drop resource unit config DELETE /api/v1/unit/config/${param0} */
export async function unitConfigDrop(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.unitConfigDropParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/unit/config/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** get resource unit config limit get resource unit config limit GET /api/v1/unit/config/limit */
export async function getUnitConfigLimit(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ClusterUnitSpecLimit }>(
    '/api/v1/unit/config/limit',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** get all resource unit configs get all resource unit configs in the cluster GET /api/v1/units/config */
export async function unitConfigList(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DbaObUnitConfig[] }>('/api/v1/units/config', {
    method: 'GET',
    ...(options || {}),
  });
}
