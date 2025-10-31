import { Tooltip } from '@oceanbase/design';
import type { ReactNode } from 'react';
import React, { useEffect, useRef, useState } from 'react';
import type { TooltipPlacement } from 'antd/es/tooltip';

type EllipsisTooltipProps = {
  title: ReactNode | (() => ReactNode);
  children: ReactNode;
  placement?: TooltipPlacement;
};

/**
 * EllipsisTooltip 组件
 * @param props EllipsisTooltipProps类型的props对象
 * @returns React元素
 */
const EllipsisTooltip: React.FC<EllipsisTooltipProps> = ({
  title,
  children,
  placement = 'top',
}) => {
  // 创建一个ref对象，用于获取文本节点的DOM元素
  const textRef = useRef<HTMLDivElement>(null);
  // 创建一个state对象，用于保存文本是否溢出的状态
  const [isOverflowed, setIsOverflowed] = useState(false);

  // 使用effect钩子函数监听文本节点的溢出状态
  useEffect(() => {
    if (textRef.current) {
      const isOverflowing = textRef.current.scrollWidth > textRef.current.clientWidth;
      setIsOverflowed(isOverflowing);
    }
  }, [textRef.current]);

  // 返回一个Tooltip和一个div元素，div元素中包含传入的children节点
  return (
    <Tooltip title={isOverflowed ? title : ''} placement={placement}>
      <div
        ref={textRef}
        style={{
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
          width: '100%',
        }}
      >
        {children}
      </div>
    </Tooltip>
  );
};

export default EllipsisTooltip;
