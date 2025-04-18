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

/** 获取用户登录相关的 key，此接口不在登录验证范围内，且当前返回的是公钥，此公钥可以通过配置管理来修改.
@return 登录相关 key 信息
 GET /api/v1/loginKey */
export async function getLoginKey(options?: { [key: string]: any }) {
  return request<API.SuccessResponse_LoginKey_>('/api/v1/loginKey', {
    method: 'GET',
    ...(options || {}),
  });
}
