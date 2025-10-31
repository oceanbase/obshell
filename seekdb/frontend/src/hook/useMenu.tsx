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
import { useSelector } from 'umi';
import Icon from '@oceanbase/icons';
import { ReactComponent as MonitorSvg } from '@/asset/monitor.svg';
import { ReactComponent as MonitorSelectedSvg } from '@/asset/monitor_selected.svg';
import { ReactComponent as AlarmSvg } from '@/asset/alarm.svg';
import { ReactComponent as AlarmSelectedSvg } from '@/asset/alarm_selected.svg';
import { ReactComponent as SystemSvg } from '@/asset/system.svg';
import { ReactComponent as SystemSelectedSvg } from '@/asset/system_selected.svg';
import { ReactComponent as InstanceSvg } from '@/asset/instance.svg';
import { ReactComponent as InstanceSelectedSvg } from '@/asset/instance_selected.svg';
import { ReactComponent as DatabaseSvg } from '@/asset/database.svg';
import { ReactComponent as DatabaseSelectedSvg } from '@/asset/database_selected.svg';
import { ReactComponent as UserSvg } from '@/asset/user.svg';
import { ReactComponent as UserSelectedSvg } from '@/asset/user_selected.svg';
import { ReactComponent as ParameterSvg } from '@/asset/parameter.svg';
import { ReactComponent as ParameterSelectedSvg } from '@/asset/parameter_selected.svg';
import { getInstanceAvailableFlag } from '@/util/instance';

export const useBasicMenu = (): MenuItem[] => {
  const { obInfoData } = useSelector((state: DefaultRootState) => state.global);

  const availableFlag = getInstanceAvailableFlag(obInfoData?.status || '');

  return [
    {
      link: '/instance',
      title: formatMessage({
        id: 'SeekDB.src.hook.useMenu.InstanceManagement',
        defaultMessage: '实例管理',
      }),
      icon: <Icon component={InstanceSvg} />,
      selectedIcon: <Icon component={InstanceSelectedSvg} />,
      accessible: true,
    },
    {
      link: `/database`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.DatabaseManagement',
        defaultMessage: '数据库管理',
      }),
      accessible: availableFlag,
      icon: <Icon component={DatabaseSvg} />,
      selectedIcon: <Icon component={DatabaseSelectedSvg} />,
    },
    {
      link: `/user`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.UserManagement',
        defaultMessage: '用户管理',
      }),
      accessible: availableFlag,
      icon: <Icon component={UserSvg} />,
      selectedIcon: <Icon component={UserSelectedSvg} />,
    },
    {
      link: '/alert',
      title: formatMessage({
        id: 'OBShell.src.hook.useMenu.AlarmCenter',
        defaultMessage: '告警中心',
      }),
      accessible: availableFlag,
      icon: <Icon component={AlarmSvg} />,
      selectedIcon: <Icon component={AlarmSelectedSvg} />,
    },

    {
      link: '/monitor',
      title: formatMessage({
        id: 'OBShell.src.hook.useMenu.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      accessible: availableFlag,
      icon: <Icon component={MonitorSvg} />,
      selectedIcon: <Icon component={MonitorSelectedSvg} />,
    },
    // {
    //   link: `/backup`,
    //   title: formatMessage({
    //     id: 'OBShell.src.hook.useMenu.BackupRecovery',
    //     defaultMessage: '备份恢复',
    //   }),
    // },
    {
      link: `/parameter`,
      title: formatMessage({
        id: 'ocp-express.src.util.menu.ParameterManagement',
        defaultMessage: '参数管理',
      }),
      accessible: availableFlag,
      icon: <Icon component={ParameterSvg} />,
      selectedIcon: <Icon component={ParameterSelectedSvg} />,
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
          accessible: availableFlag,
        },
        {
          link: `/task`,
          title: formatMessage({
            id: 'ocp-v2.src.hook.useMenu.MissionCenter',
            defaultMessage: '任务中心',
          }),
          accessible: true,
        },
      ],
    },
  ];
};
