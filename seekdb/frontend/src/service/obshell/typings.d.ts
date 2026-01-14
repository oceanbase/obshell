declare namespace API {
  type AgentIdentity = 'MASTER' | 'SINGLE' | 'CLUSTER AGENT' | 'TAKE OVER MASTER' | 'UNIDENTIFIED';

  type AgentInfo = {
    /** unused for agent start */
    ip: string;
    port: number;
  };

  type AgentInfoWithIdentity = {
    identity: AgentIdentity;
    /** unused for agent start */
    ip: string;
    port: number;
  };

  type AgentSecret = {
    /** unused for agent start */
    ip: string;
    port: number;
    public_key: string;
  };

  type AgentStatus = {
    agent?: AgentInfoWithIdentity;
    obState?: number;
    /** service pid */
    pid?: number;
    /** Ports process occupied ports */
    port?: number;
    sqlPort?: number;
    /** timestamp when service started */
    startAt?: number;
    /** service state */
    state?: number;
    underMaintenance?: boolean;
    /** service version */
    version?: string;
  };

  type AgentStatus = {
    architecture?: string;
    homePath?: string;
    identity: AgentIdentity;
    /** unused for agent start */
    ip: string;
    obVersion?: string;
    pid?: number;
    port: number;
    security?: boolean;
    startAt?: number;
    state?: number;
    supportedAuth?: string[];
    systemdUnit?: string;
    /** "seekdb" means the agent is work for seekdb */
    type?: string;
    version?: string;
  };

  type Alert = {
    description?: string;
    ends_at: number;
    fingerprint: string;
    instance: OBInstance;
    labels?: KVPair[];
    rule: string;
    severity: Severity;
    starts_at: number;
    status: Status;
    summary?: string;
    updated_at: number;
  };

  type AlertFilter = {
    end_time?: number;
    instance?: OBInstance;
    instance_type?: OBInstanceType;
    keyword?: string;
    severity?: Severity;
    start_time?: number;
  };

  type AlertmanagerConfig = {
    address?: string;
    auth?: Auth;
  };

  type ApiError = {
    /** Error code v1, deprecated */
    code?: number;
    /** Error code string */
    errCode?: string;
    /** Error message */
    message?: string;
    /** Sub errors */
    subErrors?: any[];
  };

  type Auth = {
    password?: string;
    username?: string;
  };

  type changePasswordParams = {
    /** user name */
    user: string;
  };

  type ChangeUserPasswordParam = {
    password?: string;
  };

  type CharsetInfo = {
    collations?: CollationInfo[];
    description?: string;
    maxlen?: number;
    name?: string;
  };

  type CollationInfo = {
    is_default?: boolean;
    name?: string;
  };

  type CpuInfo = {
    frequency?: string;
    logicalCores?: number;
    modelName?: string;
    physicalCores?: number;
  };

  type CreateDatabaseParam = {
    collation?: string;
    db_name: string;
    read_only?: boolean;
  };

  type CreateUserParam = {
    db_privileges?: DbPrivilegeParam[];
    global_privileges?: string[];
    host_name?: string;
    password: string;
    user_name: string;
  };

  type DagDetailDTO = {
    additional_data?: Record<string, any>;
    dag_id?: number;
    end_time?: string;
    id: string;
    maintenance_key?: string;
    maintenance_type?: number;
    max_stage?: number;
    name?: string;
    nodes?: NodeDetailDTO[];
    operator?: string;
    stage?: number;
    start_time?: string;
    state?: string;
  };

  type dagHandlerParams = {
    /** dag id */
    id: string;
  };

  type DagOperator = {
    additional_data?: Record<string, any>;
    dag_id?: number;
    end_time?: string;
    id: string;
    maintenance_key?: string;
    maintenance_type?: number;
    max_stage?: number;
    name?: string;
    nodes?: NodeDetailDTO[];
    operator: string;
    stage?: number;
    start_time?: string;
    state?: string;
  };

  type Database = {
    charset?: string;
    collation?: string;
    connection_urls?: ObproxyAndConnectionString[];
    create_time?: number;
    db_name?: string;
    read_only?: boolean;
  };

  type DbaObSysVariable = {
    info?: string;
    name?: string;
    value?: string;
  };

  type DbPrivilege = {
    db_name?: string;
    privileges?: string[];
  };

  type DbPrivilegeParam = {
    db_name: string;
    privileges: string[];
  };

  type deleteDatabaseParams = {
    /** database name */
    database: string;
  };

  type DeletePackageParam = {
    /** rpm package architecture */
    architecture: string;
    /** rpm package name */
    name: string;
    /** rpm package release distribution */
    release_distribution: string;
    /** rpm package version */
    version: string;
  };

  type DeleteSilencerParams = {
    /** silencer id */
    id: string;
  };

  type DiskInfo = {
    deviceName?: string;
    mountHash?: string;
    total?: string;
    used?: string;
  };

  type dropUserParams = {
    /** user name */
    user: string;
  };

  type GetAgentUnfinishDagsParams = {
    show_details?: boolean;
  };

  type getAllAgentDagsParams = {
    show_details?: boolean;
  };

  type getAllClusterDagsParams = {
    show_details?: boolean;
  };

  type GetClusterUnfinishDagsParams = {
    show_details?: boolean;
  };

  type getDagDetailParams = {
    /** id */
    id: string;
    show_details?: boolean;
  };

  type getDatabaseParams = {
    /** database name */
    database: string;
  };

  type getFullSubTaskLogsParams = {
    /** id */
    id: string;
  };

  type getNodeDetailParams = {
    /** id */
    id: string;
    show_details?: boolean;
  };

  type getParametersParams = {
    /** filter format */
    filter?: string;
  };

  type GetRuleParams = {
    /** rule name */
    name: string;
  };

  type GetSilencerParams = {
    /** silencer id */
    id: string;
  };

  type getStatsParams = {
    /** user name */
    user: string;
  };

  type getSubTaskDetailParams = {
    /** id */
    id: string;
    show_details?: boolean;
  };

  type GetUnfinishedDagsParams = {
    show_details?: boolean;
  };

  type getUserParams = {
    /** user name */
    user: string;
  };

  type getVariablesParams = {
    /** filter format */
    filter?: string;
  };

  type GitInfo = {
    branch?: string;
    commitId?: string;
    commitTime?: string;
    shortCommitId?: string;
  };

  type GvObParameter = {
    data_type?: string;
    edit_level?: string;
    info?: string;
    name?: string;
    value?: string;
  };

  type HostBasicInfo = {
    hostHash?: string;
    hostType?: string;
  };

  type HostInfo = {
    basic?: HostBasicInfo;
    cpuInfo?: CpuInfo;
    disks?: DiskInfo[];
    memoryInfo?: MemoryInfo;
    os?: OSInfo;
    ulimit?: UlimitInfo;
  };

  type KVPair = {
    key?: string;
    value?: string;
  };

  type ListAllMetricsParams = {
    /** metrics scope */
    scope: 'OBCLUSTER' | 'OBTENANT';
  };

  type lockUserParams = {
    /** user name */
    user: string;
  };

  type Matcher = {
    is_regex?: boolean;
    name?: string;
    value?: string;
  };

  type MemoryInfo = {
    available?: string;
    free?: string;
    total?: string;
  };

  type Metric = {
    labels?: KVPair[];
    name?: string;
  };

  type MetricClass = {
    description: string;
    metric_groups: MetricGroup[];
    name: string;
  };

  type MetricData = {
    metric: Metric;
    values: MetricValue[];
  };

  type MetricGroup = {
    description: string;
    metrics: MetricMeta[];
    name: string;
  };

  type MetricMeta = {
    description: string;
    key: string;
    name: string;
    unit: string;
  };

  type MetricQuery = {
    group_labels?: string[];
    labels?: KVPair[];
    metrics?: string[];
    query_range?: QueryRange;
  };

  type MetricValue = {
    timestamp: number;
    value: number;
  };

  type ModifyDatabaseParam = {
    collation?: string;
    read_only?: boolean;
  };

  type modifyDbPrivilegeParams = {
    /** user name */
    user: string;
  };

  type modifyGlobalPrivilegeParams = {
    /** user name */
    user: string;
  };

  type ModifyUserDbPrivilegeParam = {
    db_privileges: DbPrivilegeParam[];
  };

  type ModifyUserGlobalPrivilegeParam = {
    global_privileges?: string[];
  };

  type ModifyWhitelistParam = {
    whitelist: string;
  };

  type NodeDetailDTO = {
    additional_data?: Record<string, any>;
    end_time?: string;
    id: string;
    name?: string;
    node_id?: number;
    operator?: string;
    start_time?: string;
    state?: string;
    sub_tasks?: TaskDetailDTO[];
  };

  type nodeHandlerParams = {
    /** node id */
    id: string;
  };

  type NodeOperator = {
    additional_data?: Record<string, any>;
    end_time?: string;
    id: string;
    name?: string;
    node_id?: number;
    operator: string;
    start_time?: string;
    state?: string;
    sub_tasks?: TaskDetailDTO[];
  };

  type ObclusterStatisticInfo = {
    clusterId?: string;
    hosts?: HostInfo;
    instances?: ObserverInfo;
    obshellRevision?: string;
    obshellVersion?: string;
    reportTime?: number;
    reporter?: string;
    telemetryEnabled?: boolean;
    telemetryVersion?: number;
  };

  type OBInstance = {
    obcluster?: string;
    observer?: string;
    obtenant?: string;
    /** obzone may exist in labels */
    obzone?: string;
    type: OBInstanceType;
  };

  type OBInstanceType = 'unknown' | 'obcluster' | 'obzone' | 'obtenant' | 'observer';

  type ObproxyAndConnectionString = {
    connection_string?: string;
    obproxy_address?: string;
    obproxy_port?: number;
    type?: string;
  };

  type ObRestartParam = {
    force_pass?: boolean;
    terminate?: boolean;
  };

  type ObserverInfo = {
    cpuCount?: number;
    dataFileSize?: string;
    hostHash?: string;
    logDiskSize?: string;
    memoryLimit?: string;
    revision?: string;
    type?: string;
    version?: string;
  };

  type ObserverInfo = {
    architecture?: string;
    base_dir?: string;
    bin_path?: string;
    cluster_name?: string;
    connection_string?: string;
    cpu_count?: number;
    created_time?: string;
    data_dir?: string;
    data_disk_size?: string;
    database_count?: number;
    inner_status?: string;
    life_time?: string;
    log_dir?: string;
    log_disk_size?: string;
    memory_size?: string;
    obshell_port?: number;
    port?: number;
    redo_dir?: string;
    start_time?: string;
    status?: string;
    systemd_unit?: string;
    user?: string;
    user_count?: number;
    version?: string;
    whitelist?: string;
  };

  type ObStopParam = {
    force_pass?: boolean;
    terminate?: boolean;
  };

  type ObUser = {
    accessible_databases?: string[];
    connection_strings?: ObproxyAndConnectionString[];
    create_time?: string;
    db_privileges?: DbPrivilege[];
    global_privileges?: string[];
    granted_roles?: string[];
    is_locked?: boolean;
    user_name?: string;
  };

  type ObUserSessionStats = {
    active?: number;
    total?: number;
  };

  type ObUserStats = {
    session?: ObUserSessionStats;
  };

  type OcsAgentResponse = {
    /** Data payload when response is successful */
    data?: any;
    /** Request handling time cost (ms) */
    duration?: number;
    /** Error payload when response is failed */
    error?: ApiError;
    /** HTTP status code */
    status?: number;
    /** Whether request successful or not */
    successful?: boolean;
    /** Request handling timestamp (server time) */
    timestamp?: string;
    /** Request trace ID, contained in server logs */
    traceId?: string;
  };

  type OSInfo = {
    os?: string;
    version?: string;
  };

  type PrometheusConfig = {
    address?: string;
    auth?: Auth;
  };

  type QueryRange = {
    end_timestamp?: number;
    start_timestamp?: number;
    step?: number;
  };

  type RuleFilter = {
    instance_type?: OBInstanceType;
    keyword?: string;
    severity?: Severity;
  };

  type RuleHealth = 'unknown' | 'ok' | 'err';

  type RuleResponse = {
    description: string;
    duration: number;
    evaluation_time: number;
    health: RuleHealth;
    instance_type: OBInstanceType;
    keep_firing_for: number;
    labels: KVPair[];
    last_error?: string;
    last_evaluation: number;
    name: string;
    query: string;
    severity: Severity;
    state: RuleState;
    summary: string;
    type?: RuleType;
  };

  type RuleState = 'firing' | 'pending' | 'inactive';

  type RuleType = 'built-in' | 'customized';

  type SetParametersParam = {
    parameters: Record<string, any>;
  };

  type SetVariablesParam = {
    variables: Record<string, any>;
  };

  type Severity = 'critical' | 'major' | 'minor' | 'warning' | 'info';

  type SilencerFilter = {
    instance?: OBInstance;
    instance_type?: OBInstanceType;
    keyword?: string;
  };

  type SilencerParam = {
    comment: string;
    created_by: string;
    ends_at: number;
    id?: string;
    instances: OBInstance[];
    matchers: Matcher[];
    rules: string[];
    starts_at: number;
  };

  type SilencerResponse = {
    comment: string;
    created_by: string;
    ends_at: number;
    id: string;
    instances: OBInstance[];
    matchers: Matcher[];
    rules: string[];
    starts_at: number;
    status: Status;
    updated_at: number;
  };

  type State = 'active' | 'unprocessed' | 'suppressed';

  type State = 'active' | 'expired' | 'pending';

  type Status = {
    inhibited_by: string[];
    silenced_by: string[];
    state: State;
  };

  type Status = {
    state: State;
  };

  type TaskDetailDTO = {
    additional_data?: Record<string, any>;
    end_time?: string;
    execute_agent?: AgentInfo;
    execute_times?: number;
    id: string;
    name?: string;
    operator?: string;
    start_time?: string;
    state?: string;
    task_id?: number;
    task_logs?: string[];
  };

  type TenantCompaction = {
    frozen_scn?: number;
    frozen_time?: string;
    global_broadcast_scn?: number;
    is_error?: string;
    is_suspended?: string;
    last_finish_time?: string;
    last_scn?: number;
    start_time?: string;
    status?: string;
  };

  type UlimitInfo = {
    maxUserProcessesHard?: string;
    maxUserProcessesSoft?: string;
    nofileHard?: string;
    nofileSoft?: string;
  };

  type unlockUserParams = {
    /** user name */
    user: string;
  };

  type updateDatabaseParams = {
    /** database name */
    database: string;
  };

  type UpgradeCheckParam = {
    release: string;
    upgradeDir?: string;
    version: string;
  };

  type UpgradePkgInfo = {
    architecture?: string;
    chunk_count?: number;
    distribution?: string;
    gmt_modify?: string;
    md5?: string;
    name?: string;
    payload_size?: number;
    pkg_id?: number;
    release?: string;
    release_distribution?: string;
    size?: number;
    upgrade_dep_yaml?: string;
    version?: string;
  };
}
