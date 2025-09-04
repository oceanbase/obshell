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
import { DatePicker, Select, Space } from '@oceanbase/design';
import type { Moment } from 'moment';
import moment from 'moment';
import { ClockCircleOutlined } from '@oceanbase/icons';
import { useControllableValue } from 'ahooks';
import { DATE_TIME_FORMAT_DISPLAY } from '@/constant/datetime';
import type { RangeOption } from './constant';
import {
  NEAR_10_MINUTES,
  NEAR_1_HOURS,
  NEAR_3_HOURS,
  NEAR_6_HOURS,
  NEAR_20_MINUTES,
  NEAR_30_MINUTES,
  NEAR_5_MINUTES,
} from './constant';
import useStyles from './index.style';

export type RangeDateKey = 'customize' | string;

export type RangeDateValue = {
  key: RangeDateKey;
  range: [Moment?, Moment?];
};

interface IProps {
  selects?: RangeOption[];
  defaultValue?: RangeDateValue;
  // eslint-disable-next-line
  onChange?: (value: RangeDateValue) => void;
  // eslint-disable-next-line
  value?: RangeDateValue;
  mode?: 'normal' | 'mini';
  rangePickerStyle?: React.CSSProperties;
}

export const OCPRangePicker = (props: IProps) => {
  const { styles } = useStyles();
  const {
    selects = [
      NEAR_5_MINUTES,
      NEAR_10_MINUTES,
      NEAR_20_MINUTES,
      NEAR_30_MINUTES,
      NEAR_1_HOURS,
      NEAR_3_HOURS,
      NEAR_6_HOURS,
    ],

    defaultValue = { key: NEAR_1_HOURS.name, range: NEAR_1_HOURS.range() },
    value,
    mode = 'normal',
    rangePickerStyle = {},
  } = props;
  const [innerValue, setInnerValue] = useControllableValue<RangeDateValue>(props, {
    defaultValue: value || defaultValue,
  });
  const showBorder = mode === 'normal' || (mode === 'mini' && innerValue?.key === 'customize');
  const showRange = !(mode === 'mini' && innerValue?.key !== 'customize');

  const handleSelect = (key: RangeDateKey) => {
    if (key === 'customize') {
      setInnerValue({ ...(innerValue as RangeDateValue), key });
    } else {
      setInnerValue({
        key,
        range: selects.find(e => e.name === key)?.range() as [Moment, Moment],
      });
    }
  };

  const handleRangeChange = (range: [Moment, Moment]) => {
    if (!range) {
      setInnerValue({
        key: NEAR_30_MINUTES.name,
        range: NEAR_30_MINUTES.range(),
      });
    } else {
      setInnerValue({
        key: 'customize',
        range,
      });
    }
  };

  const disabledDate = (current: Moment) => {
    return current && current > moment().endOf('day');
  };

  return (
    <Space className={styles.ocpRangePicker}>
      <Space size={0}>
        {!showRange && <ClockCircleOutlined />}
        <Select
          bordered={showBorder}
          style={{ minWidth: 120 }}
          onSelect={handleSelect}
          value={innerValue?.key}
        >
          {selects.map(e => (
            <Select.Option key={e.name} value={e.name}>
              {e.name}
            </Select.Option>
          ))}

          <Select.Option value="customize">
            {formatMessage({
              id: 'ocp-express.component.OCPRangePicker.CustomTime',
              defaultMessage: '自定义时间',
            })}
          </Select.Option>
        </Select>
      </Space>
      {showRange && (
        <DatePicker.RangePicker
          disabledDate={disabledDate}
          format={DATE_TIME_FORMAT_DISPLAY}
          showTime={true}
          allowClear={false}
          value={innerValue?.range}
          onChange={handleRangeChange}
          style={rangePickerStyle}
        />
      )}
    </Space>
  );
};
