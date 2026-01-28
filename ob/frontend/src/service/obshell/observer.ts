// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** get observer info get observer info including data_dir and redo_dir GET /api/v1/observer/info */
export async function observerInfo(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.Observer }>('/api/v1/observer/info', {
    method: 'GET',
    ...(options || {}),
  });
}
