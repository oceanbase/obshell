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

func (s *Scheduler) advanceNode(node *task.Node) error {
	log.withScheduler(s).Infof("advance node %d", node.GetID())
	if node.IsPending() {
		if err := s.startNode(node); err != nil {
			return errors.Wrapf(err, "start node %d error", node.GetID())
		}
	}

	if node.IsFinished() {
		log.withScheduler(s).Infof("node %d is finished", node.GetID())
		return nil
	}

	isFinished, isSucceed, err := s.advanceTask(node)
	if err != nil {
		return errors.Wrapf(err, "node %d advance sub task error", node.GetID())
	}

	if isFinished {
		if isSucceed {
			node.SetState(task.SUCCEED)
		} else {
			node.SetState(task.FAILED)
		}
		if err = s.service.FinishNode(node); err != nil {
			return errors.Wrapf(err, "finish node %d error", node.GetID())
		}
	}

	return nil
}

func (s *Scheduler) mergeContext(node *task.Node) error {
	upstreamNode := node.GetUpstream()
	if upstreamNode != nil {
		subTasks, err := s.service.GetSubTasks(upstreamNode)
		if err != nil {
			return errors.Wrap(err, "get sub tasks error")
		}
		for _, subTask := range subTasks {
			node.MergeContext(subTask.GetContext())
		}
	}
	return nil
}

func (s *Scheduler) startNode(node *task.Node) error {
	if err := s.mergeContext(node); err != nil {
		return errors.Wrap(err, "merge context error")
	}
	node.SetState(task.RUNNING)
	if err := s.service.StartNode(node); err != nil {
		return errors.Wrap(err, "start node error")
	}
	return nil
}
