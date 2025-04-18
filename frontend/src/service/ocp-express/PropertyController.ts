/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/* eslint-disable */
// 该文件由 OneAPI 自动生成，请勿手动修改！
import request from '@/util/request';

/** 此处后端没有提供注释 GET /api/v1/config/properties */
export async function findNonFatalProperties(
  params: {
    // query
    keyLike?: string;
    pageable?: API.Pageable;
  },
  options?: { [key: string]: any },
) {
  return request<API.PaginatedResponse_PropertyMeta_>('/api/v1/config/properties', {
    method: 'GET',
    params: {
      ...params,
      pageable: undefined,
      ...params['pageable'],
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/config/systemInfo */
export async function getSystemInfo(options?: { [key: string]: any }) {
  return request<API.SuccessResponse_SystemInfo_>('/api/v1/config/systemInfo', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/config/refresh */
export async function refreshProperties(options?: { [key: string]: any }) {
  return request<API.NoDataResponse>('/api/v1/config/refresh', {
    method: 'POST',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 PUT /api/v1/config/properties/${param0} */
export async function updateProperty(
  params: {
    // query
    newValue?: string;
    // path
    id?: number;
  },
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.SuccessResponse_PropertyMeta_>(`/api/v1/config/properties/${param0}`, {
    method: 'PUT',
    params: {
      ...queryParams,
    },
    ...(options || {}),
  });
}
