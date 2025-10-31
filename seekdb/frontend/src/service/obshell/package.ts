// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** upload upgrade package without body encryption upload upgrade package without body encryption POST /api/v1/package */
export async function newPkgUpload(body: {}, file?: File, options?: { [key: string]: any }) {
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

  return request<API.OcsAgentResponse & { data?: API.UpgradePkgInfo }>('/api/v1/package', {
    method: 'POST',
    data: formData,
    requestType: 'form',
    ...(options || {}),
  });
}

/** delete package in ocs delete package in ocs DELETE /api/v1/upgrade/package */
export async function deletePackage(
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
