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

/** 此处后端没有提供注释 POST /api/v2/ob/clusters/${param0}/units/${param1}/migrate */
export async function migrateUnit(
  params: {
    // path
    id?: number;
    obUnitId?: number;
  },
  body?: API.ComOceanbaseOcpObopsClusterParamMigrateUnitParam,
  options?: { [key: string]: any }
) {
  const { id: param0, obUnitId: param1 } = params;
  return request<API.ComOceanbaseOcpCoreResponseNoDataResponse>(
    `/api/v2/ob/clusters/${param0}/units/${param1}/migrate`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...params },
      data: body,
      ...(options || {}),
    }
  );
}

/** 此处后端没有提供注释 POST /api/v2/ob/clusters/${param0}/units/${param1}/rollbackMigration */
export async function rollbackMigrateUnit(
  params: {
    // path
    id?: number;
    obUnitId?: number;
  },
  options?: { [key: string]: any }
) {
  const { id: param0, obUnitId: param1 } = params;
  return request<API.ComOceanbaseOcpCoreResponseNoDataResponse>(
    `/api/v2/ob/clusters/${param0}/units/${param1}/rollbackMigration`,
    {
      method: 'POST',
      params: { ...params },
      ...(options || {}),
    }
  );
}

/** 此处后端没有提供注释 POST /api/v2/ob/clusters/${param0}/units/tryDeleteUnusedUnit */
export async function tryDeleteUnusedUnit(
  params: {
    // path
    id?: number;
  },
  options?: { [key: string]: any }
) {
  const { id: param0 } = params;
  return request<API.ComOceanbaseOcpCoreResponseNoDataResponse>(
    `/api/v2/ob/clusters/${param0}/units/tryDeleteUnusedUnit`,
    {
      method: 'POST',
      params: { ...params },
      ...(options || {}),
    }
  );
}

/** 此处后端没有提供注释 GET /api/v2/ob/clusters/${param0}/units/${param1} */
export async function getUnitStats(
  params: {
    // query
    serverIp: string;
    // path
    id?: number;
    obUnitId: number;
  },
  options?: { [key: string]: any }
) {
  const { id: param0, obUnitId: param1, ...queryParams } = params;
  return request<API.ComOceanbaseOcpCoreResponseSuccessResponse_ComOceanbaseOcpObopsClusterModelClusterUnitViewOfUnit_>(
    `/api/v2/ob/clusters/${param0}/units/${param1}`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}

/** 此处后端没有提供注释 GET /api/v2/ob/clusters/${param0}/units/${param1}/migrateDestinations */
export async function getUnitMigrateDestination(
  params: {
    // path
    id?: number;
    obUnitId?: number;
  },
  options?: { [key: string]: any }
) {
  const { id: param0, obUnitId: param1 } = params;
  return request<API.ComOceanbaseOcpCoreResponseSuccessResponse_ComOceanbaseOcpObopsClusterModelUnitMigrateDestination_>(
    `/api/v2/ob/clusters/${param0}/units/${param1}/migrateDestinations`,
    {
      method: 'GET',
      params: { ...params },
      ...(options || {}),
    }
  );
}
