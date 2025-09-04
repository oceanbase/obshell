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
import React from 'react';
import { Typography, Row, Col, Divider, Checkbox, Space, Popover, theme } from '@oceanbase/design';
import type { PopoverProps } from '@oceanbase/design/es/popover';
import type { CheckboxOptionType } from '@oceanbase/design/es/checkbox';
import { groupBy, some, uniq } from 'lodash';
import { isNullValue } from '@oceanbase/util';
import { InfoCircleFilled } from '@oceanbase/icons';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import useStyles from './index.style';

export interface OptionType extends CheckboxOptionType {
  description?: string;
  span?: number;
  group?: string;
}

export type OptionValue = string | number | boolean;

export interface CheckboxPopoverProps extends PopoverProps {
  options?: OptionType[];
  defaultValue: OptionValue[];
  value: OptionValue[];
  onChange: (value: OptionValue[]) => void;
  /* 最多可选中的对象数，为空时不限制选中数 */
  maxSelectCount?: number;
  maxSelectCountLabel?: string;
  children: React.ReactNode;
}

const CheckboxPopover = ({
  title,
  options,
  defaultValue,
  value,
  onChange,
  maxSelectCount,
  maxSelectCountLabel,
  children,
  overlayClassName,
  ...restProps
}: CheckboxPopoverProps) => {
  const { styles } = useStyles();
  const { token } = theme.useToken();

  // 分组列表
  const groupList = uniq(options?.map(item => item.group));
  const groupByData = groupBy(options, 'group');
  // 分组选项存在多于 3 个的情况下，设置最大宽度，否则设置最小宽度
  const width = some(Object.keys(groupByData), key => groupByData[key].length > 3) ? 608 : 480;
  return (
    <Popover
      overlayStyle={{
        minWidth: width,
        maxWidth: width,
        padding: 0,
      }}
      overlayClassName={`${styles.container} ${overlayClassName}`}
      {...restProps}
      content={
        <Row>
          <Space style={{ width: '100%', justifyContent: 'space-between', padding: 16 }}>
            <Space>
              <Typography.Title level={5} style={{ margin: 0 }}>
                <span>{title}</span>
              </Typography.Title>
              {!isNullValue(maxSelectCount) && (
                <span>
                  <InfoCircleFilled style={{ color: token.colorPrimary }} />
                  <span style={{ marginLeft: 4, fontSize: 12, color: 'rgba(0, 0, 0, 0.45)' }}>
                    {maxSelectCountLabel ||
                      formatMessage(
                        {
                          id: 'ocp-express.component.CheckboxPopover.YouCanSelectUpTo',
                          defaultMessage: '最多可选择 {maxSelectCount} 个对象',
                        },
                        { maxSelectCount: maxSelectCount }
                      )}
                  </span>
                </span>
              )}
            </Space>
            <a
              style={{ float: 'right' }}
              onClick={() => {
                onChange(defaultValue);
              }}
            >
              {formatMessage({
                id: 'ocp-express.component.CheckboxPopover.Reset',
                defaultMessage: '重置',
              })}
            </a>
          </Space>
          <Divider style={{ margin: 0 }} />
          <Checkbox.Group
            value={value}
            onChange={newValue => {
              onChange(newValue);
            }}
            style={{
              width: '100%',
              padding: 16,
              maxHeight: 400,
              overflow: 'auto',
            }}
          >
            {groupList.map((group, index) => {
              return (
                <>
                  {group && (
                    <h4 style={{ marginTop: index === 0 ? 0 : 16, marginBottom: 8 }}>{group}</h4>
                  )}

                  <Row gutter={[16, 16]}>
                    {options
                      ?.filter(item => !group || item.group === group)
                      ?.map(item => {
                        return (
                          <Col key={item.value} span={item.span || 6}>
                            <Checkbox
                              value={item.value}
                              disabled={
                                !isNullValue(maxSelectCount) && !value.includes(item.value)
                                  ? value.length >= (maxSelectCount || 0)
                                  : false
                              }
                            >
                              <ContentWithQuestion
                                content={item.label}
                                tooltip={
                                  item.description
                                    ? {
                                        title: item.description,
                                      }
                                    : false
                                }
                              />
                            </Checkbox>
                          </Col>
                        );
                      })}
                  </Row>
                </>
              );
            })}
          </Checkbox.Group>
        </Row>
      }
    >
      {children}
    </Popover>
  );
};

export default CheckboxPopover;
