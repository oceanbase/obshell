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
import { useSelector } from 'umi';
import React, { useState } from 'react';
import { Space, Switch } from '@oceanbase/design';
import type { Moment } from 'moment';
import moment from 'moment';
import { useInterval } from 'ahooks';
import { DATE_TIME_FORMAT_DISPLAY } from '@/constant/datetime';
import { FREQUENCY } from '@/constant';
import RangeTimeDropdown from '@/component/RangeTimeDropdown';

export interface HostMonitorSearchValue {
  realtime: boolean;
  range: [Moment?, Moment?];
  menuKey: string;
}

export interface HostMonitorSearchProps {
  onChange?: (value: HostMonitorSearchValue) => void;
}

const HostMonitorSearch: React.FC<HostMonitorSearchProps> = ({ onChange }) => {
  const [realtime, setRealtime] = useState(false);
  const [range, setRange] = useState<[Moment?, Moment?]>([]);
  const [menuKey, setMenuKey] = useState<string>();
  const { systemInfo } = useSelector((state: DefaultRootState) => state.global);
  const collectInterval = systemInfo?.monitorInfo?.collectInterval || FREQUENCY;

  const handleChange = (values?: Partial<HostMonitorSearchValue>) => {
    if (onChange) {
      onChange({
        realtime,
        range,
        menuKey,
        ...values,
      });
    }
  };

  useInterval(
    () => {
      handleChange({
        range: [moment().subtract(10, 'minutes'), moment()],
        menuKey: 'custom',
      });
    },
    realtime ? collectInterval * 1000 : undefined
  );

  return (
    <Space size={16}>
      {!realtime ? (
        <>
          <RangeTimeDropdown
            menuKeys={['hour', 'day', 'week', 'custom']}
            onChange={(value, newMenuKey) => {
              setRange(value);
              setMenuKey(newMenuKey);
              handleChange({
                range: value || [],
                menuKey: newMenuKey,
              });
            }}
          />
        </>
      ) : (
        <span>
          {formatMessage({
            id: 'ocp-express.Host.Detail.Monitor.UpdateTime',
            defaultMessage: '更新时间:',
          })}

          {moment().format(DATE_TIME_FORMAT_DISPLAY)}
        </span>
      )}

      <Space>
        <span>
          {formatMessage({
            id: 'ocp-express.src.component.HostMonitorSearch.AutomaticRefresh',
            defaultMessage: '自动刷新:',
          })}
        </span>
        <Switch
          size="small"
          checked={realtime}
          style={{ marginTop: -4 }}
          onChange={value => {
            const newRange = value
              ? ([moment().subtract(10, 'minutes'), moment()] as [Moment, Moment])
              : range;
            const newMenuKey = value ? 'custom' : menuKey;
            setRealtime(value);
            setRange(newRange);
            setMenuKey(newMenuKey);
            handleChange({
              realtime: value,
              range: newRange,
              menuKey: newMenuKey,
            });
          }}
        />
      </Space>
    </Space>
  );
};

export default HostMonitorSearch;
