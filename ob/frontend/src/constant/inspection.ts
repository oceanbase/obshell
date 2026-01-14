import { formatMessage } from '@/util/intl';
export const statusList = [
  {
    label: formatMessage({ id: 'OBShell.src.constant.inspection.Success', defaultMessage: '成功' }),
    color: 'success',
    value: 'SUCCEED',
  },
  {
    label: formatMessage({ id: 'OBShell.src.constant.inspection.Failed', defaultMessage: '失败' }),
    color: 'error',
    value: 'FAILED',
  },
  {
    label: formatMessage({
      id: 'OBShell.src.constant.inspection.InOperation',
      defaultMessage: '运行中',
    }),
    color: 'processing',
    value: 'RUNNING',
  },
  {
    label: formatMessage({ id: 'OBShell.src.constant.inspection.Unknown', defaultMessage: '未知' }),
    color: 'warning',
    value: 'UNKNOWN',
  },
  {
    label: formatMessage({
      id: 'OBShell.src.constant.inspection.Deleted',
      defaultMessage: '已删除',
    }),
    color: 'default',
    value: 'DELETED',
  },
];

export const scenarioList = [
  {
    key: 'basic',
    label: formatMessage({
      id: 'OBShell.src.constant.inspection.BasicInspection',
      defaultMessage: '基础巡检',
    }),
  },
  {
    key: 'performance',
    label: formatMessage({
      id: 'OBShell.src.constant.inspection.PerformanceInspection',
      defaultMessage: '性能巡检',
    }),
  },
];
