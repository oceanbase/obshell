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

package errors

import "net/http"

type ErrorKind = int

const (
	badRequest      ErrorKind = http.StatusBadRequest
	illegalArgument ErrorKind = http.StatusBadRequest
	unauthorized    ErrorKind = http.StatusUnauthorized
	notFound        ErrorKind = http.StatusNotFound
	unexpected      ErrorKind = http.StatusInternalServerError
	notfound        ErrorKind = http.StatusNotFound
	known           ErrorKind = http.StatusInternalServerError
)

// ErrorCode includes code, kind and key.
type ErrorCode struct {
	Code    string
	Kind    ErrorKind
	key     string
	OldCode int
}

// NewErrorCode will create a new ErrorCode and append it to errorCodes
func NewErrorCode(code string, kind ErrorKind, key string, args ...interface{}) ErrorCode {
	errorCode := ErrorCode{
		Code:    code,
		Kind:    kind,
		key:     key,
		OldCode: ErrObsoleteCode,
	}
	if len(args) > 0 {
		errorCode.OldCode = args[0].(int)
	}

	return errorCode
}

var (
	ErrObsoleteCode = 0 // 0 means old code is not set

	// general error codes, range: 1000 ~ 1999
	ErrCommonBadRequest                 = NewErrorCode("Common.BadRequest", badRequest, "err.common.bad.request") // 通用错误：请求参数错误
	ErrCommonNotFound                   = NewErrorCode("Common.NotFound", notFound, "err.common.not.found")
	ErrCommonBindJsonFailed             = NewErrorCode("Common.BindJsonFailed", badRequest, "err.common.bind.json.failed")
	ErrCommonFileNotExist               = NewErrorCode("Common.FileNotExist", badRequest, "err.common.file.not.exist")   // "file '%s' is not exist"
	ErrCommonInvalidPath                = NewErrorCode("Common.InvalidPath", illegalArgument, "err.common.invalid.path") // "path '%s' is not valid: %s"
	ErrCommonIllegalArgument            = NewErrorCode("Common.IllegalArgument", illegalArgument, "err.common.illegal.argument")
	ErrCommonIllegalArgumentWithMessage = NewErrorCode("Common.IllegalArgument", illegalArgument, "err.common.illegal.argument.with.message") // the input parameter '%s' is illegal: %s
	ErrCommonInvalidPort                = NewErrorCode("Common.InvalidPort", illegalArgument, "err.common.invalid.port")                      // The port '%s' is invalid, must in (1024, 65535].
	ErrCommonInvalidIp                  = NewErrorCode("Common.InvalidIp", illegalArgument, "err.common.invalid.ip")                          // "'%s' is not a valid IP address"
	ErrCommonInvalidAddress             = NewErrorCode("Common.InvalidAddress", illegalArgument, "err.common.invalid.address")                // "'%s' is not a vaild address"
	ErrCommonDirNotEmpty                = NewErrorCode("Common.DirNotEmpty", unexpected, "err.common.dir.not.empty")                          // "dir '%s' is not empty"
	ErrCommonFilePermissionDenied       = NewErrorCode("Common.FilePermissionDenied", illegalArgument, "err.common.file.permission.denied")   // "no read/write permission for file '%s'"
	ErrCommonPathNotExist               = NewErrorCode("Common.PathNotExist", illegalArgument, "err.common.path.not.exist")                   // "'%s' is not exist"
	ErrCommonPathNotDir                 = NewErrorCode("Common.PathNotDir", illegalArgument, "err.common.path.not.dir")                       // "'%s' is not a directory"
	ErrCommonUnexpected                 = NewErrorCode("Common.Unexpected", unexpected, "err.common.unexpected")                              // "unexpected error: %s"
	ErrCommonUnauthorized               = NewErrorCode("Common.Unauthorized", unauthorized, "err.common.unauthorized")                        // "unauthorized"
	ErrCommonInvalidTimeDuration        = NewErrorCode("Common.InvalidTimeDuration", illegalArgument, "err.common.invalid.time.duration")     // "time duration '%s' is invalid: %s"
	// Log
	ErrLogWriteExceedMaxSize          = NewErrorCode("Log.WriteExceedMaxSize", unexpected, "err.log.write.exceed.max.size")                   // "write length %d exceeds maximum file size %d"
	ErrLogFileNamePrefixMismatched    = NewErrorCode("Log.FileNamePrefixMismatched", unexpected, "err.log.file.name.prefix.mismatched")       // "file name '%s' prefix mismatched"
	ErrLogFileNameExtensionMismatched = NewErrorCode("Log.FileNameExtensionMismatched", unexpected, "err.log.file.name.extension.mismatched") // "file name '%s' extension mismatched"

	// RPC
	ErrAgentRPCRequestError  = NewErrorCode("Agent.RPC.RequestError", unexpected, "err.agent.rpc.request.error")   // "request [%s]%s to %s error: %v"
	ErrAgentRPCRequestFailed = NewErrorCode("Agent.RPC.RequestFailed", unexpected, "err.agent.rpc.request.failed") // "request [%s]%s to %s failed: %s"

	// OB.Binary
	ErrObBinaryVersionUnexpected = NewErrorCode("OB.Binary.Version.Unexpected", unexpected, "err.ob.binary.version.unexpected")

	// OB.Tenant
	ErrObTenantNameInvalid                       = NewErrorCode("OB.Tenant.Name.Invalid", illegalArgument, "err.ob.tenant.name.invalid")                                                  // "tenant name '%s' is invalid: %s"
	ErrObTenantExisted                           = NewErrorCode("OB.Tenant.Existed", illegalArgument, "err.ob.tenant.existed")                                                            // "tenant %s is already existed"
	ErrObTenantNotExist                          = NewErrorCode("OB.Tenant.NotExist", badRequest, "err.ob.tenant.not.exist")                                                              // "tenant '%s' is not exist"
	ErrObTenantLocked                            = NewErrorCode("OB.Tenant.Locked", illegalArgument, "err.ob.tenant.locked")                                                              // "tenant %s is locked"
	ErrObTenantUnitNumInvalid                    = NewErrorCode("OB.Tenant.UnitNum.Invalid", illegalArgument, "err.ob.tenant.unit.num.invalid")                                           // "unit num '%d' is invalid: %s"
	ErrObTenantUnitNumExceedsLimit               = NewErrorCode("OB.Tenant.UnitNum.ExceedsLimit", illegalArgument, "err.ob.tenant.unit.num.exceeds.limit")                                // "unit num '%d' is bigger than server num '%d' in zone '%s'"
	ErrObTenantUnitNumInconsistent               = NewErrorCode("OB.Tenant.UnitNum.Inconsistent", illegalArgument, "err.ob.tenant.unit.num.inconsistent")                                 // "unit num is inconsistent, should be same in all zones"
	ErrObTenantZoneListEmpty                     = NewErrorCode("OB.Tenant.ZoneList.Empty", illegalArgument, "err.ob.tenant.zone.list.empty")                                             // "zone_list is empty"
	ErrObTenantZoneRepeated                      = NewErrorCode("OB.Tenant.Zone.Repeated", illegalArgument, "err.ob.tenant.zone.repeated")                                                // "zone '%s' is repeated"
	ErrObTenantModeNotSupported                  = NewErrorCode("OB.Tenant.Mode.NotSupported", illegalArgument, "err.ob.tenant.mode.not.supported")                                       // "tenant mode '%s' is not supported, only '%s' is supported"
	ErrObTenantNameEmpty                         = NewErrorCode("OB.Tenant.Name.Empty", illegalArgument, "err.ob.tenant.name.empty")                                                      // "tenant name is empty"
	ErrObTenantPrimaryZoneInvalid                = NewErrorCode("OB.Tenant.PrimaryZone.Invalid", illegalArgument, "err.ob.tenant.primary.zone.invalid")                                   // "primary zone '%s' is invalid: %s"
	ErrObTenantPrimaryZoneCrossRegion            = NewErrorCode("OB.Tenant.PrimaryZone.CrossRegion", illegalArgument, "err.ob.tenant.primary.zone.cross.region")                          // "primary zone '%s' is cross region, please check tenant's primary zone"
	ErrObTenantPrimaryRegionFullReplicaNotEnough = NewErrorCode("OB.Tenant.PrimaryRegion.FullReplica.NotEnough", illegalArgument, "err.ob.tenant.primary.region.full.replica.not.enough") // "The region %v where the first priority of tenant zone is located needs to have at least 2 F replicas. In fact, there are only %d full replicas."
	ErrObTenantResourceNotEnough                 = NewErrorCode("OB.Tenant.ResourceNotEnough", badRequest, "err.ob.tenant.resource.not.enough")                                           // "server %s %s resource not enough"
	ErrObTenantStatusNotNormal                   = NewErrorCode("OB.Tenant.StatusNotNormal", badRequest, "err.ob.tenant.status.not.normal")                                               // "tenant '%s' status is '%s'"                                     // "resource pool '%s' has already been granted to a tenant"
	ErrObTenantZoneAlreadyHasReplica             = NewErrorCode("OB.Tenant.ZoneAlreadyHasReplica", badRequest, "err.ob.tenant.zone.already.has.replica")                                  // "zone '%s' already has a replica"
	ErrObTenantZoneWithoutReplica                = NewErrorCode("OB.Tenant.ZoneWithoutReplica", badRequest, "err.ob.tenant.zone.without.replica")                                         // "zone '%s' does not have a replica"
	ErrObTenantHasPoolOnZone                     = NewErrorCode("OB.Tenant.HasPoolOnZone", badRequest, "err.ob.tenant.has.pool.on.zone")                                                  // "tenant already has a pool located in zone '%s'"
	ErrObTenantRebalanceDisabled                 = NewErrorCode("OB.Tenant.RebalanceDisabled", badRequest, "err.ob.tenant.rebalance.disabled")                                            // "%s is not allowed when tenant 'enable_rebalance' is disabled"
	ErrObTenantLocalityPrincipalNotAllowed       = NewErrorCode("OB.Tenant.Locality.PrincipalNotAllowed", badRequest, "err.ob.tenant.locality.principal.not.allowed")                     // "violate locality principal not allowed"
	ErrObTenantLocalityFormatUnexpected          = NewErrorCode("OB.Tenant.Locality.Format.Unexpected", unexpected, "err.ob.tenant.locality.format.unexpected")                           // "Unexpected locality format: %s"
	ErrObTenantModifyUnitNumPartially            = NewErrorCode("OB.Tenant.ModifyUnitNumPartially", unexpected, "err.ob.tenant.modify.unit.num.partially")                                // "Could not modify unit num partially"
	ErrObTenantJobFailed                         = NewErrorCode("OB.Tenant.Job.Failed", unexpected, "err.ob.tenant.job.failed")                                                           // "Job %d failed, job status is %s"
	ErrObTenantJobWaitTimeout                    = NewErrorCode("OB.Tenant.Job.WaitTimeout", unexpected, "err.ob.tenant.job.wait.timeout")                                                // "Job %d wait timeout"
	ErrObTenantJobConflict                       = NewErrorCode("OB.Tenant.Job.Conflict", unexpected, "err.ob.tenant.job.conflict")                                                       // "There is already a in-progress '%s' job"
	ErrObTenantJobNotExist                       = NewErrorCode("OB.Tenant.Job.NotExist", badRequest, "err.ob.tenant.job.not.exist")                                                      // "There is no job of '%s'"
	ErrObTenantCompactionStatusNotIdle           = NewErrorCode("OB.Tenant.Compaction.Status.NotIdle", badRequest, "err.ob.tenant.compaction.status.not.idle")                            // "tenant '%s' is in '%s' status, operation not allowed"
	ErrObTenantRootPasswordIncorrect             = NewErrorCode("OB.Tenant.RootPasswordIncorrect", badRequest, "err.ob.tenant.root.password.incorrect")                                   // "The provided password is unable to connect to the tenant."
	ErrObTenantSysOperationNotAllowed            = NewErrorCode("OB.Tenant.SysOperationNotAllowed", badRequest, "err.ob.tenant.sys.operation.not.allowed")                                // "sys tenant is not allowed to do this operation"
	ErrObTenantScenarioNotSupported              = NewErrorCode("OB.Tenant.Scenario.NotSupported", illegalArgument, "err.ob.tenant.scenario.not.supported")                               // "current observer does not support scenario"
	ErrObTenantSetScenarioNotSupported           = NewErrorCode("OB.Tenant.SetScenario.NotSupported", illegalArgument, "err.ob.tenant.set.scenario.not.supported")                        // "current observer does not support scenario"
	ErrObTenantNoPaxosReplica                    = NewErrorCode("OB.Tenant.NoPaxosReplica", illegalArgument, "err.ob.tenant.no.paxos.replica")                                            // "tenant '%s' has no paxos replica, please check tenant's replica info"
	ErrObTenantCollationInvalid                  = NewErrorCode("OB.Tenant.Collation.Invalid", illegalArgument, "err.ob.tenant.collation.invalid")                                        // "invalid collation: '%s'"
	ErrObTenantUnderMaintenance                  = NewErrorCode("OB.Tenant.UnderMaintenance", known, "err.ob.tenant.under.maintenance")                                                   // "tenant '%s' is under maintenance, please try again later"
	ErrObTenantNoActiveServer                    = NewErrorCode("OB.Tenant.NoActiveServer", badRequest, "err.ob.tenant.no.active.server")                                                 // "tenant '%s' has no active server"
	ErrObTenantEmptyVariable                     = NewErrorCode("OB.Tenant.Variable.Empty", illegalArgument, "err.ob.tenant.variable.empty")                                              // "tenant variable name or value is empty"
	ErrObTenantEmptyParameter                    = NewErrorCode("OB.Tenant.Parameter.Empty", illegalArgument, "err.ob.tenant.parameter.empty")                                            // "tenant parameter name or value is empty"
	ErrObTenantParameterNameEmpty                = NewErrorCode("OB.Tenant.Parameter.Name.Empty", illegalArgument, "err.ob.tenant.parameter.name.empty")                                  // "tenant parameter name is empty"
	ErrObTenantParameterNotExist                 = NewErrorCode("OB.Tenant.Parameter.NotExist", notFound, "err.ob.tenant.parameter.not.exist")                                            // "tenant parameter '%s' is not exist"
	ErrObTenantVariableInvalid                   = NewErrorCode("OB.Tenant.Variable.Invalid", illegalArgument, "err.ob.tenant.variable.invalid")                                          // "tenant variable '%s' is invalid: %s"
	ErrObTenantVariableNotExist                  = NewErrorCode("OB.Tenant.Variable.NotExist", notFound, "err.ob.tenant.variable.not.exist")                                              // "tenant variable '%s' is not exist"
	ErrObTenantVariableNameEmpty                 = NewErrorCode("OB.Tenant.Variable.Name.Empty", illegalArgument, "err.ob.tenant.variable.name.empty")                                    // "tenant variable name is empty"

	// OB.Recyclebin
	ErrObRecyclebinTenantNotExist = NewErrorCode("OB.Recyclebin.Tenant.NotExist", badRequest, "err.ob.recyclebin.tenant.not.exist")

	// OB.Resource.UnitConfig
	ErrObResourceUnitConfigNameEmpty = NewErrorCode("OB.Resource.UnitConfig.Name.Empty", illegalArgument, "err.ob.resource.unit.config.name.empty")
	ErrObResourceUnitConfigNotExist  = NewErrorCode("OB.Resource.UnitConfig.NotExist", illegalArgument, "err.ob.resource.unit.config.not.exist")
	ErrObResourceUnitConfigExisted   = NewErrorCode("OB.Resource.UnitConfig.Existed", illegalArgument, "err.ob.resource.unit.config.existed")

	// OB.Resource.Pool
	ErrObResourcePoolNameEmpty = NewErrorCode("OB.Resource.Pool.Name.Empty", illegalArgument, "err.ob.resource.pool.name.empty")
	ErrObResourcePoolGranted   = NewErrorCode("OB.Resource.Pool.Granted", badRequest, "err.ob.resource.pool.granted")

	// OB.Tenant.Replica
	ErrObTenantReplicaTypeInvalid = NewErrorCode("OB.Tenant.Replica.InvalidType", illegalArgument, "err.ob.tenant.replica.type.invalid")
	ErrObTenantReplicaOnlyOne     = NewErrorCode("OB.Tenant.OnlyOneReplica", illegalArgument, "err.ob.tenant.only.one.replica")
	ErrObTenantReplicaDeleteAll   = NewErrorCode("OB.Tenant.Replica.DeleteAll", badRequest, "err.ob.tenant.replica.delete.all")

	// OB.Backup
	ErrObBackupBaseUriEmpty                 = NewErrorCode("OB.Backup.BaseUriEmpty", illegalArgument, "err.ob.backup.base.uri.empty")
	ErrObBackupArchiveBaseUriEmpty          = NewErrorCode("OB.Backup.ArchiveBaseUriEmpty", illegalArgument, "err.ob.backup.archive.base.uri.empty")
	ErrObBackupDataBaseUriEmpty             = NewErrorCode("OB.Backup.DataBaseUriEmpty", illegalArgument, "err.ob.backup.data.base.uri.empty")
	ErrObBackupNoUserTenants                = NewErrorCode("OB.Backup.NoUserTenants", illegalArgument, "err.ob.backup.no.user.tenants")
	ErrObBackupLogArchiveConcurrencyInvalid = NewErrorCode("OB.Backup.LogArchiveConcurrency.Invalid", illegalArgument, "err.ob.backup.log.archive.concurrency.invalid")
	ErrObBackupHaLowThreadScoreInvalid      = NewErrorCode("OB.Backup.HaLowThreadScore.Invalid", illegalArgument, "err.ob.backup.ha.low.thread.score.invalid")
	ErrObBackupBindingInvalid               = NewErrorCode("OB.Backup.Binding.Invalid", illegalArgument, "err.ob.backup.binding.invalid")
	ErrObBackupDeletePolicyInvalid          = NewErrorCode("OB.Backup.DeletePolicy.Invalid", illegalArgument, "err.ob.backup.delete.policy.invalid")
	ErrObBackupPieceSwitchIntervalInvalid   = NewErrorCode("OB.Backup.PieceSwitchInterval.Invalid", illegalArgument, "err.ob.backup.piece.switch.interval.invalid")
	ErrObBackupArchiveLagTargetInvalid      = NewErrorCode("OB.Backup.ArchiveLagTarget.Invalid", illegalArgument, "err.ob.backup.archive.lag.target.invalid")
	ErrObBackupArchiveLagTargetForS3Invalid = NewErrorCode("OB.Backup.ArchiveLagTargetForS3.Invalid", illegalArgument, "err.ob.backup.archive.lag.target.for.s3.invalid")
	ErrObBackupModeInvalid                  = NewErrorCode("OB.Backup.Mode.Invalid", illegalArgument, "err.ob.backup.mode.invalid")
	ErrObBackupStatusInvalid                = NewErrorCode("OB.Backup.Status.Invalid", illegalArgument, "err.ob.backup.status.invalid")
	ErrObBackupArchiveLogStatusInvalid      = NewErrorCode("OB.Backup.ArchiveLogStatus.Invalid", illegalArgument, "err.ob.backup.archive.log.status.invalid")
	ErrObBackupArchiveDestEmpty             = NewErrorCode("OB.Backup.ArchiveDestEmpty", illegalArgument, "err.ob.backup.archive.dest.empty")
	ErrObBackupDataDestEmpty                = NewErrorCode("OB.Backup.DataDestEmpty", illegalArgument, "err.ob.backup.data.dest.empty")

	// Ob.Restore
	ErrObStorageURIInvalid         = NewErrorCode("OB.Storage.URI.Invalid", illegalArgument, "err.ob.storage.uri.invalid")
	ErrObRestoreNotRecovering      = NewErrorCode("OB.Restore.NotRecovering", illegalArgument, "err.ob.restore.not.recovering")
	ErrObRestoreTimeNotValid       = NewErrorCode("OB.Restore.TimeNotValid", illegalArgument, "err.ob.restore.time.not.valid")
	ErrObRestoreTaskNotExist       = NewErrorCode("OB.Restore.Task.NotExist", illegalArgument, "err.ob.restore.task.not.exist")
	ErrObRestoreTaskAlreadySucceed = NewErrorCode("OB.Restore.Task.AlreadySucceed", illegalArgument, "err.ob.restore.task.already.succeed")

	// OB.Cluster
	ErrObClusterUnderMaintenance                 = NewErrorCode("OB.Cluster.UnderMaintenance", known, "err.ob.cluster.under.maintenance")
	ErrObClusterUnderMaintenanceWithDag          = NewErrorCode("OB.Cluster.UnderMaintenanceWithDag", known, "err.ob.cluster.under.maintenance.with.dag")
	ErrObClusterPasswordEncrypted                = NewErrorCode("OB.Cluster.Password.Encrypted", illegalArgument, "err.ob.cluster.password.encrypted")
	ErrObClusterIdInvalid                        = NewErrorCode("OB.Cluster.Id.Invalid", illegalArgument, "err.ob.cluster.id.invalid")
	ErrObClusterScopeInvalid                     = NewErrorCode("OB.Cluster.Scope.Invalid", illegalArgument, "err.ob.cluster.scope.invalid")
	ErrObClusterNameEmpty                        = NewErrorCode("OB.Cluster.Name.Empty", illegalArgument, "err.ob.cluster.name.empty")
	ErrObClusterAlreadyInitialized               = NewErrorCode("OB.Cluster.AlreadyInitialized", illegalArgument, "err.ob.cluster.already.initialized")
	ErrObClusterNotInitialized                   = NewErrorCode("OB.Cluster.NotInitialized", illegalArgument, "err.ob.cluster.not.initialized")
	ErrObClusterMultiPaxosNotAlive               = NewErrorCode("OB.Cluster.MultiPaxosNotAlive", illegalArgument, "err.ob.cluster.multi.paxos.not.alive")
	ErrObClusterMysqlPortNotInitialized          = NewErrorCode("OB.Cluster.MysqlPortNotInitialized", unexpected, "err.ob.cluster.mysql.port.not.initialized")
	ErrObClusterScaleOutHigherVersion            = NewErrorCode("OB.Cluster.ScaleOutHigherVersion", illegalArgument, "err.ob.cluster.scale.out.higher.version")
	ErrObClusterScaleOutLowerVersion             = NewErrorCode("OB.Cluster.ScaleOutLowerVersion", illegalArgument, "err.ob.cluster.scale.out.lower.version")
	ErrObClusterScaleOutRetryCoordinateDagFailed = NewErrorCode("OB.Cluster.ScaleOutRetryCoordinateDagFailed", unexpected, "err.ob.cluster.scale.out.retry.coordinate.dag.failed")
	ErrObClusterMinorFreezeTimeout               = NewErrorCode("OB.Cluster.MinorFreezeTimeout", unexpected, "err.ob.cluster.minor.freeze.timeout")
	ErrObClusterAsyncOperationTimeout            = NewErrorCode("OB.Cluster.AsyncOperationTimeout", unexpected, "err.ob.cluster.async.operation.timeout")
	ErrObClusterStopModeConflict                 = NewErrorCode("OB.Cluster.StopModeConflict", illegalArgument, "err.ob.cluster.stop.mode.conflict")
	ErrObClusterForceStopRequired                = NewErrorCode("OB.Cluster.ForceStopRequired", illegalArgument, "err.ob.cluster.force.stop.required")
	ErrObClusterForceStopOrTerminateRequired     = NewErrorCode("OB.Cluster.ForceStopOrTerminateRequired", illegalArgument, "err.ob.cluster.force.stop.or.terminate.required")
	ErrObClusterPasswordIncorrect                = NewErrorCode("OB.Cluster.Password.Incorrect", illegalArgument, "err.ob.cluster.password.incorrect") // "password incorrect"
	// OB.Server
	ErrObServerDeleteSelf         = NewErrorCode("OB.Server.DeleteSelf", illegalArgument, "err.ob.server.delete.self")
	ErrObServerProcessCheckFailed = NewErrorCode("OB.Server.Process.CheckFailed", unexpected, "err.ob.server.process.check.failed")      // "check observer process exist: %s."
	ErrObServerProcessNotExist    = NewErrorCode("OB.Server.Process.NotExist", unexpected, "err.ob.server.process.not.exist")            // "observer process not exist"
	ErrObServerNotExist           = NewErrorCode("OB.Server.NotExist", notFound, "err.ob.server.not.exist")                              // "observer '%s' is not exist"
	ErrObServerNotDeleting        = NewErrorCode("OB.Server.NotDeleting", unexpected, "err.ob.server.not.deleting")                      // "observer '%s' is not deleting, status: %s"
	ErrObServerHasNotBeenStarted  = NewErrorCode("OB.Server.HasNotBeenStarted", unexpected, "err.ob.server.has.not.been.started")        // "observer has not started yet, please start it with normal way"
	ErrObServerUnavailable        = NewErrorCode("OB.Server.Unavailable", unexpected, "err.ob.server.unavailable")                       // "observer '%s' is not available"
	ErrObServerStoppedInMultiZone = NewErrorCode("OB.Server.StoppedInMultiZone", illegalArgument, "err.ob.server.stopped.in.multi.zone") // "cannot stop server or stop zone in multiple zones"

	// OB.Parameter
	ErrObParameterScopeInvalid  = NewErrorCode("OB.Parameter.Scope.Invalid", illegalArgument, "err.ob.parameter.scope.invalid")
	ErrObParameterRsListInvalid = NewErrorCode("OB.Parameter.RsList.Invalid", illegalArgument, "err.ob.parameter.rs.list.invalid")

	// OB.Zone
	ErrObZoneNotExist   = NewErrorCode("OB.Zone.NotExist", badRequest, "err.ob.zone.not.exist")          // "zone '%s' is not exist"
	ErrObZoneNotEmpty   = NewErrorCode("OB.Zone.NotEmpty", illegalArgument, "err.ob.zone.not.empty")     // "The zone '%s' is not empty and can not be deleted"
	ErrObZoneDeleteSelf = NewErrorCode("OB.Zone.DeleteSelf", illegalArgument, "err.ob.zone.delete.self") // "The current agent is in '%s', please initiate the request through another agent."
	ErrObZoneNameEmpty  = NewErrorCode("OB.Zone.Name.Empty", illegalArgument, "err.ob.zone.name.empty")

	// OB.Package
	ErrObPackageNameNotSupport = NewErrorCode("OB.Package.Name.NotSupport", illegalArgument, "err.ob.package.name.not.support")
	ErrObPackageMissingFile    = NewErrorCode("OB.Package.MissingFile", unexpected, "err.ob.package.missing.file")
	ErrObPackageNotExist       = NewErrorCode("OB.Package.NotExist", badRequest, "err.ob.package.not.exist")
	ErrObPackageCorrupted      = NewErrorCode("OB.Package.Corrupted", unexpected, "err.ob.package.corrupted")

	ErrPackageReleaseFormatInvalid    = NewErrorCode("Package.ReleaseFormat.Invalid", illegalArgument, "err.package.release.format.invalid")
	ErrPackageCompressionNotSupported = NewErrorCode("Package.Compression.NotSupported", illegalArgument, "err.package.compression.not.supported")
	ErrPackageFormatInvalid           = NewErrorCode("Package.Format.Invalid", illegalArgument, "err.package.format.invalid")

	ErrObUpgradeToLowerVersion         = NewErrorCode("OB.Upgrade.ToLowerVersion", illegalArgument, "err.ob.upgrade.to.lower.version")
	ErrObUpgradeDepYamlMissing         = NewErrorCode("OB.Upgrade.DepYamlMissing", unexpected, "err.ob.upgrade.dep.yaml.missing")
	ErrObUpgradeModeNotSupported       = NewErrorCode("OB.Upgrade.Mode.NotSupported", illegalArgument, "err.ob.upgrade.mode.not.supported")
	ErrObUpgradeUnableToRollingUpgrade = NewErrorCode("OB.Upgrade.UnableToRollingUpgrade", illegalArgument, "err.ob.upgrade.unable.to.rolling.upgrade")
	ErrObUpgradeToDeprecatedVersion    = NewErrorCode("OB.Upgrade.ToDeprecatedVersion", illegalArgument, "err.ob.upgrade.to.deprecated.version")
	ErrObUpgradePathNotExist           = NewErrorCode("OB.Upgrade.Path.NotExist", illegalArgument, "err.ob.upgrade.path.not.exist")

	// Agent
	ErrAgentCoordinatorIsFaulty         = NewErrorCode("Agent.Coordinator.IsFaulty", unexpected, "err.agent.coordinator.is.faulty")
	ErrAgentCoordinatorNotInitialized   = NewErrorCode("Agent.Coordinator.NotInitialized", unexpected, "err.agent.coordinator.not.initialized")
	ErrAgentSynchronizerNotInitialized  = NewErrorCode("Agent.Synchronizer.NotInitialized", unexpected, "err.agent.synchronizer.not.initialized")
	ErrAgentMaintainerNotActive         = NewErrorCode("Agent.Maintainer.NotActive", unexpected, "err.agent.maintainer.not.active")
	ErrAgentMaintainerNotExist          = NewErrorCode("Agent.MaintainerNotExist", unexpected, "err.agent.maintainer.not.exist")
	ErrAgentNotExist                    = NewErrorCode("Agent.NotExist", badRequest, "err.agent.not.exist") // "server '%s' not exist"
	ErrAgentOceanbaseNotHold            = NewErrorCode("Agent.OceanBase.NotHold", unexpected, "err.agent.oceanbase.not.hold")
	ErrAgentOceanbaseUesless            = NewErrorCode("Agent.OceanBase.Useless", unexpected, "err.agent.oceanbase.useless")
	ErrAgentOceanbaseDBNotOcs           = NewErrorCode("Agent.OceanBase.DB.NotOcs", unexpected, "err.agent.oceanbase.db.not.ocs")
	ErrAgentSqliteDBNotInit             = NewErrorCode("Agent.Sqlite.DB.NotInit", unexpected, "err.agent.sqlite.db.not.init")
	ErrAgentAlreadyExists               = NewErrorCode("Agent.AlreadyExists", unexpected, "err.agent.already.exists")
	ErrAgentResponseDataEmpty           = NewErrorCode("Agent.Response.DataEmpty", unexpected, "err.agent.response.data.empty")
	ErrAgentResponseDataFormatInvalid   = NewErrorCode("Agent.Response.DataFormatInvalid", unexpected, "err.agent.response.data.format.invalid")
	ErrAgentIdentifyNotSupportOperation = NewErrorCode("Agent.Identify.NotSupportOperation", badRequest, "err.agent.identify.not.support.operation")
	ErrAgentIdentifyUnknown             = NewErrorCode("Agent.Identify.Unknown", unexpected, "err.agent.identify.unknown")
	ErrAgentVersionInconsistent         = NewErrorCode("Agent.Version.Inconsistent", badRequest, "err.agent.version.inconsistent")
	ErrAgentOBVersionInconsistent       = NewErrorCode("Agent.OB.Version.Inconsistent", badRequest, "err.agent.ob.version.inconsistent")
	ErrAgentNoMaster                    = NewErrorCode("Agent.NoMaster", unexpected, "err.agent.no.master")
	ErrAgentNotUnderMaintenance         = NewErrorCode("Agent.NotUnderMaintenance", illegalArgument, "err.agent.not.under.maintenance")
	ErrAgentUnderMaintenance            = NewErrorCode("Agent.UnderMaintenance", known, "err.agent.under.maintenance")
	ErrAgentCurrentUnderMaintenance     = NewErrorCode("Agent.Current.UnderMaintenance", known, "err.agent.current.under.maintenance")
	ErrAgentUnderMaintenanceDag         = NewErrorCode("Agent.UnderMaintenanceDag", known, "err.agent.under.maintenance.dag")
	ErrAgentUnavailable                 = NewErrorCode("Agent.Unavailable", unexpected, "err.agent.unavailable")
	ErrAgentAddressInvalid              = NewErrorCode("Agent.Address.Invalid", illegalArgument, "err.agent.address.invalid")
	ErrAgentOBVersionNotSupported       = NewErrorCode("Agent.OBVersionNotSupported", badRequest, "err.agent.ob.version.not.supported")
	ErrAgentNoActiveServer              = NewErrorCode("Agent.NoActiveServer", unexpected, "err.agent.no.active.server")

	// Agent.Upgrade
	ErrAgentUpgradeToLowerVersion = NewErrorCode("Agent.Upgrade.ToLowerVersion", illegalArgument, "err.agent.upgrade.to.lower.version")
	ErrAgentPackageNotFound       = NewErrorCode("Agent.Package.NotFound", unexpected, "err.agent.package.not.found")
	ErrAgentBinaryNotFound        = NewErrorCode("Agent.Binary.NotFound", unexpected, "err.agent.binary.not.found")

	// Agent.TakeOver
	ErrAgentTakeOverHigherVersion     = NewErrorCode("Agent.TakeOver.HigherVersion", badRequest, "err.agent.take.over.higher.version")
	ErrAgentTakeOverNotExistInCluster = NewErrorCode("Agent.TakeOver.NotExistInCluster", unexpected, "err.agent.take.over.not.exist.in.cluster")
	ErrAgentRebuildPortNotSame        = NewErrorCode("Agent.Rebuild.PortNotSame", badRequest, "err.agent.rebuild.port.not.same")
	ErrAgentRebuildVersionNotSame     = NewErrorCode("Agent.Rebuild.VersionNotSame", badRequest, "err.agent.rebuild.version.not.same")

	// Agent.Environment
	ErrEnvironmentWithoutPython       = NewErrorCode("Environment.WithoutPython", unexpected, "err.environment.without.python")
	ErrEnvironmentWithoutPythonModule = NewErrorCode("Environment.WithoutModule", unexpected, "err.environment.without.module")
	ErrEnvironmentDiskSpaceNotEnough  = NewErrorCode("Environment.DiskSpaceNotEnough", unexpected, "err.environment.disk.space.not.enough")
	ErrEnvironmentWithoutObAdmin      = NewErrorCode("Environment.WithoutObAdmin", unexpected, "err.environment.without.ob.admin")

	// Ob.Database
	ErrObDatabaseNotExist = NewErrorCode("OB.Database.NotExist", notFound, "err.ob.database.not.exist")

	// Ob.User
	ErrObUserPrivilegeNotSupported = NewErrorCode("OB.User.Privilege.NotSupported", illegalArgument, "err.ob.user.privilege.not.supported")
	ErrObUserNameEmpty             = NewErrorCode("OB.User.Name.Empty", illegalArgument, "err.ob.user.name.empty")

	// Request
	ErrRequestFileMissing                        = NewErrorCode("Request.File.Missing", unexpected, "err.request.file.missing")
	ErrRequestForwardHeaderNotExist              = NewErrorCode("Request.Forward.Header.NotExist", unexpected, "err.request.forward.header.not.exist")
	ErrRequestForwardMasterAgentNotFound         = NewErrorCode("Request.Forward.MasterAgent.NotFound", unexpected, "err.request.forward.master.agent.not.found")
	ErrRequestMethodNotSupport                   = NewErrorCode("Request.Method.NotSupport", illegalArgument, "err.request.method.not.support")
	ErrRequestBodyReadFailed                     = NewErrorCode("Request.Body.ReadFailed", badRequest, "err.request.body.read.failed")
	ErrRequestBodyDecryptSm4NoKey                = NewErrorCode("Request.Body.Decrypt.SM4.NoKey", unexpected, "err.request.body.decrypt.sm4.no.key")
	ErrRequestBodyDecryptAesNoKey                = NewErrorCode("Request.Body.Decrypt.AES.NoKey", unexpected, "err.request.body.decrypt.aes.no.key")
	ErrRequestBodyDecryptAesKeyAndIvInvalid      = NewErrorCode("Request.Body.Decrypt.AES.KeyAndIv.Invalid", unexpected, "err.request.body.decrypt.aes.key.and.iv.invalid")
	ErrRequestBodyDecryptAesContentLengthInvalid = NewErrorCode("Request.Body.Decrypt.AES.ContentLength.Invalid", unexpected, "err.request.body.decrypt.aes.content.length.invalid")
	ErrRequestHeaderTypeInvalid                  = NewErrorCode("Request.Header.Type.Invalid", unexpected, "err.request.header.type.invalid")
	ErrRequestPathParamEmpty                     = NewErrorCode("Request.Path.Param.Empty", badRequest, "err.request.path.param.empty")
	ErrRequestQueryParamEmpty                    = NewErrorCode("Request.Query.Param.Empty", badRequest, "err.request.query.param.empty")
	ErrRequestQueryParamIllegal                  = NewErrorCode("Request.Query.Param.Illegal", badRequest, "err.request.query.param.illegal")
	ErrRequestHeaderNotFound                     = NewErrorCode("Request.Header.NotFound", badRequest, "err.request.header.not.found")

	// Ob.OBProxy
	ErrOBProxyAlreadyManaged                = NewErrorCode("OBProxy.AlreadyManaged", illegalArgument, "err.obproxy.already.managed")
	ErrOBProxyRsListAndConfigUrlConflicted  = NewErrorCode("OBProxy.RsListAndConfigUrl.Conflicted", illegalArgument, "err.obproxy.rs.list.and.config.url.conflicted") // "rs_list and config_url cannot be specified at the same time"
	ErrOBProxyRsListOrConfigUrlNotSpecified = NewErrorCode("OBProxy.RsListOrConfigUrl.NotSpecified", illegalArgument, "err.obproxy.rs.list.or.config.url.not.specified")
	ErrOBProxyVersionNotSupported           = NewErrorCode("OBProxy.Version.NotSupported", illegalArgument, "err.obproxy.version.not.supported")
	ErrOBProxyUpgradeToLowerVersion         = NewErrorCode("OBProxy.UpgradeToLowerVersion", badRequest, "err.obproxy.upgrade.to.lower.version")
	ErrOBProxyHealthCheckTimeout            = NewErrorCode("OBProxy.HealthCheckTimeout", illegalArgument, "err.obproxy.health.check.timeout")
	ErrOBProxyNotBeManaged                  = NewErrorCode("OBProxy.NotBeManaged", illegalArgument, "err.obproxy.not.be.managed")
	ErrOBProxyStopTimeout                   = NewErrorCode("OBProxy.StopTimeout", illegalArgument, "err.obproxy.stop.timeout")
	ErrOBProxyStopDaemonTimeout             = NewErrorCode("OBProxy.StopDaemonTimeout", illegalArgument, "err.obproxy.stop.daemon.timeout")
	ErrOBProxyHotRestartTimeout             = NewErrorCode("OBProxy.HotRestartTimeout", illegalArgument, "err.obproxy.hot.restart.timeout")
	ErrOBProxyNotRunning                    = NewErrorCode("OBProxy.NotRunning", illegalArgument, "err.obproxy.not.running")
	ErrOBProxyPackageNotFound               = NewErrorCode("OBProxy.Package.NotFound", illegalArgument, "err.obproxy.pkg.not.found")
	ErrOBProxyVersionOutputUnexpected       = NewErrorCode("OBProxy.VersionOutputUnexpected", unexpected, "err.obproxy.version.output.unexpected")
	ErrOBProxyPackageNameInvalid            = NewErrorCode("OBProxy.Package.Name.Invalid", illegalArgument, "err.obproxy.package.name.invalid")
	ErrOBProxyPackageMissingFile            = NewErrorCode("OBProxy.Package.NotFound", unexpected, "err.obproxy.package.missing.file")

	// Security
	ErrSecurityUserPermissionDenied                      = NewErrorCode("Security.User.PermissionDenied", unauthorized, "err.security.user.permission.denied")
	ErrSecurityAuthenticationUnauthorized                = NewErrorCode("Security.Authentication.Unauthorized", unauthorized, "err.security.authentication.unauthorized")
	ErrSecurityAuthenticationFileSha256Mismatch          = NewErrorCode("Security.Authentication.File.Sha256Mismatch", unauthorized, "err.security.authentication.file.sha256.mismatch")
	ErrSecurityDecryptFailed                             = NewErrorCode("Security.DecryptFailed", unexpected, "err.security.decrypt.failed")
	ErrSecurityAuthenticationHeaderDecryptFailed         = NewErrorCode("Security.Authentication.Header.DecryptFailed", unexpected, "err.security.authentication.header.decrypt.failed")
	ErrSecurityAuthenticationHeaderUriMismatch           = NewErrorCode("Security.Authentication.Header.UriMismatch", unauthorized, "err.security.authentication.header.uri.mismatch")
	ErrSecurityAuthenticationWithAgentPassword           = NewErrorCode("Security.Authentication.WithAgentPassword", illegalArgument, "err.security.authentication.with.agent.password")
	ErrSecurityAuthenticationIncorrectAgentPassword      = NewErrorCode("Security.Authentication.IncorrectAgentPassword", illegalArgument, "err.security.authentication.incorrect.agent.password")
	ErrSecurityAuthenticationIncorrectOceanbasePassword  = NewErrorCode("Security.Authentication.IncorrectOceanbasePassword", illegalArgument, "err.security.authentication.incorrect.oceanbase.password", 10008)
	ErrSecurityAuthenticationUnknownPasswordType         = NewErrorCode("Security.Authentication.UnknownPasswordType", illegalArgument, "err.security.authentication.unknown.password.type")
	ErrSecurityAuthenticationExpired                     = NewErrorCode("Security.Authentication.Expired", illegalArgument, "err.security.authentication.expired")
	ErrSecurityAuthenticationTimestampInvalid            = NewErrorCode("Security.Authentication.Timestamp.Invalid", illegalArgument, "err.security.authentication.timestamp.invalid")
	ErrSecurityAuthenticationIncorrectToken              = NewErrorCode("Security.Authentication.IncorrectToken", illegalArgument, "err.security.authentication.incorrect.token")
	ErrSecurityAuthenticationWithOceanBasePassword       = NewErrorCode("Security.Authentication.WithOceanBasePassword", illegalArgument, "err.security.authentication.with.oceanbase.password")
	ErrSecurityAuthenticationAgentPasswordNotInitialized = NewErrorCode("Security.Authentication.AgentPasswordNotInitialized", illegalArgument, "err.security.authentication.agent.password.not.initialized")

	// Task
	ErrTaskExpired                         = NewErrorCode("Task.Expired", known, "err.task.expired")
	ErrTaskNotFound                        = NewErrorCode("Task.NotFound", notFound, "err.task.not.found")
	ErrTaskNotFoundWithReason              = NewErrorCode("Task.NotFound", notFound, "err.task.not.found.with.reason")
	ErrTaskCreateFailed                    = NewErrorCode("Task.CreateFailed", known, "err.task.create.failed")
	ErrTaskEmptyTemplate                   = NewErrorCode("Task.Template.Empty", unexpected, "err.task.template.empty")
	ErrTaskParamNotSet                     = NewErrorCode("Task.Param.NotSet", unexpected, "err.task.param.not.set")
	ErrTaskParamConvertFailed              = NewErrorCode("Task.Param.ConvertFailed", unexpected, "err.task.param.convert.failed")
	ErrTaskDataNotSet                      = NewErrorCode("Task.Data.NotSet", unexpected, "err.task.data.not.set")
	ErrTaskDataConvertFailed               = NewErrorCode("Task.Data.ConvertFailed", unexpected, "err.task.data.convert.failed")
	ErrTaskAgentDataNotSet                 = NewErrorCode("Task.AgentData.NotSet", unexpected, "err.task.agent.data.not.set")
	ErrTaskAgentDataConvertFailed          = NewErrorCode("Task.AgentData.ConvertFailed", unexpected, "err.task.agent.data.convert.failed")
	ErrTaskLocalDataNotSet                 = NewErrorCode("Task.LocalData.NotSet", unexpected, "err.task.local.data.not.set")
	ErrTaskLocalDataConvertFailed          = NewErrorCode("Task.LocalData.ConvertFailed", unexpected, "err.task.local.data.convert.failed")
	ErrTaskDagExecuteTimeout               = NewErrorCode("Task.Dag.ExecuteTimeout", unexpected, "err.task.dag.execute.timeout")
	ErrTaskDagCancelTimeout                = NewErrorCode("Task.Dag.CancelTimeout", unexpected, "err.task.dag.cancel.timeout")
	ErrTaskDagPassTimeout                  = NewErrorCode("Task.Dag.PassTimeout", unexpected, "err.task.dag.pass.timeout")
	ErrTaskSubDagNotAllSucceed             = NewErrorCode("Task.SubDag.NotAllSucceed", unexpected, "err.task.sub.dag.not.all.succeed")
	ErrTaskSubDagNotAllCreated             = NewErrorCode("Task.SubDag.NotAllCreated", unexpected, "err.task.sub.dag.not.all.created")
	ErrTaskSubDagNotAllPassed              = NewErrorCode("Task.SubDag.NotAllPassed", unexpected, "err.task.sub.dag.not.all.passed")
	ErrTaskSubDagNotAllAdvanced            = NewErrorCode("Task.SubDag.NotAllAdvanced", unexpected, "err.task.sub.dag.not.all.advanced")
	ErrTaskSubDagNotAllReady               = NewErrorCode("Task.SubDag.NotAllReady", unexpected, "err.task.sub.dag.not.all.ready")
	ErrTaskDagFailed                       = NewErrorCode("Task.DagFailed", unexpected, "err.task.dag.failed")
	ErrTaskGenericIDInvalid                = NewErrorCode("Task.GenericID.Invalid", illegalArgument, "err.task.generic.id.invalid")
	ErrTaskRemoteTaskFailed                = NewErrorCode("Task.RemoteTask.Failed", unexpected, "err.task.remote.task.failed")
	ErrTaskNodeOperatorNotSupport          = NewErrorCode("Task.Node.Operator.NotSupport", illegalArgument, "err.task.node.operator.not.support")
	ErrTaskDagStateInvalid                 = NewErrorCode("Task.Dag.State.Invalid", illegalArgument, "err.task.dag.state.invalid")
	ErrTaskDagOperatorNotSupport           = NewErrorCode("Task.Dag.Operator.NotSupport", illegalArgument, "err.task.dag.operator.not.support")
	ErrTaskDagOperatorRollbackNotFailedDag = NewErrorCode("Task.Dag.Operator.RollbackNotFailedDag", illegalArgument, "err.task.dag.operator.rollback.not.failed.dag")
	ErrTaskDagOperatorRollbackNotAllowed   = NewErrorCode("Task.Dag.Operator.RollbackNotAllowed", illegalArgument, "err.task.dag.operator.rollback.not.allowed")
	ErrTaskDagOperatorCancelFinishedDag    = NewErrorCode("Task.Dag.Operator.CancelFinishedDag", illegalArgument, "err.task.dag.operator.cancel.finished.dag")
	ErrTaskDagOperatorCancelNotAllowed     = NewErrorCode("Task.Dag.Operator.CancelNotAllowed", illegalArgument, "err.task.dag.operator.cancel.not.allowed")
	ErrTaskDagOperatorPassNotFailedDag     = NewErrorCode("Task.Dag.Operator.PassNotFailedDag", illegalArgument, "err.task.dag.operator.pass.not.failed.dag")
	ErrTaskDagOperatorPassNotAllowed       = NewErrorCode("Task.Dag.Operator.PassNotAllowed", illegalArgument, "err.task.dag.operator.pass.not.allowed")
	ErrTaskDagOperatorRetryNotFailedDag    = NewErrorCode("Task.Dag.Operator.RetryNotFailedDag", illegalArgument, "err.task.dag.operator.retry.not.failed.dag")
	ErrTaskDagOperatorRetryNotAllowed      = NewErrorCode("Task.Dag.Operator.RetryNotAllowed", illegalArgument, "err.task.dag.operator.retry.not.allowed")
	ErrTaskNodeOperatorPassNotFailedDag    = NewErrorCode("Task.Node.Operator.PassNotFailedDag", illegalArgument, "err.task.node.operator.pass.not.failed.dag")
	ErrTaskNodeOperatorPassNotFailedNode   = NewErrorCode("Task.Node.Operator.PassNotFailedNode", illegalArgument, "err.task.node.operator.pass.not.failed.node")
	ErrTaskNodeOperatorPassNotAllowed      = NewErrorCode("Task.Node.Operator.PassNotAllowed", illegalArgument, "err.task.node.operator.pass.not.allowed")
	ErrTaskEngineUnexpected                = NewErrorCode("Task.Engine.Unexpected", unexpected, "err.task.engine.unexpected")

	ErrGormNoRowAffected = NewErrorCode("Gorm.NoRowAffected", unexpected, "err.gorm.no.row.affected") // "%s: no row affected"

	ErrMysqlError = NewErrorCode("Mysql.Error", badRequest, "err.mysql.error") // "%s"

	ErrPackageNameMismatch   = NewErrorCode("Package.NameMismatch", illegalArgument, "err.package.name.mismatch")     // "rpm package name %s not match %s"
	ErrPackageReleaseInvalid = NewErrorCode("Package.ReleaseInvalid", illegalArgument, "err.package.release.invalid") // "rpm package release %s not match format"

	// cli
	ErrCliFlagRequired                          = NewErrorCode("Cli.FlagRequired", illegalArgument, "err.cli.flag.required")                                                    // "required flag(s) \"%s\" not set"
	ErrCliOperationCancelled                    = NewErrorCode("Cli.OperationCancelled", known, "err.cli.operation.cancelled")                                                  // "operation cancelled"
	ErrCliUsageError                            = NewErrorCode("Cli.UsageError", illegalArgument, "err.cli.usage.error")                                                        // "Incorrect usage: %s"
	ErrCliNotFound                              = NewErrorCode("Cli.NotFound", notFound, "err.cli.not.found")                                                                   // "not found: %s"
	ErrCliUpgradePackageNotFoundInPath          = NewErrorCode("Cli.Upgrade.PackageNotFoundInPath", badRequest, "err.cli.upgrade.package.not.found.in.path")                    // "no valid %s package found in %s"
	ErrCliTakeOverWithObserverNotInConf         = NewErrorCode("Cli.TakeOver.WithObserverNotInConf", unexpected, "err.cli.take.over.with.observer.not.in.conf")                 // "server %s is not in the ob conf"
	ErrCliTakeOverMultiServerOnSameHost         = NewErrorCode("Cli.TakeOver.MultiServerOnSameHost", unexpected, "err.cli.take.over.multi.server.on.same.host")                 // "multi-server on the same host when take over by 'cluster start'"
	ErrCliObClusterNotTakenOver                 = NewErrorCode("Cli.ObCluster.NotTakenOver", unexpected, "err.cli.ob.cluster.not.taken.over")                                   // "Cluster not taken over. Run 'obshell cluster start -a' to start it."
	ErrCliUpgradeNoValidTargetBuildVersionFound = NewErrorCode("Cli.Upgrade.NoValidTargetBuildVersionFound", unexpected, "err.cli.upgrade.no.valid.target.build.version.found") // "no valid target build version found by '%s'"
	ErrCliStartRemoteAgentFailed                = NewErrorCode("Cli.StartRemoteAgentFailed", unexpected, "err.cli.start.remote.agent.failed")                                   // "failed to start remote agent"
	ErrCliUnixSocketRequestFailed               = NewErrorCode("Cli.UnixSocket.RequestFailed", unexpected, "err.cli.unix.socket.request.failed")                                // "request unix-socket [%s]%s failed: %v"
	ErrEmpty                                    = NewErrorCode("Empty", unexpected, "err.empty")                                                                                // this error code won't be display

	// 启动相关
	ErrAgentUnixSocketListenerCreateFailed = NewErrorCode("Agent.Unix.Socket.Listener.CreateFailed", unexpected, "err.agent.unix.socket.listener.create.failed") // "create unix socket listerner failed"
	ErrAgentTCPListenerCreateFailed        = NewErrorCode("Agent.TCP.Listener.CreateFailed", unexpected, "err.agent.tcp.listener.create.failed")                 // "create tcp listerner failed"
	ErrAgentAlreadyInitialized             = NewErrorCode("Agent.AlreadyInitialized", unexpected, "err.agent.already.initialized")                               // "agent already initialized"
	ErrAgentNotInitialized                 = NewErrorCode("Agent.NotInitialized", unexpected, "err.agent.not.initialized")                                       // "agent not initialized"
	ErrAgentIpInconsistentWithOBServer     = NewErrorCode("Agent.IP.InconsistentWithOBServer", unexpected, "err.agent.ip.inconsistent.with.ob.server")           // "agent ip inconsistent with observer"
	ErrAgentLoadOBConfigFailed             = NewErrorCode("Agent.Load.OBConfigFailed", unexpected, "err.agent.load.ob.config.failed")                            // "load ob config from config file failed"
	ErrAgentInfoNotEqual                   = NewErrorCode("Agent.Info.NotEqual", unexpected, "err.agent.info.not.equal")                                         // "agent info not equal"
	ErrAgentStartWithInvalidInfo           = NewErrorCode("Agent.Start.WithInvalidInfo", unexpected, "err.agent.start.with.invalid.info")                        // "agent start with invalid info: %v"
	ErrAgentNeedToTakeOver                 = NewErrorCode("Agent.NeedToTakeOver", illegalArgument, "err.agent.need.to.takeover")                                 // "obshell need to be cluster. Please do takeover first."
	ErrAgentStartObserverFailed            = NewErrorCode("Agent.Start.ObserverFailed", unexpected, "err.agent.start.observer.failed")                           // "start observer via flag failed, err: %v"
	ErrAgentTakeOverFailed                 = NewErrorCode("Agent.TakeOverFailed", unexpected, "err.agent.take.over.failed")                                      // "take over or rebuild failed: %v"
	ErrAgentServeOnUnixSocketFailed        = NewErrorCode("Agent.ServeOnUnixSocketFailed", unexpected, "err.agent.serve.on.unix.socket.failed")                  // "serve on unix listener failed: %v\n"
	ErrAgentServeOnTcpSocketFailed         = NewErrorCode("Agent.ServeOnTcpSocketFailed", unexpected, "err.agent.serve.on.tcp.socket.failed")                    // "serve on tcp listener failed: %v\n"
	ErrAgenDaemonServeOnUnixSocketFailed   = NewErrorCode("AgenDaemon.ServeOnUnixSocketFailed", unexpected, "err.agen.daemon.serve.on.unix.socket.failed")       // "daemon serve on socket listener failed: %s\n"
	ErrAgentOceanbasePasswordLoadFailed    = NewErrorCode("Agent.Oceanbase.Password.LoadFailed", unexpected, "err.agent.oceanbase.password.load.failed")         // "check password of root@sys in sqlite failed: not cluster agent"
	ErrAgentUpgradeKillOldServerTimeout    = NewErrorCode("Agent.Upgrade.KillOldServerTimeout", unexpected, "err.agent.upgrade.kill.old.server.timeout")         // "wait obshell server killed timeout"
)
