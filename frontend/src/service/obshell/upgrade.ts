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

/** upgrade ob upgrade ob POST /api/v1/ob/upgrade */
export async function obUpgrade(body: API.ObUpgradeParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/ob/upgrade', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** check ob upgrade check ob upgrade POST /api/v1/ob/upgrade/check */
export async function obUpgradeCheck(
  body: API.UpgradeCheckParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/ob/upgrade/check', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get upgrade route of target ob package get upgrade route of target ob package GET /api/v1/ob/upgrade/route */
export async function upgradePkgRoute(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.UpgradePkgRouteParams,
  options?: { [key: string]: any }
) {
  return request<API.RouteNode[]>('/api/v1/ob/upgrade/route', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** upload upgrade package upload upgrade package POST /api/v1/upgrade/package */
export async function upgradePkgUpload(body: {}, file?: File, options?: { [key: string]: any }) {
  const formData = new FormData();

  if (file) {
    formData.append('file', file);
  }

  Object.keys(body).forEach(ele => {
    const item = (body as any)[ele];

    if (item !== undefined && item !== null) {
      if (typeof item === 'object' && !(item instanceof File)) {
        if (item instanceof Array) {
          item.forEach(f => formData.append(ele, f || ''));
        } else {
          formData.append(ele, JSON.stringify(item));
        }
      } else {
        formData.append(ele, item);
      }
    }
  });

  return request<API.OcsAgentResponse & { data?: API.UpgradePkgInfo }>('/api/v1/upgrade/package', {
    method: 'POST',
    data: formData,
    requestType: 'form',
    ...(options || {}),
  });
}

/** get all upgrade package infos get all upgrade package infos GET /api/v1/upgrade/package/info */
export async function upgradePkgInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.UpgradePkgInfo[] }>(
    '/api/v1/upgrade/package/info',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** backup params backup params POST /api/v1/upgrade/params/backup */
export async function paramsBackup(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ObParameters[] }>(
    '/api/v1/upgrade/params/backup',
    {
      method: 'POST',
      ...(options || {}),
    }
  );
}

/** restore params restore params POST /api/v1/upgrade/params/restore */
export async function paramsRestore(body: API.RestoreParams, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse>('/api/v1/upgrade/params/restore', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
