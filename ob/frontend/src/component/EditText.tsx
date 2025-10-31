import { formatMessage } from '@/util/intl';
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

import React, { useState, useEffect, useRef } from 'react';
import { Input, Row, Col, Popconfirm } from '@oceanbase/design';
import { EditOutlined, CheckOutlined, CloseOutlined } from '@oceanbase/icons';
import type { InputProps } from '@oceanbase/design/es/input';

export interface EditTextProps extends Omit<InputProps, 'onChange' | 'value'> {
  /** 兼容旧用法：作为非受控模式的初始值使用 */
  address?: string;
  tipTitle?: string;
  /** 确认时回调，入参为确认后的值 */
  onOk?: (address: string) => void;
  /** 受控值，配合表单使用 */
  value?: string;
  /** 受控变更回调（在点击确认时触发） */
  onChange?: (nextValue: string) => void;
  /** 自定义确认/取消文案 */
  confirmText?: React.ReactNode;
  cancelText?: React.ReactNode;
}

const EditText: React.FC<EditTextProps> = ({
  value,
  onChange,
  address: initAddress,
  tipTitle,
  onOk,
  confirmText = formatMessage({
    id: 'OBShell.src.component.EditText.Change',
    defaultMessage: '变更',
  }),
  cancelText = formatMessage({
    id: 'OBShell.src.component.EditText.Cancel',
    defaultMessage: '取消',
  }),
  ...restProps
}) => {
  const [innerValue, setInnerValue] = useState<string>(initAddress || '');
  const [draftValue, setDraftValue] = useState<string>('');

  // 用于表单控制的隐藏 input ref
  const hiddenInputRef = useRef<HTMLInputElement>(null);

  // 确定当前显示的值
  const currentValue = value !== undefined ? value : innerValue;

  // 判断是否应该处于编辑态：没有初始值或手动进入编辑态
  const [edit, setEdit] = useState<boolean>(!currentValue);

  // 初始化草稿值
  useEffect(() => {
    setDraftValue(currentValue);
  }, [currentValue]);

  // 处理确认编辑
  const handleConfirm = () => {
    const newValue = draftValue;

    // 触发表单的 onChange
    if (onChange) {
      onChange(newValue);
    } else {
      // 非受控模式，更新内部状态
      setInnerValue(newValue);
    }

    // 触发 onOk 回调
    onOk?.(newValue);
    setEdit(false);
  };

  // 处理取消编辑
  const handleCancel = () => {
    setDraftValue(currentValue);
    setEdit(false);
  };

  // 处理开始编辑
  const handleStartEdit = () => {
    setDraftValue(currentValue);
    setEdit(true);
  };

  // 渲染确认按钮
  const renderConfirmButton = () => {
    if (tipTitle) {
      return (
        <Popconfirm
          title={tipTitle}
          onConfirm={handleConfirm}
          okText={confirmText}
          cancelText={cancelText}
          styles={{
            body: {
              width: 272,
            },
          }}
        >
          <CheckOutlined style={{ color: '#0ac185', cursor: 'pointer' }} />
        </Popconfirm>
      );
    } else {
      return (
        <CheckOutlined style={{ color: '#0ac185', cursor: 'pointer' }} onClick={handleConfirm} />
      );
    }
  };

  return (
    <div style={{ width: '100%' }}>
      {/* 隐藏的表单控件，用于 Form 获取值 */}
      <input
        ref={hiddenInputRef}
        type="hidden"
        value={currentValue}
        onChange={e => {
          // 这个 onChange 不会被调用，但 Form 需要这个控件来获取值
        }}
      />

      <Row gutter={[8, 0]} style={{ width: '100%' }}>
        {edit ? (
          <>
            <Col span={20}>
              <Input
                {...restProps}
                value={draftValue}
                onChange={e => {
                  setDraftValue(e.target.value);
                }}
                autoFocus
              />
            </Col>

            <Col style={{ lineHeight: '32px' }}>{renderConfirmButton()}</Col>
            <Col style={{ lineHeight: '32px' }}>
              <CloseOutlined
                style={{ color: '#f93939', cursor: 'pointer' }}
                onClick={handleCancel}
              />
            </Col>
          </>
        ) : (
          <>
            <Col>
              <div>
                file://{currentValue || <span style={{ color: '#bfbfbf' }}>{initAddress}</span>}
              </div>
            </Col>
            <Col>
              <EditOutlined
                style={{ cursor: 'pointer', color: '#006aff' }}
                onClick={handleStartEdit}
              />
            </Col>
          </>
        )}
      </Row>
    </div>
  );
};

export default EditText;
