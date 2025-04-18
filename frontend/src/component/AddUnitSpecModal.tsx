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

import React from 'react';
import { formatMessage } from '@/util/intl';
import { connect } from 'umi';
import { Form, InputNumber, Modal } from '@oceanbase/design';
import { isNullValue } from '@oceanbase/util';
import MyInput from '@/component/MyInput';
import type { ValueType } from '@/component/RangeInput';
import { MODAL_FORM_ITEM_LAYOUT, NAME_RULE } from '@/constant';

const FormItem = Form.Item;

export interface AddUnitSpecModalProps {
  visible: boolean;
  dispatch: any;
  onSuccess: () => void;
  /** unitSpec 为空时，为新增 unit 规格; unitSpec 不为空时，为编辑 unit 规格 */
  unitSpec?: API.UnitSpec;
  loading: boolean;
}

const AddUnitSpecModal: React.FC<AddUnitSpecModalProps> = ({
  dispatch,
  unitSpec,
  onSuccess,
  loading,
  ...restProps
}) => {
  const isEdit = !!unitSpec;
  const [form] = Form.useForm();

  const { validateFields } = form;

  const validateCPU = (rule, value: ValueType, callback) => {
    if (!isNullValue(value) && (value as number) % 0.5 !== 0) {
      callback(
        formatMessage({
          id: 'ocp-express.src.component.AddUnitSpecModal.MustBeAnIntegerMultipleOf',
          defaultMessage: '需要是 0.5 的整数倍',
        })
      );
    }
    callback();
  };

  const handleSubmit = () => {
    validateFields().then(values => {
      const { name = '', maxCpuCoreCount, maxMemorySize } = values;

      if (isEdit) {
        dispatch({
          type: 'tenant/modifyUnitSpec',
          name: unitSpec?.name,
          payload: {
            specId: unitSpec?.id,
            maxCpuCoreCount,
            maxMemorySize,
          },

          onSuccess: () => {
            if (onSuccess) {
              onSuccess();
            }
          },
        });
      } else {
        dispatch({
          type: 'tenant/addUnitSpec',
          payload: {
            name,
            maxCpuCoreCount,
            maxMemorySize,
          },

          onSuccess: () => {
            if (onSuccess) {
              onSuccess();
            }
          },
        });
      }
    });
  };

  return (
    <Modal
      title={
        isEdit
          ? formatMessage({
              id: 'ocp-express.src.component.AddUnitSpecModal.ModifyUnitSpecifications',
              defaultMessage: '编辑 Unit 规格',
            })
          : formatMessage({
              id: 'ocp-express.src.component.AddUnitSpecModal.NewUnitSpecifications',
              defaultMessage: '新增 Unit 规格',
            })
      }
      destroyOnClose={true}
      onOk={handleSubmit}
      confirmLoading={loading}
      {...restProps}
    >
      <Form
        layout="vertical"
        form={form}
        preserve={false}
        hideRequiredMark={true}
        {...MODAL_FORM_ITEM_LAYOUT}
      >
        <FormItem
          name="name"
          label={formatMessage({
            id: 'ocp-express.src.component.AddUnitSpecModal.SpecificationName',
            defaultMessage: '规格名称',
          })}
          rules={
            isEdit
              ? undefined
              : [
                  {
                    required: true,
                    message: formatMessage({
                      id: 'ocp-express.src.component.AddUnitSpecModal.PleaseEnterASpecificationName',
                      defaultMessage: '请输入规格名称',
                    }),
                  },

                  NAME_RULE,
                ]
          }
        >
          {isEdit ? unitSpec?.name : <MyInput />}
        </FormItem>
        <FormItem
          name="maxCpuCoreCount"
          initialValue={unitSpec?.maxCpuCoreCount || undefined}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.src.component.AddUnitSpecModal.PleaseEnterCpuCore',
                defaultMessage: '请输入CPU（核）',
              }),
            },

            {
              validator: validateCPU,
            },
          ]}
          label={formatMessage({
            id: 'ocp-express.src.component.AddUnitSpecModal.CpuCore',
            defaultMessage: 'CPU（核）',
          })}
          extra={formatMessage({
            id: 'ocp-express.src.component.AddUnitSpecModal.TheOptionalCpuRangeIs',
            defaultMessage: 'CPU 的可选范围为 >= 0.5 核，请输入 0.5 的整数倍。',
          })}
        >
          <InputNumber min={0.5} step={0.5} style={{ width: 200 }} />
        </FormItem>
        <FormItem
          name="maxMemorySize"
          initialValue={unitSpec?.maxMemorySize || undefined}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-express.src.component.AddUnitSpecModal.EnterMemory',
                defaultMessage: '请输入内存',
              }),
            },
          ]}
          label={formatMessage({
            id: 'ocp-express.src.component.AddUnitSpecModal.MemoryGb',
            defaultMessage: '内存（GB）',
          })}
          extra={formatMessage({
            id: 'ocp-express.src.component.AddUnitSpecModal.TheOptionalMemoryRangeIs',
            defaultMessage: '内存的可选范围为 >= 1 GB，请输入 1 的整数倍。',
          })}
        >
          <InputNumber min={1} step={1} precision={0} style={{ width: 200 }} />
        </FormItem>
      </Form>
    </Modal>
  );
};

function mapStateToProps({ loading }) {
  return {
    loading: loading.effects['tenant/modifyUnitSpec'] || loading.effects['tenant/addUnitSpec'],
  };
}

export default connect(mapStateToProps)(AddUnitSpecModal);
