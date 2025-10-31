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

package scheduler

import (
	"github.com/oceanbase/obshell/seekdb/agent/engine/task"
	"github.com/oceanbase/obshell/seekdb/agent/errors"
)

func (s *Scheduler) advanceDag(dag *task.Dag) error {
	var err error
	if dag.IsReady() {
		if err = s.startDag(dag); err != nil {
			return errors.Wrapf(err, "start dag %d error", dag.GetID())
		}
	}
	if dag.IsFinished() {
		return nil
	}

	log.withScheduler(s).Infof("advance dag %d", dag.GetID())
	stage := getCurrentStage(dag)
	node, err := s.service.GetNodeByStage(dag.GetID(), stage)
	if err != nil {
		return errors.Wrapf(err, "dag %d get node by stage error", dag.GetID())
	}

	if node.IsPending() && !node.IsRollback() {
		prevStage, ok := getPrevStage(dag)
		if ok {
			prevNode, err := s.service.GetNodeByStage(dag.GetID(), prevStage)
			if err != nil {
				return errors.Wrapf(err, "dag %d get node %d prev node error", dag.GetID(), node.GetID())
			}
			node.AddUpstream(prevNode)
		}
	}

	if err = s.advanceNode(node); err != nil {
		return errors.Wrapf(err, "dag %d advance node error", dag.GetID())
	}

	if node.IsFinished() {
		if node.IsSuccess() {
			nextStage, hasNext := getNextStage(dag)
			if hasNext {
				err = s.service.UpdateDagStage(dag, nextStage)
			} else {
				dag.SetStage(nextStage)
				err = s.service.FinishDagAsSucceed(dag)
			}
		} else {
			err = s.service.FinishDagAsFailed(dag)
		}
	}
	if err != nil {
		return errors.Wrapf(err, "update dag %d error", dag.GetID())
	}
	return nil
}

func getCurrentStage(dag *task.Dag) int {
	stage := dag.GetStage()
	maxStage := dag.GetMaxStage()
	if stage > maxStage {
		return maxStage
	}
	return stage
}

// startDag will set dag to running.
func (s *Scheduler) startDag(dag *task.Dag) error {
	dag.SetState(task.RUNNING)
	if dag.IsRollback() {
		stage := getCurrentStage(dag)
		dag.SetStage(stage)
	}
	return s.service.StartDag(dag)
}

func getNextStage(dag *task.Dag) (int, bool) {
	stage := getCurrentStage(dag)
	if dag.IsRollback() {
		if stage <= 1 {
			return 1, false
		}
		return stage - 1, true
	} else {
		maxStage := dag.GetMaxStage()
		if stage >= maxStage {
			return maxStage, false
		}
		return stage + 1, true
	}
}

func getPrevStage(dag *task.Dag) (int, bool) {
	stage := getCurrentStage(dag)
	if dag.IsRollback() {
		maxStage := dag.GetMaxStage()
		if stage >= maxStage {
			return maxStage, false
		}
		return stage + 1, true
	} else {
		if stage <= 1 {
			return 1, false
		}
		return stage - 1, true
	}
}
