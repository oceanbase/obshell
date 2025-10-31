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

import MyInput from '@/component/MyInput';
import MySelect from '@/component/MySelect';
import { isEnglish } from '@/util';
import { formatMessage } from '@/util/intl';
import { Button, Col, Form, Row } from '@oceanbase/design';
import React from 'react';
import { sortByString } from '@oceanbase/util';
import { DeleteOutlined, PlusOutlined } from '@oceanbase/icons';
import type { FormInstance } from '@oceanbase/design/es/form';

const { Option } = MySelect;

export interface StartParamSelectAndInputProps {
  form: FormInstance;
  loading?: boolean;
  parameType: string;
  dataSource: API.ParameterInfo[];
  addText: string;
  lableText: string;
}

const StartParamSelectAndInput: React.FC<StartParamSelectAndInputProps> = ({
  form,
  loading,
  parameType,
  dataSource,
  addText,
  lableText,
}) => {
  return (
    <Form.List name={parameType} initialValue={[{}]}>
      {(fields, { add, remove }) => {
        return (
          <>
            {fields.map((field, index: number) => {
              return (
                <div key={field.key}>
                  <Row gutter={8}>
                    <Col span={7}>
                      <Form.Item shouldUpdate noStyle>
                        {() => {
                          const parameters = form.getFieldValue([parameType]);
                          let otherParameterNames: string[] = [];
                          if (Array.isArray(parameters)) {
                            otherParameterNames = parameters
                              .filter((parameter, i) => parameter?.name && i !== index)
                              .map(item => item.name);
                          }
                          return (
                            <Form.Item
                              style={{ marginBottom: 8 }}
                              {...field}
                              label={index === 0 && lableText}
                              name={[field.name, 'name']}
                              fieldKey={[field.fieldKey, 'name']}
                              rules={[
                                {
                                  required: true,
                                  message: formatMessage({
                                    id: 'ocp-express.src.component.StartParamSelectAndInput.SelectAParameter',
                                    defaultMessage: '请选择参数',
                                  }),
                                },
                              ]}
                            >
                              <MySelect
                                loading={loading}
                                showSearch={true}
                                optionFilterProp="label"
                                dropdownMatchSelectWidth={false}
                                dropdownClassName="select-dropdown-with-description"
                                placeholder={formatMessage({
                                  id: 'ocp-express.src.component.StartParamSelectAndInput.SelectAParameter',
                                  defaultMessage: '请选择参数',
                                })}
                              >
                                {dataSource
                                  // 对参数名称进行排序
                                  .sort((a, b) => (sortByString(a, b, 'name') ? 1 : -1))
                                  .filter(
                                    (item: API.ParameterInfo) =>
                                      !otherParameterNames.includes(item.name)
                                  )
                                  .map((item: API.ParameterInfo) => {
                                    return (
                                      <Option key={item.id} value={item.name} label={item.name}>
                                        <span>{item.name}</span>
                                        {/* 为保证回填到选择框时，展示仍然符合预期，需要设置相关样式 */}
                                        <span
                                          style={{
                                            fontSize: 12,
                                            color: 'rgba(0, 0, 0, 0.45)',
                                            opacity: 1,
                                            float: 'right',
                                          }}
                                        >
                                          <span>{item.type}</span>
                                        </span>
                                      </Option>
                                    );
                                  })}
                              </MySelect>
                            </Form.Item>
                          );
                        }}
                      </Form.Item>
                    </Col>
                    <Col span={isEnglish() ? 4 : 3}>
                      <Form.Item noStyle shouldUpdate>
                        {() => {
                          return (
                            <Form.Item
                              style={{ marginBottom: 8 }}
                              {...field}
                              label={index === 0 && ' '}
                              name={[field.name, 'value']}
                              fieldKey={[field.fieldKey, 'value']}
                              rules={[
                                {
                                  required: true,
                                  message: formatMessage({
                                    id: 'ocp-express.src.component.StartParamSelectAndInput.EnterParameterValues',
                                    defaultMessage: '请输入参数值',
                                  }),
                                },
                              ]}
                            >
                              <MyInput
                                disabled={!form.getFieldValue([parameType, index, 'name'])}
                                placeholder={formatMessage({
                                  id: 'ocp-express.src.component.StartParamSelectAndInput.EnterParameterValues',
                                  defaultMessage: '请输入参数值',
                                })}
                              />
                            </Form.Item>
                          );
                        }}
                      </Form.Item>
                    </Col>
                    <Col span={1}>
                      <Form.Item label={index === 0 && ' '} style={{ marginBottom: 8 }}>
                        <DeleteOutlined
                          style={{ color: 'rgba(0, 0, 0, .45)' }}
                          onClick={() => {
                            remove(field.name);
                          }}
                        />
                      </Form.Item>
                    </Col>
                  </Row>
                </div>
              );
            })}
            <Row gutter={8}>
              <Col span={isEnglish() ? 11 : 10}>
                <Form.Item style={{ marginBottom: 8 }}>
                  <Button
                    type="dashed"
                    onClick={() => {
                      add();
                    }}
                    style={{ width: '100%' }}
                  >
                    <PlusOutlined /> {addText}
                  </Button>
                </Form.Item>
              </Col>
            </Row>
          </>
        );
      }}
    </Form.List>
  );
};

export default StartParamSelectAndInput;
