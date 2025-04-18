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

/** 此处后端没有提供注释 POST /api/v1/hosts/logs/download */
export async function downloadLog(body?: API.DownloadLogRequest, options?: { [key: string]: any }) {
  return request<API.ResponseEntity_Resource_>('/api/v1/hosts/logs/download', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/obagent */
export async function getAgentDetail(
  params: {
    // query
    ip?: string;
    obSvrPort?: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.SuccessResponse_ObAgentDetail_>('/api/v1/obagent', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/hosts */
export async function getHostInfo(
  params: {
    // query
    ip?: string;
    obSvrPort?: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.SuccessResponse_HostInfo_>('/api/v1/hosts', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/hosts/logs/query */
export async function queryLog(body?: API.QueryLogRequest, options?: { [key: string]: any }) {
  return request<API.SuccessResponse_QueryLogResult_>('/api/v1/hosts/logs/query', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/obagent/restart */
export async function restartHostAgent(
  params: {
    // query
    ip?: string;
    obSvrPort?: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.SuccessResponse_TaskInstance_>('/api/v1/obagent/restart', {
    method: 'POST',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}
