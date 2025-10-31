// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** join the specified agent join the specified agent POST /api/v1/agent */
export async function postAgent(body: API.JoinApiParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/agent', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** remove the specified agent remove the specified agent DELETE /api/v1/agent */
export async function deleteAgent(body: API.AgentInfo, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/agent', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get host info get host info GET /api/v1/agent/host-info */
export async function getHostInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.HostInfo }>('/api/v1/agent/host-info', {
    method: 'GET',
    ...(options || {}),
  });
}

/** join the specified agent join the specified agent POST /api/v1/agent/join */
export async function postAgentJoin(body: API.JoinApiParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/agent/join', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** remove the specified agent remove the specified agent POST /api/v1/agent/remove */
export async function postAgentRemove(body: API.AgentInfo, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/agent/remove', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
