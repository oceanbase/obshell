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
	"github.com/oceanbase/obshell/seekdb/agent/api/common"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/executor/upgrade"
	"github.com/oceanbase/obshell/seekdb/agent/meta"
	"github.com/oceanbase/obshell/seekdb/param"
)

// @ID getUpgradePkgInfo
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
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}

	data, agentErr := upgrade.GetAllUpgradePkgInfos()
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
func pkgUploadHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrRequestFileMissing, "file", err.Error()))
		return
	}
	defer file.Close()
	data, agentErr := upgrade.UpgradePkgUpload(file)
	if agentErr != nil || data == nil {
		common.SendResponse(c, nil, agentErr)
		return
	}
	bo := data.ToBO()
	common.SendResponse(c, &bo, agentErr)
}

// @ID deletePackage
// @Summary delete package in ocs
// @Description delete package in ocs
// @Tags package
// @Accept application/json
// @Produce application/json
// @Param body body param.DeletePackageParam true "delete package param"
// @Success 200 object http.OcsAgentResponse
// @Success 204 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/upgrade/package [delete]
func pkgDeleteHandler(c *gin.Context) {
	if !meta.OCS_AGENT.IsClusterAgent() {
		common.SendResponse(c, nil, errors.Occur(errors.ErrAgentIdentifyNotSupportOperation, meta.OCS_AGENT.String(), meta.OCS_AGENT.GetIdentity(), meta.CLUSTER_AGENT))
		return
	}
	var p param.DeletePackageParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, nil, upgrade.DeletePackage(p))
}
