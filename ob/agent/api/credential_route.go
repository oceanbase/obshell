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

	"github.com/oceanbase/obshell/ob/agent/api/common"
	"github.com/oceanbase/obshell/ob/agent/constant"
)

func InitCredentialRoutes(parentGroup *gin.RouterGroup, isLocalRoute bool) {
	security := parentGroup.Group(constant.URI_SECURITY_GROUP)
	credential := security.Group(constant.URI_CREDENTIAL)
	credentials := security.Group(constant.URI_CREDENTIALS)

	if !isLocalRoute {
		security.Use(common.Verify())
		credential.Use(common.Verify())
		credentials.Use(common.Verify())
	}

	// Credential CRUD operations
	credential.POST("", createCredentialHandler)
	credential.GET(constant.URI_PATH_PARAM_ID, checkClusterAgentWrapper(getCredentialHandler))
	credential.PATCH(constant.URI_PATH_PARAM_ID, checkClusterAgentWrapper(updateCredentialHandler))
	credential.DELETE(constant.URI_PATH_PARAM_ID, checkClusterAgentWrapper(deleteCredentialHandler))

	// Credential validation
	credential.POST(constant.URI_VALIDATE, checkClusterAgentWrapper(validateCredentialHandler))

	// Credential AES key update
	credential.PUT(constant.URI_ENCRYPT_SECRETKEY, checkClusterAgentWrapper(updateCredentialEncryptSecretKeyHandler))

	// Batch operations
	credentials.DELETE("", checkClusterAgentWrapper(batchDeleteCredentialsHandler))
	credentials.GET("", checkClusterAgentWrapper(listCredentialsHandler))
	credentials.POST(constant.URI_VALIDATE, checkClusterAgentWrapper(batchValidateCredentialsHandler))
}
