declare namespace Monitor {
  type Label =
    | 'ob_cluster_name'
    | 'ob_cluster_id'
    | 'tenant_name'
    | 'tenant_id'
    | 'svr_ip'
    | 'obzone'
    | 'cluster';

  type LabelType = {
    key: Label;
    value: string;
  };

  type OptionType = {
    label: string;
    value: string | number;
    zone?: string;
  };

  type FilterDataType = {
    tenantList?: OptionType[];
    zoneList?: OptionType[];
    serverList?: OptionType[];
    date?: any;
  };

  type QueryRangeType = {
    end_timestamp: number;
    start_timestamp: number;
    step: number;
  };

  type EventObjectType =
    | 'OBCLUSTER'
    | 'OBTENANT'
    | 'OBBACKUPPOLICY'
    | 'OBPROXY'
    | EventObjectType[];

  type LabelKeys =
    | 'ob_cluster_name'
    | 'ob_cluster_id'
    | 'tenant_name'
    | 'tenant_id'
    | 'svr_ip'
    | 'obzone'
    | 'cluster';

  type MetricsLabels = { key: LabelKeys; value: string }[];
  type MonitorUseTarget = 'OVERVIEW' | 'DETAIL' | 'OVERVIEW_IN_CLUSTER';
  type MonitorUseFor = 'cluster' | 'tenant';
}
