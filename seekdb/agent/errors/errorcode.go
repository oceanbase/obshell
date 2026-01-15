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
	ErrCommonUnauthorized               = NewErrorCode("Common.Unauthorized", unauthorized, "err.common.unauthorized", 10008)                 // "unauthorized"
	ErrCommonInvalidTimeDuration        = NewErrorCode("Common.InvalidTimeDuration", illegalArgument, "err.common.invalid.time.duration")     // "time duration '%s' is invalid: %s"
	ErrJsonMarshal                      = NewErrorCode("Common.JsonMarshal", unexpected, "err.common.json.marshal")                           // "json marshal failed: %s"
	ErrJsonUnmarshal                    = NewErrorCode("Common.JsonUnmarshal", unexpected, "err.common.json.unmarshal")                       // "json unmarshal failed: %s"

	// Log
	ErrLogWriteExceedMaxSize          = NewErrorCode("Log.WriteExceedMaxSize", unexpected, "err.log.write.exceed.max.size")                   // "write length %d exceeds maximum file size %d"
	ErrLogFileNamePrefixMismatched    = NewErrorCode("Log.FileNamePrefixMismatched", unexpected, "err.log.file.name.prefix.mismatched")       // "file name '%s' prefix mismatched"
	ErrLogFileNameExtensionMismatched = NewErrorCode("Log.FileNameExtensionMismatched", unexpected, "err.log.file.name.extension.mismatched") // "file name '%s' extension mismatched"

	// RPC
	ErrAgentRPCRequestError  = NewErrorCode("Agent.RPC.RequestError", unexpected, "err.agent.rpc.request.error")   // "request [%s]%s to %s error: %v"
	ErrAgentRPCRequestFailed = NewErrorCode("Agent.RPC.RequestFailed", unexpected, "err.agent.rpc.request.failed") // "request [%s]%s to %s failed: %s"

	// OB.Binary
	ErrObBinaryVersionUnexpected = NewErrorCode("seekdb.Binary.Version.Unexpected", unexpected, "err.seekdb.binary.version.unexpected")

	// OB.Tenant
	ErrObTenantCompactionStatusNotIdle = NewErrorCode("seekdb.Compaction.Status.NotIdle", badRequest, "err.seekdb.compaction.status.not.idle") // "Instance is in '%s' status, operation not allowed"
	ErrObEmptyVariable                 = NewErrorCode("seekdb.Variable.Empty", illegalArgument, "err.seekdb.variable.empty")                   // "variable name or value is empty"
	ErrObEmptyParameter                = NewErrorCode("seekdb.Parameter.Empty", illegalArgument, "err.seekdb.parameter.empty")                 // "parameter name or value is empty"
	ErrObParameterNotExist             = NewErrorCode("seekdb.Parameter.NotExist", badRequest, "err.seekdb.parameter.not.exist")               // "parameter '%s' is not exist"
	ErrObVariableInvalid               = NewErrorCode("seekdb.Variable.Invalid", illegalArgument, "err.seekdb.variable.invalid")               // "variable '%s' is invalid: %s"
	ErrObVariableNotExist              = NewErrorCode("seekdb.Variable.NotExist", notFound, "err.seekdb.variable.not.exist")                   // "variable '%s' is not exist"

	// OB.Cluster
	ErrObClusterUnderMaintenance        = NewErrorCode("seekdb.UnderMaintenance", known, "err.seekdb.under.maintenance")
	ErrObClusterUnderMaintenanceWithDag = NewErrorCode("seekdb.UnderMaintenanceWithDag", known, "err.seekdb.under.maintenance.with.dag")
	ErrObClusterNotInitialized          = NewErrorCode("seekdb.NotInitialized", illegalArgument, "err.seekdb.not.initialized")
	ErrObClusterMysqlPortNotInitialized = NewErrorCode("seekdb.MysqlPortNotInitialized", unexpected, "err.seekdb.mysql.port.not.initialized")
	ErrObClusterMinorFreezeTimeout      = NewErrorCode("seekdb.MinorFreezeTimeout", unexpected, "err.seekdb.minor.freeze.timeout")
	ErrObClusterAsyncOperationTimeout   = NewErrorCode("seekdb.AsyncOperationTimeout", unexpected, "err.seekdb.async.operation.timeout")
	ErrObClusterPasswordIncorrect       = NewErrorCode("seekdb.Password.Incorrect", illegalArgument, "err.seekdb.password.incorrect") // "password incorrect"

	// OB.Server
	ErrObServerProcessCheckFailed = NewErrorCode("seekdb.Process.CheckFailed", unexpected, "err.seekdb.process.check.failed") // "check seekdb process exist: %s."
	ErrObServerProcessNotExist    = NewErrorCode("seekdb.Process.NotExist", unexpected, "err.seekdb.process.not.exist")       // "seekdb process not exist"
	ErrObServerHasNotBeenStarted  = NewErrorCode("seekdb.HasNotBeenStarted", unexpected, "err.seekdb.has.not.been.started")   // "seekdb has not started yet, please start it with normal way"

	// OB.Package
	ErrObPackageNameNotSupport = NewErrorCode("seekdb.Package.Name.NotSupport", illegalArgument, "err.seekdb.package.name.not.support")
	ErrObPackageMissingFile    = NewErrorCode("seekdb.Package.MissingFile", unexpected, "err.seekdb.package.missing.file")

	ErrPackageReleaseFormatInvalid    = NewErrorCode("Package.ReleaseFormat.Invalid", illegalArgument, "err.package.release.format.invalid")
	ErrPackageCompressionNotSupported = NewErrorCode("Package.Compression.NotSupported", illegalArgument, "err.package.compression.not.supported")
	ErrPackageFormatInvalid           = NewErrorCode("Package.Format.Invalid", illegalArgument, "err.package.format.invalid")

	// Agent
	ErrAgentOceanbaseNotHold            = NewErrorCode("Agent.OceanBase.NotHold", unexpected, "err.agent.oceanbase.not.hold")
	ErrAgentOceanbaseUesless            = NewErrorCode("Agent.OceanBase.Useless", unexpected, "err.agent.oceanbase.useless")
	ErrAgentOceanbaseDBNotOcs           = NewErrorCode("Agent.OceanBase.DB.NotOcs", unexpected, "err.agent.oceanbase.db.not.ocs")
	ErrAgentSqliteDBNotInit             = NewErrorCode("Agent.Sqlite.DB.NotInit", unexpected, "err.agent.sqlite.db.not.init")
	ErrAgentResponseDataEmpty           = NewErrorCode("Agent.Response.DataEmpty", unexpected, "err.agent.response.data.empty")
	ErrAgentResponseDataFormatInvalid   = NewErrorCode("Agent.Response.DataFormatInvalid", unexpected, "err.agent.response.data.format.invalid")
	ErrAgentIdentifyNotSupportOperation = NewErrorCode("Agent.Identify.NotSupportOperation", badRequest, "err.agent.identify.not.support.operation")
	ErrAgentIdentifyUnknown             = NewErrorCode("Agent.Identify.Unknown", unexpected, "err.agent.identify.unknown")
	ErrAgentUnderMaintenance            = NewErrorCode("Agent.UnderMaintenance", known, "err.agent.under.maintenance")
	ErrAgentCurrentUnderMaintenance     = NewErrorCode("Agent.Current.UnderMaintenance", known, "err.agent.current.under.maintenance")
	ErrAgentUnderMaintenanceDag         = NewErrorCode("Agent.UnderMaintenanceDag", known, "err.agent.under.maintenance.dag")
	ErrAgentOBVersionNotSupported       = NewErrorCode("Agent.OBVersionNotSupported", badRequest, "err.agent.ob.version.not.supported")
	ErrAgentBaseDirInvalid              = NewErrorCode("Agent.BaseDir.Invalid", illegalArgument, "err.agent.base.dir.invalid")

	// Agent.Upgrade
	ErrAgentUpgradeToLowerVersion = NewErrorCode("Agent.Upgrade.ToLowerVersion", illegalArgument, "err.agent.upgrade.to.lower.version")
	ErrAgentPackageNotFound       = NewErrorCode("Agent.Package.NotFound", unexpected, "err.agent.package.not.found")

	// Agent.TakeOver
	ErrAgentRebuildPortNotSame    = NewErrorCode("Agent.Rebuild.PortNotSame", badRequest, "err.agent.rebuild.port.not.same")
	ErrAgentRebuildVersionNotSame = NewErrorCode("Agent.Rebuild.VersionNotSame", badRequest, "err.agent.rebuild.version.not.same")

	// Agent.Environment
	ErrEnvironmentDiskSpaceNotEnough = NewErrorCode("Environment.DiskSpaceNotEnough", unexpected, "err.environment.disk.space.not.enough")

	// Ob.Database
	ErrObDatabaseNotExist    = NewErrorCode("seekdb.Database.NotExist", notFound, "err.seekdb.database.not.exist")
	ErrObDatabaseNameInvalid = NewErrorCode("seekdb.Database.Name.Invalid", illegalArgument, "err.seekdb.database.name.invalid")

	// Ob.User
	ErrObUserPrivilegeNotSupported = NewErrorCode("seekdb.User.Privilege.NotSupported", illegalArgument, "err.seekdb.user.privilege.not.supported")
	ErrObUserNameEmpty             = NewErrorCode("seekdb.User.Name.Empty", illegalArgument, "err.seekdb.user.name.empty")
	ErrObUserNameInvalid           = NewErrorCode("seekdb.User.Name.Invalid", illegalArgument, "err.seekdb.user.name.invalid")
	ErrObUserNotExists             = NewErrorCode("seekdb.User.NotExists", notFound, "err.seekdb.user.not.exist")
	ErrObUserPasswordError         = NewErrorCode("seekdb.User.Password.Error", illegalArgument, "err.seekdb.user.password.error")

	// Request
	ErrRequestFileMissing                        = NewErrorCode("Request.File.Missing", unexpected, "err.request.file.missing")
	ErrRequestMethodNotSupport                   = NewErrorCode("Request.Method.NotSupport", illegalArgument, "err.request.method.not.support")
	ErrRequestBodyReadFailed                     = NewErrorCode("Request.Body.ReadFailed", badRequest, "err.request.body.read.failed")
	ErrRequestBodyDecryptSm4NoKey                = NewErrorCode("Request.Body.Decrypt.SM4.NoKey", unexpected, "err.request.body.decrypt.sm4.no.key")
	ErrRequestBodyDecryptAesNoKey                = NewErrorCode("Request.Body.Decrypt.AES.NoKey", unexpected, "err.request.body.decrypt.aes.no.key")
	ErrRequestBodyDecryptAesKeyAndIvInvalid      = NewErrorCode("Request.Body.Decrypt.AES.KeyAndIv.Invalid", unexpected, "err.request.body.decrypt.aes.key.and.iv.invalid")
	ErrRequestBodyDecryptAesContentLengthInvalid = NewErrorCode("Request.Body.Decrypt.AES.ContentLength.Invalid", unexpected, "err.request.body.decrypt.aes.content.length.invalid")
	ErrRequestHeaderTypeInvalid                  = NewErrorCode("Request.Header.Type.Invalid", unexpected, "err.request.header.type.invalid")
	ErrRequestHeaderNotFound                     = NewErrorCode("Request.Header.NotFound", badRequest, "err.request.header.not.found")

	// Security
	ErrSecurityUserPermissionDenied                     = NewErrorCode("Security.User.PermissionDenied", unauthorized, "err.security.user.permission.denied")
	ErrSecurityAuthenticationUnauthorized               = NewErrorCode("Security.Authentication.Unauthorized", unauthorized, "err.security.authentication.unauthorized", 10008)
	ErrSecurityAuthenticationFileSha256Mismatch         = NewErrorCode("Security.Authentication.File.Sha256Mismatch", unauthorized, "err.security.authentication.file.sha256.mismatch")
	ErrSecurityAuthenticationHeaderDecryptFailed        = NewErrorCode("Security.Authentication.Header.DecryptFailed", unauthorized, "err.security.authentication.header.decrypt.failed")
	ErrSecurityAuthenticationHeaderUriMismatch          = NewErrorCode("Security.Authentication.Header.UriMismatch", unauthorized, "err.security.authentication.header.uri.mismatch", 10008)
	ErrSecurityAuthenticationIncorrectOceanbasePassword = NewErrorCode("Security.Authentication.IncorrectseekdbPassword", unauthorized, "err.security.authentication.incorrect.seekdb.password", 10008)
	ErrSecurityAuthenticationExpired                    = NewErrorCode("Security.Authentication.Expired", unauthorized, "err.security.authentication.expired", 10008)
	ErrSecurityAuthenticationTimestampInvalid           = NewErrorCode("Security.Authentication.Timestamp.Invalid", unauthorized, "err.security.authentication.timestamp.invalid", 10008)

	// Task
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
	ErrTaskGenericIDInvalid                = NewErrorCode("Task.GenericID.Invalid", illegalArgument, "err.task.generic.id.invalid")
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

	ErrGormNoRowAffected = NewErrorCode("Gorm.NoRowAffected", unexpected, "err.gorm.no.row.affected") // "%s: no row affected"

	ErrMysqlError = NewErrorCode("MySQL.Error", badRequest, "err.mysql.error") // "%s"

	ErrPackageNameMismatch   = NewErrorCode("Package.NameMismatch", illegalArgument, "err.package.name.mismatch")     // "rpm package name %s not match %s"
	ErrPackageReleaseInvalid = NewErrorCode("Package.ReleaseInvalid", illegalArgument, "err.package.release.invalid") // "rpm package release %s not match format"

	// cli
	ErrCliFlagRequired                          = NewErrorCode("Cli.FlagRequired", illegalArgument, "err.cli.flag.required")                                                    // "required flag(s) \"%s\" not set"
	ErrCliOperationCancelled                    = NewErrorCode("Cli.OperationCancelled", known, "err.cli.operation.cancelled")                                                  // "operation cancelled"
	ErrCliUsageError                            = NewErrorCode("Cli.UsageError", illegalArgument, "err.cli.usage.error")                                                        // "Incorrect usage: %s"
	ErrCliNotFound                              = NewErrorCode("Cli.NotFound", notFound, "err.cli.not.found")                                                                   // "not found: %s"
	ErrCliUpgradePackageNotFoundInPath          = NewErrorCode("Cli.Upgrade.PackageNotFoundInPath", badRequest, "err.cli.upgrade.package.not.found.in.path")                    // "no valid %s package found in %s"
	ErrCliUpgradeNoValidTargetBuildVersionFound = NewErrorCode("Cli.Upgrade.NoValidTargetBuildVersionFound", unexpected, "err.cli.upgrade.no.valid.target.build.version.found") // "no valid target build version found by '%s'"
	ErrCliUnixSocketRequestFailed               = NewErrorCode("Cli.UnixSocket.RequestFailed", unexpected, "err.cli.unix.socket.request.failed")                                // "request unix-socket [%s]%s failed: %v"
	ErrCliTargetObshellUnavailable              = NewErrorCode("Cli.TargetObshellUnavailable", unexpected, "err.cli.target.obshell.unavailable")                                // "target obshell unavailable"
	ErrEmpty                                    = NewErrorCode("Empty", unexpected, "err.empty")                                                                                // this error code won't be display

	// 启动相关
	ErrAgentUnixSocketListenerCreateFailed = NewErrorCode("Agent.Unix.Socket.Listener.CreateFailed", unexpected, "err.agent.unix.socket.listener.create.failed") // "create unix socket listerner failed"
	ErrAgentTCPListenerCreateFailed        = NewErrorCode("Agent.TCP.Listener.CreateFailed", unexpected, "err.agent.tcp.listener.create.failed")                 // "create tcp listerner failed"
	ErrAgentAlreadyInitialized             = NewErrorCode("Agent.AlreadyInitialized", unexpected, "err.agent.already.initialized")                               // "agent already initialized"
	ErrAgentNotInitialized                 = NewErrorCode("Agent.NotInitialized", unexpected, "err.agent.not.initialized")                                       // "agent not initialized"
	ErrAgentIpInconsistentWithOBServer     = NewErrorCode("Agent.IP.InconsistentWithOBServer", unexpected, "err.agent.ip.inconsistent.with.ob.server")           // "agent ip inconsistent with seekdb"
	ErrAgentLoadOBConfigFailed             = NewErrorCode("Agent.Load.OBConfigFailed", unexpected, "err.agent.load.ob.config.failed")                            // "load ob config from config file failed"
	ErrAgentInfoNotEqual                   = NewErrorCode("Agent.Info.NotEqual", unexpected, "err.agent.info.not.equal")                                         // "agent info not equal"
	ErrAgentStartWithInvalidInfo           = NewErrorCode("Agent.Start.WithInvalidInfo", unexpected, "err.agent.start.with.invalid.info")                        // "agent start with invalid info: %v"
	ErrAgentSeekDBNotExists                = NewErrorCode("Agent.seekdb.Not.Exists", illegalArgument, "err.agent.seekdb.not.exists")                             // "seekdb not exists in current directory. Please do takeover first."
	ErrAgentStartObserverFailed            = NewErrorCode("Agent.Start.ObserverFailed", unexpected, "err.agent.start.observer.failed")                           // "start seekdb via flag failed, err: %v"
	ErrAgentTakeOverFailed                 = NewErrorCode("Agent.TakeOverFailed", unexpected, "err.agent.take.over.failed")                                      // "take over or rebuild failed: %v"
	ErrAgentServeOnUnixSocketFailed        = NewErrorCode("Agent.ServeOnUnixSocketFailed", unexpected, "err.agent.serve.on.unix.socket.failed")                  // "serve on unix listener failed: %v\n"
	ErrAgentServeOnTcpSocketFailed         = NewErrorCode("Agent.ServeOnTcpSocketFailed", unexpected, "err.agent.serve.on.tcp.socket.failed")                    // "serve on tcp listener failed: %v\n"
	ErrAgentDaemonServeOnUnixSocketFailed  = NewErrorCode("Agent.Daemon.ServeOnUnixSocketFailed", unexpected, "err.agent.daemon.serve.on.unix.socket.failed")    // "daemon serve on socket listener failed\n"
	ErrAgentOceanbasePasswordLoadFailed    = NewErrorCode("Agent.Oceanbase.Password.LoadFailed", unexpected, "err.agent.oceanbase.password.load.failed")         // "check password of root@sys in sqlite failed: not cluster agent"
	ErrAgentUpgradeKillOldServerTimeout    = NewErrorCode("Agent.Upgrade.KillOldServerTimeout", unexpected, "err.agent.upgrade.kill.old.server.timeout")         // "wait obshell server killed timeout"
	ErrAgentDaemonStartFailed              = NewErrorCode("Agent.Daemon.StartFailed", unexpected, "err.agent.daemon.start.failed")                               // "daemon start failed: %v

	// config
	ErrConfigGetFailed = NewErrorCode("Config.GetFailed", unexpected, "err.config.get.failed")
	ErrConfigNotFound  = NewErrorCode("Config.NotFound", unexpected, "err.config.not.found")

	// external
	ErrExternalComponentNotReady = NewErrorCode("External.Component.Not.Ready", unexpected, "err.external.component.not.ready")

	// alarm related
	ErrAlarmClientFailed                 = NewErrorCode("Alarm.ClientFailed", unexpected, "err.alarm.client.failed")
	ErrAlarmQueryFailed                  = NewErrorCode("Alarm.QueryFailed", unexpected, "err.alarm.query.failed")
	ErrAlarmUnexpectedStatus             = NewErrorCode("Alarm.UnexpectedStatus", unexpected, "err.alarm.unexpected.status")
	ErrAlarmRuleNotFound                 = NewErrorCode("Alarm.RuleNotFound", notFound, "err.alarm.rule.not.found")
	ErrAlarmSilencerInstanceTypeMismatch = NewErrorCode("Alarm.Silencer.InstanceTypeMismatch", illegalArgument, "err.alarm.silencer.instance.type.mismatch")
	ErrAlarmSilencerOBClusterMismatch    = NewErrorCode("Alarm.Silencer.OBClusterMismatch", illegalArgument, "err.alarm.silencer.obcluster.mismatch")
	ErrAlarmSilencerUnknownInstanceType  = NewErrorCode("Alarm.Silencer.UnknownInstanceType", illegalArgument, "err.alarm.silencer.unknown.instance.type")

	// metric related
	ErrMetricConfigNotFound           = NewErrorCode("Metric.ConfigNotFound", unexpected, "err.metric.config.not.found")
	ErrMetricPrometheusConfigNotFound = NewErrorCode("Metric.PrometheusConfigNotFound", unexpected, "err.metric.prometheus.config.not.found")
	ErrMetricQueryFailed              = NewErrorCode("Metric.QueryFailed", unexpected, "err.metric.query.failed")
	ErrMetricUnexpectedStatus         = NewErrorCode("Metric.UnexpectedStatus", unexpected, "err.metric.unexpected.status")
	ErrMetricParseValueFailed         = NewErrorCode("Metric.ParseValueFailed", unexpected, "err.metric.parse.value.failed")
)
