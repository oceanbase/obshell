import React, { useEffect, useState } from 'react';
import { Segmented } from '@oceanbase/design';
import { SegmentedProps } from '@oceanbase/design/es/segmented';

interface MySegmentedProps extends SegmentedProps {}

//从组件内部抽离成单独组件，维护自己的value状态，防止Segmented动画切换时每次都从第一项跳转到对应项
const MySegmented: React.FC<MySegmentedProps> = ({ value, onChange, options, style, ...rest }) => {
  const [segmentedValue, setSegmentedValue] = useState(value);

  useEffect(() => {
    setSegmentedValue(value);
  }, [value]);

  return (
    <Segmented
      value={segmentedValue}
      onChange={onChange}
      options={options}
      style={style}
      {...rest}
    />
  );
};

export default MySegmented;
