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
import { connect } from 'umi';
import { InputNumber } from '@oceanbase/design';
import type { SelectValue } from '@oceanbase/design/es/select';
import UnitSpecSelect from '@/component/UnitSpecSelect';
import styles from './index.less';

export type ValueType =
  | {
      unitSpecName?: string | undefined;
      unitCount?: number | undefined;
    }
  | undefined;

export interface ResourcePoolSelectProps {
  value?: ValueType;
  onChange?: (value: ValueType) => void;
  obVersion?: string;
}

interface ResourcePoolSelectState {
  value: ValueType;
}

class ResourcePoolSelect extends React.Component<ResourcePoolSelectProps, ResourcePoolSelectState> {
  static getDerivedStateFromProps(props: ResourcePoolSelectProps) {
    const newState: any = {
      prevProps: props,
    };

    const { value } = props;
    if ('value' in props) {
      // 将下发的 value 赋值给 FormEditTable
      newState.value = value;
    }
    return newState;
  }

  static validate = (rule, value, callback) => {
    if (value) {
      if (!value.unitSpecName) {
        callback(
          formatMessage({
            id: 'ocp-express.component.ResourcePoolSelect.PleaseSelectTheUnitSpecification',
            defaultMessage: '请选择 Unit 规格',
          })
        );
      }
      if (!value.unitCount) {
        callback(
          formatMessage({
            id: 'ocp-express.component.ResourcePoolSelect.PleaseEnterTheUnitQuantity',
            defaultMessage: '请输入 Unit 数量',
          })
        );
      }
      callback();
    }
    callback();
  };

  constructor(props: ResourcePoolSelectProps) {
    super(props);
    this.state = {
      value: undefined,
    };
  }

  public handleChange = (value: ValueType) => {
    const { onChange } = this.props;
    if (onChange) {
      // 通过下发的 onChange 属性函数收集 ResourcePoolSelect 的值
      onChange(value);
    }
  };

  public handleSelectChange = (val: SelectValue) => {
    const { value } = this.state;
    this.handleChange({
      ...value,
      unitSpecName: val as string | undefined,
    });
  };

  public handleInputNumberChange = (val: number | undefined) => {
    const { value } = this.state;

    this.handleChange({
      ...value,
      unitCount: val,
    });
  };

  public render() {
    const { value } = this.state;
    const { obVersion } = this.props;

    return (
      <span className={styles.container}>
        <UnitSpecSelect
          value={value && value.unitSpecName}
          obVersion={obVersion}
          onChange={this.handleSelectChange}
        />

        <InputNumber
          min={1}
          value={value && value.unitCount}
          onChange={this.handleInputNumberChange}
          placeholder={formatMessage({
            id: 'ocp-express.component.ResourcePoolSelect.PleaseEnterTheUnitQuantity',
            defaultMessage: '请输入 Unit 数量',
          })}
        />
      </span>
    );
  }
}

export default connect(() => ({}))(ResourcePoolSelect);
