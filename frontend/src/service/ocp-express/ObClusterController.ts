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

/** 此处后端没有提供注释 GET /api/v1/ob/cluster */
export async function getClusterInfo(options?: { [key: string]: any }) {
  return request<API.SuccessResponse_ClusterInfo_>('/api/v1/ob/cluster', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/cluster/unitSpecLimit */
export async function getClusterUnitSpecLimit(options?: { [key: string]: any }) {
  return request<API.SuccessResponse_ClusterUnitSpecLimit_>('/api/v1/ob/cluster/unitSpecLimit', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/cluster/server */
export async function getServerInfo(
  params: {
    // query
    ip?: string;
    obSvrPort?: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.SuccessResponse_Server_>('/api/v1/ob/cluster/server', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/ob/cluster/init */
export async function initObCluster(body?: API.ClusterInitParam, options?: { [key: string]: any }) {
  return request<API.SuccessResponse_BasicCluster_>('/api/v1/ob/cluster/init', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ob/cluster/charsets */
export async function listCharsets(
  params: {
    // query
    /** Tenant compatible mode. */
    tenantMode?: API.TenantMode;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_Charset_>('/api/v1/ob/cluster/charsets', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}
