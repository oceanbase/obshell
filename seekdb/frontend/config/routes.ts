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

export default [
  {
    path: '/',
    component: 'Layout',
    name: '系统布局',
    routes: [
      {
        path: 'login',
        component: 'Login',
        name: '登录页',
      },

      {
        path: '/',
        component: 'Layout/BasicLayout',
        routes: [
          {
            path: '/',
            redirect: '/instance',
          },
          {
            path: 'instance',
            component: 'Tenant/Detail/Overview',
            name: '实例管理',
          },
          {
            path: 'database',
            component: 'Tenant/Detail/Database',
            name: '租户数据库',
          },
          {
            path: 'user',
            component: 'Tenant/Detail/User',
            name: '租户用户',
          },
          {
            path: 'parameter',
            component: 'Tenant/Detail/Parameter',
            name: '租户参数',
          },
          {
            path: 'task',
            component: 'Task',
            name: '任务中心',
          },
          {
            path: 'task/:taskId',
            component: 'Task/Detail',
            name: '任务详情',
          },
          {
            path: 'package',
            component: 'Package',
            name: '软件包管理',
          },
          // 性能监控
          {
            path: 'monitor',
            component: 'Monitor',
            name: '性能监控',
          },
          // 告警
          {
            path: 'alert',
            component: 'Alert',
            name: '告警',
            routes: [
              {
                path: '/alert',
                redirect: '/alert/event',
              },
              {
                path: 'event',
                component: 'Alert/Event',
                name: '告警事件',
              },
              {
                path: 'shield',
                component: 'Alert/Shield',
                name: '告警屏蔽',
              },
              {
                path: 'rules',
                component: 'Alert/Rules',
                name: '告警规则',
              },
            ],
          },
          {
            path: 'error',
            name: '错误页',
            routes: [
              {
                path: '/error',
                redirect: '/error/404',
              },
              {
                path: '403',
                component: 'Error/403',
                name: '403 错误',
              },
              {
                path: '404',
                component: 'Error/404',
                name: '404 错误',
              },
            ],
          },
        ],
      },
      {
        // 前面的路径都没有匹配到，说明是 404 页面
        // 使用 Error/404 覆盖默认的 404 路径
        component: 'Error/404',
        name: '404 页面不存在',
      },
    ],
  },
];
