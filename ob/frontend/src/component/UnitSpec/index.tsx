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
import { Col, InputNumber, Row, theme } from '@oceanbase/design';
import { byte2GB, isNullValue } from '@oceanbase/util';
import React, { useState } from 'react';
import styles from './index.less';

export interface unitSpec {
  cpuCore: number;
  memorySize: number;
  dataDiskSize: number;
}

export interface UnitSpecProps {
  isSingleServer: boolean;
  unitSpecLimit?: any;
  idleUnitSpec?: any;
  defaultUnitSpec?: API.ObUnitConfig;
  onChange?: (value?: unitSpec) => void;
  isSharedStorage?: boolean;
}

const UnitSpec: React.FC<UnitSpecProps> = ({
  isSingleServer,
  unitSpecLimit,
  idleUnitSpec,
  defaultUnitSpec,
  onChange,
  isSharedStorage,
}) => {
  const { token } = theme.useToken();
  const { max_cpu, memory_size, data_disk_size } = defaultUnitSpec || {};
  const defaultCpuCoreValue = max_cpu || 0;
  const defaultMemorySizeValue = byte2GB(memory_size || 0);
  const defaultDataDiskSizeValue = byte2GB(data_disk_size || 0);

  const [cpuCoreValue, setCpuCoreValue] = useState(defaultCpuCoreValue);
  const [memorySizeValue, setMemorySizeValue] = useState(defaultMemorySizeValue);
  const [dataDiskSizeValue, setDataDiskSizeValue] = useState(defaultDataDiskSizeValue);

  const sharedStorageColSpan = isSharedStorage ? 8 : 12;

  const onValueChange = (cpuCore: number, memorySize: number, dataDiskSize: number) => {
    if (onChange) {
      onChange({ cpuCore, memorySize, dataDiskSize });
    }
  };
  const extraStyle = {
    color: token.colorTextTertiary,
    lineHeight: '10px',
  };

  const { cpuLowerLimit, memoryLowerLimit } = unitSpecLimit;
  const { idleCpuCore, idleMemoryInBytes } = idleUnitSpec;

  // 修改 unit 时，CUP可配置范围上限，当前 unit 已分配CUP + 剩余空闲CUP
  const currentMaxCpuCoreCount = idleCpuCore + defaultCpuCoreValue;
  // 修改 unit 时，可配置范围上限，当前 unit 已分配内存 + 剩余空闲内存
  const currentMaxMemorySize = idleMemoryInBytes + defaultMemorySizeValue;
  return (
    <Row
      gutter={8}
      style={{
        flex: 1,
        marginTop: isSingleServer ? -22 : -32,
      }}
    >
      <Col span={sharedStorageColSpan}>
        <div className={styles.unitSpecTitle}>CPU</div>
        <InputNumber
          step={0.5}
          min={isSingleServer ? cpuLowerLimit : 0}
          max={isSingleServer ? currentMaxCpuCoreCount : undefined}
          addonAfter={formatMessage({
            id: 'ocp-express.component.UnitSpec.Nuclear',
            defaultMessage: '核',
          })}
          defaultValue={defaultCpuCoreValue}
          onChange={(value: number) => {
            setCpuCoreValue(value);
            onValueChange(value, memorySizeValue, dataDiskSizeValue);
          }}
        />

        {isSingleServer && !isNullValue(cpuLowerLimit) && !isNullValue(currentMaxCpuCoreCount) && (
          <div style={extraStyle}>
            {cpuLowerLimit < currentMaxCpuCoreCount
              ? `[${cpuLowerLimit}，${currentMaxCpuCoreCount}]`
              : formatMessage({
                  id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                  defaultMessage: '当前可配置资源不足',
                })}
          </div>
        )}
      </Col>
      <Col span={sharedStorageColSpan}>
        <div className={styles.unitSpecTitle}>
          {formatMessage({ id: 'OBShell.component.UnitSpec.Memory', defaultMessage: '内存' })}
        </div>
        <InputNumber
          min={isSingleServer ? memoryLowerLimit : 0}
          max={isSingleServer ? currentMaxMemorySize : undefined}
          addonAfter="GB"
          defaultValue={defaultMemorySizeValue}
          onChange={(value: number) => {
            setMemorySizeValue(value);
            onValueChange(cpuCoreValue, value, dataDiskSizeValue);
          }}
        />

        {isSingleServer && !isNullValue(memoryLowerLimit) && !isNullValue(currentMaxMemorySize) && (
          <div style={extraStyle}>
            {memoryLowerLimit < currentMaxMemorySize
              ? `[${memoryLowerLimit}，${currentMaxMemorySize}]`
              : formatMessage({
                  id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                  defaultMessage: '当前可配置资源不足',
                })}
          </div>
        )}
      </Col>
      {isSharedStorage && (
        <Col span={8}>
          <div className={styles.unitSpecTitle}>
            {formatMessage({
              id: 'OBShell.component.UnitSpec.DataCaching',
              defaultMessage: '数据缓存',
            })}
          </div>
          <InputNumber
            addonAfter="GB"
            min={0}
            defaultValue={defaultDataDiskSizeValue}
            onChange={(value: number) => {
              setDataDiskSizeValue(value);
              onValueChange(cpuCoreValue, memorySizeValue, value);
            }}
          />
        </Col>
      )}
    </Row>
  );
};

export default UnitSpec;
