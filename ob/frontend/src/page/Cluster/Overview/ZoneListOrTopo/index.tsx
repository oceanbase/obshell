import MyCard from '@/component/MyCard';
import { formatMessage } from '@/util/intl';
import React from 'react';
import ZoneList from './ZoneList';

const ZoneListOrTopo: React.FC = () => {
  return (
    <div>
      <MyCard
        id="ocp-express-topo-card"
        title={formatMessage({
          id: 'ocp-express.Component.ZoneListOrTopo.TopologicalStructure',
          defaultMessage: '拓扑结构',
        })}
      >
        <ZoneList />
      </MyCard>
    </div>
  );
};

export default ZoneListOrTopo;
