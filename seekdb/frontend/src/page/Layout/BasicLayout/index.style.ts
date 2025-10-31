import { createStyles } from 'antd-style';

const useStyles = createStyles(() => {
  return {
    desktopLayoutMac: {
      '& .ant-pro-basic-layout-header': {
        padding: '10px 24px 10px 100px',
      },
    },
    desktopLayoutWindows: {
      '& .ant-pro-basic-layout-header': {
        padding: '10px 140px 10px 24px',
      },
    },
    headerExtra: {
      '& .ob-layout-header .ob-layout-header-extra > .ob-layout-header-extra-item': {
        marginRight: '24px',
      },
    },
  };
});
export default useStyles;
