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

package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/lib/http"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/client/command"
	clientconst "github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/global"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/param"
)

type ClusterScaleInFlags struct {
	server string // the server would be delete from the cluster
	zone   string // the zone would be delete from the cluster
	force  bool   // force to delete the server
	global.DropFlags
}

func NewScaleInCmd() *cobra.Command {
	opts := &ClusterScaleInFlags{}
	scaleInCmd := command.NewCommand(&cobra.Command{
		Use:   CMD_SCALE_IN,
		Short: "Delete a observer or a zone from OceanBase cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			ocsagentlog.SetDBLoggerLevel(ocsagentlog.Silent)
			stdio.SetSkipConfirmMode(opts.SkipConfirm)
			stdio.SetVerboseMode(opts.Verbose)
			if err := clusterScaleIn(cmd, opts); err != nil {
				stdio.LoadFailedWithoutMsg()
				stdio.Error(err.Error())
				return err
			}
			return nil
		},
		Example: `  obshell cluster scale-in -s 192.168.1.1:2886
  obshell cluster scale-in -z zone1`,
	})

	scaleInCmd.Flags().SortFlags = false
	// Setup of required flags for 'scale-in' command.
	scaleInCmd.VarsPs(&opts.server, []string{FLAG_SERVER_SH, FLAG_SERVER}, "", "The address of the server holding the observer to be deleted from the cluster. If the port is unspecified, it will be 2886.", false)
	scaleInCmd.VarsPs(&opts.zone, []string{FLAG_ZONE_SH, FLAG_ZONE}, "", "The zone to be deleted from the cluster.", false)
	scaleInCmd.VarsPs(&opts.force, []string{FLAG_FORCE_SH, FLAG_FORCE}, false, "Forcefully kill the observer.", false)

	scaleInCmd.VarsPs(&opts.SkipConfirm, []string{clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH}, false, "Skip the confirmation prompt", false)
	scaleInCmd.VarsPs(&opts.Verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)

	return scaleInCmd.Command
}

func clusterScaleIn(cmd *cobra.Command, flags *ClusterScaleInFlags) (err error) {
	if flags.server == "" && flags.zone == "" {
		err = errors.New("Please specify one of the server or zone to scale in.")
	}
	if flags.server != "" && flags.zone != "" {
		err = errors.New("Please specify only one of the server or zone to scale in.")
	}
	if err != nil {
		cmd.SilenceUsage = false
		return err
	}

	var dag *task.DagDetailDTO
	if flags.server != "" {
		if dag, err = deleteServer(flags.server, flags.force); err != nil {
			return err
		}
	} else if flags.zone != "" {
		if dag, err = deleteZone(flags.zone); err != nil {
			return err
		}
	}

	if dag.GenericDTO == nil {
		return
	}
	if err = api.NewDagHandler(dag).PrintDagStage(); err != nil {
		return err
	}
	return
}

func deleteServer(server string, forceKill bool) (*task.DagDetailDTO, error) {
	targetAgentInfo, err := meta.ConvertAddressToAgentInfo(server)
	if err != nil {
		return nil, err
	}
	message := fmt.Sprintf("Please confirm if you need to delete '%s' with observer from cluster.", server)
	if forceKill {
		message = fmt.Sprintf("Warning: The observer will be killed before '%s' delete from obcluster, Please confirm.", server)
	}
	pass, err := stdio.Confirm(message)
	if err != nil {
		return nil, errors.New("ask for scale-in confirmation failed")
	}
	if !pass {
		return nil, errors.New("scale-in cancelled")
	}

	scaleInParam := param.ClusterScaleInParam{
		AgentInfo: *targetAgentInfo,
		ForceKill: forceKill,
	}
	stdio.StartLoading("Calling API to delete server")
	var dag task.DagDetailDTO
	if err := api.CallApiWithMethod(http.DELETE, constant.URI_OBSERVER_API_PREFIX, scaleInParam, &dag); err != nil {
		return nil, err
	}
	stdio.StopLoading()
	if dag.GenericDTO == nil {
		stdio.Print("There is no server need to be deleted.")
	}
	return &dag, nil
}

func deleteZone(zone string) (*task.DagDetailDTO, error) {
	pass, err := stdio.Confirm(fmt.Sprintf("Please confirm if you need to delete '%s' from obcluster.", zone))
	if err != nil {
		return nil, errors.Wrap(err, "ask for scale-in confirmation failed")
	}
	if !pass {
		return nil, errors.New("scale-in cancelled")
	}

	stdio.StartLoading("Calling API to delete zone")
	var dag task.DagDetailDTO
	if err := api.CallApiWithMethod(http.DELETE, constant.URI_ZONE_API_PREFIX+"/"+zone, nil, &dag); err != nil {
		return nil, err
	}
	stdio.StopLoading()

	if dag.GenericDTO == nil {
		stdio.Print("There is no zone need to be deleted.")
	}
	return &dag, nil
}
