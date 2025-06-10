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

import { formatMessage } from '@/util/intl';
import React from 'react';
import { Lottie } from '@oceanbase/ui';
import type { MenuItem } from '@oceanbase/ui/es/BasicLayout';
import Icon from '@oceanbase/icons';
import IconFont from '@/component/IconFont';
import { ReactComponent as MonitorSvg } from '@/asset/monitor.svg';
import { ReactComponent as DiagnosisSvg } from '@/asset/diagnosis.svg';

export const useBasicMenu = (): MenuItem[] => {
  return [
    {
      link: '/overview',
      title: formatMessage({
        id: 'ocp-v2.src.hook.useMenu.ClusterManagement',
        defaultMessage: '集群管理',
      }),
      icon: <IconFont type="overview" />,
      selectedIcon: <Lottie path="/lottie/overview.json" mode="icon" speed={3} loop={false} />,
    },

    {
      link: '/tenant',
      title: formatMessage({
        id: 'ocp-express.src.util.menu.TenantManagement',
        defaultMessage: '租户管理',
      }),
      icon: <IconFont type="tenant" />,
      selectedIcon: <Lottie path="/lottie/tenant.json" mode="icon" speed={3} loop={false} />,
    },

    // {
    //   link: '/monitor',
    //   title: formatMessage({
    //     id: 'ocp-express.src.hook.useMenu.ClusterMonitoring',
    //     defaultMessage: '集群监控',
    //   }),
    //   icon: <Icon component={MonitorSvg} />,
    //   selectedIcon: <Lottie path="/lottie/monitor.json" mode="icon" speed={3} loop={false} />,
    // },

    // {
    //   link: '/diagnosis',
    //   title: formatMessage({
    //     id: 'ocp-express.src.hook.useMenu.DiagnosticAnalysis',
    //     defaultMessage: '诊断分析',
    //   }),
    //   icon: <Icon component={DiagnosisSvg} />,
    //   selectedIcon: <Lottie path="/lottie/diagnosis.json" mode="icon" speed={3} loop={false} />,
    //   children: [
    //     {
    //       link: `/diagnosis/session`,
    //       title: formatMessage({
    //         id: 'ocp-express.Diagnosis.Session.SessionManagement',
    //         defaultMessage: '会话诊断',
    //       }),
    //     },
    //     {
    //       link: `/diagnosis/sql`,
    //       title: formatMessage({
    //         id: 'ocp-express.src.hook.useMenu.SqlDiagnosis',
    //         defaultMessage: 'SQL 诊断',
    //       }),
    //     },
    //   ],
    // },

    // {
    //   link: '/log',
    //   title: formatMessage({
    //     id: 'ocp-express.page.Log.LogQuery',
    //     defaultMessage: '日志查询',
    //   }),
    //   icon: <IconFont type="log" />,
    //   selectedIcon: <Lottie path="/lottie/log.json" mode="icon" speed={3} loop={false} />,
    //   divider: true,
    // },

    {
      link: '/system',
      title: formatMessage({
        id: 'ocp-express.src.util.menu.SystemParameters',
        defaultMessage: '系统配置',
      }),
      icon: <IconFont type="property" />,
      selectedIcon: <Lottie path="/lottie/property.json" mode="icon" speed={3} loop={false} />,
      children: [
        {
          link: `/package`,
          title: formatMessage({
            id: 'ocp-v2.src.hook.useMenu.SoftwarePackageManagement',
            defaultMessage: '软件包管理',
          }),
        },
        {
          link: `/task`,
          title: formatMessage({
            id: 'ocp-v2.src.hook.useMenu.MissionCenter',
            defaultMessage: '任务中心',
          }),
        },
      ],
    },
  ];
};

export const useTenantMenu = (
  tenantId: number,
  tenantMode: API.TenantMode,
  oraclePrivilegeManagementSupported: boolean
): MenuItem[] => {
  const menus = [
    {
      link: `/tenant/${tenantId}`,
      title: formatMessage({ id: 'ocp-express.src.util.menu.Overview', defaultMessage: '总览' }),
    },
    // {
    //   link: `/tenant/${tenantId}/monitor`,
    //   title: formatMessage({
    //     id: 'ocp-express.src.hook.useMenu.PerformanceMonitoring',
    //     defaultMessage: '性能监控',
    //   }),
    // },

    {
      link: `/tenant/${tenantId}/database`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.DatabaseManagement',
        defaultMessage: '数据库管理',
      }),
      // MySQL 租户才有数据库管理
      accessible: tenantMode === 'MYSQL',
    },

    {
      link: `/tenant/${tenantId}/user`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.UserManagement',
        defaultMessage: '用户管理',
      }),
      // MySQL 租户或者支持权限管理的 Oracle 租户才有用户管理
      accessible: tenantMode === 'MYSQL' || oraclePrivilegeManagementSupported,
    },
    {
      link: `/tenant/${tenantId}/parameter`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.ParameterManagement',
        defaultMessage: '参数管理',
      }),
    },
  ];

  return menus;
};

export const useDiagnosisMenu = (clusterId: number): MenuItem[] => {
  const menus = [
    {
      link: `/diagnosis/${clusterId}/realtime`,
      title: formatMessage({
        id: 'ocp-express.src.hook.useMenu.RealTimeDiagnosis',
        defaultMessage: '实时诊断',
      }),
    },

    {
      link: `/diagnosis/${clusterId}/capacity`,
      title: formatMessage({
        id: 'ocp-express.src.hook.useMenu.CapacityCenter',
        defaultMessage: '容量中心',
      }),
    },

    {
      link: `/diagnosis/${clusterId}/report`,
      title: formatMessage({
        id: 'ocp-express.src.hook.useMenu.ReportCenter',
        defaultMessage: '报告中心',
      }),
    },

    {
      link: `/diagnosis/${clusterId}/optimize`,
      title: formatMessage({
        id: 'ocp-express.src.hook.useMenu.OptimizationHistory',
        defaultMessage: '优化历史',
      }),
    },
  ];

  return menus;
};
