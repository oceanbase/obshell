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
import { InputNumber } from '@oceanbase/design';
import type { InputNumberProps } from '@oceanbase/design/es/input-number';
import { isNullValue } from '@oceanbase/util';
import styles from './index.less';

export type ValueType = {
  min?: number;
  max?: number;
};

export interface RangeInputProps {
  value?: ValueType;
  onChange?: (value: ValueType) => void;
  minProps?: InputNumberProps;
  maxProps?: InputNumberProps;
}

class RangeInput extends React.Component<RangeInputProps> {
  static validate = (rule, value: ValueType, callback) => {
    const { min, max } = value || {};
    if (isNullValue(min) && isNullValue(max)) {
      callback(
        formatMessage({
          id: 'ocp-express.component.RangeInput.PleaseEnterTheMaximumAnd',
          defaultMessage: '请输入最大值和最小值',
        })
      );
    } else if (isNullValue(min)) {
      callback(
        formatMessage({
          id: 'ocp-express.component.RangeInput.PleaseEnterTheMinimumValue',
          defaultMessage: '请输入最小值',
        })
      );
    } else if (isNullValue(max)) {
      callback(
        formatMessage({
          id: 'ocp-express.component.RangeInput.PleaseEnterTheMaximumValue',
          defaultMessage: '请输入最大值',
        })
      );
    } else if ((min as number) > (max as number)) {
      callback(
        formatMessage({
          id: 'ocp-express.component.RangeInput.MinimumNeedLessThanEqual',
          defaultMessage: '最小值需要小于等于最大值',
        })
      );
    }
    callback();
  };

  public handleMinChange = (min: number | undefined) => {
    const { value, onChange } = this.props;
    if (onChange) {
      // 通过下发的 onChange 属性函数收集 WhitelistInput 的值
      onChange({
        ...(value || {}),
        min,
      });
    }
  };

  public handleMaxChange = (max: number | undefined) => {
    const { value, onChange } = this.props;
    if (onChange) {
      // 通过下发的 onChange 属性函数收集 WhitelistInput 的值
      onChange({
        ...(value || {}),
        max,
      });
    }
  };

  public render() {
    const { value, minProps = {}, maxProps = {} } = this.props;
    const { min, max } = value || {};
    return (
      <span className={styles.container}>
        <InputNumber value={min} onChange={this.handleMinChange} {...minProps} />
        <span className={styles.seperator}>~</span>
        <InputNumber value={max} onChange={this.handleMaxChange} {...maxProps} />
      </span>
    );
  }
}

export default RangeInput;
