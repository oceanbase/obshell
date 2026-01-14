import React from 'react';
import { Progress, Typography } from '@oceanbase/design';
import type { ProgressProps } from '@oceanbase/design/es/progress';
import { useTheme } from '@oceanbase/charts';
import styles from './index.less';
export interface MyProgressProps extends ProgressProps {
  /* 外层容器的类型 */
  wrapperClassName?: string;
  /* 前缀 */
  prefix?: React.ReactNode;
  prefixWidth?: number;
  prefixTooltip?: boolean;
  /* 后缀 */
  affix?: React.ReactNode;
  affixWidth?: number;
}

const MyProgress: React.FC<MyProgressProps> = ({
  wrapperClassName,
  strokeColor,
  strokeLinecap = 'square',
  prefix,
  prefixWidth = 40,
  prefixTooltip = true,
  affix,
  affixWidth = 40,
  ...restProps
}: MyProgressProps) => {
  const theme = useTheme();
  return (
    <span className={`${styles.progress} ${wrapperClassName}`}>
      {prefix && (
        <Typography.Text
          className={styles.prefix}
          style={{ width: prefixWidth }}
          ellipsis={{ tooltip: prefixTooltip }}
        >
          {prefix}
        </Typography.Text>
      )}

      <span className={styles.wrapper}>
        <Progress
          strokeColor={strokeColor || theme.defaultColor}
          strokeLinecap={strokeLinecap}
          {...restProps}
        />
      </span>
      {affix && (
        <span className={styles.affix} style={{ width: affixWidth }}>
          {affix}
        </span>
      )}
    </span>
  );
};

export default MyProgress;
