import { formatMessage } from '@/util/intl';
import { ALERT_STATE_MAP, SHILED_STATUS_MAP } from '@/constant/alert';
import { Alert } from '@/typing/env/alert';
import { flatten, uniq } from 'lodash';

export interface OptionsType {
  value: string;
  label: string;
}

export const getInstancesFromRes = (resInstances: API.OBInstance[]): Alert.InstancesType => {
  const getInstanceValues = (type: Alert.InstancesKey) => {
    return flatten(
      resInstances.filter(instance => instance.type === type).map(item => item[type]!)
    );
  };
  const res: Alert.InstancesType = {
    obcluster: getInstanceValues('obcluster'),
    type: 'obcluster',
  };
  const types = resInstances.map(item => item.type);
  if (types.includes('observer')) {
    res.type = 'observer';
    res.observer = getInstanceValues('observer');
    res.obcluster = [resInstances[0].obcluster!];
  } else if (types.includes('obtenant')) {
    res.type = 'obtenant';
    res.obtenant = getInstanceValues('obtenant');
    res.obcluster = [resInstances[0].obcluster!];
  }
  return res;
};

/**
 * @description
 * Sort alarms by status and time
 */
const sortAlarm = (
  alarms: AlertAlert[] | SilenceSilencerResponse[],
  map: { [T: string]: { text: string; color: string; weight: number } }
) => {
  if (!alarms || !alarms.length) return [];
  const types = uniq(alarms.map(item => item.status.state));
  types.sort((pre, cur) => map[cur].weight - map[pre].weight);
  let res: AlertAlert[] | SilenceSilencerResponse[] = [];
  for (const type of types) {
    const events = alarms.filter(event => event.status.state === type);
    events.sort((preEvent, curEvent) => curEvent.startsAt - preEvent.startsAt);
    res = [...res, ...events];
  }
  return res;
};

export const sortEvents = (alertEvents: AlertAlert[]) => {
  return sortAlarm(alertEvents, ALERT_STATE_MAP);
};

export const sortAlarmShielding = (listSilencers: SilenceSilencerResponse[]) => {
  return sortAlarm(listSilencers, SHILED_STATUS_MAP);
};

/**
 * @description Each item in LabelInput can be empty, but cannot be incomplete.
 */
export const validateLabelValues = (labelValues: AlarmMatcher[] | CommonKVPair[]) => {
  for (const labelValue of labelValues) {
    const tempArr = Object.keys(labelValue).map(key =>
      Boolean(labelValue[key as keyof (AlarmMatcher | CommonKVPair)])
    );
    const stautsArr = tempArr.filter(item => item === true);
    if (stautsArr.length === 1) return false;
    if (stautsArr.length === 2 && tempArr.length > 2 && (labelValue as AlarmMatcher).isRegex) {
      return false;
    }
  }
  return true;
};

/**
 *
 * @param duration The unit is seconds
 */
export const formatDuration = (duration: number) => {
  if (Math.floor(duration / 60) > 180) {
    return `${Math.floor(duration / 3600)}小时`;
  }
  if (duration > 600) {
    return `${Math.floor(duration / 60)}分钟`;
  }
  return formatMessage(
    {
      id: 'OBShell.page.Alert.helper.DurationSeconds',
      defaultMessage: '{duration}秒',
    },
    { duration: duration }
  );
};

export const isListSame = (pre: string[], cur: string[]) => {
  if (pre.length !== cur.length) return false;
  for (const tenant of pre) {
    if (!cur.includes(tenant)) return false;
  }
  return true;
};
