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

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/errors"
)

const (
	CMD_ARG_COUNT = 1
)

func ValidateArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.Occurf(errors.ErrCliUsageError, "unspecified arguments: %v", args)
	}
	return nil
}

func ValidateArgTenantName(cmd *cobra.Command, args []string) (err error) {
	length := len(args)
	if length == 0 {
		err = errors.Occur(errors.ErrCliUsageError, "tenant name is required")
	} else if length > CMD_ARG_COUNT {
		err = errors.Occurf(errors.ErrCliUsageError, "too many arguments, expected %d, got %d", CMD_ARG_COUNT, length)
	}
	return err
}
