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

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/oceanbase/obshell/ob/agent/engine/task"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/client/lib/stdio"
)

type DagHandler struct {
	GenericID   string
	Dag         *task.DagDetailDTO
	TargetAgent meta.AgentInfoInterface

	retryTimes   int
	currentStage int
	forUpgrade   bool
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewDagHandler(dag *task.DagDetailDTO) *DagHandler {
	return &DagHandler{
		GenericID: dag.GenericID,
		Dag:       dag,
	}
}

func NewDagHandlerWithAgent(dag *task.DagDetailDTO, agent meta.AgentInfoInterface) *DagHandler {
	return &DagHandler{
		GenericID:   dag.GenericID,
		Dag:         dag,
		TargetAgent: agent,
	}
}

func (dh *DagHandler) SetRetryTimes(retryTimes int) {
	dh.retryTimes = retryTimes
}
func (dh *DagHandler) SetForUpgrade() {
	dh.forUpgrade = true
}

func (dh *DagHandler) GetDag() (*task.DagDetailDTO, error) {
	var err error
	if dh.TargetAgent == nil {
		dh.Dag, err = GetDagDetail(dh.GenericID)
		if dh.Dag == nil && dh.forUpgrade {
			dh.Dag, err = GetDagDetailForUpgrade(dh.GenericID)
			// Double check by attempting regular retrieval if the upgrade-specific retrieval returns nil.
			if dh.Dag == nil {
				stdio.Verbose(err.Error())
				dh.Dag, err = GetDagDetail(dh.GenericID)
			}
		}
	} else {
		dh.Dag, err = GetDagDetailViaTCP(dh.TargetAgent, dh.GenericID)
	}
	return dh.Dag, err
}

func (dh *DagHandler) Retry() error {
	return sendDagOperatorRequest(task.RETRY, dh.GenericID)
}

func (dh *DagHandler) PassDag() error {
	return sendDagOperatorRequest(task.PASS, dh.GenericID)
}

func (dh *DagHandler) Rollback() error {
	return sendDagOperatorRequest(task.ROLLBACK, dh.GenericID)
}

func (dh *DagHandler) CancelDag() error {
	err := sendDagOperatorRequest(task.CANCEL, dh.GenericID)
	if err != nil {
		return err
	}
	if dh.cancel != nil {
		dh.cancel()
	}
	return nil
}

func (dh *DagHandler) waitDagFinished() error {
	for i := 0; i < 30; i++ {
		if dag, err := dh.GetDag(); err != nil {
			return err
		} else if dag.IsFinished() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.Occurf(errors.ErrCommonUnexpected, "Wait dag %s finished time out", dh.GenericID)
}

func (dh *DagHandler) PrintDagStage() (err error) {
	var failed bool
	dh.ctx, dh.cancel = context.WithCancel(context.Background())
	for i := 0; i <= dh.Dag.MaxStage && !dh.Dag.IsSucceed(); i = dh.Dag.Stage {
		failed, err = dh.waitDagFinishStage(i)
		if err != nil || failed {
			return
		}
	}
	return
}

// waitDagFinishStage will wait for the dag to finish the stage.
func (dh *DagHandler) waitDagFinishStage(stage int) (failed bool, err error) {
	for {
		select {
		case <-dh.ctx.Done():
			err = dh.waitDagFinished()
			return dh.Dag.IsFailed(), err
		case <-time.After(1 * time.Second):
			if finished, err := dh.chaseToLatestStage(stage); err != nil {
				return false, err
			} else if finished {
				return dh.Dag.IsFailed(), nil
			}
		}
	}
}

func (dh *DagHandler) getLoadMessage(stage int) string {
	switch dh.Dag.Operator {
	case task.RUN_STR:
		return fmt.Sprintf("%s [%d/%d]", dh.Dag.Nodes[stage-1].Name, stage, dh.Dag.MaxStage)
	case task.ROLLBACK_STR:
		return fmt.Sprintf("Rollback `%s` [%d/%d]", dh.Dag.Nodes[stage-1].Name, stage, dh.Dag.MaxStage)
	default:
		return ""
	}
}

// chaseToLatestStage will chase the dag to the latest stage, and return whether the prevStage is finished.
// When exiting the function, the current print must be the latest stage.
// when prevStage finished, return true, else return false.
func (dh *DagHandler) chaseToLatestStage(prevStage int) (finished bool, err error) {
	var msg string
	stage := prevStage
	_, err = dh.GetDag()
	if err != nil {
		if dh.retryTimes > 0 {
			stdio.Verbosef("%v, retry times: %d", err, dh.retryTimes)
			dh.retryTimes--
			time.Sleep(1 * time.Second)
			return false, nil
		}
		return false, err
	}

	if stage == 0 {
		switch dh.Dag.Operator {
		case task.RUN_STR:
			stage = 1
		case task.ROLLBACK_STR:
			stage = dh.Dag.Stage
		default:
			return false, errors.Occur(errors.ErrTaskDagOperatorNotSupport, dh.Dag.Operator)
		}
		stdio.StartOrUpdateLoading(dh.getLoadMessage(stage))
	}

	nextStage := func(stage int) int {
		if dh.Dag.Operator == task.ROLLBACK_STR {
			stage--
		} else {
			stage++
		}

		if stage < 1 {
			stage = 1
		} else if stage > dh.Dag.MaxStage {
			stage = dh.Dag.MaxStage
		}
		return stage
	}

	isLatestStage := func(stage int) bool {
		return dh.Dag.Stage == stage
	}

	for !isLatestStage(stage) {
		stdio.LoadStageSuccess(dh.getLoadMessage(stage))

		stage = nextStage(stage)
		stdio.StartOrUpdateLoading(dh.getLoadMessage(stage))
	}

	if dh.Dag.IsSucceed() {
		stdio.LoadSuccess(dh.getLoadMessage(stage))
		msg = fmt.Sprintf("Congratulations! %s '%s' task completed successfully.", dh.Dag.Operator, dh.Dag.Name)
		stdio.Success(msg)
		return true, nil
	}

	if dh.Dag.IsFailed() {
		stdio.LoadFailedWithoutMsg()
		for _, log := range GetFailedDagLastLog(dh.Dag) {
			stdio.Error(log)
		}
		return true, errors.Occurf(errors.ErrEmpty, "Sorry, task '%s' failed", dh.Dag.Name)
	}
	return true, nil
}
