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

/** 查询SlowSQL GET /api/v1/ob/tenants/${param0}/slowSql */
export async function slowSql(
  params: {
    // query
    startTime?: string;
    endTime?: string;
    server?: string;
    inner?: boolean;
    sqlText?: string;
    limit?: number;
    sqlTextLength?: number;
    filterExpression?: string;
    // path
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, ...queryParams } = params;
  return request<API.IterableResponse_SqlAuditStatSummary_>(
    `/api/v1/ob/tenants/${param0}/slowSql`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    },
  );
}

/** 查询slowSql数量排名前n的租户信息 GET /api/v1/ob/slowSql/top */
export async function slowSqlRank(
  params: {
    // query
    startTime?: string;
    endTime?: string;
    top?: number;
  },
  options?: { [key: string]: any },
) {
  return request<API.IterableResponse_SlowSqlRankInfo_>('/api/v1/ob/slowSql/top', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 查询SQL的文本
@param tenantId 租户Id
@param sqlId SQL_ID
 GET /api/v1/ob/tenants/${param0}/dbName/${param1}/sqls/${param2}/text */
export async function sqlText(
  params: {
    // query
    startTime?: string;
    endTime?: string;
    // path
    /** 租户Id */
    tenantId?: number;
    dbName?: string;
    /** SQL_ID */
    sqlId?: string;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, dbName: param1, sqlId: param2, ...queryParams } = params;
  return request<API.SuccessResponse_SqlText_>(
    `/api/v1/ob/tenants/${param0}/dbName/${param1}/sqls/${param2}/text`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    },
  );
}

/** 查询TopSQL**
@param tenantId 租户Id*
@param startTime 期间开始时间*
@param endTime 期间结束时间*
@param server 服务Ip+port
@param inner 是否包含内部SQL*
@param sqlText SQL的文本*
 GET /api/v1/ob/tenants/${param0}/topSql */
export async function topSql(
  params: {
    // query
    /** 期间开始时间* */
    startTime?: string;
    /** 期间结束时间* */
    endTime?: string;
    /** 服务Ip+port */
    server?: string;
    /** 是否包含内部SQL* */
    inner?: boolean;
    /** SQL的文本* */
    sqlText?: string;
    limit?: number;
    filterExpression?: string;
    // path
    /** 租户Id* */
    tenantId?: number;
  },
  options?: { [key: string]: any },
) {
  const { tenantId: param0, ...queryParams } = params;
  return request<API.IterableResponse_SqlAuditStatSummary_>(`/api/v1/ob/tenants/${param0}/topSql`, {
    method: 'GET',
    params: {
      ...queryParams,
    },
    ...(options || {}),
  });
}
