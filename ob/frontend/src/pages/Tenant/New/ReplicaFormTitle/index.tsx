import ContentWithQuestion from '@/component/ContentWithQuestion';
import { formatMessage } from '@/util/intl';
import { Col, Row, TooltipProps } from '@oceanbase/design';
import { useModel } from '@umijs/max';
import React from 'react';
import styles from './index.less';

interface ReplicaFormTitleProps {
  replicaTypeTooltipConfig?: TooltipProps;
}
const ReplicaFormTitle: React.FC<ReplicaFormTitleProps> = ({ replicaTypeTooltipConfig }) => {
  const { isSharedStorage } = useModel('cluster');
  const sharedStorageColSpan = isSharedStorage ? 4 : 5;
  return (
    <Row gutter={16} className={styles.replicaFormTitle}>
      <Col span={sharedStorageColSpan}>
        {formatMessage({
          id: 'ocp-express.Tenant.New.ZoneName',
          defaultMessage: 'Zone 名称',
        })}
      </Col>
      <Col span={sharedStorageColSpan}>
        <ContentWithQuestion
          content={formatMessage({
            id: 'ocp-express.Tenant.New.ReplicaType',
            defaultMessage: '副本类型',
          })}
          tooltip={replicaTypeTooltipConfig}
        />
      </Col>
      <Col span={sharedStorageColSpan}>
        {formatMessage({
          id: 'ocp-express.Tenant.New.UnitSpecifications',
          defaultMessage: 'Unit 规格',
        })}
      </Col>
      <Col span={sharedStorageColSpan}></Col>
      {isSharedStorage && <Col span={4}></Col>}
      <Col span={3}>
        {formatMessage({
          id: 'ocp-express.Tenant.New.UnitQuantity',
          defaultMessage: 'Unit 数量',
        })}
      </Col>
    </Row>
  );
};

export default ReplicaFormTitle;
