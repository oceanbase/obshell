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
        path: 'tenant/result/:taskId',
        component: 'Tenant/Result/Success',
        name: '新建租户任务提交成功',
      },
      {
        // 避免与租户详情冲突
        path: 'tenantCreate/new',
        component: 'Tenant/New',
        name: '新建租户',
      },
      {
        // 单个租户管理
        path: 'cluster/tenant/:tenantName',
        component: 'Tenant/Detail',
        routes: [
          {
            path: '/cluster/tenant/:tenantName',
            redirect: 'overview',
          },
          {
            path: 'overview',
            component: 'Tenant/Detail/Overview',
            name: '租户详情',
          },
          {
            path: 'monitor',
            component: 'Tenant/Detail/Monitor',
            name: '租户性能监控',
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
            path: 'user/role',
            component: 'Tenant/Detail/User',
            name: 'Oracle 租户角色',
          },
          {
            path: 'user/:username',
            component: 'Tenant/Detail/User/Oracle/UserOrRoleDetail',
            name: 'Oracle 租户用户详情',
          },
          {
            path: 'user/role/:roleName',
            component: 'Tenant/Detail/User/Oracle/UserOrRoleDetail',
            name: 'Oracle 租户角色详情',
          },
          {
            path: 'parameter',
            component: 'Tenant/Detail/Parameter',
            name: '租户参数',
          },
          {
            path: 'session',
            component: 'Tenant/Detail/Session',
            name: '租户会话',
          },
          {
            path: 'session',
            component: 'Tenant/Detail/Session',
            name: '租户会话',
          },
          {
            path: 'backup',
            component: 'Tenant/Detail/Backup',
            name: '租户备份恢复',
          },
          // {
          //   path: 'backup/addBackupStrategy',
          //   component: 'Tenant/Detail/Backup/BackupStrategy/New',
          // },
          // {
          //   path: 'backup/editBackupStrategy',
          //   component: 'Tenant/Detail/Backup/BackupStrategy/Edit',
          // },
          {
            path: 'backup/restoreNow',
            component: 'Tenant/Detail/Backup/RestoreNow',
            name: '租户发起恢复',
          },
          {
            path: 'backup/nowBackupTenant',
            component: 'Backup/Backup/NowBackupTenant',
            name: '立即备份', // 租户
          },
          // {
          //   path: 'backup/copy',
          //   component: 'Backup/Restore/Copy',
          //   name: '复制恢复任务',
          // },
        ],
      },

      {
        path: '/',
        component: 'Layout/BasicLayout',
        routes: [
          {
            path: '/',
            redirect: '/overview',
          },
          {
            path: 'overview',
            component: 'Cluster/Overview',
            name: '集群详情',
          },
          {
            path: 'overview/parameter',
            component: 'Cluster/Parameter',
            name: '集群参数管理',
          },
          {
            path: 'tenant/result/:taskId',
            component: 'Tenant/Result/Success',
            name: '新建租户任务提交成功',
          },
          {
            // 避免与租户详情冲突
            path: 'tenantCreate/new',
            component: 'Tenant/New',
            name: '新建租户',
          },
          {
            path: 'tenant',
            component: 'Tenant',
            name: '租户列表',
          },
          {
            path: 'tenant/restore',
            component: 'Tenant/Detail/Backup/RestoreNow',
            name: '租户发起恢复',
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
          {
            path: 'credential',
            component: 'Credential',
            name: '凭据管理',
          },
          // 性能监控
          {
            path: 'monitor',
            component: 'Monitor',
            routes: [
              {
                path: '/monitor',
                redirect: '/monitor/cluster',
              },
              {
                path: 'cluster',
                component: 'Monitor/ClusterMonitor',
                name: 'OceanBase 集群监控',
              },
              {
                path: 'tenant',
                component: 'Monitor/TenantMonitor',
                name: 'OceanBase 租户监控',
              },
            ],
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
            path: 'inspection',
            name: '巡检服务',
            routes: [
              {
                path: '/inspection',
                component: 'Inspection',
              },
              {
                path: 'report/:id',
                component: 'Inspection/Result',
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
