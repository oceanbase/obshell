// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** get all agent status get all agent status GET /api/v1/agents/status */
export async function getAllAgentsStatus(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: Record<string, any> }>('/api/v1/agents/status', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get git info get git info GET /api/v1/git-info */
export async function getGitInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.GitInfo }>('/api/v1/git-info', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get agent info get agent info GET /api/v1/info */
export async function getAgentInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.AgentStatus }>('/api/v1/info', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get secret get secret GET /api/v1/secret */
export async function getSecret(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.AgentSecret }>('/api/v1/secret', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get agent status get agent status GET /api/v1/status */
export async function getStatus(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.AgentStatus }>('/api/v1/status', {
    method: 'GET',
    ...(options || {}),
  });
}

/** get current time get current time GET /api/v1/time */
export async function getTime(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: string }>('/api/v1/time', {
    method: 'GET',
    ...(options || {}),
  });
}
