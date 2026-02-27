import { useSelector } from '@umijs/max';
import { useMemo } from 'react';

const useUiMode = () => {
  const { uiMode } = useSelector((state: DefaultRootState) => state.global);
  const isDesktopMode = useMemo(() => uiMode === 'isDesktop', [uiMode]);

  return {
    uiMode,
    isDesktopMode,
  };
};

export default useUiMode;
