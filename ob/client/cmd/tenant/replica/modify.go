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
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/client/command"
	clientconst "github.com/oceanbase/obshell/ob/client/constant"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
	"github.com/oceanbase/obshell/ob/client/utils/api"
	"github.com/oceanbase/obshell/ob/param"
)

type replicaModifyFlags struct {
	verbose bool
	ZoneParamsFlags
}

func newModifyCmd() *cobra.Command {
	opts := &replicaModifyFlags{}
	modifyCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_MODIFY,
		Short: "Modify tenant replicas.",
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				return errors.Occur(errors.ErrCliUsageError, "tenant name is required")
			}
			stdio.SetVerboseMode(opts.verbose)
			return replicaModify(cmd, args[0], &opts.ZoneParamsFlags)
		}),
		Example: `  obshell tenant replica modify t1 --unit s2 --unit_num 2
  obshell tenant replica modify t1 -z zone1,zone2 --zone1.replica_type=READONLY --zone2.unit=s2`,
	})

	modifyCmd.Annotations = map[string]string{clientconst.ANNOTATION_ARGS: "<tenant-name>"}
	modifyCmd.Flags().SortFlags = false
	modifyCmd.VarsPs(&opts.Zones, []string{FLAG_ZONE, FLAG_ZONE_SH}, "", "The zones of the tenant.", false)
	modifyCmd.VarsPs(&opts.UnitNum, []string{FLAG_UNIT_NUM}, 0, "The number of units in each zone.", false)
	modifyCmd.VarsPs(&opts.UnitConfigName, []string{FLAG_UNIT, FLAG_UNIT_SH}, "", "The unit config name.", false)
	modifyCmd.VarsPs(&opts.ReplicaType, []string{FLAG_REPLICA_TYPE}, "", "The replica type.", false)
	modifyCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	modifyCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return FlagErrorFunc(cmd, err, &opts.ZoneParamsFlags)
	})

	return modifyCmd.Command
}

func BuildModifyReplicaZoneParams(cmd *cobra.Command, tenantName string, opts *ZoneParamsFlags) ([]param.ModifyReplicaZoneParam, error) {
	// build ZoneParam
	zoneList := make([]param.ModifyReplicaZoneParam, 0)
	var zones []string
	if !cmd.Flags().Changed(FLAG_ZONE) && !cmd.Flags().Changed(FLAG_ZONE_SH) {
		obInfo, err := api.GetTenantOverView(tenantName)
		if err != nil {
			return nil, err
		}
		arr := strings.Split(obInfo.Locality, ",")
		for _, v := range arr {
			split := strings.SplitN(v, "@", 2)
			if len(split) != 2 {
				return nil, errors.Occur(errors.ErrObTenantLocalityFormatUnexpected, obInfo.Locality)
			} else {
				zones = append(zones, split[1])
			}
		}
	} else {
		zones = strings.Split(opts.Zones, ",")
	}
	zoneParams := make(map[string]*param.ModifyReplicaZoneParam, 0)
	for _, zone := range zones {
		zoneParam := param.ModifyReplicaZoneParam{
			Name: zone,
		}
		if cmd.Flags().Changed(FLAG_UNIT_NUM) {
			zoneParam.UnitNum = &opts.UnitNum
		}
		if cmd.Flags().Changed(FLAG_UNIT) || cmd.Flags().Changed(FLAG_UNIT_SH) {
			zoneParam.UnitConfigName = &opts.UnitConfigName
		}
		if cmd.Flags().Changed(FLAG_REPLICA_TYPE) {
			zoneParam.ReplicaType = &opts.ReplicaType
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
			zoneParams[arr[0]].UnitConfigName = &kv[1]
		case "replica_type":
			zoneParams[arr[0]].ReplicaType = &kv[1]
		case "unit_num":
			num, err := strconv.Atoi(kv[1])
			if err != nil {
				return nil, errors.Occurf(errors.ErrCliUsageError, "bad flag syntax: %s", flag)
			}
			zoneParams[arr[0]].UnitNum = &num
		default:
			return nil, errors.Occurf(errors.ErrCliUsageError, "bad flag syntax: %s", flag)
		}
	}
	for _, zoneParam := range zoneParams {
		zoneList = append(zoneList, *zoneParam)
	}
	return zoneList, err
}

func replicaModify(cmd *cobra.Command, tenantName string, opts *ZoneParamsFlags) (err error) {
	modifyZoneParams, err := BuildModifyReplicaZoneParams(cmd, tenantName, opts)
	if err != nil {
		return err
	}
	params := param.ModifyReplicasParam{
		ZoneList: modifyZoneParams,
	}

	var dag task.DagDetailDTO
	if err := api.CallApiWithMethod(http.PATCH, constant.URI_TENANT_API_PREFIX+"/"+tenantName+constant.URI_REPLICAS, params, &dag); err != nil {
		return err
	}
	if dag.GenericDTO == nil { // No content
		stdio.Print("nothing need to be modified")
		return nil
	}
	if err = api.NewDagHandler(&dag).PrintDagStage(); err != nil {
		return err
	}
	return nil
}
