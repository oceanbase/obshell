/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package constant

import "time"

const (
	SCENARIO_BASIC       = "BASIC"
	SCENARIO_PERFORMANCE = "PERFORMANCE"

	// Inspection report status
	INSPECTION_STATUS_RUNNING = "RUNNING"
	INSPECTION_STATUS_SUCCEED = "SUCCEED"
	INSPECTION_STATUS_FAILED  = "FAILED"
	INSPECTION_STATUS_DELETED = "DELETED"
	INSPECTION_STATUS_UNKNOWN = "UNKNOWN"

	CONFIG_DB_HOST                              = "db_host"
	CONFIG_DB_PORT                              = "db_port"
	CONFIG_TENANT_SYS_USER                      = "tenant_sys.user"
	CONFIG_TENANT_SYS_PASSWORD                  = "tenant_sys.password"
	CONFIG_CLUSTER_NAME                         = "obcluster.ob_cluster_name"
	CONFIG_OBCLUSTER_SERVERS_NODES_SSH_USERNAME = "obcluster.servers.nodes[%d].ssh_username"
	CONFIG_OBCLUSTER_SERVERS_NODES_SSH_PASSWORD = "obcluster.servers.nodes[%d].ssh_password"
	CONFIG_OBCLUSTER_SERVERS_NODES_HOME_PATH    = "obcluster.servers.nodes[%d].home_path"
	CONFIG_OBCLUSTER_SERVERS_NODES_DATA_DIR     = "obcluster.servers.nodes[%d].data_dir"
	CONFIG_OBCLUSTER_SERVERS_NODES_REDO_DIR     = "obcluster.servers.nodes[%d].redo_dir"
	CONFIG_OBCLUSTER_SERVERS_NODES_IP           = "obcluster.servers.nodes[%d].ip"

	BINARY_OBDIAG = "obdiag"

	OBDIAG_VERSION_MIN = "3.7.2"
)

var ZERO_TIME = time.Unix(0, 0)

var INSPECTION_BASIC_OBSERVER_TASKS = []string{
	"bugs.*",
	"err_code.*",
	"clog.*",
	"ls.*",
	"log.log_size",
	"log.log_size_with_ocp",
	"disk.data_disk_full",
	"disk.disk_full",
	"disk.disk_hole",
	"disk.clog_abnormal_file",
	"disk.sstable_abnormal_file",
	"disk.mount_disk_full",
	"disk.xfs_repair",
	"table.information_schema_tables_two_data",
	"table.auto_split_error",
	"tenant.tenant_min_resource",
	"tenant.writing_throttling_trigger_percentage",
	"tenant.ddl_operation_table_size",
	"tenant.tenant_locality_consistency_check",
	"tenant.max_stale_time_for_weak_consistency",
	"cluster.auto_increment_cache_size",
	"cluster.datafile_next",
	"cluster.data_path_settings",
	"cluster.deadlocks",
	"cluster.freeze_trigger_percentage",
	"cluster.global_indexes_too_much",
	"cluster.logons_check",
	"cluster.ls_number",
	"cluster.major",
	"cluster.major_suspended",
	"cluster.memory_chunk_cache_size",
	"cluster.memory_limit_percentage",
	"cluster.memory_limit_vs_phy_mem",
	"cluster.memstore_limit_percentage",
	"cluster.mod_too_large",
	"cluster.no_leader",
	"cluster.ob_enable_plan_cache_bad_version",
	"cluster.ob_query_timeout",
	"cluster.observer_not_active",
	"cluster.observer_port",
	"cluster.optimizer_better_inlist_costing_parmmeter",
	"cluster.part_trans_action_max",
	"cluster.resource_limit_max_session_num",
	"cluster.server_permanent_offline_time",
	"cluster.session_limit",
	"cluster.sys_log_level",
	"cluster.sys_obcon_health",
	"cluster.table_history_too_many",
	"cluster.task_opt_stat_gather_fail",
	"cluster.tenant_locks",
	"cluster.tenant_memory_tablet_count",
	"cluster.tenant_number",
	"cluster.upgrade_finished",
	"cluster.upper_trans_version",
	"cluster.zone_not_active",
	"cluster.core_file_find",
	"network.local_ip_check",
	"network.TCP-retransmission",
	"system.aio",
	"system.arm_smmu",
	"system.check_command",
	"system.check_system_language",
	"system.clock_source",
	"system.clock_source_check",
	"system.core_pattern",
	"system.dependent_software",
	"system.dependent_software_swapon",
	"system.getenforce",
	"system.instruction_set_avx",
	"system.kernel_bad_version",
	"system.mount_options",
	"system.parameter",
	"system.parameter_ip_local_port_range",
	"system.parameter_tcp_rmem",
	"system.parameter_tcp_wmem",
	"system.ulimit_parameter",
	"version.*",
}

var INSPECTION_PERFORMANCE_OBSERVER_TASKS = []string{
	"index.global_index_unpartitioned",
	"cluster.autoinc_cache_refresh_interval",
	"cluster.clog_sync_time_warn_threshold",
	"cluster.cpu_quota_concurrency",
	"cluster.default_compress_func",
	"cluster.enable_lock_priority",
	"cluster.large_query_threshold",
	"cluster.memstore_usage",
	"cluster.ob_enable_prepared_statement",
	"cluster.syslog_io_bandwidth_limit",
	"cluster.task_opt_stat",
	"cluster.trace_log_slow_query_watermark",
	"cpu.oversold",
	"disk.disk_iops",
	"tenant.tenant_threshold",
	"network.log_easy_slow",
	"network.network_speed",
	"network.network_drop",
	"network.network_speed_diff",
	"network.network_offset",
	"network.network_write_cond_wakeup",
	"system.tcp_tw_reuse",
	"system.cgroup_version",
}
