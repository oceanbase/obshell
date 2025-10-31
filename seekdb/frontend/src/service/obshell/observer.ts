// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** get observer charsets get observer charsets GET /api/v1/observer/charsets */
export async function getObserverCharsets(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.CharsetInfo[] }>('/api/v1/observer/charsets', {
    method: 'GET',
    ...(options || {}),
  });
}

/** trigger  major compaction trigger  major compaction POST /api/v1/observer/compact */
export async function majorCompaction(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse>('/api/v1/observer/compact', {
    method: 'POST',
    ...(options || {}),
  });
}

/** get major compaction info get major compaction info GET /api/v1/observer/compaction */
export async function getCompaction(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.TenantCompaction }>(
    '/api/v1/observer/compaction',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** clear major compaction error clear major compaction error DELETE /api/v1/observer/compaction-error */
export async function clearCompactionError(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse>('/api/v1/observer/compaction-error', {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** get ob and agent info get ob and agent info GET /api/v1/observer/info */
export async function getObInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ObserverInfo }>('/api/v1/observer/info', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get parameters get parameters GET /api/v1/observer/parameters */
export async function getParameters(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getParametersParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.GvObParameter[] }>(
    '/api/v1/observer/parameters',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** set parameters set parameters PATCH /api/v1/observer/parameters */
export async function setParameters(
  body: API.SetParametersParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/observer/parameters', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** restart observer restart observer POST /api/v1/observer/restart */
export async function restartObserver(body: API.ObRestartParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/observer/restart', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** start observers start observers POST /api/v1/observer/start */
export async function startObserver(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/observer/start', {
    method: 'POST',
    ...(options || {}),
  });
}

/** get ob state get ob state GET /api/v1/observer/state */
export async function getObState(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: string }>('/api/v1/observer/state', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get statistics data get statistics data GET /api/v1/observer/statistics */
export async function getStatistics(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ObclusterStatisticInfo }>(
    '/api/v1/observer/statistics',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** stop observer stop observer POST /api/v1/observer/stop */
export async function stopObserver(body: API.ObStopParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/observer/stop', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get variables get variables GET /api/v1/observer/variables */
export async function getVariables(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getVariablesParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DbaObSysVariable[] }>(
    '/api/v1/observer/variables',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** set variables set variables PATCH /api/v1/observer/variables */
export async function setVariables(body: API.SetVariablesParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse>('/api/v1/observer/variables', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** modify whitelist modify whitelist PUT /api/v1/observer/whitelist */
export async function modifyWhitelist(
  body: API.ModifyWhitelistParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/observer/whitelist', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
