// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** List all alerts List all alerts POST /api/v1/alarm/alerts */
export async function listAlerts(body: API.AlertFilter, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.Alert[] }>('/api/v1/alarm/alerts', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Get a rule Get a rule by name GET /api/v1/alarm/rule/${param0} */
export async function getRule(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.GetRuleParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.RuleResponse }>(
    `/api/v1/alarm/rule/${param0}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** List all rules List all rules POST /api/v1/alarm/rules */
export async function listRules(body: API.RuleFilter, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.RuleResponse[] }>('/api/v1/alarm/rules', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Create or update a silencer Create or update a silencer PUT /api/v1/alarm/silencer */
export async function createOrUpdateSilencer(
  body: API.SilencerParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.SilencerResponse }>('/api/v1/alarm/silencer', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Get a silencer Get a silencer by id GET /api/v1/alarm/silencer/${param0} */
export async function getSilencer(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.GetSilencerParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.SilencerResponse }>(
    `/api/v1/alarm/silencer/${param0}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** Delete a silencer Delete a silencer by id DELETE /api/v1/alarm/silencer/${param0} */
export async function deleteSilencer(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.DeleteSilencerParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/alarm/silencer/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** List all silencers List all silencers POST /api/v1/alarm/silencers */
export async function listSilencers(body: API.SilencerFilter, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.SilencerResponse[] }>(
    '/api/v1/alarm/silencers',
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
