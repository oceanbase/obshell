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

package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obshell/agent/api/common"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/lib/crypto"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/agent/secure"
	"github.com/oceanbase/obshell/param"
)

// When the pwd is shorter than encrypted pwd, return directly. Otherwise decrypt the pwd:
// 1. decrypt failed: if isForward, return error; otherwise return the original pwd.
// 2. decrypt success: return the decrypted pwd.
func parseRootPwd(pwd string, isForward bool) (string, error) {
	strSize := (crypto.KEY_SIZE/8 + 2) / 3 * 4
	if len(pwd) >= strSize {
		// Try decode pwd.
		parsedPwd, err := secure.Crypter.Decrypt(pwd)
		if err != nil {
			if isForward {
				return "", errors.New("Please don't encrypt the password, and send /api/v1/obcluster/config to master agent.")
			}
			return pwd, nil
		}
		return parsedPwd, nil
	}
	return pwd, nil
}

// StatisticsHandler returns the statistics data
//
// @ID GetStatistics
// @Summary get statistics data
// @Description get statistics data
// @Tags ob
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=bo.ObclusterStatisticInfo}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/obcluster/statistics [GET]
func GetStatistics(c *gin.Context) {
	statisticsData := ob.GetStatisticsInfo()
	common.SendResponse(c, statisticsData, nil)
}

// @ID obclusterConfig
// @Summary put ob cluster configs
// @Description put ob cluster configs
// @Tags ob
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ObClusterConfigParams true "obcluster configs"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 401 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/obcluster/config [POST]
func obclusterConfigHandler(deleteAll bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params param.ObClusterConfigParams
		if err := c.BindJSON(&params); err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		if params.ClusterId == nil || *params.ClusterId <= 0 {
			common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "cluster id is invalid"))
			return
		}
		if params.ClusterName == nil || *params.ClusterName == "" {
			common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "cluster name is empty"))
			return
		}

		switch meta.OCS_AGENT.GetIdentity() {
		case meta.FOLLOWER:
			master := agentService.GetMasterAgentInfo()
			if master == nil {
				common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "master is not found"))
				return
			}
			common.ForwardRequest(c, master, params)
		case meta.CLUSTER_AGENT:
			if deleteAll {
				common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "Cluster has already initialized."))
				return
			}
			fallthrough
		case meta.MASTER:
			_, isForward := c.Get(common.IsAutoForwardedFlag)
			if params.RootPwd != nil && *params.RootPwd != "" {
				var err error
				*params.RootPwd, err = parseRootPwd(*params.RootPwd, isForward)
				if err != nil {
					common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
					return
				}
				// encrypt root pwd
				pwd, err := secure.Crypter.Encrypt(*params.RootPwd)
				if err != nil {
					log.WithContext(common.NewContextWithTraceId(c)).WithError(err).Error("request from local route, encrypt password failed")
					common.SendResponse(c, nil, err)
					return
				}
				params.RootPwd = &pwd

			}
			dag, err := ob.CreateUpdateOBClusterConfigDag(params, deleteAll)
			common.SendResponse(c, dag, err)

		default:
			common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s agent can not update ob cluster config", meta.OCS_AGENT.GetIdentity()))
		}
	}
}

// @ID obServerConfig
// @Summary put observer configs
// @Description put observer configs
// @Tags ob
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ObServerConfigParams true "ob server configs"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/observer/config [POST]
func obServerConfigHandler(deleteAll bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params param.ObServerConfigParams
		if err := c.BindJSON(&params); err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		if params.ObServerConfig == nil || len(params.ObServerConfig) == 0 {
			common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "config is empty"))
			return
		}

		if err := ob.CheckOBServerConfigParams(params); err != nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err.Error()))
			return
		}

		switch meta.OCS_AGENT.GetIdentity() {
		case meta.FOLLOWER:
			master := agentService.GetMasterAgentInfo()
			if master == nil {
				common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "master is not found"))
				return
			}
			common.ForwardRequest(c, master, params)
		case meta.CLUSTER_AGENT:
			if deleteAll {
				common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "Cluster has already initialized."))
				return
			}
			fallthrough
		case meta.MASTER:
			if err := ob.IsValidScope(&params.Scope); err != nil {
				common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, err))
				return
			}
			dag, err := ob.CreateUpdateOBServerConfigDag(params, deleteAll)
			common.SendResponse(c, dag, err)
		default:
			common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s agent can not update ob cluster config", meta.OCS_AGENT.GetIdentity()))
		}
	}
}

// @ID obInit
// @Summary init ob cluster
// @Description init ob cluster
// @Tags ob
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/ob/init [post]
func obInitHandler(c *gin.Context) {
	var param param.ObInitParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	switch meta.OCS_AGENT.GetIdentity() {
	case meta.MASTER:
		data, err := ob.CreateInitDag(param)
		common.SendResponse(c, data, err)
	case meta.FOLLOWER:
		master := agentService.GetMasterAgentInfo()
		if master == nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "master is not found"))
			return
		}
		common.ForwardRequest(c, master, nil)
	case meta.CLUSTER_AGENT:
		common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "Failed to initialize cluster, cluster has already initialized."))
	default:
		common.SendResponse(c, nil, errors.Occurf(errors.ErrBadRequest, "%s can not init", meta.OCS_AGENT.GetIdentity()))
	}
}

// @ID obStop
// @Summary stop observers
// @Description stop observers or the whole cluster, use param to specify
// @Tags ob
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ObStopParam true "use 'Scope' to specify the servers/zones/cluster, use 'Force'(optional) to specify whether alter system, use 'ForcePassDag'(optional) to force pass the prev stop dag if need"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/ob/stop [post]
func obStopHandler(c *gin.Context) {
	var param param.ObStopParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	emergencyMode, err := isEmergencyMode(c, &param.Scope)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if emergencyMode && param.Force {
		data, err := ob.EmergencyStop()
		common.SendResponse(c, data, err)
	} else {
		data, err := ob.HandleObStop(param)
		common.SendResponse(c, data, err)
	}
}

// @ID obStart
// @Summary start observers
// @Description start observers or the whole cluster, use param to specify
// @Tags ob
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.StartObParam true "use 'Scope' to specify the servers/zones/cluster, use 'ForcePassDag'(optional) to force pass the prev start dag if need"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/ob/start [post]
func obStartHandler(c *gin.Context) {
	var param param.StartObParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	emergencyMode, err := isEmergencyMode(c, &param.Scope)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	hasStart, e := ob.HasStarted()
	if e != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrUnexpected, e))
		return
	}
	if !hasStart {
		if meta.OCS_AGENT.IsSingleAgent() {
			err = errors.Occur(errors.ErrKnown, "There is no 'OceanBase' cluster present. Please execute the 'join' operation first.")
		} else if meta.OCS_AGENT.IsMasterAgent() || meta.OCS_AGENT.IsFollowerAgent() {
			err = errors.Occur(errors.ErrKnown, "There is no 'OceanBase' cluster present. Please execute the 'init' operation first.")
		} else {
			err = errors.Occur(errors.ErrKnown, "Observer has not been previously started. Please initialize the cluster first.")
		}
		common.SendResponse(c, nil, err)
		return
	}

	if emergencyMode {
		data, err := ob.EmergencyStart()
		common.SendResponse(c, data, err)
	} else {
		data, err := ob.HandleObStart(param)
		common.SendResponse(c, data, err)
	}
}

// @ID ScaleOut
// @Summary cluster scale-out
// @Description cluster scale-out
// @Tags ob
// @Accept application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ClusterScaleOutParam true "scale-out param"
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/ob/scale_out [POST]
func obClusterScaleOutHandler(c *gin.Context) {
	var param param.ClusterScaleOutParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	data, err := ob.HandleClusterScaleOut(param)
	common.SendResponse(c, data, err)
}

// @Summary cluster scale-in
// @Description cluster scale-in
// @Tags ob
// @Accept application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ClusterScaleInParam true "scale-in param"
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Success 204 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/ob/scale_in [post]
// @Router /api/v1/observer [delete]
func obClusterScaleInHandler(c *gin.Context) {
	var param param.ClusterScaleInParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	if meta.OCS_AGENT.IsSingleAgent() && meta.OCS_AGENT.Equal(&param.AgentInfo) {
		common.SendNoContentResponse(c, nil)
		return
	}
	if meta.OCS_AGENT.Equal(&param.AgentInfo) {
		common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "Cannot delete the current server."))
		return
	}
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	dag, err := ob.ClusterScaleIn(param)
	if dag == nil && err == nil {
		common.SendNoContentResponse(c, nil)
	} else {
		common.SendResponse(c, dag, err)
	}
}

// @ID GetObInfo
// @Summary get ob and agent info
// @Description get ob and agent info
// @Tags ob
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=param.ObInfoResp}
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/ob/info [get]
func obInfoHandler(c *gin.Context) {
	if meta.OCS_AGENT.IsFollowerAgent() {
		master := agentService.GetMasterAgentInfo()
		if master == nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "master is not found"))
			return
		}
		common.ForwardRequest(c, master, nil)
		return
	}
	data, err := ob.GetObInfo()
	common.SendResponse(c, data, err)
}

// @ID				obclusterInfo
// @Summary		get obcluster info
// @Description	get obcluster info
// @Tags			obcluster
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse{data=bo.ClusterInfo}
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/info [get]
func obclusterInfoHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	clusterInfo, err := ob.GetObclusterSummary()
	common.SendResponse(c, clusterInfo, err)
}

// @ID				obclusterParameters
// @Summary		get obcluster parameters
// @Description	get obcluster parameters
// @Tags			obcluster
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string	true	"Authorization"
// @Success		200				object	http.OcsAgentResponse{data=[]bo.ClusterParameter}
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/parameters [get]
func obclusterParametersHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	clusterInfo, err := ob.GetAllParameters()
	common.SendResponse(c, clusterInfo, err)
}

// @ID				obclusterSetParameters
// @Summary		set obcluster parameters
// @Description	set obcluster parameters
// @Tags			obcluster
// @Accept			application/json
// @Produce		application/json
// @Param			X-OCS-Header	header	string								true	"Authorization"
// @Param			body			body	param.SetObclusterParametersParam	true	"obcluster parameters"
// @Success		204				object	http.OcsAgentResponse
// @Failure		400				object	http.OcsAgentResponse
// @Failure		401				object	http.OcsAgentResponse
// @Failure		500				object	http.OcsAgentResponse
// @Router			/api/v1/obcluster/parameters [patch]
func obclusterSetParametersHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	var param param.SetObclusterParametersParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, ob.SetObclusterParameters(param.Params))
}

func isEmergencyMode(c *gin.Context, scope *param.Scope) (res bool, agentErr *errors.OcsAgentError) {
	if common.IsLocalRoute(c) && ob.ScopeOnlySelf(scope) && !meta.OCS_AGENT.IsClusterAgent() {
		return true, nil
	}
	if err := ob.IsValidScope(scope); err != nil {
		return false, errors.Occur(errors.ErrIllegalArgument, err)
	}
	return false, nil
}

// @ID agentUpgradeCheck
// @Summary check agent upgrade
// @Description check agent upgrade
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.UpgradeCheckParam true "agent upgrade check params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/agent/upgrade/check [post]
func agentUpgradeCheckHandler(c *gin.Context) {
	var param param.UpgradeCheckParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	task, err := ob.AgentUpgradeCheck(param)
	common.SendResponse(c, task, err)
}

// @ID obUpgradeCheck
// @Summary check ob upgrade
// @Description check ob upgrade
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.UpgradeCheckParam true "ob upgrade check params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/ob/upgrade/check [post]
func obUpgradeCheckHandler(c *gin.Context) {
	var param param.UpgradeCheckParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	task, err := ob.ObUpgradeCheck(param)
	common.SendResponse(c, task, err)
}

// @ID UpgradePkgRoute
// @Summary get upgrade route of target ob package
// @Description get upgrade route of target ob package
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param version query	string true "version"
// @Param release query	string true "release"
// @Success 200	object []ob.RouteNode
// @Failure 401	object http.OcsAgentResponse
// @Failure 500	object http.OcsAgentResponse
// @Router /api/v1/ob/upgrade/route [get]
func obPkgUpgradeRouteHandler(c *gin.Context) {
	version := c.Query("version")
	release := c.Query("release")
	if version == "" || release == "" {
		common.SendResponse(c, nil, errors.Occur(errors.ErrIllegalArgument, "version or release is empty"))
		return
	}
	routes, err := ob.GetObPackageUpgradeDepYaml(version, release)
	common.SendResponse(c, routes, err)
}

// @ID UpgradePkgUpload
// @Summary upload upgrade package
// @Description upload upgrade package
// @Tags upgrade
// @Accept multipart/form-data
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param file formData file true "ob upgrade package"
// @Success 200 object http.OcsAgentResponse{data=oceanbase.UpgradePkgInfo}
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/upgrade/package [post]
func pkgUploadHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObclusterNotFound, "Unable to proceed with package upload. Please ensure the 'init' command is executed before attempting to upload."))
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "get file failed: %s", err))
		return
	}
	defer file.Close()
	data, agentErr := ob.UpgradePkgUpload(file)
	common.SendResponse(c, &data, agentErr)
}

// @ID UpgradePkgInfo
// @Summary get all upgrade package infos
// @Description get all upgrade package infos
// @Tags upgrade
// @Accept application/json
// @Produce	 application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=[]bo.UpgradePkgInfo}
// @Failure	401	object http.OcsAgentResponse
// @Failure 500	object http.OcsAgentResponse
// @Router /api/v1/upgrade/package/info [get]
func pkgInfoHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObclusterNotFound, "Unable to get package infos from ocs database. Please ensure the 'init' command is executed before attempting to upload."))
		return
	}

	data, agentErr := ob.GetAllUpgradePkgInfos()
	common.SendResponse(c, data, agentErr)
}

// @ID NewPkgUpload
// @Summary upload upgrade package without body encryption
// @Description upload upgrade package without body encryption
// @Tags package
// @Accept multipart/form-data
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param X-OCS-File-SHA256 header string true "SHA256 of the file"
// @Param file formData file true "ob upgrade package"
// @Success 200 object http.OcsAgentResponse{data=bo.UpgradePkgInfo}
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/package [post]
func newPkgUploadHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrObclusterNotFound, "Unable to proceed with package upload. Please ensure the 'init' command is executed before attempting to upload."))
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "get file failed: %s", err))
		return
	}
	defer file.Close()
	data, agentErr := ob.UpgradePkgUpload(file)
	if agentErr != nil || data == nil {
		common.SendResponse(c, nil, agentErr)
		return
	}
	bo := data.ToBO()
	common.SendResponse(c, &bo, agentErr)
}

// @ID ParamsBackup
// @Summary backup params
// @Description backup params
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=[]oceanbase.ObParameters}
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/upgrade/params/backup [post]
func paramsBackupHandler(c *gin.Context) {
	data, err := ob.ParamsBackup()
	common.SendResponse(c, data, err)
}

// @ID ParamsRestore
// @Summary restore params
// @Description restore params
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.RestoreParams true "restore params"
// @Success 200 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/upgrade/params/restore [post]
func paramsRestoreHandler(c *gin.Context) {
	var param param.RestoreParams
	err := ob.ParamsRestore(param)
	common.SendResponse(c, nil, err)
}

// @ID agentUpgrade
// @Summary upgrade agent
// @Description upgrade agent
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.UpgradeCheckParam true "agent upgrade check params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/agent/upgrade [post]
func agentUpgradeHandler(c *gin.Context) {
	var param param.UpgradeCheckParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := ob.AgentUpgrade(param)
	common.SendResponse(c, dag, err)
}

// @ID obUpgrade
// @Summary upgrade ob
// @Description upgrade ob
// @Tags upgrade
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ObUpgradeParam true "ob upgrade params"
// @Success 200 object http.OcsAgentResponse{data=task.DagDetailDTO}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/ob/upgrade [post]
func obUpgradeHandler(c *gin.Context) {
	var param param.ObUpgradeParam
	if err := c.BindJSON(&param); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	dag, err := ob.CheckAndUpgradeOb(param)
	common.SendResponse(c, dag, err)
}

func obAgentsHandler(c *gin.Context) {
	if meta.OCS_AGENT.IsFollowerAgent() {
		master := agentService.GetMasterAgentInfo()
		if master == nil {
			common.SendResponse(c, nil, errors.Occur(errors.ErrBadRequest, "master is not found"))
			return
		}
		common.ForwardRequest(c, master, nil)
		return
	}
	data, err := ob.GetObAgents()
	common.SendResponse(c, data, err)
}

// @ID getObclusterCharsets
// @Summary get obcluster charsets
// @Description get obcluster charsets
// @Tags ob
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Success 200 object http.OcsAgentResponse{data=[]bo.CharsetInfo}
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/obcluster/charsets [get]
func getObclusterCharsets(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occurf(errors.ErrKnown, "%s is not cluster agent.", meta.OCS_AGENT.String()))
		return
	}
	charsets, err := ob.GetObclusterCharsets()
	common.SendResponse(c, charsets, err)
}

// @ID getUnitConfigLimit
// @Summary get resource unit config limit
// @Description get resource unit config limit
// @Tags obcluster
// @Accept application/json
// @Produce application/json
// @Success 200 object http.OcsAgentResponse{data=param.ClusterUnitConfigLimit}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/obcluster/unit-config-limit [get]
func getUnitConfigLimitHandler(c *gin.Context) {
	unit := ob.GetClusterUnitSpecLimit()
	common.SendResponse(c, unit, nil)
}
