import { useMemo } from 'react';
import { useSelector } from 'umi';

const useUiMode = () => {
  const { uiMode } = useSelector((state: DefaultRootState) => state.global);
  const isDesktopMode = useMemo(() => uiMode === 'isDesktop', [uiMode]);

  return {
    uiMode,
    isDesktopMode,
  };
};

export default useUiMode;
