import { createStyles } from 'antd-style';

const useStyles = createStyles(() => {
  return {
    hideHeaderLayout: {
      '& .ant-pro-basic-layout-header': {
        display: 'none !important',
      },
      '& .ant-pro-basic-layout-content-layout': {
        marginTop: '0 !important',
      },
      '& .ant-pro-page-container': {
        minHeight: '100vh !important',
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
