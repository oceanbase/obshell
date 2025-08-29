import { getSelects } from '@/constant/backup';
import { DATE_TIME_FORMAT_DISPLAY, RFC3339_DATE_TIME_FORMAT } from '@/constant/datetime';
import { formatMessage } from '@/util/intl';
import Data from '@/page/Backup/Backup/Data';
import Log from '@/page/Backup/Backup/Log';
import RestoreTask from '@/page/Backup/Restore/RestoreTask';
import React, { useState } from 'react';
import { useLocation } from 'umi';
import type { Moment } from 'moment';
import { Card, Segmented } from '@oceanbase/design';
import { Ranger } from '@oceanbase/ui';
export interface TaskProps {
  tenantName?: string;
  obVersion?: string;
  taskTab?: string;
  taskSubTab?: string;
  status?: string;
  // 是否展示恢复任务
  showRestoreTask?: boolean;
  onTabChange?: (key: string) => void;
  onSubTabChange?: (key: string) => void;
  onLogBackupTaskSuccess?: (logBackupTaskList: API.LogBackupTask[]) => void;
  logOperationsFilter?: (operations: Global.OperationItem[]) => Global.OperationItem[]; // 是否允许操作
}

const Task: React.FC<TaskProps> = ({
  tenantName,
  obVersion,
  taskTab,
  taskSubTab,
  status,
  showRestoreTask = true,
  onTabChange,
  onSubTabChange,
  onLogBackupTaskSuccess,
  logOperationsFilter,
}) => {
  const { pathname } = useLocation();

  const [tab, setTab] = useState(taskTab || 'backup');
  const [subTab, setSubTab] = useState(taskSubTab || 'data');
  const [range, setRange] = useState<Moment[]>([]);
  const startTime = range && range[0] && range[0].format(RFC3339_DATE_TIME_FORMAT);
  const endTime = range && range[1] && range[1].format(RFC3339_DATE_TIME_FORMAT);

  return (
    <Card
      // 设置元素 ID，用于获取滚动对象，方便实现自动滚动到当前位置的交互效果
      id="ocp-backup-task-card"
      {...(tab === 'restore' ? { className: 'card-without-padding' } : {})}
      bordered={false}
      activeTabKey={tab}
      onTabChange={key => {
        setTab(key);
        setSubTab('data');
        if (onTabChange) {
          onTabChange(key);
        }
      }}
      tabList={[
        {
          key: 'backup',
          tab: formatMessage({
            id: 'ocp-v2.Backup.Overview.Task.BackupTask',
            defaultMessage: '备份任务',
          }),
        },
        ...(showRestoreTask
          ? [
              {
                key: 'restore',
                tab: formatMessage({
                  id: 'ocp-v2.Backup.Overview.Task.RestoreTask',
                  defaultMessage: '恢复任务',
                }),
              },
            ]
          : []),
      ]}
      tabBarExtraContent={
        <Ranger
          value={range}
          format={DATE_TIME_FORMAT_DISPLAY}
          selects={getSelects()}
          onChange={(value: Moment[]) => {
            setRange(value || []);
          }}
        />
      }
    >
      <Segmented
        value={subTab}
        onChange={(key: string) => {
          setSubTab(key);
          if (onSubTabChange) {
            onSubTabChange(key);
          }
        }}
        options={[
          tab === 'backup' && {
            label: formatMessage({
              id: 'ocp-v2.Backup.Overview.Task.DataBackup',
              defaultMessage: '数据备份',
            }),
            value: 'data',
          },
          tab === 'backup' && {
            label: formatMessage({
              id: 'ocp-v2.Backup.Overview.Task.LogBackup',
              defaultMessage: '日志备份',
            }),
            value: 'log',
          },
        ].filter(Boolean)}
      />
      {tab === 'backup' && (
        <>
          {subTab === 'data' && (
            <Data
              tenantName={tenantName}
              obVersion={obVersion}
              status={status}
              startTime={startTime}
              endTime={endTime}
            />
          )}

          {subTab === 'log' && (
            <Log
              tenantName={tenantName}
              startTime={startTime}
              endTime={endTime}
              onSuccess={onLogBackupTaskSuccess}
              obVersion={obVersion}
              operationsFilter={logOperationsFilter}
            />
          )}
        </>
      )}

      {tab === 'restore' && (
        <div
          style={{
            // 恢复任务没有 subTab，需要设置 marginTop: 16，以抵消 Tabs 占据的空间
            marginTop: -24,
          }}
        >
          <RestoreTask tenantName={tenantName} startTime={startTime} endTime={endTime} />
        </div>
      )}
    </Card>
  );
};

export default Task;
