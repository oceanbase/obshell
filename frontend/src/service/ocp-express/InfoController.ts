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
import request from "@/util/request";

/** 此处后端没有提供注释 GET /api/v1/druid/stat */
export async function druidStat(options?: { [key: string]: any }) {
  return request<API.OneApiResult_object_>("/api/v1/druid/stat", {
    method: "GET",
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/git-info */
export async function gitInfo(options?: { [key: string]: any }) {
  return request<API.Map_String_Object_>("/api/v1/git-info", {
    method: "GET",
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/ */
export async function home(options?: { [key: string]: any }) {
  return request<API.OneApiResult_string_>("/api/v1/", {
    method: "GET",
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/info */
export async function info(options?: { [key: string]: any }) {
  return request<API.Map_String_Object_>("/api/info", {
    method: "GET",
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/status */
export async function status(options?: { [key: string]: any }) {
  return request<API.Map_String_Object_>("/api/v1/status", {
    method: "GET",
    ...(options || {}),
  });
}

/** 此处后端没有提供注释 GET /api/v1/time */
export async function time(options?: { [key: string]: any }) {
  return request<API.OneApiResult_offsetdatetime_>("/api/v1/time", {
    method: "GET",
    ...(options || {}),
  });
}
