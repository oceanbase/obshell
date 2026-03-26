import RefreshProgress from '@/component/MonitorSearch/AutoFresh/RefreshProgress';
import { DATE_TIME_FORMAT_DISPLAY } from '@/constant/datetime';
import { formatMessage } from '@/util/intl';
import { Button, Dropdown, Form, Space } from '@oceanbase/design';
import { DownOutlined, SyncOutlined } from '@oceanbase/icons';
import { DateRanger } from '@oceanbase/ui';
import { useInterval } from 'ahooks';
import dayjs from 'dayjs';
import { toString } from 'lodash';
import React, { useState } from 'react';

import { AUTO_FRESH_ITEMS, DATE_RANGER_SELECTS } from '@/constant/monitor';
import styles from './index.less';

const FormItem = Form.Item;

export type FREQUENCY_TYPE = number | 'off';
export interface AUTO_FRESH_ITEMS_TYPE {
  key: FREQUENCY_TYPE;
  label: string;
}

interface AutoFreshProps {
  queryRange?: dayjs.Dayjs[];
  currentMoment?: dayjs.Dayjs;
  isRealtime?: boolean;
  frequency: FREQUENCY_TYPE;
  style?: React.CSSProperties;
  onRefresh: (mergeDefault?: boolean) => void;
  withRanger?: boolean;
  onMenuClick: (key: string) => void;
  onRefreshClick: () => void;
  customRefreshItems?: AUTO_FRESH_ITEMS_TYPE[];
}

const AutoFresh: React.FC<AutoFreshProps> = ({
  queryRange,
  currentMoment,
  isRealtime,
  frequency,
  onRefresh,
  style,
  withRanger = true,
  onMenuClick,
  onRefreshClick,
  customRefreshItems,
}) => {
  const [second, setSecond] = useState<number>(0);
  const [loading, setLoading] = useState(false);

  useInterval(
    () => {
      if (second === (frequency as number) - 1) {
        setSecond(0);
        onRefresh(true);
      } else {
        setSecond(second + 1);
      }
    },
    isRealtime ? 1000 : undefined
  );

  const frequencyItems = customRefreshItems || AUTO_FRESH_ITEMS;
  const frequencyItem = frequencyItems.find(item => item.key === frequency);

  return (
    <div className={styles.autoFreshContainer} style={style}>
      <Dropdown.Button
        buttonsRender={([leftButton, rightButton]) => [
          leftButton,
          <Button key="frequency">
            <Space>
              {isRealtime
                ? frequencyItem?.label
                : formatMessage({
                    id: 'OBShell.MonitorSearch.AutoFresh.Close',
                    defaultMessage: '关闭',
                  })}
              <DownOutlined />
            </Space>
          </Button>,
        ]}
        menu={{
          items: frequencyItems.map(item => ({
            key: item.key,
            label: item.label,
          })),
          selectedKeys: [toString(frequency)],
          onClick: item => {
            setSecond(0);
            onMenuClick?.(item.key);
            onRefresh();
          },
        }}
        onClick={() => {
          if (isRealtime) return;
          setLoading(true);
          // 确保旋转动画至少完成一圈（1秒）
          setTimeout(() => {
            setLoading(false);
          }, 500);
          setSecond(0);
          onRefreshClick?.();
          onRefresh();
        }}
      >
        <Space>
          {isRealtime ? (
            <RefreshProgress value={second / ((frequency as number) - 1)} />
          ) : (
            <SyncOutlined spin={loading} />
          )}
          <span>
            {formatMessage({
              id: 'OBShell.MonitorSearch.AutoFresh.Refresh',
              defaultMessage: '刷新',
            })}
          </span>
        </Space>
      </Dropdown.Button>

      {withRanger && (
        <FormItem
          name="range"
          initialValue={queryRange}
          style={{ width: '470px', marginRight: 0, marginBottom: 0 }}
        >
          <DateRanger
            selects={DATE_RANGER_SELECTS}
            hasRewind={false}
            hasForward={false}
            hasSync={false}
            allowClear={false}
            defaultQuickValue={DateRanger.NEAR_1_HOURS.name}
            disabledDate={date => (currentMoment ? date.isAfter(currentMoment) : false)}
            format={DATE_TIME_FORMAT_DISPLAY}
            style={{ width: '100%' }}
          />
        </FormItem>
      )}
    </div>
  );
};

export default AutoFresh;
