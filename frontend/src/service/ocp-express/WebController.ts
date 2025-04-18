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

/** 此处后端没有提供注释 GET / */
export async function index(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 DELETE / */
export async function indexUsingDELETE(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/', {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 HEAD / */
export async function indexUsingHEAD(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/', {
    method: 'HEAD',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 PATCH / */
export async function indexUsingPATCH(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/', {
    method: 'PATCH',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST / */
export async function indexUsingPOST(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/', {
    method: 'POST',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 PUT / */
export async function indexUsingPUT(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/', {
    method: 'PUT',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /login */
export async function login(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/login', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 DELETE /login */
export async function loginUsingDELETE(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/login', {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 HEAD /login */
export async function loginUsingHEAD(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/login', {
    method: 'HEAD',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 PATCH /login */
export async function loginUsingPATCH(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/login', {
    method: 'PATCH',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 POST /login */
export async function loginUsingPOST(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/login', {
    method: 'POST',
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 PUT /login */
export async function loginUsingPUT(options?: { [key: string]: any }) {
  return request<API.ModelAndView>('/login', {
    method: 'PUT',
    ...(options || {}),
  });
}
