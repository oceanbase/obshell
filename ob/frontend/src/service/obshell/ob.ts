// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** get ob and agent info get ob and agent info GET /api/v1/ob/info */
export async function getObInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ObInfoResp }>('/api/v1/ob/info', {
    method: 'GET',
    ...(options || {}),
  });
}

/** init ob cluster init ob cluster POST /api/v1/ob/init */
export async function obInit(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/ob/init', {
    method: 'POST',
    ...(options || {}),
  });
}

/** cluster scale-in cluster scale-in POST /api/v1/ob/scale_in */
export async function postObScaleIn(
  body: API.ClusterScaleInParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/ob/scale_in', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** cluster scale-out cluster scale-out POST /api/v1/ob/scale_out */
export async function scaleOut(body: API.ClusterScaleOutParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/ob/scale_out', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** start observers start observers or the whole cluster, use param to specify POST /api/v1/ob/start */
export async function obStart(body: API.StartObParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/ob/start', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** stop observers stop observers or the whole cluster, use param to specify POST /api/v1/ob/stop */
export async function obStop(body: API.ObStopParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/ob/stop', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get obcluster charsets get obcluster charsets GET /api/v1/obcluster/charsets */
export async function getObclusterCharsets(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getObclusterCharsetsParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.CharsetInfo[] }>(
    '/api/v1/obcluster/charsets',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** put ob cluster configs put ob cluster configs POST /api/v1/obcluster/config */
export async function obclusterConfig(
  body: API.ObClusterConfigParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/obcluster/config', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get statistics data get statistics data GET /api/v1/obcluster/statistics */
export async function getStatistics(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ObclusterStatisticInfo }>(
    '/api/v1/obcluster/statistics',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** cluster scale-in cluster scale-in DELETE /api/v1/observer */
export async function deleteObserver(
  body: API.ClusterScaleInParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/observer', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** put observer configs put observer configs POST /api/v1/observer/config */
export async function obServerConfig(
  body: API.ObServerConfigParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/observer/config', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** delete zone delete zone DELETE /api/v1/zone/${param0} */
export async function deleteZone(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.DeleteZoneParams,
  options?: { [key: string]: any }
) {
  const { zoneName: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(`/api/v1/zone/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}
