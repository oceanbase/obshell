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

import { theme } from '@oceanbase/design';
import { useMouse, useSize } from 'ahooks';
import React, { useRef, useState } from 'react';
import ReactStickyMouseTooltip from 'react-sticky-mouse-tooltip';

interface MouseTooltipProps {
  children: React.ReactElement;
  overlay: React.ReactNode;
  /* 外部传入的 visible 并不完全控制显示和隐藏，只是作为是否显示的前提条件 */
  visible?: boolean;
  offsetX?: number;
  offsetY?: number;
  style?: React.CSSProperties;
}

const MouseTooltip: React.FC<MouseTooltipProps> = ({
  children,
  overlay,
  visible: outerVisible = true,
  style,
  ...restProps
}) => {
  const [visible, setVisible] = useState(false);

  const { token } = theme.useToken();

  // 获取鼠标位置
  const mouse = useMouse();
  const ref = useRef<HTMLDivElement>(null);
  const size = useSize(ref);
  // tooltip 宽高，由于 ref 是设置在内容区上的，因此还需要加上外部的 padding
  const tooltipWidth = (size?.width || 0) + 24;
  const tooltipHeight = (size?.height || 0) + 16;
  // 页面宽高
  const pageWidth = document.body.scrollWidth || 0;
  const pageHeight = document.body.scrollHeight || 0;

  // 判断 tooltip 尺寸是否已经计算完成，避免首次渲染时尺寸为 0 导致位置判断错误
  const isSizeReady = size && size.width > 0;

  // 判断右侧是否有足够空间，如果没有则显示在鼠标左侧
  const rightSpace = pageWidth - mouse.pageX;
  const hasEnoughRightSpace = rightSpace > tooltipWidth + 24;

  // 计算横向偏移量
  let offsetX = 15;
  if (!hasEnoughRightSpace) {
    // 显示在左侧，需要负偏移
    offsetX = -(tooltipWidth + 15);
  }

  // 避免纵向超出浏览器，需要计算出纵向偏移量，16 为多留出的间隙
  const offsetHeight =
    mouse.pageY + tooltipHeight > pageHeight ? mouse.pageY + tooltipHeight - pageHeight + 16 : 0;

  const realChildren = React.cloneElement(children, {
    onMouseEnter: () => {
      if (outerVisible) {
        setVisible(true);
      }
    },
    onMouseLeave: () => {
      setVisible(false);
    },
  });

  return (
    <span>
      {realChildren}
      <ReactStickyMouseTooltip
        visible={visible}
        offsetX={offsetX}
        offsetY={10 - offsetHeight}
        style={{
          // 需要大于 Popover 的 1030 index 值，否则会被遮挡
          zIndex: 1100,
          boxShadow: token.boxShadowSecondary,
          padding: '16px 24px',
          borderRadius: token.borderRadius,
          background: token.colorBgContainer,
          maxWidth: 300,
          // 在尺寸未准备好时隐藏，避免位置跳动
          opacity: isSizeReady ? 1 : 0,
          transition: 'none',
          ...style,
        }}
        {...restProps}
      >
        <div ref={ref}>{overlay}</div>
      </ReactStickyMouseTooltip>
    </span>
  );
};

export default MouseTooltip;
