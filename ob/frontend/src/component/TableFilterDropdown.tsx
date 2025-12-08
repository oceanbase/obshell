import MyInput from '@/component/MyInput';
import { formatMessage } from '@/util/intl';
import React, { useEffect, useState } from 'react';
import { Button, Space } from '@oceanbase/design';
import type { FilterDropdownProps } from '@oceanbase/design/es/table/interface';

export interface TableFilterDropdownProps extends FilterDropdownProps {
  /* 确定筛选后的回调函数 */
  onConfirm?: (value: React.Key) => void;
}

const TableFilterDropdown: React.FC<TableFilterDropdownProps> = ({
  setSelectedKeys,
  selectedKeys,
  confirm,
  clearFilters,
  visible,
  onConfirm,
}) => {
  const [value, setValue] = useState('');
  const defaultValue = selectedKeys?.[0];

  const confirmFilter = () => {
    confirm();
    if (onConfirm) {
      onConfirm(selectedKeys && selectedKeys[0]);
    }
  };

  useEffect(() => {
    if (!visible) {
      confirmFilter();
    }
  }, [visible]);

  useEffect(() => {
    if (defaultValue) setValue(defaultValue);
  }, [defaultValue]);

  return (
    <div style={{ padding: 8 }}>
      <MyInput
        // value={selectedKeys?.[0]}
        onChange={e => {
          setValue(e.target.value);
          setSelectedKeys(e.target.value ? [e.target.value] : []);
        }}
        value={value}
        onPressEnter={() => {
          confirmFilter();
        }}
        style={{ width: 188, marginBottom: 8, display: 'block' }}
      />
      <Space>
        <Button
          type="primary"
          onClick={() => {
            confirmFilter();
          }}
          size="small"
          style={{ width: 90 }}
        >
          {formatMessage({ id: 'ocp-v2.src.util.component.Search', defaultMessage: '搜索' })}
        </Button>
        <Button
          onClick={() => {
            if (clearFilters) {
              clearFilters();
              setValue('');
            }
            confirmFilter();
          }}
          size="small"
          style={{ width: 90 }}
        >
          {formatMessage({ id: 'ocp-v2.src.util.component.Reset', defaultMessage: '重置' })}
        </Button>
      </Space>
    </div>
  );
};

export default TableFilterDropdown;
