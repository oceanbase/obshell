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

package instance

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obshell/seekdb/agent/constant"
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
	"github.com/oceanbase/obshell/seekdb/agent/lib/path"

	"github.com/oceanbase/obshell/seekdb/client/command"
	clientconst "github.com/oceanbase/obshell/seekdb/client/constant"
	cmdlib "github.com/oceanbase/obshell/seekdb/client/lib/cmd"
	"github.com/oceanbase/obshell/seekdb/client/lib/stdio"
	"github.com/oceanbase/obshell/seekdb/client/utils/api"
	"github.com/oceanbase/obshell/seekdb/client/utils/printer"
)

type StartObserverFlags struct {
	verbose     bool
	skipConfirm bool
}

func newStartCmd() *cobra.Command {
	opts := &StartObserverFlags{}
	startCmd := command.NewCommand(&cobra.Command{
		Use:     CMD_START,
		Short:   "Start observer.",
		PreRunE: cmdlib.ValidateArgs,
		RunE: command.WithErrorHandler(func(cmd *cobra.Command, args []string) error {
			stdio.SetVerboseMode(opts.verbose)
			stdio.SetSilenceMode(false)
			return observerStart()
		}),
		Example: startCmdExample(),
	})

	startCmd.Flags().SortFlags = false
	startCmd.VarsPs(&opts.verbose, []string{clientconst.FLAG_VERBOSE, clientconst.FLAG_VERBOSE_SH}, false, "Activate verbose output", false)
	return startCmd.Command
}

func observerStart() (err error) {
	agentStatus, err := api.GetMyAgentStatus()
	if err != nil {
		return
	}

	if agentStatus.Agent.IsUnidentified() {
		err = handleTakeoverForStart()
	} else if agentStatus.Agent.IsClusterAgent() {
		if agentStatus.UnderMaintenance {
			if err = CheckAgentMaintenance(); err != nil {
				log.Errorf("check all agent maintain status failed: %v", err)
				return err
			}
		}
		stdio.Verbosef("start seekdb in %s", path.AgentDir())
		err = callStartApi()
	} else {
		err = errors.Occur(errors.ErrCommonUnexpected, "current my agent is %s", agentStatus.Agent.GetIdentity())
	}
	return
}

func callStartApi() (err error) {
	uri := constant.URI_SEEKDB_API_PREFIX + constant.URI_START
	if err = callEmerTypeApi(uri, nil); err != nil {
		log.Errorf("call start api failed: %v", err)
		return
	}
	return
}

func CheckAgentMaintenance() error {
	stdio.Verbose("check agent's maintenance")
	maintainDag, err := api.GetAgentLastMaintenanceDag()
	if err != nil {
		return err
	}
	if maintainDag == nil {
		return nil
	} else if IsInstanceLifecycleDag(maintainDag) {
		return hanldUnderMaintenance(maintainDag)
	}
	return errors.Occur(errors.ErrAgentUnderMaintenanceDag)
}

func hanldUnderMaintenance(maintainDag *task.DagDetailDTO) error {
	log.Info("current under maintenance")
	if maintainDag != nil {
		stdio.Warn("The observer is currently under maintenance.")
		printer.PrintDagsTable(maintainDag)
		autoPass, err := stdio.Confirm("Would you like to automatically finish prerequisite tasks, regardless of whether they are currently running?")
		if err != nil {
			return errors.Wrap(err, "ask for auto finish dag confirmation failed")
		}
		if autoPass {
			return autoFinishMaintainDag(maintainDag)
		}
	}
	return nil
}

func autoFinishMaintainDag(dag *task.DagDetailDTO) error {
	stdio.StartLoadingf("Auto finish task '%s'", dag.GenericID)
	currDag, err := api.GetDagDetail(dag.GenericID)
	if err != nil {
		stdio.LoadErrorf("Sorry, get task '%s' failed", dag.GenericID)
		return err
	}
	if currDag.IsSucceed() {
		stdio.LoadSuccessf("Task '%s' has been finished successfully.", dag.GenericID)
		return nil
	}

	if currDag.IsFailed() {
		if err = passDag(dag.GenericID); err != nil {
			stdio.Verbosef("pass dag %s failed with error %v", dag.GenericID, err)
			return err
		}
	}
	if currDag.IsRunning() {
		if err = cancelAndPassDag(dag.GenericID); err != nil {
			return err
		}
	}
	if err != nil {
		stdio.LoadErrorf("Sorry, auto finish task '%s' failed", dag.GenericID)
		return err
	}
	stdio.LoadSuccessf("Task '%s' has been finished successfully.", dag.GenericID)
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
		dag, err1 := api.GetDagDetail(id)
		if err1 != nil {
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
		dag, err1 := api.GetDagDetail(id)
		if err1 != nil {
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
	return errors.Occurf(errors.ErrCommonUnexpected, "wait dag %s finished time out", id)
}

func startCmdExample() string {
	return `  obshell seekdb start
  obshell seekdb start --port 2886`
}

func IsInstanceLifecycleDag(dag *task.DagDetailDTO) bool {
	return dag.Name == DAG_START_OBSERVER ||
		dag.Name == DAG_STOP_OBSERVER ||
		dag.Name == DAG_RESTART_OBSERVER
}
