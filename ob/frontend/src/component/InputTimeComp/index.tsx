import { formatMessage } from '@/util/intl';
import { InputNumber, Select } from '@oceanbase/design';
import React, { useEffect, useState } from 'react';

interface InputTimeCompProps {
  onChange?: (value: number | null) => void;
  value?: number;
}

type UnitType = 'second' | 'minute' | 'hour';

const SELECT_OPTIONS = [
  {
    label: formatMessage({ id: 'OBShell.component.InputTimeComp.Seconds', defaultMessage: '秒' }),
    value: 'second',
  },
  {
    label: formatMessage({ id: 'OBShell.component.InputTimeComp.Minutes', defaultMessage: '分钟' }),
    value: 'minute',
  },
  {
    label: formatMessage({ id: 'OBShell.component.InputTimeComp.Hours', defaultMessage: '小时' }),
    value: 'hour',
  },
];

export default function InputTimeComp({ onChange, value }: InputTimeCompProps) {
  const [unit, setUnit] = useState<UnitType>('minute');

  const durationChange = (value: number | null) => {
    if (!value || unit === 'second') onChange?.(value);
    if (unit === 'minute') {
      onChange?.(value! * 60);
    }
    if (unit === 'hour') {
      onChange?.(value! * 3600);
    }
  };
  const getValue = (value: number | undefined) => {
    if (!value) return value;
    if (unit === 'minute') {
      return Math.floor(value / 60);
    }
    if (unit === 'hour') {
      return Math.floor(value / 3600);
    }
    return value;
  };

  useEffect(() => {
    if (value && value < 60 && unit === 'minute') {
      setUnit('second');
    }
  }, [value]);

  return (
    <InputNumber
      min={1}
      onChange={durationChange}
      value={getValue(value)}
      placeholder={formatMessage({
        id: 'OBShell.component.InputTimeComp.PleaseEnter',
        defaultMessage: '请输入',
      })}
      addonAfter={
        <Select
          defaultValue={'minute'}
          value={unit}
          onChange={(val: UnitType) => setUnit(val)}
          options={SELECT_OPTIONS}
        />
      }
    />
  );
}
