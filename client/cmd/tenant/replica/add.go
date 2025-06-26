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

package replica

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	"github.com/oceanbase/obshell/agent/repository/model/bo"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type replicaAddFlags struct {
	verbose bool
	ZoneParamsFlags
}

type ZoneParamsFlags struct {
	Zones          string
	UnitNum        int
	UnitConfigName string
	ReplicaType    string
	UnknownFlags   []string
}

var zoneFlags = []string{"unit", "replica_type", "unit_num"}

func contain(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func FlagErrorFunc(cmd *cobra.Command, err error, opts *ZoneParamsFlags) error {
	cmd.SetFlagErrorFunc(nil)
	newArgs := []string{}
	for _, arg := range os.Args {
		if len(arg) > 2 && arg[:2] == "--" {
			if len(arg[2:]) == 0 || arg[2:][0] == '-' || arg[2:][0] == '=' {
				return err
			}
			// parse
			kv := strings.Split(arg[2:], "=")
			if len(kv) != 2 || len(kv[0]) == 0 || len(kv[1]) == 0 || !strings.Contains(kv[0], ".") {
				newArgs = append(newArgs, arg)
				continue
			}

			opts.UnknownFlags = append(opts.UnknownFlags, arg)
			continue
		}
		newArgs = append(newArgs, arg)
	}
	os.Args = newArgs
	err = cmd.Execute()
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return err
}

func newAddCmd() *cobra.Command {
	opts := &replicaAddFlags{}
	addCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_ADD,
		Short: "Add tenant replicas.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				return errors.Occur(errors.ErrCliUsageError, "tenant name is required")
			}
			stdio.SetVerboseMode(opts.verbose)
			return replicaAdd(cmd, args[0], &opts.ZoneParamsFlags)
		}),
		Example: `  obshell tenant replica add t1 -z zone4,zone5 --unit s1
  obshell tenant replica add t1 -z zone4,zone5 --zone4.unit=s4 --zone5.unit=s5
  obshell tenant replica add t1 -z zone4,zone5 --zone4.replica_type=FULL --zone5.replica_type=READONLY --unit s1`,
	})

	addCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	addCmd.Flags().SortFlags = false
	addCmd.VarsPs(&opts.Zones, []string{FLAG_ZONE, FLAG_ZONE_SH}, "", "The zones of the tenant.", true)
	addCmd.VarsPs(&opts.UnitConfigName, []string{FLAG_UNIT, FLAG_UNIT_SH}, "", "The unit config name.", false)
	addCmd.VarsPs(&opts.ReplicaType, []string{FLAG_REPLICA_TYPE}, "", "The replica type.", false)
	addCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	addCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return FlagErrorFunc(cmd, err, &opts.ZoneParamsFlags)
	})

	return addCmd.Command
}

func BuildZoneParams(cmd *cobra.Command, opts *ZoneParamsFlags) ([]param.ZoneParam, error) {
	// build ZoneParam
	zoneList := make([]param.ZoneParam, 0)
	var zones []string
	if !cmd.Flags().Changed(FLAG_ZONE) && !cmd.Flags().Changed(FLAG_ZONE_SH) {
		obInfo, err := api.GetObInfo()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get ob info")
		}
		for z := range obInfo.Config.ZoneConfig {
			zones = append(zones, z)
		}
	} else {
		zones = strings.Split(opts.Zones, ",")
	}
	zoneParams := make(map[string]*param.ZoneParam, 0)
	for _, zone := range zones {
		zoneParam := param.ZoneParam{
			Name: zone,
		}
		zoneParam.UnitNum = opts.UnitNum
		if cmd.Flags().Changed(FLAG_UNIT) || cmd.Flags().Changed(FLAG_UNIT_SH) {
			zoneParam.UnitConfigName = opts.UnitConfigName
		}
		if cmd.Flags().Changed(FLAG_REPLICA_TYPE) {
			zoneParam.ReplicaType = opts.ReplicaType
		}
		zoneParams[zone] = &zoneParam
	}
	var err error
	for _, flag := range opts.UnknownFlags {
		kv := strings.Split(flag[2:], "=")
		arr := strings.Split(kv[0], ".")
		if len(arr) != 2 || !contain(zones, arr[0]) || !contain(zoneFlags, arr[1]) {
			return nil, errors.Occurf(errors.ErrCliUsageError, "bad flag syntax: %s", flag)
		}
		if _, ok := zoneParams[arr[0]]; !ok {
			return nil, errors.Occurf(errors.ErrCliUsageError, "bad flag syntax: %s", flag)
		}
		switch arr[1] {
		case "unit":
			zoneParams[arr[0]].UnitConfigName = kv[1]
		case "replica_type":
			zoneParams[arr[0]].ReplicaType = kv[1]
		case "unit_num":
			num, err := strconv.Atoi(kv[1])
			if err != nil {
				return nil, errors.Occurf(errors.ErrCliUsageError, "bad flag syntax: %s", flag)
			}
			zoneParams[arr[0]].UnitNum = num
		default:
			return nil, errors.Occurf(errors.ErrCliUsageError, "bad flag syntax: %s", flag)
		}
	}
	for _, zoneParam := range zoneParams {
		zoneList = append(zoneList, *zoneParam)
	}
	return zoneList, err
}

func replicaAdd(cmd *cobra.Command, tenantName string, opts *ZoneParamsFlags) (err error) {
	// get tenant info for unit num
	info := bo.TenantInfo{}
	err = api.CallApiWithMethod(http.GET, constant.URI_TENANT_API_PREFIX+"/"+tenantName, nil, &info)
	if err != nil {
		return err
	}
	if len(info.Pools) > 0 {
		opts.UnitNum = info.Pools[0].UnitNum
	}

	// build ZoneParam
	zoneParams, err := BuildZoneParams(cmd, opts)
	if err != nil {
		return err
	}
	params := param.ScaleOutTenantReplicasParam{
		ZoneList: zoneParams,
	}

	dag := task.DagDetailDTO{}
	if err := api.CallApiWithMethod(http.POST, constant.URI_TENANT_API_PREFIX+"/"+tenantName+constant.URI_REPLICAS, params, &dag); err != nil {
		return err
	}
	if err = api.NewDagHandler(&dag).PrintDagStage(); err != nil {
		return err
	}
	return nil
}
