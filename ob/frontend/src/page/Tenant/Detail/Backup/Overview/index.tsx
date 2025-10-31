import Task from '@/page/Backup/Backup/Task';
import { useSelector } from 'umi';
import React, { useEffect } from 'react';
import { Col, Row } from '@oceanbase/design';
import scrollIntoView from 'scroll-into-view';
import styles from './index.less';

interface DetailProps {
  tenantName: string;
  // 是否有备份策略
  hasBackupStrategy?: boolean;
  taskTab?: string;
  taskSubTab?: string;
}

const Detail: React.FC<DetailProps> = ({ tenantName, hasBackupStrategy, taskTab, taskSubTab }) => {
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  // 支持租户数据备份且未指定 taskTab，才默认展示数据备份任务
  // const [tab, setTab] = useState(taskTab || 'data');
  // const [range, setRange] = useState<Moment[]>([]);
  // const startTime = range && range[0] && range[0].format(RFC3339_DATE_TIME_FORMAT);
  // const endTime = range && range[1] && range[1].format(RFC3339_DATE_TIME_FORMAT);

  useEffect(() => {
    if (taskTab) {
      const taskCard = document.getElementById('ocp-backup-task-card');
      if (taskCard) {
        scrollIntoView(taskCard, {
          align: {
            top: 0,
          },
        });
      }
    }
  }, []);

  return (
    <Row gutter={[0, 16]} className={styles.container}>
      <Col span={24}>
        <Task
          tenantName={tenantName}
          obVersion={tenantData.obVersion}
          taskTab={taskTab}
          taskSubTab={taskSubTab}
          logOperationsFilter={(opts = []) => {
            // 没有备份策略时，不允许进行日志操作，只能查看错误信息
            if (!hasBackupStrategy) return opts.filter(opt => opt.value === 'viewErrorMessage');
            return opts;
          }}
        />
      </Col>
    </Row>
  );
};

export default Detail;
