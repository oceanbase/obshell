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
import { Button, Checkbox, Menu, Switch, Tooltip } from '@oceanbase/design';
import { isObject } from 'lodash';
import type { TooltipProps } from '@oceanbase/design/es/tooltip';

export interface AccessProps {
  /* 是否可访问 */
  accessible: boolean;
  /* 所见即所得模式下，无权限时的展示内容，与 tooltip 属性是互斥的 (优先级: tooltip > fallback)，不能同时设置 */
  fallback?: React.ReactNode;
  /* 默认是所见即所得模式，如果想展示为 disabled + Tooltip 模式，则需要设置 tooltip 属性 */
  tooltip?:
    | boolean
    | (Omit<TooltipProps, 'title'> & {
        // 将 title 改为可选属性
        title?: React.ReactNode;
      });
  children: React.ReactElement;
}

export default ({
  accessible = true,
  fallback,
  tooltip = false,
  children,
  ...restProps
}: AccessProps) => {
  const childrenProps = children.props || {};
  const disabled = !accessible || childrenProps.disabled;
  const tooltipProps = isObject(tooltip) ? tooltip || {} : {};
  const { title, ...restTooltipProps } = tooltipProps;
  const element = React.cloneElement(children, {
    style: {
      ...(disabled &&
      (children.type === Button || children.type === Checkbox || children.type === Switch)
        ? { pointerEvents: 'none' }
        : {}),
      ...childrenProps.style,
    },

    ...(children.type === Menu.Item || children.type === Menu.SubMenu
      ? { eventKey: children.key }
      : {}),
    ...(tooltip
      ? {
          // 根据 accessible 设置 disabled
          disabled,
        }
      : {}),
  });
  return tooltip ? (
    // disabled + Tooltip 模式
    <Tooltip
      title={
        !accessible &&
        tooltip &&
        (title ||
          formatMessage({
            id: 'ocp-express.src.component.Access.NoOperationPermissionIsAvailable',
            defaultMessage: '暂无操作权限，请联系管理员开通权限',
          }))
      }
      {...restTooltipProps}
      {...restProps}
    >
      <span
        style={{
          ...(children.type === Button ? { display: 'inline-block' } : {}),
          // 设置 disabled 状态下鼠标的样式
          ...(disabled ? { cursor: 'not-allowed' } : {}),
        }}
      >
        {element}
      </span>
    </Tooltip>
  ) : accessible ? (
    <span
      {...restProps}
      style={{
        ...(children.type === Button ? { display: 'inline-block' } : {}),
        // 设置 disabled 状态下鼠标的样式
        ...(disabled ? { cursor: 'not-allowed' } : {}),
      }}
    >
      {element}
    </span>
  ) : (
    <>{fallback || null}</>
  );
};
