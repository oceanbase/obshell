// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** get seekdb charsets get seekdb charsets GET /api/v1/seekdb/charsets */
export async function getCharsets(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.CharsetInfo[] }>('/api/v1/seekdb/charsets', {
    method: 'GET',
    ...(options || {}),
  });
}

/** trigger  major compaction trigger  major compaction POST /api/v1/seekdb/compact */
export async function majorCompaction(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse>('/api/v1/seekdb/compact', {
    method: 'POST',
    ...(options || {}),
  });
}

/** get major compaction info get major compaction info GET /api/v1/seekdb/compaction */
export async function getCompaction(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.TenantCompaction }>(
    '/api/v1/seekdb/compaction',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** clear major compaction error clear major compaction error DELETE /api/v1/seekdb/compaction-error */
export async function clearCompactionError(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse>('/api/v1/seekdb/compaction-error', {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** get seekdb info get seekdb info GET /api/v1/seekdb/info */
export async function getSeekdbInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ObserverInfo }>('/api/v1/seekdb/info', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get parameters get parameters GET /api/v1/seekdb/parameters */
export async function getParameters(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getParametersParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.GvObParameter[] }>(
    '/api/v1/seekdb/parameters',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** set parameters set parameters PATCH /api/v1/seekdb/parameters */
export async function setParameters(
  body: API.SetParametersParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/seekdb/parameters', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** restart seekdb restart seekdb POST /api/v1/seekdb/restart */
export async function restartSeekdb(body: API.ObRestartParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/seekdb/restart', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** start seekdb start seekdb POST /api/v1/seekdb/start */
export async function startSeekdb(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/seekdb/start', {
    method: 'POST',
    ...(options || {}),
  });
}

/** get statistics data get statistics data GET /api/v1/seekdb/statistics */
export async function getStatistics(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ObclusterStatisticInfo }>(
    '/api/v1/seekdb/statistics',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** stop seekdb stop seekdb POST /api/v1/seekdb/stop */
export async function stopSeekdb(body: API.ObStopParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/seekdb/stop', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get variables get variables GET /api/v1/seekdb/variables */
export async function getVariables(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getVariablesParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DbaObSysVariable[] }>(
    '/api/v1/seekdb/variables',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** set variables set variables PATCH /api/v1/seekdb/variables */
export async function setVariables(body: API.SetVariablesParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse>('/api/v1/seekdb/variables', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** modify whitelist modify whitelist PUT /api/v1/seekdb/whitelist */
export async function modifyWhitelist(
  body: API.ModifyWhitelistParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/seekdb/whitelist', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
