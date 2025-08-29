import { formatMessage } from '@/util/intl';
export interface TimezoneOption {
  value: string;
  label: string;
  gmt: string;
  description: string;
}

export const TIMEZONE_LIST: TimezoneOption[] = [
  {
    value: '-12:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.WestOfTheDateLine',
      defaultMessage: '日界线西',
    }),
    gmt: 'GTM-12:00',
    description: 'International Date Line West',
  },
  {
    value: '-11:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.Samoa',
      defaultMessage: '萨摩亚群岛',
    }),
    gmt: 'GTM-11:00',
    description: 'Samoa Islands',
  },
  {
    value: '-10:00',
    label: formatMessage({ id: 'OBShell.src.constant.timezone.Hawaii', defaultMessage: '夏威夷' }),
    gmt: 'GTM-10:00',
    description: 'Hawaii',
  },
  {
    value: '-09:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.Alaska',
      defaultMessage: '阿拉斯加',
    }),
    gmt: 'GTM-09:00',
    description: 'Alaska',
  },
  {
    value: '-08:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.PacificTime',
      defaultMessage: '太平洋时间',
    }),
    gmt: 'GTM-08:00',
    description: 'Pacific Time (US and Canada)',
  },
  {
    value: '-07:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.Arizona',
      defaultMessage: '亚利桑那',
    }),
    gmt: 'GTM-07:00',
    description: 'Arizona',
  },
  {
    value: '-06:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.CentralTime',
      defaultMessage: '中部时间',
    }),
    gmt: 'GTM-06:00',
    description: 'Central Time (US and Canada)',
  },
  {
    value: '-05:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.EasternTime',
      defaultMessage: '东部时间',
    }),
    gmt: 'GTM-05:00',
    description: 'Eastern Time (US and Canada)',
  },
  {
    value: '-04:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.AtlanticTime',
      defaultMessage: '大西洋时间',
    }),
    gmt: 'GTM-04:00',
    description: 'Atlantic Time (US and Canada)',
  },
  {
    value: '-03:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.Brasilia',
      defaultMessage: '巴西利亚',
    }),
    gmt: 'GTM-03:00',
    description: 'Brasilia',
  },
  {
    value: '-02:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.MidAtlantic',
      defaultMessage: '中大西洋',
    }),
    gmt: 'GTM-02:00',
    description: 'Mid-Atlantic',
  },
  {
    value: '-01:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.Azores',
      defaultMessage: '亚速尔群岛',
    }),
    gmt: 'GTM-01:00',
    description: 'Azores',
  },
  {
    value: '+00:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.GreenwichMeanTime',
      defaultMessage: '格林威治标准时间',
    }),
    gmt: 'GTM',
    description: 'Greenwich Mean Time',
  },
  {
    value: '+01:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.Sarajevo',
      defaultMessage: '萨拉热窝',
    }),
    gmt: 'GTM+01:00',
    description: 'Sarajevo',
  },
  {
    value: '+02:00',
    label: formatMessage({ id: 'OBShell.src.constant.timezone.Cairo', defaultMessage: '开罗' }),
    gmt: 'GTM+02:00',
    description: 'Cairo',
  },
  {
    value: '+03:00',
    label: formatMessage({ id: 'OBShell.src.constant.timezone.Moscow', defaultMessage: '莫斯科' }),
    gmt: 'GTM+03:00',
    description: 'Moscow',
  },
  {
    value: '+04:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.AbuDhabi',
      defaultMessage: '阿布扎比',
    }),
    gmt: 'GTM+04:00',
    description: 'Abu Dhabi',
  },
  {
    value: '+05:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.Islamabad',
      defaultMessage: '伊斯兰堡',
    }),
    gmt: 'GTM+05:00',
    description: 'Islamabad',
  },
  {
    value: '+06:00',
    label: formatMessage({ id: 'OBShell.src.constant.timezone.Dhaka', defaultMessage: '达卡' }),
    gmt: 'GTM+06:00',
    description: 'Dhaka',
  },
  {
    value: '+07:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.BangkokHanoi',
      defaultMessage: '曼谷, 河内',
    }),
    gmt: 'GTM+07:00',
    description: 'Bangkok, Hanoi',
  },
  {
    value: '+08:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.ChinaStandardTime',
      defaultMessage: '中国标准时间',
    }),
    gmt: 'GTM+08:00',
    description: 'China Standard Time',
  },
  {
    value: '+09:00',
    label: formatMessage({ id: 'OBShell.src.constant.timezone.Seoul', defaultMessage: '首尔' }),
    gmt: 'GTM+09:00',
    description: 'Seoul',
  },
  {
    value: '+10:00',
    label: formatMessage({ id: 'OBShell.src.constant.timezone.Guam', defaultMessage: '关岛' }),
    gmt: 'GTM+10:00',
    description: 'Guam',
  },
  {
    value: '+11:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.SolomonIslands',
      defaultMessage: '所罗门群岛',
    }),
    gmt: 'GTM+11:00',
    description: 'Solomon Islands',
  },
  {
    value: '+12:00',
    label: formatMessage({ id: 'OBShell.src.constant.timezone.Fiji', defaultMessage: '斐济' }),
    gmt: 'GTM+12:00',
    description: 'Fiji',
  },
  {
    value: '+13:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.NukuAlfa',
      defaultMessage: '努库阿勒法',
    }),
    gmt: 'GTM+13:00',
    description: "Nuku'alofa",
  },
  {
    value: '+14:00',
    label: formatMessage({
      id: 'OBShell.src.constant.timezone.Kiribati',
      defaultMessage: '基里巴斯',
    }),
    gmt: 'GTM+14:00',
    description: 'Kiribati',
  },
];

// 根据值获取时区选项
export const getTimezoneByValue = (value: string): TimezoneOption | undefined => {
  return TIMEZONE_LIST.find(item => item.value === value);
};

// 根据GMT偏移获取时区选项
export const getTimezoneByGMT = (gmt: string): TimezoneOption | undefined => {
  return TIMEZONE_LIST.find(item => item.gmt === gmt);
};

// 获取显示文本
export const getTimezoneDisplayText = (value: string): string => {
  const timezone = getTimezoneByValue(value);
  if (!timezone) return '';
  return `${timezone.gmt} ${timezone.label}`;
};

// 根据租户模式过滤时区列表
export const filterTimezonesByMode = (
  timezoneList: TimezoneOption[],
  mode: 'MYSQL' | 'ORACLE'
): TimezoneOption[] => {
  return timezoneList.filter(item => {
    const timezoneValue = item.value;
    const hours = parseInt(timezoneValue.substring(1, 3));

    if (mode === 'MYSQL') {
      // MySQL模式：时区区间为 -12:59 到 +13:00
      return timezoneValue.startsWith('-') ? hours <= 12 : hours <= 13;
    } else {
      // Oracle模式：时区区间为 -15:59 到 +15:00
      return timezoneValue.startsWith('-') ? hours <= 12 : hours <= 14;
    }
  });
};
