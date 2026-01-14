import { formatMessage } from '@/util/intl';
import React from 'react';
import MyCard from '@/component/MyCard';
import ZoneList from './ZoneList';

export interface ZoneListOrTopoProps {
  clusterData: API.ClusterInfo;
}

const ZoneListOrTopo: React.FC<ZoneListOrTopoProps> = ({ clusterData }) => {
  return (
    <div>
      <MyCard
        id="ocp-express-topo-card"
        title={formatMessage({
          id: 'ocp-express.Component.ZoneListOrTopo.TopologicalStructure',
          defaultMessage: '拓扑结构',
        })}
      >
        <ZoneList clusterData={clusterData} />
      </MyCard>
    </div>
  );
};

export default ZoneListOrTopo;
