import { formatMessage } from '@/util/intl';
import { SelectProps } from '@oceanbase/design';
import { RuleObject } from 'antd/es/form';
import { DefaultOptionType } from 'antd/es/select';

export const ALERT_STATE_MAP = {
  active: {
    text: formatMessage({ id: 'OBShell.src.constant.alert.Active', defaultMessage: '活跃' }),
    color: 'green',
    weight: 0,
  },
  unprocessed: {
    text: formatMessage({ id: 'OBShell.src.constant.alert.NotTreated', defaultMessage: '未处理' }),
    color: 'default',
    weight: 1,
  },
  suppressed: {
    text: formatMessage({ id: 'OBShell.src.constant.alert.Inhibit', defaultMessage: '抑制' }),
    color: 'red',
    weight: 2,
  },
};

export const SEVERITY_MAP = {
  critical: {
    color: 'red',
    label: formatMessage({ id: 'OBShell.src.constant.alert.Serious', defaultMessage: '严重' }),
    weight: 3,
  },
  warning: {
    color: 'gold',
    label: formatMessage({ id: 'OBShell.src.constant.alert.Warning', defaultMessage: '警告' }),
    weight: 2,
  },
  caution: {
    color: 'blue',
    label: formatMessage({ id: 'OBShell.src.constant.alert.Attention', defaultMessage: '注意' }),
    weight: 1,
  },
  info: {
    color: 'green',
    label: formatMessage({ id: 'OBShell.src.constant.alert.Reminder', defaultMessage: '提醒' }),
    weight: 0,
  },
};

export const LEVER_OPTIONS_ALARM: SelectProps['options'] = [
  {
    label: formatMessage({ id: 'OBShell.src.constant.alert.Serious', defaultMessage: '严重' }),
    value: 'critical',
  },
  {
    label: formatMessage({ id: 'OBShell.src.constant.alert.Warning', defaultMessage: '警告' }),
    value: 'warning',
  },
  {
    label: formatMessage({ id: 'OBShell.src.constant.alert.Attention', defaultMessage: '注意' }),
    value: 'caution',
  },
  {
    label: formatMessage({ id: 'OBShell.src.constant.alert.Reminder', defaultMessage: '提醒' }),
    value: 'info',
  },
];

export const OBJECT_OPTIONS_ALARM: DefaultOptionType[] = [
  {
    label: 'obcluster',
    value: 'obcluster',
  },
  {
    label: 'obtenant',
    value: 'obtenant',
  },
  {
    label: 'observer',
    value: 'observer',
  },
];

export const SHILED_STATUS_MAP = {
  expired: {
    text: formatMessage({ id: 'OBShell.src.constant.alert.Expired', defaultMessage: '过期' }),
    color: 'gold',
    weight: 1,
  },
  pending: {
    text: formatMessage({ id: 'OBShell.src.constant.alert.NotInForce', defaultMessage: '未生效' }),
    color: 'default',
    weight: 0,
  },
  active: {
    text: formatMessage({ id: 'OBShell.src.constant.alert.Active', defaultMessage: '活跃' }),
    color: 'green',
    weight: 2,
  },
};

export const LABELNAME_REG = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

export const VALIDATE_DEBOUNCE = 1000;

export const LABEL_NAME_RULE = {
  validator: (_: RuleObject, value: any[]) => {
    if (
      value.some(item => (item?.name || item?.key) && !LABELNAME_REG.test(item?.name || item?.key))
    ) {
      return Promise.reject(
        formatMessage({
          id: 'OBShell.src.constant.alert.TheLabelNameMustStart',
          defaultMessage: '标签名需满足以字母或下划线开头，包含字母，数字，下划线',
        })
      );
    }
    return Promise.resolve();
  },
};
