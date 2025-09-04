import { createStyles } from 'antd-style';

export const useStyles = createStyles(({ token, css }) => ({
  container: css`
    margin-bottom: ${token.marginLG}px;
  `,

  zoneRow: css`
    display: flex;
    margin-bottom: ${token.margin}px;
  `,

  checkboxItem: css`
    margin-bottom: ${token.marginSM}px;
    width: 24px;
  `,

  formRow: css`
    flex: 1;
  `,

  formItem: css`
    margin-bottom: ${token.marginSM}px;
  `,

  alert: css`
    margin-top: -8px;
    margin-bottom: ${token.margin}px;
  `,

  readonlyInput: css`
    .ant-input-number-input {
      background-color: ${token.colorFillAlter};
      cursor: not-allowed;
    }
  `,
}));
