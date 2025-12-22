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
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	credentialexecutor "github.com/oceanbase/obshell/ob/agent/executor/credential"
	"github.com/oceanbase/obshell/ob/param"
)

// @ID updateCredentialEncryptSecretKey
// @Summary update AES key for credential encryption
// @Description update AES key and re-encrypt all stored credential passphrases
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.UpdateCredentialEncryptSecretKeyParam true "new AES key (Base64 encoded raw key bytes)"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credential/encrypt-secret-key [put]
func updateCredentialEncryptSecretKeyHandler(c *gin.Context) {
	var p param.UpdateCredentialEncryptSecretKeyParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	err := credentialexecutor.UpdateCredentialEncryptSecretKey(&p)
	common.SendResponse(c, nil, err)
}

// @ID createCredential
// @Summary create credential
// @Description create a new credential with SSH validation and encryption
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.CreateCredentialParam true "create credential params"
// @Success 200 object http.OcsAgentResponse{data=bo.Credential}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credential [post]
func createCredentialHandler(c *gin.Context) {
	var p param.CreateCredentialParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	data, err := credentialexecutor.CreateCredential(&p)
	common.SendResponse(c, data, err)
}

// @ID getCredential
// @Summary get credential
// @Description get credential by ID
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param id path int true "Credential ID"
// @Success 200 object http.OcsAgentResponse{data=bo.Credential}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credential/{id} [get]
func getCredentialHandler(c *gin.Context) {
	idStr := c.Param(constant.URI_PARAM_ID)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonIllegalArgument, "invalid credential id"))
		return
	}

	data, err := credentialexecutor.GetCredential(id)
	common.SendResponse(c, data, err)
}

// @ID updateCredential
// @Summary update credential
// @Description update credential with SSH validation and encryption
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param id path int true "Credential ID"
// @Param body body param.UpdateCredentialParam true "update credential params"
// @Success 200 object http.OcsAgentResponse{data=bo.Credential}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credential/{id} [patch]
func updateCredentialHandler(c *gin.Context) {
	idStr := c.Param(constant.URI_PARAM_ID)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonIllegalArgument, "invalid credential id"))
		return
	}

	var p param.UpdateCredentialParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	data, err := credentialexecutor.UpdateCredential(id, &p)
	common.SendResponse(c, data, err)
}

// @ID deleteCredential
// @Summary delete credential
// @Description delete credential by ID (hard delete)
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param id path int true "Credential ID"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 404 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credential/{id} [delete]
func deleteCredentialHandler(c *gin.Context) {
	idStr := c.Param(constant.URI_PARAM_ID)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.SendResponse(c, nil, errors.Occur(errors.ErrCommonIllegalArgument, "invalid credential id"))
		return
	}

	err = credentialexecutor.DeleteCredential(id)
	common.SendResponse(c, nil, err)
}

// @ID batchDeleteCredentials
// @Summary batch delete credentials
// @Description delete multiple credentials by ID list (hard delete)
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.BatchDeleteCredentialParam true "batch delete credential params"
// @Success 200 object http.OcsAgentResponse
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credentials [delete]
func batchDeleteCredentialsHandler(c *gin.Context) {
	var p param.BatchDeleteCredentialParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	ids := make([]int64, len(p.CredentialIdList))
	for i, id := range p.CredentialIdList {
		ids[i] = int64(id)
	}

	err := credentialexecutor.BatchDeleteCredential(ids)
	common.SendResponse(c, nil, err)
}

// @ID listCredentials
// @Summary list credentials
// @Description list credentials with filtering, pagination, and sorting
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param credential_id query int false "Credential ID"
// @Param target_type query string false "Target type, currently only supports HOST"
// @Param key_word query string false "Keyword for searching"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param sort query string false "Sort field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Success 200 object http.OcsAgentResponse{data=bo.PaginatedCredentialResponse}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credentials [get]
func listCredentialsHandler(c *gin.Context) {
	var p param.ListCredentialQueryParam
	if err := c.BindQuery(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	p.Format()

	response, err := credentialexecutor.ListCredentials(&p)
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	common.SendResponse(c, response, nil)
}

// @ID validateCredential
// @Summary validate credential
// @Description validate credential without storing it
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.ValidateCredentialParam true "validate credential params"
// @Success 200 object http.OcsAgentResponse{data=bo.ValidationResult}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credential/validate [post]
func validateCredentialHandler(c *gin.Context) {
	var p param.ValidateCredentialParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	data, err := credentialexecutor.ValidateCredential(&p)
	common.SendResponse(c, data, err)
}

// @ID batchValidateCredentials
// @Summary batch validate credentials
// @Description validate multiple credentials by ID list
// @Tags security
// @Accept application/json
// @Produce application/json
// @Param X-OCS-Header header string true "Authorization"
// @Param body body param.BatchValidateCredentialParam true "batch validate credential params"
// @Success 200 object http.OcsAgentResponse{data=[]bo.ValidationResult}
// @Failure 400 object http.OcsAgentResponse
// @Failure 401 object http.OcsAgentResponse
// @Failure 500 object http.OcsAgentResponse
// @Router /api/v1/security/credentials/validate [post]
func batchValidateCredentialsHandler(c *gin.Context) {
	var p param.BatchValidateCredentialParam
	if err := c.BindJSON(&p); err != nil {
		common.SendResponse(c, nil, err)
		return
	}

	data, err := credentialexecutor.BatchValidateCredential(p.CredentialIdList)
	common.SendResponse(c, data, err)
}
