import React, { useEffect } from 'react';
import { theme } from '@oceanbase/design';
import { useSpring, animated } from 'react-spring';

const RefreshProgress: React.FC<any> = ({ value }) => {
  const { token } = theme.useToken();

  const r = 50;
  const totalLength = 2 * Math.PI * r;
  const [springProps, setSpringProps] = useSpring(() => ({
    value: 0,
  }));

  useEffect(() => {
    setSpringProps({
      value,
    });
  }, [setSpringProps, value]);

  return (
    <svg
      viewBox="0 0 120 120"
      width="1em"
      height="1em"
      className="anticon"
      style={{
        transform: 'rotate(-90deg)',
      }}
    >
      <circle
        cx="60"
        cy="60"
        r={r}
        fill="none"
        stroke={token.colorFillSecondary}
        strokeWidth="20"
      />
      <animated.circle
        cx="60"
        cy="60"
        r={r}
        fill="none"
        stroke={springProps.value.interpolate({
          range: [0, 1],
          output: [token.colorFillSecondary, token.colorInfo],
        })}
        strokeWidth="20"
        strokeDasharray={totalLength}
        strokeDashoffset={springProps.value.interpolate({
          range: [0, 1],
          output: [totalLength, 0],
        })}
      />
    </svg>
  );
};

export default RefreshProgress;
