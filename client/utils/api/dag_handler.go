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

	"github.com/oceanbase/obshell/agent/engine/task"
	"github.com/oceanbase/obshell/agent/meta"
	"github.com/oceanbase/obshell/client/lib/stdio"
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
	return fmt.Errorf("Wait dag %s finished time out", dh.GenericID)
}

func (dh *DagHandler) PrintDagStage() (err error) {
	var failed bool
	dh.ctx, dh.cancel = context.WithCancel(context.Background())
	for i := 1; i <= dh.Dag.MaxStage && !dh.Dag.IsSucceed(); i = dh.Dag.Stage {
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

	if stage == 1 {
		msg = fmt.Sprintf("%s [%d/%d]", dh.Dag.Nodes[0].Name, dh.Dag.Stage, dh.Dag.MaxStage)
		stdio.StartOrUpdateLoading(msg)
	}

	for ; stage+1 < dh.Dag.Stage; stage++ {
		msg = fmt.Sprintf("%s [%d/%d]", dh.Dag.Nodes[stage-1].Name, stage, dh.Dag.MaxStage)
		stdio.LoadStageSuccess(msg)
		msg = fmt.Sprintf("%s [%d/%d]", dh.Dag.Nodes[stage].Name, stage+1, dh.Dag.MaxStage)
		stdio.StartOrUpdateLoading(msg)
	}

	if dh.Dag.Stage > stage {
		msg = fmt.Sprintf("%s [%d/%d]", dh.Dag.Nodes[stage-1].Name, stage, dh.Dag.MaxStage)
		stdio.LoadStageSuccess(msg)
		msg = fmt.Sprintf("%s [%d/%d]", dh.Dag.Nodes[stage].Name, dh.Dag.Stage, dh.Dag.MaxStage)
		stdio.StartOrUpdateLoading(msg)
		return true, nil
	}

	if dh.Dag.IsSucceed() {
		msg = fmt.Sprintf("%s [%d/%d]", dh.Dag.Nodes[stage-1].Name, dh.Dag.Stage, dh.Dag.MaxStage)
		stdio.LoadSuccess(msg)
		msg = fmt.Sprintf("Congratulations! '%s' task completed successfully.", dh.Dag.Name)
		stdio.Success(msg)
		return true, nil
	}

	if dh.Dag.IsFailed() {
		stdio.LoadFailedWithoutMsg()
		for _, log := range GetFailedDagLastLog(dh.Dag) {
			stdio.Error(log)
		}
		return true, fmt.Errorf("Sorry, task '%s' failed", dh.Dag.Name)
	}
	return false, nil
}
