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

package parameter

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

const (
	CMD_PARAMETER = "parameter"

	// obshell tenant parameter show
	CMD_SHOW = "show"

	// obshell tenant parameter set
	CMD_SET = "set"
)

func NewParameterCmd() *cobra.Command {
	parameterCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_PARAMETER,
		Short: "Display and manage the tenant parameters.",
	})
	parameterCmd.AddCommand(newShowCmd())
	parameterCmd.AddCommand(newSetCmd())

	return parameterCmd.Command
}

func newShowCmd() *cobra.Command {
	var verbose bool
	showCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SHOW,
		Short: "Show speciaic parameter(s).",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.Occur(errors.ErrCliUsageError, "tenant is required")
			}
			if len(args) < 2 {
				return errors.Occur(errors.ErrCliUsageError, "parameter is required")
			}
			stdio.SetVerboseMode(verbose)
			return showParameter(cmd, args[0], args[1])
		}),
		Example: `  obshell tenant parameter show t1 cpu_quota_concurrency`,
	})
	showCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name> [parameter]"}
	showCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return showCmd.Command
}

func showParameter(cmd *cobra.Command, tenant string, parameter string) error {
	info := make([]oceanbase.GvObParameter, 0)
	fuzzy := map[string]string{
		"filter": parameter,
	}
	if err := api.CallApiWithMethod(http.GET, constant.URI_TENANT_API_PREFIX+"/"+tenant+constant.URI_PARAMETERS, fuzzy, &info); err != nil {
		return err
	}
	data := make([][]string, 0)
	for _, p := range info {
		data = append(data, []string{p.Name, p.Value})
	}

	if len(data) != 0 {
		stdio.PrintTable([]string{"Name", "Value"}, data)
	} else {
		return errors.Occur(errors.ErrCliNotFound, parameter)
	}
	return nil
}

func BuildVariableOrParameterMap(str string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if len(str) == 0 {
		return m, nil
	}
	items := strings.Split(str, ",")
	for _, item := range items {
		kv := strings.Split(item, "=")
		if len(kv) != 2 || len(kv[0]) == 0 || len(kv[1]) == 0 {
			return nil, errors.Occurf(errors.ErrCliUsageError, "error format: %s, show be name=value", item)
		}
		m[kv[0]] = kv[1]
		if number, err := strconv.Atoi(kv[1]); err == nil {
			m[kv[0]] = number
		} else if floatValue, err := strconv.ParseFloat(kv[1], 64); err == nil {
			m[kv[0]] = floatValue
		} else if strings.ToLower(kv[1]) == "true" {
			m[kv[0]] = true
		} else if strings.ToLower(kv[1]) == "false" {
			m[kv[0]] = false
		}
	}
	return m, nil
}

func newSetCmd() *cobra.Command {
	var verbose bool
	setCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SET,
		Short: "Set speciaic parameters.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.Occur(errors.ErrCliUsageError, "tenant is required")
			}
			if len(args) < 2 {
				return errors.Occur(errors.ErrCliUsageError, "parameter is required")
			}
			stdio.SetVerboseMode(verbose)
			return setParameter(cmd, args[0], args[1])
		}),
		Example: `  obshell tenant parameter set t1 cpu_quota_concurrency=10,_rowsets_enabled=true`,
	})
	setCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name> <name=value>"}
	setCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return setCmd.Command
}

func setParameter(cmd *cobra.Command, tenant string, str string) error {
	parameters, err := BuildVariableOrParameterMap(str)
	if err != nil {
		return err
	}
	params := param.SetTenantParametersParam{
		Parameters: parameters,
	}

	stdio.StartLoading("set tenant parameter(s)")
	if err := api.CallApiWithMethod(http.PUT, constant.URI_TENANT_API_PREFIX+"/"+tenant+constant.URI_PARAMETERS, params, nil); err != nil {
		return err
	}
	stdio.LoadSuccess("set tenant parameter(s)")
	return nil
}
