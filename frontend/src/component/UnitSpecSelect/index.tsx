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
import { useDispatch, useSelector } from 'umi';
import { Tooltip } from '@oceanbase/design';
import type { SelectProps } from '@oceanbase/design/es/select';
import { getUnitSpecLimitText } from '@/util/tenant';
import MySelect from '@/component/MySelect';
import SelectDropdownRender from '@/component/SelectDropdownRender';
import AddUnitSpecModal from '@/component/AddUnitSpecModal';
import useStyles from './index.style';

const { Option } = MySelect;

export interface UnitSpecSelectProps extends SelectProps<any> {
  name?: string;
  // type 支持多选，因此直接透传给接口参数即可
  type?: API.UnitSpecType[];
  obVersion?: string;
}

const UnitSpecSelect: React.FC<UnitSpecSelectProps> = ({ name, type, obVersion, ...restProps }) => {
  const { styles } = useStyles();
  const dispatch = useDispatch();
  const { unitSpecList } = useSelector((state: DefaultRootState) => state.tenant);
  const loading = useSelector(
    (state: DefaultRootState) => state.loading.effects['tenant/getUnitSpecList']
  );

  const [visible, setVisible] = useState(false);

  const getUnitSpecList = () => {
    dispatch({
      type: 'tenant/getUnitSpecList',
      payload: {
        name,
        type,
        obVersion,
      },
    });
  };

  // 参数变化，重新发起请求
  useEffect(() => {
    getUnitSpecList();
    // 组件卸载时，不能重置状态，否则删除所在行时，其他行的 UnitSpecSelect 下拉选项会被清空
    // return () => {
    //   dispatch({
    //     type: 'tenant/update',
    //     payload: {
    //       unitSpecList: [],
    //     },
    //   });
    // };
  }, [name, type?.join(','), obVersion]);

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
        {unitSpecList.map(
          ({
            maxMemorySize: maxMemorySizeGB,
            maxCpuCoreCount,
            name: unitSpecName,
            notRecommended,
            unitSpecLimit,
          }) => {
            return (
              // Unit 规格，直接将规格名称作为值
              // notRecommended为 true 时，表示 OB 集群不适合使用该 Unit 规格
              <Option
                key={unitSpecName}
                value={unitSpecName}
                disabled={notRecommended}
                // disabled 的选项 hover 时也需要出现背景，便于识别 tooltip 的位置
                className={styles.option}
              >
                <Tooltip
                  title={notRecommended && getUnitSpecLimitText(unitSpecLimit)}
                  popupAlign={{
                    offset: [0, 10],
                  }}
                >
                  <div style={{ width: '100%' }}>
                    <span>{unitSpecName}</span>
                    {/* 设置 float right， Option 被选中，在 Select Content 显示也是左右分开的，Option 中默认 justify-content space-between 所以不需要设置 float */}
                    <span
                      style={{
                        fontSize: 12,
                        color: 'rgba(0, 0, 0, 0.45)',
                        opacity: 1,
                        float: 'right',
                      }}
                    >
                      {`${maxCpuCoreCount}C${maxMemorySizeGB}G`}
                    </span>
                  </div>
                </Tooltip>
              </Option>
            );
          }
        )}
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
