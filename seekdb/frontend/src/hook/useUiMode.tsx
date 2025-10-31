import { useMemo } from 'react';
import { DESKTOP_UI_MODE, DESKTOP_WINDOW } from '@/constant';

const useUiMode = () => {
  const uiMode = localStorage.getItem(DESKTOP_UI_MODE);
  const window = localStorage.getItem(DESKTOP_WINDOW);
  const isDesktopMode = useMemo(() => uiMode === 'ob-desktop', [uiMode]);

  return {
    uiMode,
    window,
    isDesktopMode,
  };
};

export default useUiMode;
