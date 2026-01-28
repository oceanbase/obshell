import { formatMessage } from '@/util/intl';
import { Button } from '@oceanbase/design';
import { directTo } from '@oceanbase/util';
import React from 'react';
import { history } from 'umi';

export interface TaskLabelProps {
  taskId?: number | string;
  openMode?: 'current' | 'new' | false;
  mode?: 'link' | 'button';
}

const TaskLabel: React.FC<TaskLabelProps> = ({ mode = 'button', taskId, openMode = 'current' }) => {
  return (
    <>
      {mode === 'button' ? (
        <Button
          size="small"
          onClick={() => {
            if (openMode === 'new') {
              directTo(`/task/${taskId}`);
            } else {
              history.push(`/task/${taskId}`);
            }
          }}
        >
          {formatMessage({
            id: 'OBShell.src.component.TaskLabel.ViewTasks',
            defaultMessage: '查看任务',
          })}
        </Button>
      ) : (
        <a
          onClick={() => {
            if (openMode === 'new') {
              directTo(`/task/${taskId}`);
            } else {
              history.push(`/task/${taskId}`);
            }
          }}
        >
          {formatMessage({
            id: 'OBShell.src.component.TaskLabel.ViewTasks',
            defaultMessage: '查看任务',
          })}
        </a>
      )}
    </>
  );
};

export default TaskLabel;
