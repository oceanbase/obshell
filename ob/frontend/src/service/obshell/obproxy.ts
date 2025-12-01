// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** Add obproxy POST /api/v1/obproxy */
export async function obproxyAdd(body: API.AddObproxyParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/obproxy', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Delete obproxy DELETE /api/v1/obproxy */
export async function obproxyDelete(options?: { [key: string]: any }) {
  return request<any>('/api/v1/obproxy', {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** Upload obproxy package POST /api/v1/obproxy/package */
export async function obproxyPkgUpload(body: {}, file?: File, options?: { [key: string]: any }) {
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
          formData.append(ele, new Blob([JSON.stringify(item)], { type: 'application/json' }));
        }
      } else {
        formData.append(ele, item);
      }
    }
  });

  return request<API.OcsAgentResponse & { data?: API.UpgradePkgInfo }>('/api/v1/obproxy/package', {
    method: 'POST',
    data: formData,
    requestType: 'form',
    ...(options || {}),
  });
}

/** Start obproxy POST /api/v1/obproxy/start */
export async function obproxyStart(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/obproxy/start', {
    method: 'POST',
    ...(options || {}),
  });
}

/** Stop obproxy POST /api/v1/obproxy/stop */
export async function obproxyStop(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/obproxy/stop', {
    method: 'POST',
    ...(options || {}),
  });
}

/** Upgrade obproxy POST /api/v1/obproxy/upgrade */
export async function obproxyUpgrade(
  body: API.UpgradeObproxyParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/obproxy/upgrade', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
