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
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type unitConfigCreateFlags struct {
	MemorySize  string
	MaxCpu      float64
	MinCpu      float64
	LogDiskSize string
	MinIops     int
	MaxIops     int
	Verbose     bool
}

func newCreateCmd() *cobra.Command {
	opts := unitConfigCreateFlags{}
	createCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_CREATE,
		Short: "Create a resource unit config.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			// get unit config name
			if len(args) <= 0 {
				return errors.Occur(errors.ErrCliUsageError, "unit config name is required")
			}
			stdio.SetVerboseMode(opts.Verbose)
			return unitConfigCreate(cmd, args[0], &opts)
		}),
		Example: `  obshell unit create s1 -m 5G -c 2`,
	})

	createCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<unit-config-name>"}
	createCmd.Flags().SortFlags = false
	// Setup of required flags for 'obshell unit create'.
	createCmd.VarsPs(&opts.MemorySize, []string{FLAG_MEMORY_SIZE, FLAG_MEMORY_SIZE_SH}, "", "Unit Config memory size.", true)
	createCmd.VarsPs(&opts.MaxCpu, []string{FLAG_MAX_CPU, FLAG_MAX_CPU_SH}, float64(0), "Unit Config max cpu.", true)

	// Configuration of optional flags for more detailed setup.
	createCmd.VarsPs(&opts.MinCpu, []string{FLAG_MIN_CPU}, float64(0), "Unit Config min cpu. Default to max_cpu", false)
	createCmd.VarsPs(&opts.LogDiskSize, []string{FLAG_LOG_DISK_SIZE}, "", "Unit Config log disk size.", false)
	createCmd.VarsPs(&opts.MinIops, []string{FLAG_MIN_IOPS}, 0, "Unit Config min iops.", false)
	createCmd.VarsPs(&opts.MaxIops, []string{FLAG_MAX_IOPS}, 0, "Unit Config max iops.", false)
	createCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Show verbose output.", false)

	return createCmd.Command
}

func unitConfigCreate(cmd *cobra.Command, name string, opts *unitConfigCreateFlags) error {
	params := buildCreateUnitConfigParams(cmd, name, opts)
	stdio.StartLoadingf("create unit config %s", name)
	if err := api.CallApiWithMethod(http.POST, constant.URI_UNIT_GROUP_PREFIX, params, nil); err != nil {
		return err
	}
	stdio.LoadSuccessf("create unit config %s", name)
	return nil
}

func buildCreateUnitConfigParams(cmd *cobra.Command, name string, opts *unitConfigCreateFlags) *param.CreateResourceUnitConfigParams {
	params := param.CreateResourceUnitConfigParams{}
	params.Name = &name
	params.MemorySize = &opts.MemorySize
	params.MaxCpu = &opts.MaxCpu
	if cmd.Flags().Changed(FLAG_MIN_CPU) {
		params.MinCpu = &opts.MinCpu
	}
	if cmd.Flags().Changed(FLAG_LOG_DISK_SIZE) {
		params.LogDiskSize = &opts.LogDiskSize
	}
	if cmd.Flags().Changed(FLAG_MIN_IOPS) {
		params.MinIops = &opts.MinIops
	}
	if cmd.Flags().Changed(FLAG_MAX_IOPS) {
		params.MaxIops = &opts.MaxIops
	}
	return &params
}
