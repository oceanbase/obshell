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

package command

import (
	"fmt"

	"github.com/oceanbase/obshell/ob/agent/config"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/log"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/spf13/cobra"
)

func WithErrorHandler(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		log.InitLogger(config.DefaultClientLoggerConifg())
		err := fn(cmd, args)
		if err != nil {
			stdio.LoadFailedWithoutMsg()
			if ocsAgentError, ok := err.(errors.OcsAgentErrorInterface); ok {
				if ocsAgentError.ErrorCode().Code == errors.ErrEmpty.Code {
					stdio.Error(err.Error())
				} else {
					stdio.Error(ocsAgentError.ErrorMessage())
				}
				if ocsAgentError.ErrorCode().Code == errors.ErrCliUsageError.Code || ocsAgentError.ErrorCode().Code == errors.ErrCliFlagRequired.Code {
					cmd.SilenceUsage = false
				}
			} else if ok, apiError := http.GetApiError(err); ok {
				stdio.Error(fmt.Sprintf("[%s]: %s", apiError.ErrCode, err.Error()))
			} else {
				stdio.Error(fmt.Sprintf("[%s]: %s", errors.ErrCommonUnexpected.Code, err.Error()))
			}
		} else {
			stdio.StopLoading()
		}
		return err
	}
}
