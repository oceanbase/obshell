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

/** 此处后端没有提供注释 PUT /api/v1/profiles/me/changePassword */
export async function changePassword(
  body?: API.ChangePasswordRequest,
  options?: { [key: string]: any }
) {
  return request<API.SuccessResponse_Map_String_String__>('/api/v1/profiles/me/changePassword', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/profiles/me */
export async function userInfo(options?: { [key: string]: any }) {
  return request<API.SuccessResponse_AuthenticatedUser_>('/api/v1/profiles/me', {
    method: 'GET',
    ...(options || {}),
  });
}
