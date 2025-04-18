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

  type ApiError = {
    /** Error code */
    code?: number;
    /** Error message */
    message?: string;
    /** Sub errors */
    subErrors?: any[];
  };

  type ArchiveLogStatusParam = {
    status?: string;
  };

  type BackupDeletePolicy = {
    policy?: string;
    recovery_window?: string;
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
    community_edition?: boolean;
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

  type CollationInfo = {
    is_default?: boolean;
    name?: string;
  };

  type CreateDatabaseParam = {
    collation?: string;
    db_name: string;
    read_only?: boolean;
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
    root_password?: string;
    user_name: string;
  };

  type createUserParams = {
    /** tenant name */
    name: string;
  };

  type DagDetailDTO = {
    additional_data?: Record<string, any>;
    dag_id?: number;
    end_time?: string;
    id: string;
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

  type deleteDatabaseParams = {
    /** tenant name */
    name: string;
    /** database name */
    database: string;
  };

  type DeleteZoneParams = {
    /** zone name */
    zoneName: string;
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

  type getRestoreOverviewParams = {
    /** Tenant name */
    tenantName: string;
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

  type getTenantInfoParams = {
    /** tenant name */
    name: string;
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

  type getTenantTopCompactionParams = {
    /** top n */
    top?: string;
  };

  type getTenantTopSlowSqlRankParams = {
    /** start time */
    start_time: string;
    /** end time */
    end_time: string;
    /** top n */
    top?: string;
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

  type GvObParameter = {
    data_type?: string;
    edit_level?: string;
    info?: string;
    name?: string;
    value?: string;
  };

  type JoinApiParam = {
    agentInfo: AgentInfo;
    masterPassword?: string;
    zoneName: string;
  };

  type listDatabasesParams = {
    /** tenant name */
    name: string;
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

  type ModifyDatabaseParam = {
    collation?: string;
    read_only?: boolean;
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
    db_privileges?: DbPrivilegeParam[];
  };

  type ModifyUserGlobalPrivilegeParam = {
    global_privileges?: string[];
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

  type ObClusterConfigParams = {
    clusterId?: number;
    clusterName?: string;
    rootPwd?: string;
    rsList?: string;
  };

  type ObInfoResp = {
    agent_info?: AgentInstance[];
    obcluster_info?: ClusterConfig;
  };

  type ObjectPrivilege = {
    object?: string;
    privileges?: string[];
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

  type Observer = {
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

  type ObStopParam = {
    force?: boolean;
    forcePassDag?: ForcePassDagParam;
    scope: Scope;
    terminate?: boolean;
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
    db_privileges?: DbPrivilege[];
    global_privileges?: string[];
    granted_roles?: string[];
    is_locked?: boolean;
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

  type patchTenantArchiveLogParams = {
    /** Tenant name */
    name: string;
  };

  type patchTenantBackupParams = {
    /** Tenant name */
    name: string;
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
    pool_id?: number;
    pool_name?: string;
    unit_config?: ObUnitConfig;
    unit_num?: number;
    zone_list?: string;
  };

  type RestoreOverview = {
    backup_cluster_name?: string;
    backup_cluster_version?: number;
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
    timestamp?: string;
    /** Tenant zone list with unit config. */
    zone_list: ZoneParam[];
  };

  type RestoreParams = {
    params: ObParameters[];
  };

  type RestoreWindowsParam = {
    archive_log_uri?: string;
    data_backup_uri: string;
  };

  type RootServer = {
    ip?: string;
    role?: string;
    svr_port?: number;
    zone?: string;
  };

  type RouteNode = {
    build_version?: string;
    deprecated_info?: string[];
    release?: string;
    require_from_binary?: boolean;
    version?: string;
  };

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

  type StartObParam = {
    forcePassDag?: ForcePassDagParam;
    scope: Scope;
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
    created_time?: string;
    in_recyclebin?: string;
    locality?: string;
    locked?: string;
    mode?: string;
    pools?: ResourcePoolWithUnit[];
    primary_zone?: string;
    /** Default to false. */
    read_only?: boolean;
    status?: string;
    tenant_id?: number;
    tenant_name?: string;
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
    cpu_used_percent?: number;
    data_disk_usage?: number;
    memory_used_percent?: number;
    tenant_id?: number;
    tenant_name?: string;
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

  type UpgradePkgRouteParams = {
    /** version */
    version: string;
    /** release */
    release: string;
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
