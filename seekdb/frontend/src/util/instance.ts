import { formatMessage } from '@/util/intl';
export const getInstanceAvailableFlag = (status: string) => {
  return ['AVAILABLE', 'SWITCHING_OVER', 'ACTIVATING'].includes(status);
};

export const INSTANCE_TYPE_LIST = [
  {
    value: 'PRIMARY',
    signLabel: formatMessage({ id: 'seekdb.src.util.instance.Main', defaultMessage: '主' }),
    backgroundColor: '#3B55A3',
  },

  {
    value: 'STANDBY',
    signLabel: formatMessage({ id: 'seekdb.src.util.instance.Prepare', defaultMessage: '备' }),
    backgroundColor: '#a3a3a3',
  },
];
