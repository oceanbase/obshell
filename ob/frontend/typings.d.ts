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

import type { ButtonProps } from '@oceanbase/design/es/button';
import type { PopconfirmProps } from '@oceanbase/design/es/popconfirm';
import type { TooltipProps } from '@oceanbase/design/es/tooltip';

declare module '*.less';
declare module '@/config/*';

declare global {
  namespace NodeJS {
    interface ProcessEnv {
      NODE_ENV?: 'development' | 'production';
      OBSHELL_PROXY?: string;
    }
  }

  interface Window {
    __OCP_REPORT_DATA?: any;
    __OCP_REPORT_LOCALE?: 'zh-CN' | 'en-US';
    Tracert: any;
  }

  namespace Global {
    interface AppInfo {
      buildVersion: string;
      buildTime: string;
    }

    interface ChartDataItem {
      value?: number;
      [key: string]: any;
    }

    type ChartData = ChartDataItem[];

    type AccessibleField = '';

    type MonitorApp = 'OB' | 'ODP' | 'HOST';

    type MonitorScope =
      | 'app'
      | 'obregion'
      | 'ob_cluster_id'
      | 'obproxy_cluster'
      | 'obproxy_cluster_id'
      | 'tenant_name'
      | 'obzone'
      | 'svr_ip'
      | 'mount_point'
      | 'mount_label'
      | 'task_type'
      | 'process'
      | 'device'
      | 'cpu';

    interface OperationItem {
      value: string;
      label: string;
      modalTitle?: string;
      divider?: boolean;
      isDanger?: boolean;
      buttonProps?: ButtonProps;
      tooltip?: Omit<TooltipProps, 'overlay'>;
      popconfirm?: PopconfirmProps;
    }

    interface StatusItem {
      value: string | boolean;
      label: string;
      badgeStatus?: 'success' | 'processing' | 'error' | 'default' | 'warning';
      tagColor?: string;
      color?: string;
      operations?: OperationItem[];
      [key: string]: any;
    }
  }
}
