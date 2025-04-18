import React, { useState } from 'react';
import { flatten } from 'lodash';
import { Alert, Typography } from '@oceanbase/design';
import { joinComponent } from '@oceanbase/util';

const { Paragraph } = Typography;

export interface BinlogAssociationMsgProps {
  clusterType: string;
  type: string;
  binlogService: API.TenantBinlogService[] | [];
}

const BinlogAssociationMsg: React.FC<BinlogAssociationMsgProps> = ({
  clusterType,
  type,
  binlogService = [],
}) => {
  const [expanded, setExpanded] = useState(false);

  const instanceList = flatten((binlogService || []).map(item => item?.nodes)) || [];

  const instanceNames = instanceList.map(item => item.instanceName) || [];

  const TEXT = (
    <>
      <Paragraph
        style={{ marginBottom: 0 }}
        ellipsis={{
          rows: 2,
          expandable: 'collapsible',
          expanded,
          onExpand: (_, info) => setExpanded(info.expanded),
          tooltip: {
            placement: 'top',
          },
        }}
      >
        当前 {clusterType} 与 Binlog 实例：
        {joinComponent(instanceNames, item => (
          <span style={{ marginBottom: 0 }}>{item}</span>
        ))}
      </Paragraph>
      关联，本操作会造成 Binlog 服务与下游业务中断， 请谨慎评估业务影响后再进行操作。
    </>
  );

  return (
    <>
      {instanceNames?.length > 0 ? (
        <>
          {type === 'text' ? (
            TEXT
          ) : (
            <Alert
              type="warning"
              showIcon={true}
              message={TEXT}
              style={{ marginBottom: 24, color: '#132039' }}
            />
          )}
        </>
      ) : null}
    </>
  );
};

export default BinlogAssociationMsg;
