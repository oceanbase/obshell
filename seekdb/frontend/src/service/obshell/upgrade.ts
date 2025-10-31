// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** upgrade agent upgrade agent POST /api/v1/agent/upgrade */
export async function agentUpgrade(body: API.UpgradeCheckParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/agent/upgrade', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** check agent upgrade check agent upgrade POST /api/v1/agent/upgrade/check */
export async function agentUpgradeCheck(
  body: API.UpgradeCheckParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    '/api/v1/agent/upgrade/check',
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

/** get all upgrade package infos get all upgrade package infos GET /api/v1/upgrade/package/info */
export async function getUpgradePkgInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.UpgradePkgInfo[] }>(
    '/api/v1/upgrade/package/info',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** delete upgrade package delete upgrade package DELETE /api/v1/upgrade/package */
export async function deleteUpgradePackage(
  body: API.DeletePackageParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/upgrade/package', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
