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

import React, { useState, useEffect } from 'react';
import { formatMessage } from '@/util/intl';
import { Tooltip } from '@oceanbase/design';
import * as ObUnitController from '@/service/obshell/unit';
import type { SelectProps } from '@oceanbase/design/es/select';
import { getUnitSpecLimitText } from '@/util/tenant';
import { useRequest } from 'ahooks';
import { byte2GB } from '@oceanbase/util';
import MySelect from '@/component/MySelect';
import SelectDropdownRender from '@/component/SelectDropdownRender';
import AddUnitSpecModal from '@/component/AddUnitSpecModal';
import useStyles from './index.style';

const { Option } = MySelect;

export interface UnitSpecSelectProps extends SelectProps<any> {
  name?: string;
  obVersion?: string;
}

const UnitSpecSelect: React.FC<UnitSpecSelectProps> = ({ name, type, obVersion, ...restProps }) => {
  const { styles } = useStyles();

  const [visible, setVisible] = useState(false);

  // 获取目标集群的数据
  const {
    data,
    run: getUnitSpecList,
    loading,
  } = useRequest(ObUnitController.unitConfigList, {
    manual: true,
  });

  const unitSpecList = data?.data?.contents || [];

  // 参数变化，重新发起请求
  useEffect(() => {
    getUnitSpecList();
  }, [name, obVersion]);

  return (
    <span>
      <MySelect
        loading={loading}
        placeholder={formatMessage({
          id: 'ocp-express.src.component.UnitSpecSelect.PleaseSelectTheUnitSpecification',
          defaultMessage: '请选择 Unit 规格',
        })}
        dropdownRender={menu => (
          <SelectDropdownRender
            menu={menu}
            text={formatMessage({
              id: 'ocp-express.src.component.UnitSpecSelect.NewSpecifications',
              defaultMessage: '新增规格',
            })}
            onClick={() => {
              setVisible(true);
            }}
          />
        )}
        dropdownClassName="select-dropdown-with-description"
        {...restProps}
      >
        {unitSpecList.map(({ memory_size, max_cpu, name, unitSpecLimit }) => {
          return (
            // Unit 规格，直接将规格名称作为值
            // notRecommended为 true 时，表示 OB 集群不适合使用该 Unit 规格
            <Option
              key={name}
              value={name}
              // disabled 的选项 hover 时也需要出现背景，便于识别 tooltip 的位置
              className={styles.option}
            >
              <Tooltip
                title={unitSpecLimit && getUnitSpecLimitText(unitSpecLimit)}
                popupAlign={{
                  offset: [0, 10],
                }}
              >
                <div style={{ width: '100%' }}>
                  <span>{name}</span>
                  {/* 设置 float right， Option 被选中，在 Select Content 显示也是左右分开的，Option 中默认 justify-content space-between 所以不需要设置 float */}
                  <span
                    style={{
                      fontSize: 12,
                      color: 'rgba(0, 0, 0, 0.45)',
                      opacity: 1,
                      float: 'right',
                    }}
                  >
                    {`${max_cpu}C${byte2GB(memory_size)}G`}
                  </span>
                </div>
              </Tooltip>
            </Option>
          );
        })}
      </MySelect>
      <AddUnitSpecModal
        visible={visible}
        onCancel={() => {
          setVisible(false);
        }}
        onSuccess={() => {
          setVisible(false);
          getUnitSpecList();
        }}
      />
    </span>
  );
};

export default UnitSpecSelect;
