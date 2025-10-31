// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** list all metrics list all metrics meta info, return by groups GET /api/v1/metrics */
export async function listAllMetrics(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.ListAllMetricsParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.MetricClass[] }>('/api/v1/metrics', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** query metrics query metric data POST /api/v1/metrics/query */
export async function queryMetrics(body: API.MetricQuery, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.MetricData[] }>('/api/v1/metrics/query', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
