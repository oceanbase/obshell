import { createStyles } from 'antd-style';

const useStyles = createStyles(() => {
  return {
    item: {
      display: 'inline-flex',
      alignItems: 'center',
    },
    prefix: {
      marginRight: '8px',
    },
    anticon: {
      verticalAlign: 'middle',
    },
    affix: {
      marginLeft: '8px',
      lineHeight: '10px',
    },
    pointable: {
      cursor: 'pointer',
    },
  };
});
export default useStyles;
