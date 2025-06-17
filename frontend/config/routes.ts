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
    spmb: 'b55539',
    routes: [
      {
        path: 'login',
        component: 'Login',
        name: '登录页',
        spmb: 'b55540',
      },
      {
        path: 'tenant/result/:taskId',
        component: 'Tenant/Result/Success',
        name: '新建租户任务提交成功',
        spmb: 'b55555',
      },
      {
        // 避免与租户详情冲突
        path: 'tenantCreate/new',
        component: 'Tenant/New',
        name: '新建租户',
        spmb: 'b55554',
      },
      {
        // 单个租户管理
        path: 'tenant/:tenantName',
        component: 'Tenant/Detail',
        routes: [
          {
            path: '/tenant/:tenantName',
            redirect: 'overview',
          },
          {
            path: 'overview',
            component: 'Tenant/Detail/Overview',
            name: '租户详情',
            spmb: 'b55541',
          },
          {
            path: 'monitor',
            component: 'Tenant/Detail/Monitor',
            name: '租户性能监控',
            spmb: 'b55575',
          },
          {
            path: 'database',
            component: 'Tenant/Detail/Database',
            name: '租户数据库',
            spmb: 'b55542',
          },
          {
            path: 'user',
            component: 'Tenant/Detail/User',
            name: '租户用户',
            spmb: 'b55543',
          },
          {
            path: 'user/role',
            component: 'Tenant/Detail/User',
            name: 'Oracle 租户角色',
            spmb: 'b55544',
          },
          {
            path: 'user/:username',
            component: 'Tenant/Detail/User/Oracle/UserOrRoleDetail',
            name: 'Oracle 租户用户详情',
            spmb: 'b55545',
          },
          {
            path: 'user/role/:roleName',
            component: 'Tenant/Detail/User/Oracle/UserOrRoleDetail',
            name: 'Oracle 租户角色详情',
            spmb: 'b55546',
          },
          {
            path: 'parameter',
            component: 'Tenant/Detail/Parameter',
            name: '租户参数',
            spmb: 'b55547',
          },
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
            spmb: 'b55548',
          },
          {
            path: 'overview/server/:ip/:serverPort',
            component: 'Cluster/Host',
            name: 'OBServer 详情',
            spmb: 'b55549',
          },
          {
            path: 'overview/unit',
            component: 'Cluster/Unit',
            name: 'Unit 分布',
            spmb: 'b55550',
          },
          {
            path: 'overview/parameter',
            component: 'Cluster/Parameter',
            name: '参数管理',
            spmb: 'b55551',
          },
          {
            path: 'tenant',
            component: 'Tenant',
            name: '租户列表',
            spmb: 'b55552',
          },
          {
            path: 'monitor',
            component: 'Monitor',
            name: '集群监控',
            spmb: 'b55553',
            ignoreMergeRoute: true,
            queryTitle: 'scope',
          },
          // 创建租户改为异步任务 结果页
          {
            path: 'diagnosis/session',
            component: 'Diagnosis/Session',
            name: '会话诊断',
            spmb: 'b55576',
            ignoreMergeRoute: true,
            queryTitle: 'tab',
          },
          {
            path: 'diagnosis/sql',
            component: 'Diagnosis/SQLDiagnosis',
            name: 'SQL 诊断',
            spmb: 'b55577',
            ignoreMergeRoute: true,
            queryTitle: 'tab',
          },
          {
            path: 'task',
            component: 'Task',
            name: '任务中心',
            spmb: 'b55556',
          },
          {
            path: 'task/:taskId',
            component: 'Task/Detail',
            name: '任务详情',
            spmb: 'b55557',
          },
          {
            path: 'package',
            component: 'Package',
            name: '软件包管理',
          },
          {
            path: 'property',
            component: 'Property',
            name: '系统管理',
            spmb: 'b55558',
          },
          {
            path: 'log',
            component: 'Log',
            name: '日志服务',
            spmb: 'b55559',
          },
          {
            path: 'error',
            name: '错误页',
            spmb: 'b55560',
            routes: [
              {
                path: '/error',
                redirect: '/error/404',
              },
              {
                path: '403',
                component: 'Error/403',
                name: '403 错误',
                spmb: 'b55561',
              },
              {
                path: '404',
                component: 'Error/404',
                name: '404 错误',
                spmb: 'b55562',
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
        spmb: 'b55563',
      },
    ],
  },
];
