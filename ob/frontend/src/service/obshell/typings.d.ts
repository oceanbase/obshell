declare namespace API {
  type AddObproxyParam = {
    config_url?: string;
    /** Default to 2885. */
    exporter_port?: number;
    home_path: string;
    name?: string;
    obproxy_sys_password?: string;
    parameters?: Record<string, any>;
    proxyro_password?: string;
    /** Default to 2884. */
    rpc_port?: number;
    rs_list?: string;
    /** Default to 2883. */
    sql_port?: number;
  };

  type AgentIdentity =
    | 'MASTER'
    | 'FOLLOWER'
    | 'SINGLE'
    | 'CLUSTER AGENT'
    | 'TAKE OVER MASTER'
    | 'TAKE OVER FOLLOWER'
    | 'SCALING OUT'
    | 'SCALING IN'
    | 'UNIDENTIFIED';

  type AgentInfo = {
    ip: string;
    port: number;
  };

  type AgentInfoWithIdentity = {
    identity: AgentIdentity;
    ip: string;
    port: number;
  };

  type AgentInstance = {
    identity: AgentIdentity;
    ip: string;
    port: number;
    version?: string;
    zone: string;
  };

  type AgentSecret = {
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
    ip: string;
    isObproxyAgent?: boolean;
    obVersion?: string;
    pid?: number;
    port: number;
    security?: boolean;
    startAt?: number;
    state?: number;
    supportedAuth?: string[];
    version?: string;
    zone: string;
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

  type ArchiveLogStatusParam = {
    status?: string;
  };

  type ArchiveLogTask = {
    checkpoint?: string;
    /** seconds */
    delay?: number;
    path?: string;
    round_id?: number;
    start_time?: string;
    status?: string;
    tenant_id?: number;
    tenant_name?: string;
  };

  type Auth = {
    password?: string;
    username?: string;
  };

  type BackupDeletePolicy = {
    policy?: string;
    recovery_window?: string;
  };

  type BackupJob = {
    backup_set_id?: number;
    backup_type?: string;
    comment?: string;
    description?: string;
    end_timestamp?: string;
    job_id?: number;
    path?: string;
    plus_archivelog?: string;
    result?: number;
    start_timestamp?: string;
    status?: string;
    tenant_id?: number;
  };

  type BackupOverview = {
    statuses?: CdbObBackupTask[];
  };

  type BackupParam = {
    encryption?: string;
    mode?: string;
    plus_archive?: boolean;
  };

  type BackupStatusParam = {
    status?: string;
  };

  type BaseResourceStats = {
    cpu_core_assigned?: number;
    cpu_core_assigned_percent?: number;
    cpu_core_total?: number;
    disk_assigned?: string;
    disk_assigned_percent?: number;
    disk_in_bytes_assigned?: number;
    disk_in_bytes_total?: number;
    disk_total?: string;
    memory_assigned?: string;
    memory_assigned_percent?: number;
    memory_in_bytes_assigned?: number;
    memory_in_bytes_total?: number;
    memory_total?: string;
  };

  type cancelRestoreTaskParams = {
    /** Tenant name */
    tenantName: string;
  };

  type CdbObBackupTask = {
    backup_set_id?: number;
    comment?: string;
    data_turn_id?: number;
    encryption_mode?: string;
    end_scn?: number;
    end_timestamp?: string;
    extra_meta_bytes?: number;
    file_count?: number;
    finish_macro_block_count?: number;
    finish_tablet_count?: number;
    incarnation?: number;
    input_bytes?: number;
    job_id?: number;
    macro_block_count?: number;
    meta_turn_id?: number;
    output_bytes?: number;
    output_rate_bytes?: number;
    path?: string;
    result?: number;
    start_scn?: number;
    start_timestamp?: string;
    status?: string;
    tablet_count?: number;
    task_id?: number;
    tenant_id?: number;
    user_ls_start_scn?: number;
  };

  type CdbObSysVariable = {
    info?: string;
    name?: string;
    value?: string;
  };

  type changePasswordParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type ChangeUserPasswordParam = {
    new_password?: string;
    root_password?: string;
  };

  type CharsetInfo = {
    collations?: CollationInfo[];
    description?: string;
    maxlen?: number;
    name?: string;
  };

  type ClusterBackupConfigParam = {
    archive_lag_target?: string;
    backup_base_uri?: string;
    binding?: string;
    delete_policy?: BackupDeletePolicy;
    ha_low_thread_score?: number;
    location?: string;
    log_archive_concurrency?: number;
    piece_switch_interval?: string;
  };

  type ClusterConfig = {
    id?: number;
    name?: string;
    topology?: Record<string, any>;
    version?: string;
  };

  type ClusterInfo = {
    cluster_id?: number;
    cluster_name?: string;
    is_community_edition?: boolean;
    is_standalone?: boolean;
    /** only for standalone cluster */
    license?: ObLicense;
    ob_version?: string;
    stats?: BaseResourceStats;
    status?: string;
    tenant_stats?: TenantResourceStat[];
    tenants?: TenantInfo[];
    zones?: Zone[];
  };

  type ClusterParameter = {
    data_type?: string;
    default_value?: string;
    edit_level?: string;
    info?: string;
    is_single_value?: boolean;
    name?: string;
    ob_parameters?: ObParameterValue[];
    scope?: string;
    section?: string;
    tenant_value?: TenantParameterValue[];
    values?: string[];
  };

  type ClusterScaleInParam = {
    agent_info: AgentInfo;
    /** default to false */
    force_kill?: boolean;
  };

  type ClusterScaleOutParam = {
    agentInfo: AgentInfo;
    obConfigs: Record<string, any>;
    targetAgentPassword?: string;
    zone: string;
  };

  type ClusterUnitConfigLimit = {
    min_cpu?: number;
    min_memory?: number;
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
    root_password?: string;
  };

  type createDatabaseParams = {
    /** tenant name */
    name: string;
  };

  type CreateResourceUnitConfigParams = {
    /** log disk size, greater than or equal to '2G' */
    log_disk_size?: string;
    /** max cpu cores, greater than 0 */
    max_cpu: number;
    /** max iops, greater than or equal to 1024 */
    max_iops?: number;
    /** memory size, greater than or equal to '1G' */
    memory_size: string;
    /** min cpu cores, smaller than or equal 'max_cpu_cores' */
    min_cpu?: number;
    /** min iops, smaller than or equal to 'max_iops' */
    min_iops?: number;
    /** unit config name */
    name: string;
  };

  type CreateRoleParam = {
    global_privileges?: string[];
    role_name: string;
    roles?: string[];
    root_password?: string;
  };

  type createRoleParams = {
    /** tenant name */
    name: string;
  };

  type CreateTenantParam = {
    charset?: string;
    collation?: string;
    /** Messages. */
    comment?: string;
    /** whether to import script. */
    import_script?: boolean;
    /** Tenant mode, "MYSQL"(default) or "ORACLE". */
    mode?: string;
    /** Tenant name. */
    name: string;
    /** Tenant parameters. */
    parameters?: Record<string, any>;
    /** Tenant primary_zone. */
    primary_zone?: string;
    /** Default to false. */
    read_only?: boolean;
    /** Root password. */
    root_password?: string;
    /** Tenant scenario. */
    scenario?: string;
    /** Tenant global variables. */
    variables?: Record<string, any>;
    /** Tenant whitelist. */
    whitelist?: string;
    /** Tenant zone list with unit config. */
    zone_list: ZoneParam[];
  };

  type CreateUserParam = {
    db_privileges?: DbPrivilegeParam[];
    global_privileges?: string[];
    host_name?: string;
    password: string;
    roles?: string[];
    root_password?: string;
    user_name: string;
  };

  type createUserParams = {
    /** tenant name */
    name: string;
  };

  type CustomPage = {
    number?: number;
    size?: number;
    total_elements?: number;
    total_pages?: number;
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

  type DbaObjectBo = {
    full_name?: string;
    name?: string;
    owner?: string;
    type?: string;
  };

  type DbaObResourcePool = {
    createTime?: string;
    id?: number;
    modifyTime?: string;
    name?: string;
    replica_type?: string;
    tenant_id?: number;
    unitCount?: number;
    unit_config_id?: number;
    unit_num?: number;
    zone_list?: string;
  };

  type DbaObTenant = {
    created_time?: string;
    in_recyclebin?: string;
    locality?: string;
    locked?: string;
    mode?: string;
    primary_zone?: string;
    read_only?: boolean;
    status?: string;
    tenant_id?: number;
    tenant_name?: string;
  };

  type DbaObUnitConfig = {
    create_time?: string;
    log_disk_size?: number;
    max_cpu?: number;
    max_iops?: number;
    memory_size?: number;
    min_cpu?: number;
    min_iops?: number;
    modify_time?: string;
    name?: string;
    unit_config_id?: number;
  };

  type DbaRecyclebin = {
    can_purge?: string;
    can_undrop?: string;
    object_name?: string;
    original_tenant_name?: string;
  };

  type DbPrivilege = {
    db_name?: string;
    privileges?: string[];
  };

  type DbPrivilegeParam = {
    db_name: string;
    privileges: string[];
  };

  type DeadLock = {
    event_id?: string;
    nodes?: DeadLockNode[];
    report_time?: string;
    size?: number;
  };

  type DeadLockNode = {
    idx?: number;
    resource?: string;
    roll_backed?: boolean;
    sql?: string;
    svr_ip?: string;
    svr_port?: number;
    transaction_hash?: string;
  };

  type deleteDatabaseParams = {
    /** tenant name */
    name: string;
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

  type DeleteZoneParams = {
    /** zone name */
    zoneName: string;
  };

  type DiskInfo = {
    deviceName?: string;
    mountHash?: string;
    total?: string;
    used?: string;
  };

  type DropRoleParam = {
    root_password?: string;
  };

  type dropRoleParams = {
    /** tenant name */
    name: string;
    /** role name */
    role: string;
  };

  type DropTenantParam = {
    /** Whether to recycle tenant(can be flashback). */
    need_recycle?: boolean;
  };

  type DropUserParam = {
    root_password?: string;
  };

  type dropUserParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type FlashBackTenantParam = {
    new_name?: string;
  };

  type ForcePassDagParam = {
    id?: string[];
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
    /** tenant name */
    name: string;
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

  type getObclusterCharsetsParams = {
    /** tenant mode */
    tenant_mode?: string;
  };

  type getRestoreOverviewParams = {
    /** Tenant name */
    tenantName: string;
  };

  type getRoleParams = {
    /** tenant name */
    name: string;
    /** role name */
    role: string;
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
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type getSubTaskDetailParams = {
    /** id */
    id: string;
    show_details?: boolean;
  };

  type getTenantArchiveLogOverviewParams = {
    /** Tenant name */
    name: string;
  };

  type getTenantBackupConfigParams = {
    /** Tenant name */
    name: string;
  };

  type getTenantBackupInfoParams = {
    /** Tenant name */
    name: string;
  };

  type getTenantInfoParams = {
    /** tenant name */
    name: string;
  };

  type getTenantOverViewParams = {
    /** tenant compitable mode: MYSQL or ORACLE */
    mode?: string;
  };

  type getTenantParameterParams = {
    /** tenant name */
    name: string;
    /** parameter name */
    para: string;
  };

  type getTenantParametersParams = {
    /** tenant name */
    name: string;
    /** filter format */
    filter?: string;
  };

  type getTenantSessionParams = {
    /** tenant name */
    name: string;
    /** session id */
    sessionId: string;
  };

  type getTenantSessionsParams = {
    /** tenant name */
    name: string;
    /** page */
    page?: number;
    /** size */
    size?: number;
    /** db user */
    user?: string;
    /** db name */
    db?: string;
    /** client ip */
    client_ip?: string;
    /** session id */
    id?: string;
    /** active only */
    active_only?: boolean;
    /** server ip */
    svr_ip?: string;
    /** sort */
    sort?: string;
  };

  type getTenantSessionsStatsParams = {
    /** tenant name */
    name: string;
  };

  type getTenantTopCompactionParams = {
    /** top n */
    limit?: string;
  };

  type getTenantTopSlowSqlRankParams = {
    /** start time */
    start_time: string;
    /** end time */
    end_time: string;
    /** top n */
    limit?: string;
  };

  type getTenantVariableParams = {
    /** tenant name */
    name: string;
    /** variable name */
    var: string;
  };

  type getTenantVariablesParams = {
    /** tenant name */
    name: string;
    /** filter format */
    filter?: string;
  };

  type GetUnfinishedDagsParams = {
    show_details?: boolean;
  };

  type getUserParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type GitInfo = {
    branch?: string;
    commitId?: string;
    commitTime?: string;
    shortCommitId?: string;
  };

  type GrantObjectPrivilegeParam = {
    object_privileges: ObjectPrivilegeParam[];
    root_password?: string;
  };

  type grantRoleObjectPrivilegeParams = {
    /** tenant name */
    name: string;
    /** role name */
    role: string;
  };

  type grantUserObjectPrivilegeParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
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

  type JoinApiParam = {
    agentInfo: AgentInfo;
    masterPassword?: string;
    zoneName: string;
  };

  type KillTenantSessionQueryParam = {
    session_ids?: number[];
  };

  type killTenantSessionQueryParams = {
    /** tenant name */
    name: string;
  };

  type KillTenantSessionsParam = {
    session_ids?: number[];
  };

  type killTenantSessionsParams = {
    /** tenant name */
    name: string;
  };

  type KVPair = {
    key?: string;
    value?: string;
  };

  type ListAllMetricsParams = {
    /** metrics scope */
    scope: 'OBCLUSTER' | 'OBTENANT';
  };

  type listDatabasesParams = {
    /** tenant name */
    name: string;
  };

  type listObjectsParams = {
    /** tenant name */
    name: string;
  };

  type listRestoreTasksParams = {
    /** Tenant name */
    tenantName: string;
  };

  type listRolesParams = {
    /** tenant name */
    name: string;
  };

  type listTenantBackupTasksParams = {
    /** Tenant name */
    name: string;
  };

  type listTenantDeadlocksParams = {
    /** tenant name */
    name: string;
    /** page */
    page?: number;
    /** size */
    size?: number;
  };

  type listUsersParams = {
    /** tenant name */
    name: string;
  };

  type lockUserParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type LogoutSessionParam = {
    session_id: string;
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
    root_password?: string;
  };

  type modifyDbPrivilegeParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type modifyGlobalPrivilegeParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type ModifyObjectPrivilegeParam = {
    object_privileges: ObjectPrivilegeParam[];
    root_password?: string;
  };

  type ModifyReplicasParam = {
    /** Tenant zone list with unit config. */
    zone_list: ModifyReplicaZoneParam[];
  };

  type ModifyReplicaZoneParam = {
    /** Replica type, "FULL"(default) or "READONLY". */
    replica_type?: string;
    unit_config_name?: string;
    unit_num?: number;
    zone_name: string;
  };

  type ModifyRoleGlobalPrivilegeParam = {
    global_privileges?: string[];
    root_password?: string;
  };

  type modifyRoleGlobalPrivilegeParams = {
    /** tenant name */
    name: string;
    /** role name */
    role: string;
  };

  type modifyRoleObjectPrivilegeParams = {
    /** tenant name */
    name: string;
    /** role name */
    role: string;
  };

  type ModifyRoleParam = {
    roles: string[];
    root_password?: string;
  };

  type modifyRoleParams = {
    /** tenant name */
    name: string;
    /** role name */
    role: string;
  };

  type ModifyTenantPrimaryZoneParam = {
    primary_zone: string;
  };

  type ModifyTenantRootPasswordParam = {
    new_password: string;
    old_password?: string;
  };

  type ModifyTenantWhitelistParam = {
    whitelist: string;
  };

  type ModifyUserDbPrivilegeParam = {
    db_privileges: DbPrivilegeParam[];
    root_password?: string;
  };

  type ModifyUserGlobalPrivilegeParam = {
    global_privileges?: string[];
    root_password?: string;
  };

  type modifyUserObjectPrivilegeParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type modifyUserRolesParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
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

  type ObClusterConfigParams = {
    clusterId?: number;
    clusterName?: string;
    rootPwd?: string;
    rsList?: string;
  };

  type ObclusterStatisticInfo = {
    clusterId?: string;
    hosts?: HostInfo[];
    instances?: ObserverInfo[];
    obshellRevision?: string;
    obshellVersion?: string;
    reportTime?: number;
    reporter?: string;
    telemetryEnabled?: boolean;
    telemetryVersion?: number;
  };

  type ObInfoResp = {
    agent_info?: AgentInstance[];
    obcluster_info?: ClusterConfig;
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

  type ObjectPrivilege = {
    object?: DbaObjectBo;
    privileges?: string[];
  };

  type ObjectPrivilegeParam = {
    object_name: string;
    object_type?: string;
    owner: string;
    privileges: string[];
  };

  type ObLicense = {
    activation_time?: string;
    cluster_ulid?: string;
    end_user?: string;
    expired_time?: string;
    issuance_date?: string;
    license_code?: string;
    license_id?: string;
    license_type?: string;
    node_num?: number;
    options?: string;
    product_type?: string;
  };

  type ObParameters = {
    dataType?: string;
    defaultValue?: string;
    editLevel?: string;
    info?: string;
    name?: string;
    scope?: string;
    section?: string;
    svrIp?: string;
    svrPort?: number;
    tenantId?: number;
    value?: string;
    zone?: string;
  };

  type ObParameterValue = {
    svr_ip?: string;
    svr_port?: number;
    tenant_id?: number;
    tenant_name?: string;
    value?: string;
    zone?: string;
  };

  type ObproxyAndConnectionString = {
    connection_string?: string;
    obproxy_address?: string;
    obproxy_port?: number;
    type?: string;
  };

  type ObRole = {
    global_privileges?: string[];
    gmt_create?: string;
    gmt_modified?: string;
    granted_roles?: string[];
    object_privileges?: ObjectPrivilege[];
    role?: string;
    role_grantees?: string[];
    user_grantees?: string[];
  };

  type Observer = {
    architecture?: string;
    id?: number;
    inner_status?: string;
    ip?: string;
    sql_port?: number;
    start_time?: string;
    stats?: ServerResourceStats;
    status?: string;
    stop_time?: string;
    svr_port?: number;
    version?: string;
    with_rootserver?: boolean;
  };

  type ObServerConfigParams = {
    observerConfig: Record<string, any>;
    restart?: boolean;
    scope: Scope;
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

  type ObStopParam = {
    force?: boolean;
    forcePassDag?: ForcePassDagParam;
    scope: Scope;
    terminate?: boolean;
  };

  type ObTenantPreCheckResult = {
    is_connectable?: boolean;
    is_empty_root_password?: boolean;
    is_password_exists?: boolean;
  };

  type ObUnitConfig = {
    create_time?: string;
    log_disk_size?: number;
    max_cpu?: number;
    max_iops?: number;
    memory_size?: number;
    min_cpu?: number;
    min_iops?: number;
    modify_time?: string;
    name?: string;
    unit_config_id?: number;
  };

  type ObUpgradeParam = {
    mode: string;
    release: string;
    upgradeDir?: string;
    version: string;
  };

  type ObUser = {
    accessible_databases?: string[];
    connection_strings?: ObproxyAndConnectionString[];
    create_time?: string;
    db_privileges?: DbPrivilege[];
    global_privileges?: string[];
    granted_roles?: string[];
    is_locked?: boolean;
    /** only for oracle tenant */
    object_privileges?: ObjectPrivilege[];
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

  type PaginatedBackupJobResponse = {
    contents?: BackupJob[];
    page?: CustomPage;
  };

  type PaginatedDeadLocks = {
    contents?: DeadLock[];
    page?: CustomPage;
  };

  type PaginatedRestoreTaskResponse = {
    contents?: RestoreTask[];
    page?: CustomPage;
  };

  type PaginatedTenantSessions = {
    contents?: TenantSession[];
    page?: CustomPage;
  };

  type ParameterTemplate = {
    description?: string;
    scenario?: string;
  };

  type patchRoleObjectPrivilegeParams = {
    /** tenant name */
    name: string;
    /** role name */
    role: string;
  };

  type patchTenantArchiveLogParams = {
    /** Tenant name */
    name: string;
  };

  type patchTenantBackupConfigParams = {
    /** Tenant name */
    name: string;
  };

  type patchTenantBackupParams = {
    /** Tenant name */
    name: string;
  };

  type patchUserObjectPrivilegeParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type PersistTenantRootPasswordParam = {
    password?: string;
  };

  type persistTenantRootPasswordParams = {
    /** tenant name */
    name: string;
  };

  type poolDropParams = {
    /** resource pool name */
    name: string;
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

  type recyclebinFlashbackTenantParams = {
    /** original tenant name or object name in recyclebin */
    name: string;
  };

  type recyclebinTenantPurgeParams = {
    /** original tenant name or object name in recyclebin */
    name: string;
  };

  type RenameTenantParam = {
    /** New tenant name. */
    new_name: string;
  };

  type ResourcePoolWithUnit = {
    observer_list?: string;
    pool_id?: number;
    pool_name?: string;
    unit_config?: ObUnitConfig;
    unit_num?: number;
    zone_list?: string;
  };

  type RestoreOverview = {
    backup_cluster_name?: string;
    backup_cluster_version?: number;
    backup_data_uri?: string;
    backup_log_uri?: string;
    backup_piece_list?: string;
    backup_set_list?: string;
    backup_tenant_id?: number;
    backup_tenant_name?: string;
    comment?: string;
    description?: string;
    finish_bytes?: number;
    finish_bytes_display?: string;
    finish_ls_count?: number;
    finish_tablet_count?: number;
    finish_timestamp?: string;
    job_id?: number;
    ls_count?: number;
    recover_progress?: string;
    recover_scn?: number;
    recover_scn_display?: string;
    restore_option?: string;
    restore_progress?: string;
    restore_scn?: number;
    restore_scn_display?: string;
    restore_tenant_id?: number;
    restore_tenant_name?: string;
    start_timestamp?: string;
    status?: string;
    tablet_count?: number;
    tenant_id?: number;
    total_bytes?: number;
    total_bytes_display?: string;
  };

  type RestoreParam = {
    archive_log_uri?: string;
    concurrency?: number;
    data_backup_uri: string;
    decryption?: string[];
    ha_high_thread_score?: number;
    kms_encrypt_info?: string;
    primary_zone?: string;
    restore_tenant_name: string;
    scn?: number;
    /** time_format does not cause precision loss */
    timestamp?: string;
    /** Tenant zone list with unit config. */
    zone_list: ZoneParam[];
  };

  type RestoreParams = {
    params: ObParameters[];
  };

  type RestoreStorageParam = {
    archive_log_uri?: string;
    data_backup_uri: string;
  };

  type RestoreTask = {
    backup_cluster_name?: string;
    backup_cluster_version?: number;
    backup_data_uri?: string;
    backup_log_uri?: string;
    backup_piece_list?: string;
    backup_set_list?: string;
    backup_tenant_id?: number;
    backup_tenant_name?: string;
    comment?: string;
    description?: string;
    finish_bytes?: number;
    finish_bytes_display?: string;
    finish_ls_count?: number;
    finish_tablet_count?: number;
    finish_timestamp?: string;
    job_id?: number;
    ls_count?: number;
    recover_progress?: string;
    recover_scn?: number;
    recover_scn_display?: string;
    restore_option?: string;
    restore_progress?: string;
    restore_scn?: number;
    restore_scn_display?: string;
    restore_tenant_id?: number;
    restore_tenant_name?: string;
    start_timestamp?: string;
    status?: string;
    tablet_count?: number;
    tenant_id?: number;
    total_bytes?: number;
    total_bytes_display?: string;
  };

  type RestoreTenantInfo = {
    cluster_name?: string;
    restore_windows?: RestoreWindows;
    tenant_name?: string;
  };

  type RestoreWindow = {
    end_time?: string;
    start_time?: string;
  };

  type RestoreWindows = {
    restore_windows?: RestoreWindow[];
  };

  type RestoreWindowsParam = {
    archive_log_uri?: string;
    data_backup_uri: string;
  };

  type RevokeObjectPrivilegeParam = {
    object_privileges: ObjectPrivilegeParam[];
    root_password?: string;
  };

  type revokeRoleObjectPrivilegeParams = {
    /** tenant name */
    name: string;
    /** role name */
    role: string;
  };

  type revokeUserObjectPrivilegeParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type RootServer = {
    ip?: string;
    role?: string;
    svr_port?: number;
    zone?: string;
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

  type ScaleInTenantReplicasParam = {
    zones: string[];
  };

  type ScaleOutTenantReplicasParam = {
    /** Tenant zone list with unit config. */
    zone_list: ZoneParam[];
  };

  type Scope = {
    target?: string[];
    type?: string;
  };

  type ServerConfig = {
    agent_port?: number;
    build_version?: string;
    sql_port?: number;
    status?: string;
    svr_ip?: string;
    svr_port?: number;
    with_rootserver?: string;
  };

  type ServerResourceStats = {
    cpu_core_assigned?: number;
    cpu_core_assigned_percent?: number;
    cpu_core_total?: number;
    disk_assigned?: string;
    disk_assigned_percent?: number;
    disk_free?: string;
    disk_in_bytes_assigned?: number;
    disk_in_bytes_free?: number;
    disk_in_bytes_total?: number;
    disk_in_bytes_used?: number;
    disk_total?: string;
    disk_used?: string;
    disk_used_percent?: number;
    ip?: string;
    memory_assigned?: string;
    memory_assigned_percent?: number;
    memory_in_bytes_assigned?: number;
    memory_in_bytes_total?: number;
    memory_total?: string;
    port?: number;
    zone?: string;
  };

  type SetObclusterParametersParam = {
    params: SetSingleObclusterParameterParam[];
  };

  type SetSingleObclusterParameterParam = {
    /** Whether to set all tenants, if true, the Tenants field will be ignored. */
    all_user_tenant?: boolean;
    name: string;
    /** Scope can be “CLUSTER" or "TENANT". */
    scope: string;
    servers?: string[];
    /** Tenant name list, if not set, it means all tenants. */
    tenants?: string[];
    value: string;
    zones?: string[];
  };

  type SetTenantParametersParam = {
    parameters: Record<string, any>;
  };

  type SetTenantVariablesParam = {
    tenant_password?: string;
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

  type StartObParam = {
    forcePassDag?: ForcePassDagParam;
    scope: Scope;
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

  type tenantAddReplicasParams = {
    /** tenant name */
    name: string;
  };

  type TenantBackupConfigParam = {
    archive_base_uri?: string;
    archive_lag_target?: string;
    binding?: string;
    data_base_uri?: string;
    delete_policy?: BackupDeletePolicy;
    ha_low_thread_score?: number;
    location?: string;
    log_archive_concurrency?: number;
    piece_switch_interval?: string;
  };

  type tenantBackupConfigParams = {
    /** Tenant name */
    name: string;
  };

  type TenantBackupInfo = {
    archive_log_delay?: number;
    lastest_archive_log_checkpoint?: string;
    lastest_data_backup_time?: string;
    tenant_id?: number;
  };

  type tenantBackupOverviewParams = {
    /** Tenant name */
    name: string;
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
    tenant_id?: number;
  };

  type TenantCompactionHistory = {
    cost_time?: number;
    last_finish_time?: string;
    start_time?: string;
    status?: string;
    tenant_id?: number;
    tenant_name?: string;
  };

  type tenantDropParams = {
    /** tenant name */
    name: string;
  };

  type TenantInfo = {
    /** Only for ORACLE tenant */
    charset?: string;
    /** Only for ORACLE tenant */
    collation?: string;
    connection_strings?: ObproxyAndConnectionString[];
    created_time?: string;
    in_recyclebin?: string;
    locality?: string;
    locked?: string;
    lower_case_table_names?: string;
    mode?: string;
    pools?: ResourcePoolWithUnit[];
    primary_zone?: string;
    /** Default to false. */
    read_only?: boolean;
    status?: string;
    tenant_id?: number;
    tenant_name?: string;
    time_zone?: string;
    whitelist?: string;
  };

  type tenantLockParams = {
    /** tenant name */
    name: string;
  };

  type tenantModifyPasswordParams = {
    /** tenant name */
    name: string;
  };

  type tenantModifyPrimaryZoneParams = {
    /** tenant name */
    name: string;
  };

  type tenantModifyReplicasParams = {
    /** tenant name */
    name: string;
  };

  type tenantModifyWhitelistParams = {
    /** tenant name */
    name: string;
  };

  type TenantParameterValue = {
    tenant_id?: number;
    tenant_name?: string;
    value?: string;
  };

  type tenantPreCheckParams = {
    /** tenant name */
    name: string;
  };

  type tenantRemoveReplicasParams = {
    /** tenant name */
    name: string;
  };

  type tenantRenameParams = {
    /** tenant name */
    name: string;
  };

  type TenantResourceStat = {
    cpu_core_total?: number;
    cpu_used_percent?: number;
    data_disk_usage?: number;
    memory_in_bytes_total?: number;
    memory_used_percent?: number;
    tenant_id?: number;
    tenant_name?: string;
  };

  type TenantRootPasswordParam = {
    root_password?: string;
  };

  type TenantSession = {
    action?: string;
    client_info?: string;
    command?: string;
    db?: string;
    host?: string;
    id?: number;
    info?: string;
    level?: number;
    memory_usage?: number;
    module?: string;
    /** parse from host when ProxySessId is not null */
    proxy_ip?: string;
    proxy_sess_id?: number;
    record_policy?: string;
    sample_percentage?: number;
    sql_id?: string;
    sql_port?: number;
    state?: string;
    svr_ip?: string;
    svr_port?: number;
    tenant?: string;
    time?: number;
    total_cpu_time?: number;
    user?: string;
  };

  type TenantSessionClientStats = {
    active_count?: number;
    client_ip?: string;
    total_count?: number;
  };

  type TenantSessionDbStats = {
    active_count?: number;
    db_name?: string;
    total_count?: number;
  };

  type TenantSessionStats = {
    active_count?: number;
    client_stats?: TenantSessionClientStats[];
    db_stats?: TenantSessionDbStats[];
    max_active_time?: number;
    total_count?: number;
    user_stats?: TenantSessionUserStats[];
  };

  type TenantSessionUserStats = {
    active_count?: number;
    total_count?: number;
    user_name?: string;
  };

  type tenantSetParametersParams = {
    /** tenant name */
    name: string;
  };

  type tenantSetVariableParams = {
    /** tenant name */
    name: string;
  };

  type TenantSlowSqlCount = {
    count?: number;
    tenant_id?: number;
    tenant_name?: string;
  };

  type tenantStartBackupParams = {
    /** Tenant name */
    name: string;
  };

  type tenantUnlockParams = {
    /** tenant name */
    name: string;
  };

  type UlimitInfo = {
    maxUserProcessesHard?: string;
    maxUserProcessesSoft?: string;
    nofileHard?: string;
    nofileSoft?: string;
  };

  type unitConfigDropParams = {
    /** resource unit name */
    name: string;
  };

  type unitConfigGetParams = {
    /** resource unit name */
    name: string;
  };

  type unlockUserParams = {
    /** tenant name */
    name: string;
    /** user name */
    user: string;
  };

  type updateDatabaseParams = {
    /** tenant name */
    name: string;
    /** database name */
    database: string;
  };

  type UpgradeCheckParam = {
    release: string;
    upgradeDir?: string;
    version: string;
  };

  type UpgradeObproxyParam = {
    release: string;
    upgrade_dir?: string;
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

  type UpgradePkgInfo = {
    architecture?: string;
    chunkCount?: number;
    distribution?: string;
    gmtModify?: string;
    md5?: string;
    name?: string;
    payloadSize?: number;
    pkgId?: number;
    release?: string;
    releaseDistribution?: string;
    size?: number;
    upgradeDepYaml?: string;
    version?: string;
  };

  type UpgradePkgInfo = {
    architecture?: string;
    chunkCount?: number;
    distribution?: string;
    gmtModify?: string;
    md5?: string;
    name?: string;
    payloadSize?: number;
    pkgId?: number;
    release?: string;
    releaseDistribution?: string;
    size?: number;
    version?: string;
  };

  type Zone = {
    idc_name?: string;
    inner_status?: string;
    name?: string;
    region_name?: string;
    root_server?: RootServer;
    servers?: Observer[];
    status?: string;
  };

  type ZoneParam = {
    name: string;
    /** Replica type, "FULL"(default) or "READONLY". */
    replica_type?: string;
    unit_config_name: string;
    unit_num: number;
  };
}
