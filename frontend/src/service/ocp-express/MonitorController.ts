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

/** 此处后端没有提供注释 GET /api/v1/monitor/cache/stats */
export async function getMonitorStats(options?: { [key: string]: any }) {
  return request<API.SuccessResponse_Map_String_Number__>('/api/v1/monitor/cache/stats', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/monitor/metricGroups */
export async function listMetricClasses(
  params: {
    // query
    pageable?: API.Pageable;
    type?: API.MetricType;
    scope?: API.MetricScope;
  },
  options?: { [key: string]: any },
) {
  return request<API.PaginatedResponse_MetricClass_>('/api/v1/monitor/metricGroups', {
    method: 'GET',
    params: {
      ...params,
      pageable: undefined,
      ...params['pageable'],
    },
    ...(options || {}),
  });
}

/** 根据exporter id请求agent数据，排查问题使用 GET /api/v1/monitor/metric/queryByExporterId */
export async function queryByExporterId(
  params: {
    // query
    exporterId: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.SuccessResponse_string_>('/api/v1/monitor/metric/queryByExporterId', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /api/v1/monitor/debugQuery */
export async function queryForDebug(
  body?: API.MetricQueryDebugServiceDebugQueryParam,
  options?: { [key: string]: any },
) {
  return request<API.SuccessResponse_object_>('/api/v1/monitor/debugQuery', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/monitor/metric */
export async function queryMetric(
  params: {
    // query
    startTime?: string;
    endTime?: string;
    metrics?: string;
    groupBy?: string;
    interval?: number;
    minStep?: number;
    maxPoints?: number;
    limit?: number;
    labels?: string;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_Map_Object_Object__>('/api/v1/monitor/metric', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/monitor/metric/evalExpr */
export async function queryMetricDataWithExpr(
  params: {
    // query
    expr: string;
    startTime: number;
    endTime: number;
    step: number;
    isRouted?: boolean;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_OcpPrometheusQueryResult_>(
    '/api/v1/monitor/metric/evalExpr',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    },
  );
}

/** 此处后端没有提供注释 GET /api/v1/monitor/metric/series */
export async function queryMetricSeries(
  params: {
    // query
    startTime?: string;
    endTime?: string;
    metrics?: string;
    groupBy?: string;
    interval?: number;
    minStep?: number;
    max_points?: number;
    limit?: number;
    labels?: string;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_SeriesMetricValues_>('/api/v1/monitor/metric/series', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/monitor/top */
export async function queryMetricTop(
  params: {
    // query
    startTime?: string;
    endTime?: string;
    metrics?: string;
    groupBy?: string;
    interval?: number;
    minStep?: number;
    limit?: number;
    maxPoints?: number;
    labels?: string;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_Map_Object_Object__>('/api/v1/monitor/top', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/monitor/metricsWithLabel */
export async function queryMetricsWithLabel(
  params: {
    // query
    startTime?: string;
    endTime?: string;
    metrics?: string;
    groupBy?: string;
    interval?: number;
    minStep?: number;
    maxPoints?: number;
    labels?: string;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_Map_Object_Object__>('/api/v1/monitor/metricsWithLabel', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/monitor/top/realtime */
export async function queryRealtimeTopMetrics(
  params: {
    // query
    endTime?: string;
    duration?: number;
    metrics?: string;
    groupBy?: string;
    interval?: number;
    minStep?: number;
    maxPoints?: number;
    limit?: number;
    labels?: string;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_Map_Object_Object__>('/api/v1/monitor/top/realtime', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}
