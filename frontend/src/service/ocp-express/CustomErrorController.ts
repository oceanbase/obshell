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

/** 此处后端没有提供注释 GET /error */
export async function handleError(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/error', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 DELETE /error */
export async function handleErrorUsingDELETE(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/error', {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 HEAD /error */
export async function handleErrorUsingHEAD(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/error', {
    method: 'HEAD',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 PATCH /error */
export async function handleErrorUsingPATCH(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/error', {
    method: 'PATCH',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /error */
export async function handleErrorUsingPOST(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/error', {
    method: 'POST',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 PUT /error */
export async function handleErrorUsingPUT(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/error', {
    method: 'PUT',
    ...(options || {}),
  });
}
