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
import type { ReactElement, SyntheticEvent } from 'react';
import React from 'react';
import { PlusOutlined } from '@oceanbase/icons';
import { Button, Table } from '@oceanbase/design';
import type { ButtonProps } from '@oceanbase/design/es/button';
import type { TableProps, ColumnProps } from '@oceanbase/design/es/table';
import styles from './index.less';

export interface CellValueType<T> extends SyntheticEvent<T> {
  currentTarget: EventTarget & T;
  target: EventTarget & { value: any };
}

export interface AobEditFormTableProps<T> extends TableProps<T> {
  disabled?: boolean;
  value: T[];
  onEditChange: (value: T[]) => void; // 编辑值变化
  addButtonProps?: ButtonProps;
}

interface AobEditFormTableState {
  data: any[];
}

class AobEditFormTable<T> extends React.Component<AobEditFormTableProps<T>, AobEditFormTableState> {
  public static getDerivedStateFromProps<T2>(props: AobEditFormTableProps<T2>) {
    const newState: any = {
      prevProps: props,
    };

    const { value } = props;
    if ('value' in props) {
      // 将下发的 value 赋值给 AobEditFormTable，这里将 data 作为 AobEditFormTable 的值
      newState.data = value;
    }
    return newState;
  }

  public state = {
    data: [],
  };

  public handleEditChange = (record: any, index: number) => {
    const { onEditChange } = this.props;
    const { data } = this.state;
    let newData: any[] = [...data];
    const prevRecord: any = newData[index];
    newData = newData.map((item, itemIndex) => {
      if (index === itemIndex) {
        return {
          ...prevRecord,
          ...record,
        };
      }
      return item;
    });
    if (onEditChange) {
      // 通过下发的 onEditChange 属性函数收集 AobEditFormTable 的值
      onEditChange(newData);
    }
  };

  public handleCellChange = (item: any, record: any, index: number) => {
    this.handleEditChange(
      {
        ...record,
        ...item,
      },

      index
    );
  };

  public handleAddChange = (record: any) => {
    const { onEditChange } = this.props;
    const { data } = this.state;
    if (onEditChange) {
      // 通过下发的 onEditChange 属性函数收集 AobEditFormTable 的值
      onEditChange([...data, record]);
    }
  };

  public handleAdd = () => {
    const { columns } = this.props;
    const record = {};
    columns!.forEach(column => {
      if (column) {
        record[column.dataIndex] = '';
      }
    });
    this.handleAddChange(record);
  };

  public render() {
    const { disabled, addButtonProps, columns, ...restProps } = this.props;
    const newColumns = columns!.map((column: ColumnProps<T>) => {
      const { dataIndex, render } = column;
      const newColumn = {
        ...column,
        render: (text: string, record: any, index: number) => {
          const element = (render ? render!(text, record, index) : <span />) as ReactElement;
          const { onChange } = element.props;
          const nodeWithRender = React.cloneElement(element, {
            disabled: element.props.disabled || disabled,
            onChange: (value: CellValueType<T>) => {
              if (onChange) {
                onChange(value);
              }
              this.handleCellChange(
                {
                  [dataIndex!]: value.nativeEvent // 判断 value 是否为 Raect 合成事件
                    ? (value.target as HTMLInputElement) && value.target.value
                    : value,
                },

                record,
                index
              );
            },
          });

          const nodeWithoutRender = React.cloneElement(
            element,
            {
              onChange: (value: CellValueType<T>) => {
                this.handleCellChange(
                  {
                    [dataIndex!]: value.nativeEvent // 判断 value 是否为 Raect 合成事件
                      ? (value.target as HTMLInputElement) && value.target.value
                      : value,
                  },

                  record,
                  index
                );
              },
            },

            text
          );

          return render ? nodeWithRender : nodeWithoutRender;
        },
      };

      return newColumn;
    });
    const newProps = {
      ...restProps,
      columns: newColumns,
    };

    return (
      <div>
        <Table {...newProps} />
        <Button
          block={true}
          icon={<PlusOutlined />}
          className={styles.addButton}
          {...addButtonProps}
        >
          {formatMessage({ id: 'ocp-express.component.FormTable.Add', defaultMessage: '新增' })}
        </Button>
      </div>
    );
  }
}

export default AobEditFormTable;
