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

package variable

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/client/cmd/tenant/parameter"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

const (
	CMD_VARIABLE = "variable"

	// obshell tenant variable show
	CMD_SHOW = "show"

	// obshell tenant variable set
	CMD_SET = "set"

	FLAG_TENANT_PASSWORD = "tenant_password"
)

func NewVariableCmd() *cobra.Command {
	variableCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_VARIABLE,
		Short: "Display and manage the tenant global variables.",
	})
	variableCmd.AddCommand(newShowCmd())
	variableCmd.AddCommand(newSetCmd())

	return variableCmd.Command
}

func newShowCmd() *cobra.Command {
	var verbose bool
	showCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SHOW,
		Short: "Show speciaic variable.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.Occur(errors.ErrCliUsageError, "tenant name is required")
			}
			if len(args) < 2 {
				return errors.Occur(errors.ErrCliUsageError, "variable is required)")
			}
			stdio.SetVerboseMode(verbose)
			return showVariable(cmd, args[0], args[1])
		}),
		Example: `  obshell tenant variable show t1 max_connections`,
	})
	showCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name> [variable]"}
	showCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return showCmd.Command
}

func showVariable(cmd *cobra.Command, tenant string, variable string) error {
	info := make([]oceanbase.CdbObSysVariable, 0)
	fuzzy := map[string]string{
		"filter": variable,
	}
	if err := api.CallApiWithMethod(http.GET, constant.URI_TENANT_API_PREFIX+"/"+tenant+constant.URI_VARIABLES, fuzzy, &info); err != nil {
		return err
	}
	data := make([][]string, 0)
	for _, p := range info {
		data = append(data, []string{p.Name, p.Value})
	}

	if len(data) != 0 {
		stdio.PrintTable([]string{"Name", "Value"}, data)
	} else {
		return errors.Occur(errors.ErrCliNotFound, variable)
	}
	return nil
}

func newSetCmd() *cobra.Command {
	var verbose bool
	var tenantPassword string
	setCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SET,
		Short: "Set speciaic variables.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.Occur(errors.ErrCliUsageError, "tenant name is required")
			}
			if len(args) < 2 {
				return errors.Occur(errors.ErrCliUsageError, "variable is required)")
			}
			stdio.SetVerboseMode(verbose)
			return setVariable(cmd, args[0], args[1], tenantPassword)
		}),
		Example: `  obshell tenant variable set t1 max_connections=10000,recyclebin=true`,
	})
	setCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name> <name=value>"}
	setCmd.VarsPs(&verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	setCmd.VarsPs(&tenantPassword, []string{FLAG_TENANT_PASSWORD}, "", "Tenant password", false)
	return setCmd.Command
}

func setVariable(cmd *cobra.Command, tenant string, str string, tenantPassword string) error {
	variables, err := parameter.BuildVariableOrParameterMap(str)
	if err != nil {
		return err
	}
	params := param.SetTenantVariablesParam{
		Variables:      variables,
		TenantPassword: tenantPassword,
	}
	stdio.StartLoading("set tenant variables")
	if err := api.CallApiWithMethod(http.PUT, constant.URI_TENANT_API_PREFIX+"/"+tenant+constant.URI_VARIABLES, params, nil); err != nil {
		return err
	}
	stdio.LoadSuccess("set tenant variables")
	return nil
}
