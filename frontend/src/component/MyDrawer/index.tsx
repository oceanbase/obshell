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
import type { MouseEvent } from 'react';
import React from 'react';
import { Drawer } from '@oceanbase/design';
import type { DrawerProps } from '@oceanbase/design/es/drawer';

export type EventType = MouseEvent<HTMLElement, MouseEvent<Element, MouseEvent>>;

export interface MyDrawerProps extends DrawerProps {
  onOk?: (e: EventType) => void;
  onCancel?: (e: EventType) => void;
}

const MyDrawer: React.FC<MyDrawerProps> = ({
  children,
  onClose,
  onOk,
  onCancel,
  footer,
  ...restProps
}) => {

  return (
    <Drawer
      destroyOnClose={true}
      cancelText={formatMessage({ id: 'ocp-v2.component.MyDrawer.Cancel', defaultMessage: '取消' })}
      okText={formatMessage({ id: 'ocp-v2.component.MyDrawer.Determine', defaultMessage: '确定' })}
      onOk={onOk}
      onClose={onClose || onCancel}
      onCancel={onClose || onCancel}
      footer={footer}
      {...restProps}
    >
      {children}
    </Drawer>
  );
};

export default MyDrawer;
