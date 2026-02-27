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

import type { ModalFuncProps } from '@oceanbase/design';
import { Button, Dropdown, Menu, Popconfirm, Space, Tooltip } from '@oceanbase/design';
import type { ButtonProps } from '@oceanbase/design/es/button';
import type { PopconfirmProps } from '@oceanbase/design/es/popconfirm';
import { EllipsisOutlined } from '@oceanbase/icons';
import { history } from '@umijs/max';
import React from 'react';

import type { TooltipProps } from '@oceanbase/design/es/tooltip';

export interface getConfirmModalProps extends ModalFuncProps {
  operationType: 'confirm' | 'delete';
  subTitle?: React.ReactNode;
}

export function breadcrumbItemRender(route, params, routes) {
  const last = routes.indexOf(route) === routes.length - 1;

  return last ? (
    <span>{route.title || route.breadcrumbName}</span>
  ) : (
    <a
      onClick={() => {
        history.push(
          route.query
            ? {
                pathname: route.path,
                query: route.query,
              }
            : route.path
        );
      }}
    >
      {route.title || route.breadcrumbName}
    </a>
  );
}

export type Operation = {
  value: string;
  label: string;
  // 按钮样式
  buttonProps?: ButtonProps;
  tooltip?: Omit<TooltipProps, 'overlay'>;
  popconfirm?: PopconfirmProps;
  // 是否增加一个 divider
  divider?: boolean;
};

export function getOperationComponent<T>({
  operations = [],
  handleOperation,
  record,
  mode = 'link',
  // link 模式，默认露出一个操作项
  // button 模式，默认露出三个操作项
  displayCount = mode === 'link' ? 1 : 3,
}: {
  operations?: Operation[];
  handleOperation: (key: string, record: T, operations: Operation[]) => void;
  record: T; // link: 链接模式，常用于表格的操作列
  // buttom: 按钮模式，常用于页头的 extra 操作区域
  mode?: 'link' | 'button';
  displayCount?: number;
}) {
  const operations1 =
    mode === 'link'
      ? operations.slice(0, displayCount)
      : // button 模式的操作优先级与 link 模式正好相反，且只针对 operations1 生效
        // 由于 reverse 会改变原数组，因此这里采用解构浅拷贝一份数据进行操作
        [...operations.slice(0, displayCount)].reverse();
  const operations2 = operations.slice(displayCount, operations.length);

  return (
    // link 和 button 模式的间距做差异化处理
    <Space size={mode === 'link' ? 16 : 8}>
      {operations1.map(({ value, label, buttonProps, tooltip, popconfirm }, index) => {
        return (
          <Tooltip key={value} title={tooltip?.title} {...tooltip}>
            {mode === 'link' ? (
              popconfirm ? (
                <Popconfirm placement="bottomLeft" {...popconfirm}>
                  <a>{label}</a>
                </Popconfirm>
              ) : (
                <a onClick={() => handleOperation(value, record, operations)}>{label}</a>
              )
            ) : (
              <Button
                // 最后一个按钮设置为 primary
                type={index === operations1.length - 1 ? 'primary' : 'default'}
                onClick={() => handleOperation(value, record, operations)}
                {...(buttonProps || {})}
              >
                {label}
              </Button>
            )}
          </Tooltip>
        );
      })}

      {/* 操作项不为空，且具有其中一个操作项的权限 */}
      {operations2.length > 0 && (
        <Dropdown
          overlay={
            <Menu
              onClick={({ key }) => {
                handleOperation(key, record, operations);
              }}
            >
              {operations2.map(({ value, label, divider, tooltip }) => {
                return (
                  <>
                    <Menu.Item key={value}>
                      <Tooltip title={tooltip?.title} {...tooltip}>
                        <div>{label}</div>
                      </Tooltip>
                    </Menu.Item>
                    {divider && <Menu.Divider />}
                  </>
                );
              })}
            </Menu>
          }
        >
          {mode === 'link' ? (
            <a>
              <Button icon={<EllipsisOutlined />} size="small" />
            </a>
          ) : (
            <Button>
              <EllipsisOutlined />
            </Button>
          )}
        </Dropdown>
      )}
    </Space>
  );
}
