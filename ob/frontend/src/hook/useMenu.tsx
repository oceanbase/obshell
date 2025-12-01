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
import type { MenuItem } from '@oceanbase/ui/es/BasicLayout';
import Icon from '@oceanbase/icons';
import { ReactComponent as MonitorSvg } from '@/asset/monitor.svg';
import { ReactComponent as MonitorSelectedSvg } from '@/asset/monitor_selected.svg';
import { ReactComponent as AlarmSvg } from '@/asset/alarm.svg';
import { ReactComponent as AlarmSelectedSvg } from '@/asset/alarm_selected.svg';
import { ReactComponent as TenantSvg } from '@/asset/tenant.svg';
import { ReactComponent as TenantSelectedSvg } from '@/asset/tenant_selected.svg';
import { ReactComponent as ClusterSvg } from '@/asset/cluster.svg';
import { ReactComponent as ClusterSelectedSvg } from '@/asset/cluster_selected.svg';
import { ReactComponent as SystemSvg } from '@/asset/system.svg';
import { ReactComponent as SystemSelectedSvg } from '@/asset/system_selected.svg';
import useUiMode from './useUiMode';

export const useBasicMenu = (): MenuItem[] => {
  return [
    {
      link: '/overview',
      title: formatMessage({
        id: 'ocp-v2.src.hook.useMenu.ClusterManagement',
        defaultMessage: '集群管理',
      }),
      icon: <Icon component={ClusterSvg} />,
      selectedIcon: <Icon component={ClusterSelectedSvg} />,
    },

    {
      link: '/tenant',
      title: formatMessage({
        id: 'ocp-express.src.util.menu.TenantManagement',
        defaultMessage: '租户管理',
      }),
      icon: <Icon component={TenantSvg} />,
      selectedIcon: <Icon component={TenantSelectedSvg} />,
    },

    {
      link: '/monitor',
      title: formatMessage({
        id: 'OBShell.src.hook.useMenu.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      icon: <Icon component={MonitorSvg} />,
      selectedIcon: <Icon component={MonitorSelectedSvg} />,
    },

    {
      link: '/alert',
      title: formatMessage({
        id: 'OBShell.src.hook.useMenu.AlarmCenter',
        defaultMessage: '告警中心',
      }),
      icon: <Icon component={AlarmSvg} />,
      selectedIcon: <Icon component={AlarmSelectedSvg} />,
    },
    {
      link: '/system',
      title: formatMessage({
        id: 'ocp-express.src.util.menu.SystemManagement',
        defaultMessage: '系统管理',
      }),
      icon: <Icon component={SystemSvg} />,
      selectedIcon: <Icon component={SystemSelectedSvg} />,
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

export const useTenantMenu = (tenantName: string, tenantMode: API.TenantMode): MenuItem[] => {
  const { isDesktopMode } = useUiMode();
  const menus = [
    {
      link: `/cluster/tenant/${tenantName}`,
      title: formatMessage({ id: 'ocp-express.src.util.menu.Overview', defaultMessage: '总览' }),
    },
    {
      link: `/cluster/tenant/${tenantName}/database`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.DatabaseManagement',
        defaultMessage: '数据库管理',
      }),
      // MySQL 租户才有数据库管理
      accessible: tenantMode === 'MYSQL',
    },

    {
      link: `/cluster/tenant/${tenantName}/user`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.UserManagement',
        defaultMessage: '用户管理',
      }),
      accessible: true,
    },
    {
      link: `/cluster/tenant/${tenantName}/monitor`,
      title: formatMessage({
        id: 'OBShell.src.hook.useMenu.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      hidden: isDesktopMode,
    },
    {
      link: `/cluster/tenant/${tenantName}/backup`,
      title: formatMessage({
        id: 'OBShell.src.hook.useMenu.BackupRecovery',
        defaultMessage: '备份恢复',
      }),
      hidden: tenantName === 'sys' || isDesktopMode,
    },
    {
      link: `/cluster/tenant/${tenantName}/parameter`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.ParameterManagement',
        defaultMessage: '参数管理',
      }),
    },
  ];

  return (menus || []).filter(item => !item.hidden);
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
