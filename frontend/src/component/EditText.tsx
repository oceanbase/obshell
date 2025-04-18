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
import { Input, Row, Col } from '@oceanbase/design';
import { EditOutlined, CheckOutlined, CloseOutlined } from '@oceanbase/icons';

export interface EditTextProps {
  address: string;
  onOk: (address: string) => void;
  style?: React.CSSProperties;
}

const EditText: React.FC<EditTextProps> = ({ address: initAddress, onOk, ...restProps }) => {
  const [edit, setEdit] = useState<boolean>(false);
  const [address, setAddress] = useState<string>(initAddress as string);
  const [oldAddress, setOldAddress] = useState<string>(initAddress as string);
  // 初始化默认值
  useEffect(() => {
    if (initAddress) {
      setAddress(initAddress);
      setOldAddress(initAddress);
    }
  }, [initAddress]);

  return (
    <Row gutter={[8, 0]} style={{ width: '70%' }}>
      {edit ? (
        <>
          <Col span={18}>
            <Input
              {...restProps}
              value={address}
              onChange={({ target: { value } }) => {
                setAddress(value);
              }}
            />
          </Col>
          <Col style={{ lineHeight: '32px' }}>
            <CloseOutlined
              onClick={() => {
                setAddress(oldAddress);
                setEdit(false);
              }}
            />
          </Col>
          <Col style={{ lineHeight: '32px' }}>
            <CheckOutlined
              onClick={() => {
                if (onOk) {
                  onOk(address);
                }
              }}
            />
          </Col>
        </>
      ) : (
        <>
          <Col>{address}</Col>
          <Col>
            <a>
              <EditOutlined
                onClick={() => {
                  setEdit(true);
                }}
              />
            </a>
          </Col>
        </>
      )}
    </Row>
  );
};

export default EditText;
