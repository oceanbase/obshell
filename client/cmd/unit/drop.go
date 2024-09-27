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

package unit

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/repository/model/oceanbase"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/global"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
)

func newDropCmd() *cobra.Command {
	opts := &global.DropFlags{}
	dropCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_DROP,
		Short: "Drop a resource unit config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			// get unit config name
			if len(args) <= 0 {
				stdio.Error("unit config name is required")
				return errors.New("unit config name is required")
			}
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetSkipConfirmMode(opts.SkipConfirm)
			stdio.SetVerboseMode(opts.Verbose)
			if err := unitConfigDrop(args[0], opts); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: `  obshell unit drop s1`,
	})

	dropCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<unit-config-name>"}
	dropCmd.Flags().SortFlags = false
	dropCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	dropCmd.VarsPs(&opts.SkipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation of drop unit config operation", false)
	return dropCmd.Command
}

func unitConfigDrop(name string, opts *global.DropFlags) error {
	pass, err := stdio.Confirmf("Please confirm if you need to drop unit config %s", name)
	if err != nil {
		return errors.New("ask for confirmation failed")
	}
	if !pass {
		return nil
	}
	if opts.IfExist {
		unitConfigs := make([]oceanbase.DbaObUnitConfig, 0)
		// show all
		stdio.StartLoading("get all unit configs")
		err := api.CallApiWithMethod(http.GET, constant.URI_API_V1+constant.URI_UNITS_GROUP, nil, &unitConfigs)
		if err != nil {
			return err
		}
		stdio.LoadSuccess("get all unit configs")
		for _, unitConfig := range unitConfigs {
			if unitConfig.Name == name {
				// Drop unit config
				stdio.StartLoadingf("drop unit config %s", name)
				if err := api.CallApiWithMethod(http.DELETE, constant.URI_UNIT_GROUP_PREFIX+"/"+name, nil, nil); err != nil {
					return err
				}
				stdio.LoadSuccessf("drop unit config %s", name)
			}
		}
		return nil
	}
	// Drop unit config
	stdio.StartLoadingf("drop unit config %s", name)
	if err := api.CallApiWithMethod(http.DELETE, constant.URI_UNIT_GROUP_PREFIX+"/"+name, nil, nil); err != nil {
		return err
	}
	stdio.LoadSuccessf("drop unit config %s", name)
	return nil
}
