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

import React from 'react';
import { Col, Row } from '@oceanbase/design';
import TaskGraph from './TaskGraph';
import LogCard from './LogCard';

export interface LogProps {
  ref?: any;
  taskData?: API.TaskInstance;
  onOperationSuccess: () => void;
  subtask?: API.SubtaskInstance;
  logData?: any[];
  logLoading?: boolean;
  logPolling?: boolean;
  onSubtaskChange: (subtaskId?: number | undefined) => void;
}

const Log: React.FC<LogProps> = React.forwardRef(
  (
    { taskData, onOperationSuccess, subtask, logData, logLoading, logPolling, onSubtaskChange },
    ref
  ) => {
    return (
      <Row>
        <Col xxl={{ span: 9 }} xl={{ span: 12 }} md={{ span: 12 }} style={{ paddingRight: 8 }}>
          <TaskGraph
            ref={ref}
            taskData={taskData}
            onOperationSuccess={onOperationSuccess}
            subtask={subtask}
            onSubtaskChange={onSubtaskChange}
          />
        </Col>
        <Col xxl={{ span: 15 }} xl={{ span: 12 }} md={{ span: 12 }}>
          <LogCard
            taskData={taskData}
            onOperationSuccess={onOperationSuccess}
            subtask={subtask}
            logData={logData}
            logLoading={logLoading}
            logPolling={logPolling}
          />
        </Col>
      </Row>
    );
  }
);

export default Log;
