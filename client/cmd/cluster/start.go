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
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/agent/config"
	"github.com/oceanbase/obshell/agent/constant"
	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/errors"
	"github.com/oceanbase/obshell/agent/executor/ob"
	"github.com/oceanbase/obshell/agent/global"
	ocsagentlog "github.com/oceanbase/obshell/agent/log"
	clientconst "github.com/oceanbase/obshell/client/constant"
	cmdlib "github.com/oceanbase/obshell/client/lib/cmd"
	"github.com/oceanbase/obshell/client/lib/stdio"
	"github.com/oceanbase/obshell/client/utils/api"
	"github.com/oceanbase/obshell/client/utils/printer"
	"github.com/oceanbase/obshell/param"
)

type SSHFlags struct {
	user       string
	port       string
	password   string
	keyfile    string
	passphrase string
}

type ClusterStartFlags struct {
	scopeFlags
	SSHFlags
	id          string
	verbose     bool
	skipConfirm bool
}

type scopeFlags struct {
	server string
	zone   string
	global bool
}

func newStartCmd() *cobra.Command {
	opts := &ClusterStartFlags{}
	startCmd := &cobra.Command{
		Use:     CMD_START,
		Short:   "Start observers within the specified range.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			ocsagentlog.InitLogger(config.DefaultClientLoggerConifg())
			stdio.SetSkipConfirmMode(opts.skipConfirm)
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			if err := clusterStart(opts); err != nil {
				if stdio.IsBusy() {
					stdio.LoadFailed(err.Error())
				} else {
					stdio.Error(err.Error())
				}
				return err
			}
			return nil
		},
		Example: startCmdExample(),
	}

	startCmd.Flags().SortFlags = false
	startCmd.Flags().StringVarP(&opts.server, FLAG_SERVER, FLAG_SERVER_SH, "", "The operations address of the target server to start. Separated by commas if multiple servers are specified. The format should be ip:port")
	startCmd.Flags().StringVarP(&opts.zone, FLAG_ZONE, FLAG_ZONE_SH, "", "Start all servers within the specified zone. Separated by commas if multiple zones are specified")
	startCmd.Flags().BoolVarP(&opts.global, FLAG_ALL, FLAG_ALL_SH, false, "Start all servers within the cluster")
	startCmd.Flags().StringVarP(&opts.id, FLAG_ID, FLAG_ID_SH, "", "ID of the previous start/stop task. Separated by commas if multiple tasks are specified")
	startCmd.Flags().BoolVarP(&opts.verbose, clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH, false, "Activate verbose output")
	startCmd.Flags().BoolVarP(&opts.skipConfirm, clientconst.FLAG_SKIP_CONFIRM, clientconst.FLAG_SKIP_CONFIRM_SH, false, "Skip the confirmation of start operation")

	// Define flags used for configuring SSH connections.
	startCmd.Flags().StringVarP(&opts.user, FLAG_SSH_USER, "", "", "The user name for the ssh connection.")
	startCmd.Flags().StringVarP(&opts.port, FLAG_SSH_PORT, "", "", "The port for the ssh connection.")
	startCmd.Flags().StringVarP(&opts.password, FLAG_USER_PASSWORD, "", "", "The password of the ssh user.")
	startCmd.Flags().StringVarP(&opts.keyfile, FLAG_SSH_KEY_FILE, "", "", "The private key file for the SSH connection.(only make sense when user_password is empty)")
	startCmd.Flags().StringVarP(&opts.passphrase, FLAG_SSH_KEY_PASSPHRASE, "", "", "The passphrase for the private key file.")

	startCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printer.PrintHelpFunc(cmd, []string{})
	})
	return startCmd
}

func clusterStart(flags *ClusterStartFlags) (err error) {
	select {
	case <-statusCh:
	case err = <-errorCh:
		return err
	}

	if err = validateScopeFlags(&flags.scopeFlags); err != nil {
		return
	}

	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return
	}

	if flags.server == "" && flags.zone == "" && !flags.global {
		flags.server = fmt.Sprintf("%s:%d", agentStatus.Agent.GetIp(), agentStatus.Agent.GetPort())
	}

	stdio.Verbosef("current my agent is %s", agentStatus.Agent.GetIdentity())
	if agentStatus.Agent.IsUnidentified() {
		err = handleTakeoverForStart(flags)
	} else if agentStatus.Agent.IsClusterAgent() {
		err = callStartApi(flags)
	} else {
		// Call start APIs depending on the role or state of the agent (takeover, master, follower).
		err = callStartEachApi(flags)
	}
	return
}

func callStartEachApi(flags *ClusterStartFlags) (err error) {
	switch getScopeType(&flags.scopeFlags) {
	case ob.SCOPE_SERVER:
		err = handleTakeoverForStart(flags)
	case ob.SCOPE_ZONE:
		err = errors.New("not support zone scope for non-cluster agent, please use -s or -a")
	case ob.SCOPE_GLOBAL:
		err = handleTakeoverForStart(flags)
	default:
		err = errors.New("invalid scope type")
	}
	return err
}

func callStartApi(flags *ClusterStartFlags) (err error) {
	if err = CheckAllAgentMaintenance(); err != nil {
		log.Errorf("check all agent maintain status failed: %v", err)
		return err
	}

	param := &param.StartObParam{
		Scope:             newScopeParam(&flags.scopeFlags),
		ForcePassDagParam: *newForcePassIdParam(flags.id),
	}

	uri := constant.URI_OB_API_PREFIX + constant.URI_START
	if err = callEmerTypeApi(uri, param); err != nil {
		log.Errorf("call start api failed: %v", err)
		return
	}
	return
}

func CheckAllAgentMaintenance() error {
	stdio.Verbose("check all agents' maintenance")
	mainDags, maintainDags, err := api.GetAllMainAndMaintainDag()
	if err != nil {
		return err
	}
	if len(mainDags)+len(maintainDags) == 0 {
		return nil
	} else {
		return hanldUnderMaintenance(mainDags, maintainDags)
	}
}

func hanldUnderMaintenance(mainDags, maintainDags []*task.DagDetailDTO) error {
	log.Info("current under maintenance")
	allDags := append(mainDags, maintainDags...)

	// If there are active maintenance DAGs, halt further maintenance activities.
	if len(maintainDags) > 0 {
		stdio.Error("Due to the cluster currently being under maintenance, other maintenance operations cannot be performed at this time.\nPlease address the ongoing maintenance tasks before attempting further actions.")
		printer.PrintDagsTable(allDags)
		stdio.Printf("Please view the task details by '%s/bin/obshell task show -i <ID> -d'", global.HomePath)
		return errors.New("Cluster is under maintenance")
	} else {
		// If there are non-emergency DAGs, offer to auto-finish tasks if confirmed by the user.
		stdio.Warn("The cluster is currently under maintenance.")
		printer.PrintDagsTable(mainDags)
		autoPass, err := stdio.Confirm("Would you like to automatically finish prerequisite tasks, regardless of whether they are currently running?")
		if err != nil {
			return err
		}
		if autoPass {
			return autoFinishMainDag(mainDags)
		}
	}
	return nil
}

func autoFinishMainDag(dags []*task.DagDetailDTO) error {
	for _, dag := range dags {
		stdio.StartLoadingf("Auto finish task '%s'", dag.GenericID)
		currDag, err := api.GetDagDetail(dag.GenericID)
		if err != nil {
			return err
		}
		if currDag.IsSucceed() {
			stdio.LoadSuccessf("Task '%s' has been finished successfully.", dag.GenericID)
			continue
		}

		if currDag.IsFailed() {
			err = cancelAndPassSubDags(currDag)
		}
		if currDag.IsRunning() {
			err = cancelMainAndPassSubDags(currDag)
		}
		if err != nil {
			stdio.LoadErrorf("Sorry, auto finish task '%s' failed", dag.GenericID)
			return err
		}
		stdio.LoadSuccessf("Task '%s' has been finished successfully.", dag.GenericID)
	}
	return nil
}

func cancelAndPassSubDags(dag *task.DagDetailDTO) (err error) {
	subDagIDs, ok := api.GetSubDagIDs(dag)
	if !ok {
		return fmt.Errorf("get sub dags of %s failed", dag.GenericID)
	}
	stdio.Verbosef("sub dags of %s is %v", dag.GenericID, subDagIDs)

	for _, id := range subDagIDs {
		subDag, err := api.GetDagDetail(id)
		if err != nil {
			return errors.Wrapf(err, "get sub dag %s failed", id)
		}
		stdio.Verbosef("sub dag %s state is %s", id, subDag.State)
		if subDag.IsSucceed() {
			continue
		}

		if subDag.IsFailed() {
			if err = passDag(id); err != nil {
				return err
			}
		}

		if err = cancelAndPassDag(id); err != nil {
			return err
		}
	}
	return nil
}

// cancelMainAndPassSubDags attempts to cancel the main DAG and ensures subsequent
// passing of its sub-DAGs if the cancellation is successful.
func cancelMainAndPassSubDags(dag *task.DagDetailDTO) (err error) {
	// Attempt to cancel the main DAG using its generic ID and check if it is already succeeded.
	// If main dag isSucceed, not need to continue
	succeed, err := cancelDag(dag.GenericID)
	if err != nil {
		return err
	}
	if succeed {
		return nil
	}

	// Wait for the main DAG to be cancelled before proceeding.
	if err = waitDagFinished(dag.GenericID); err != nil {
		return err
	}

	// Once the main DAG is finished, attempt to cancel and pass the associated sub-DAGs.
	if err = cancelAndPassSubDags(dag); err != nil {
		return err
	}

	return nil
}

func cancelAndPassDag(id string) (err error) {
	succeed, err := cancelDag(id)
	if err != nil {
		return err
	}
	if succeed {
		return nil
	}

	if err = waitDagFinished(id); err != nil {
		return err
	}

	return passDag(id)
}

func cancelDag(id string) (succeed bool, err error) {
	stdio.Verbosef("try to cancel %s", id)
	if err = api.CancelDag(id); err != nil {
		log.WithError(err).Warnf("cancel %s failed", id)
		dag, err := api.GetDagDetail(id)
		if err != nil {
			return false, errors.Wrapf(err, "get dag %s failed", dag.GenericID)
		}
		if dag.IsSucceed() {
			stdio.Verbosef("%s is succeed", dag.GenericID)
			return true, nil
		}
		if !dag.IsFailed() {
			return false, errors.Wrapf(err, "cancel dag %s failed", dag.GenericID)
		}
	}
	stdio.Verbosef("cancel %s successfully", id)
	return false, nil
}

func passDag(id string) (err error) {
	stdio.Verbosef("try to pass %s", id)
	if err = api.PassDag(id); err != nil {
		dag, err := api.GetDagDetail(id)
		if err != nil {
			return errors.Wrapf(err, "get dag %s failed", dag.GenericID)
		}
		if !dag.IsSucceed() {
			return errors.Wrapf(err, "pass dag %s failed", dag.GenericID)
		}
	}
	stdio.Verbosef("pass %s successfully", id)
	return nil
}

func waitDagFinished(id string) (err error) {
	stdio.Verbosef("wait dag %s finished", id)
	for i := 0; i < 3; i++ {
		dag, err := api.GetDagDetail(id)
		if err != nil {
			return errors.Wrapf(err, "get dag details %s failed", dag.GenericID)
		}
		if dag.IsFinished() {
			stdio.Verbosef("%s is finished", id)
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("wait dag %s finished time out", id)
}

func newForcePassIdParam(idStr string) (p *param.ForcePassDagParam) {
	p = &param.ForcePassDagParam{}
	if idStr != "" {
		p.ID = strings.Split(strings.TrimSpace(idStr), ",")
	}
	return
}

func getScopeType(flags *scopeFlags) string {
	if flags.server != "" {
		return ob.SCOPE_SERVER
	}
	if flags.zone != "" {
		return ob.SCOPE_ZONE
	}
	if flags.global {
		return ob.SCOPE_GLOBAL
	}
	return ob.SCOPE_SERVER
}

func validateScopeFlags(flags *scopeFlags) error {
	stdio.Verbosef("validate cmd flags %+v", flags)
	if flags.server != "" && flags.zone != "" && flags.global {
		return errors.New("-s, -z and -a cannot be specified at the same time")
	}
	if flags.server != "" && flags.zone != "" {
		return errors.New("-s and -z cannot be specified at the same time")
	}
	if flags.server != "" && flags.global {
		return errors.New("-s and -a cannot be specified at the same time")
	}
	if flags.zone != "" && flags.global {
		return errors.New("-z and -a cannot be specified at the same time")
	}
	return nil
}

func startCmdExample() string {
	return `  obshell cluster start -s 192.168.1.1:2886
  obshell cluster start -z zone1,zone2
  obshell cluster start -a`
}
