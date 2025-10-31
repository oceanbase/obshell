// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** get obcluster info get obcluster info GET /api/v1/obcluster/info */
export async function obclusterInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ClusterInfo }>('/api/v1/obcluster/info', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get obcluster parameters get obcluster parameters GET /api/v1/obcluster/parameters */
export async function obclusterParameters(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ClusterParameter[] }>(
    '/api/v1/obcluster/parameters',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** set obcluster parameters set obcluster parameters PATCH /api/v1/obcluster/parameters */
export async function obclusterSetParameters(
  body: API.SetObclusterParametersParam,
  options?: { [key: string]: any }
) {
  return request<any>('/api/v1/obcluster/parameters', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get resource unit config limit get resource unit config limit GET /api/v1/obcluster/unit-config-limit */
export async function getUnitConfigLimit(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ClusterUnitConfigLimit }>(
    '/api/v1/obcluster/unit-config-limit',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}
