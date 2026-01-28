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

/** trigger cluster inspection trigger cluster inspection with specified scenario (basic or performance) POST /api/v1/obcluster/inspection */
export async function triggerInspection(
  body: API.InspectionParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    '/api/v1/obcluster/inspection',
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

/** get inspection report by id get detailed inspection report by report id GET /api/v1/obcluster/inspection/report/${param0} */
export async function getInspectionReport(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getInspectionReportParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.InspectionReport }>(
    `/api/v1/obcluster/inspection/report/${param0}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** get inspection history get paginated inspection history with filters GET /api/v1/obcluster/inspection/reports */
export async function getInspectionHistory(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getInspectionHistoryParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.PaginatedInspectionHistoryResponse }>(
    '/api/v1/obcluster/inspection/reports',
    {
      method: 'GET',
      params: {
        // page has a default value: 1
        page: '1',
        // size has a default value: 10
        size: '10',

        // sort has a default value: start_time,desc
        sort: 'start_time,desc',
        ...params,
      },
      ...(options || {}),
    }
  );
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

/** get obcluster topology info get obcluster topology info GET /api/v1/obcluster/topology/info */
export async function obclusterTopologyInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ClusterTopology }>(
    '/api/v1/obcluster/topology/info',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
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
