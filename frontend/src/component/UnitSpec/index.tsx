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
import React, { useState } from 'react';
import { Col, Row, InputNumber } from '@oceanbase/design';
import { byte2GB } from '@oceanbase/util';

export interface unitSpec {
  cpuCore: number;
  memorySize: number;
}

export interface UnitSpecProps {
  unitSpecLimit?: any;
  idleUnitSpec?: any;
  defaultUnitSpec?: API.ObUnitConfig;
  onChange?: (value?: unitSpec) => void;
}

const UnitSpec: React.FC<UnitSpecProps> = ({
  unitSpecLimit,
  idleUnitSpec,
  defaultUnitSpec,
  onChange,
}) => {
  const [cpuCoreValue, setCpuCoreValue] = useState(defaultUnitSpec?.max_cpu);
  const [memorySizeValue, setMemorySizeValue] = useState(byte2GB(defaultUnitSpec?.memory_size));

  const onValueChange = (cpuCore: number, memorySize: number) => {
    if (onChange) {
      onChange({ cpuCore, memorySize });
    }
  };
  const extraStyle = {
    height: 22,
    fontSize: 14,
    color: '#8592AD',
    lineHeight: '22px',
  };

  const { cpuLowerLimit, memoryLowerLimit } = unitSpecLimit;
  const { idleCpuCore, idleMemoryInBytes } = idleUnitSpec;

  // 修改 unit 时，CUP可配置范围上限，当前 unit 已分配CUP + 剩余空闲CUP
  const currentMaxCpuCoreCount = idleCpuCore + defaultUnitSpec?.max_cpu;
  // 修改 unit 时，可配置范围上限，当前 unit 已分配内存 + 剩余空闲内存
  const currentMaxMemorySize = idleMemoryInBytes + byte2GB(defaultUnitSpec?.memory_size);

  return (
    <Row
      gutter={8}
      style={{
        flex: 1,
        paddingTop: 16,
      }}
    >
      <Col span={12}>
        <InputNumber
          min={cpuLowerLimit || 0.5}
          max={currentMaxCpuCoreCount}
          step={0.5}
          addonAfter={formatMessage({
            id: 'ocp-express.component.UnitSpec.Nuclear',
            defaultMessage: '核',
          })}
          defaultValue={defaultUnitSpec?.max_cpu}
          onChange={value => {
            setCpuCoreValue(value);
            onValueChange(value, memorySizeValue);
          }}
        />

        {cpuLowerLimit && currentMaxCpuCoreCount && (
          <div style={extraStyle}>
            {formatMessage(
              {
                id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueCpulowerlimitIdlecpucore',
                defaultMessage: '当前可配置范围值 {cpuLowerLimit}~{idleCpuCore}',
              },
              { cpuLowerLimit: cpuLowerLimit, idleCpuCore: currentMaxCpuCoreCount }
            )}
          </div>
        )}
      </Col>
      <Col span={12}>
        <InputNumber
          addonAfter="GB"
          min={memoryLowerLimit}
          max={currentMaxMemorySize}
          defaultValue={byte2GB(defaultUnitSpec?.memory_size)}
          onChange={(value: number) => {
            setMemorySizeValue(value);
            onValueChange(cpuCoreValue, value);
          }}
        />

        {memoryLowerLimit !== undefined && currentMaxMemorySize !== undefined && (
          <div style={extraStyle}>
            {memoryLowerLimit < currentMaxMemorySize
              ? formatMessage(
                  {
                    id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes',
                    defaultMessage: '当前可配置范围值 {memoryLowerLimit}~{idleMemoryInBytes}',
                  },
                  { memoryLowerLimit, idleMemoryInBytes: currentMaxMemorySize }
                )
              : formatMessage({
                  id: 'ocp-express.component.UnitSpec.CurrentConfigurableRangeValueMemorylowerlimitIdlememoryinbytes2',
                  defaultMessage: '当前可配置资源不足',
                })}
          </div>
        )}
      </Col>
    </Row>
  );
};

export default UnitSpec;
