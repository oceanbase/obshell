import { theme } from '@oceanbase/design';
import React from 'react';

export interface BackGroundContainerProps {
  children: React.ReactNode;
  style?: React.CSSProperties;
}
/**
 * 背景容器
 * @param children 子组件
 * @param style 样式 默认值: { backgroundColor: token.colorFillQuaternary, padding: 16, borderRadius: token.borderRadius }
 * @param restProps 其他属性，会透传到 div 上
 * @returns 背景容器
 */
const BackGroundContainer: React.FC<BackGroundContainerProps> = ({ children, ...props }) => {
  const { token } = theme.useToken();
  const { style, ...restProps } = props;

  const defaultStyles = {
    backgroundColor: token.colorFillQuaternary,
    padding: 16,
    borderRadius: token.borderRadius,
  };

  return (
    <div style={{ ...defaultStyles, ...style }} {...restProps}>
      {children}
    </div>
  );
};

export default BackGroundContainer;
