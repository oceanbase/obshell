import { Input, Menu } from '@oceanbase/design';
import { SearchOutlined } from '@oceanbase/icons';
import React, { useEffect, useState } from 'react';
import EllipsisTooltip from '../common/EllipsisTooltip';
import styles from './index.less';

interface MonitorUnitSelectProps {
  units: Monitor.OptionType[];
  onSelect: (unit: Monitor.OptionType) => void;
}

export default (props: MonitorUnitSelectProps) => {
  const { units, onSelect } = props;
  const [selectedUnit, setSelectedUnit] = useState<Monitor.OptionType>();
  const [unitSearchStr, setUnitSearchStr] = useState('');

  // 给 selectedUnit 赋初值
  useEffect(() => {
    if (!units?.length || units?.find(u => u?.value === selectedUnit?.value)) return;
    handleSelect(units[0]);
  }, [units]);

  const handleSearch = e => {
    const { value } = e?.target || {};
    setUnitSearchStr(value);
  };

  const handleSelect = (unit: Monitor.OptionType) => {
    setSelectedUnit(unit);
    onSelect?.(unit);
  };

  const searchedUnits = units?.filter(s =>
    unitSearchStr ? !!String(s.value).includes(String(unitSearchStr)) : true
  );

  return (
    <div className={styles.search}>
      <div className={styles.searchInput}>
        <Input
          allowClear
          onClear={handleSearch}
          prefix={<SearchOutlined />}
          onPressEnter={handleSearch}
        />
      </div>
      <Menu
        selectedKeys={[String(selectedUnit?.value) || '']}
        items={searchedUnits.map(u => ({
          label: (
            <EllipsisTooltip title={u.value || ''} placement="right">
              {u.value || ''}
            </EllipsisTooltip>
          ),
          key: u.value || '',
          onClick: () => handleSelect(u),
        }))}
      />
    </div>
  );
};
