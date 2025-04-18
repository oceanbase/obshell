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
import React, { useEffect, useRef } from 'react';
import type { FilterDropdownProps } from '@oceanbase/design/es/table/interface';
import { theme, Button, Space } from '@oceanbase/design';
import { TreeSearch } from '@oceanbase/ui';
import type { TreeSearchRef, Node } from '@oceanbase/ui/es/TreeSearch';

export interface TableTreeFilterDropdownProps extends FilterDropdownProps {
  /* 确定筛选后的回调函数 */
  onConfirm?: (value?: React.Key[]) => void;
  treeData: Node[];
  // TODO: 待从 ob-ui 中导出TreeSearch 的属性类型定义并直接引用
  height?: number;
  width?: number;
  defaultExpandAll?: boolean;
}

const TableTreeFilterDropdown: React.FC<TableTreeFilterDropdownProps> = ({
  setSelectedKeys,
  selectedKeys,
  confirm,
  clearFilters,
  visible,
  onConfirm,
  ...restProps
}) => {
  const ref = useRef<TreeSearchRef>(null);
  const { token } = theme.useToken();

  const confirmFilter = () => {
    confirm();
    if (onConfirm) {
      onConfirm(selectedKeys);
    }
  };

  useEffect(() => {
    if (!visible) {
      confirmFilter();
    }
  }, [visible]);

  return (
    <div style={{ paddingTop: '12px' }}>
      <TreeSearch
        ref={ref}
        onChange={nodes => {
          setSelectedKeys(nodes.map(node => node.value));
        }}
        height={300}
        {...restProps}
      />

      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          padding: '5px 12px',
          borderTop: '1px solid #f0f0f0',
        }}
      >
        <Button
          type="link"
          size="small"
          onClick={() => {
            if (ref.current?.invertSelect) {
              ref.current?.invertSelect();
            }
            confirmFilter();
          }}
        >
          {formatMessage({
            id: 'ocp-express.Detail.Component.Event.InvertSelection',
            defaultMessage: '反选',
          })}
        </Button>
        <Space>
          <Button
            size="small"
            onClick={() => {
              if (ref.current?.reset) {
                ref.current?.reset();
              }
              if (clearFilters) {
                clearFilters();
              }
              if (onConfirm) {
                onConfirm(selectedKeys);
              }
            }}
            disabled={selectedKeys?.length === 0}
            style={
              selectedKeys?.length > 0
                ? {
                  color: token.colorPrimary,
                }
                : {}
            }
          >
            {formatMessage({
              id: 'ocp-express.Detail.Component.Event.Reset',
              defaultMessage: '重置',
            })}
          </Button>
          <Button
            type="primary"
            size="small"
            onClick={() => {
              confirmFilter();
            }}
          >
            {formatMessage({
              id: 'ocp-express.Detail.Component.Event.Determine',
              defaultMessage: '确定',
            })}
          </Button>
        </Space>
      </div>
    </div>
  );
};

export default TableTreeFilterDropdown;
