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

import { formatMessage } from '@/util/intl';
import React, { useState, useRef, useImperativeHandle } from 'react';
import MyCard from '@/component/MyCard';
import type { ZoneListRef } from './ZoneList';
import ZoneList from './ZoneList';
import Topo from './Topo';

export interface ZoneListOrTopoRef extends ZoneListRef {
  setType: (type: string) => void;
}

export interface ZoneListOrTopoProps {
  clusterData: API.ClusterInfo;
  serverList: any[];
}

const ZoneListOrTopo = React.forwardRef<ZoneListOrTopoRef, ZoneListOrTopoProps>(
  ({ clusterData, serverList }, ref) => {
    const [type, setType] = useState('LIST');
    const zoneListRef = useRef<ZoneListRef>();

    // 向组件外部暴露 refresh 属性函数，可通过 ref 引用
    useImperativeHandle(ref, () => ({
      setType: (newType: string) => {
        setType(newType);
      },
      expandAll: () => {
        zoneListRef.current?.expandAll();
      },
      setStatusList: (statusList: API.ObServerStatus[]) => {
        zoneListRef.current?.setStatusList(statusList);
      },
    }));

    return (
      <div data-aspm="c304258" data-aspm-desc="集群拓扑" data-aspm-param={``} data-aspm-expo>
        <MyCard
          id="ocp-express-topo-card"
          title={formatMessage({
            id: 'ocp-express.Component.ZoneListOrTopo.TopologicalStructure',
            defaultMessage: '拓扑结构',
          })}
          // extra={
          //   <Radio.Group
          //     optionType="button"
          //     value={type}
          //     onChange={e => {
          //       setType(e.target.value);
          //       // 自动滚动到底部，保证拓扑图展示完全
          //       setTimeout(() => {
          //         const topoCardElement = document.getElementById('ocp-express-topo-card');
          //         if (topoCardElement) {
          //           scrollIntoView(topoCardElement, {
          //             align: {
          //               topOffset: 50,
          //             },
          //           });
          //         }
          //       }, 0);
          //     }}
          //   >
          //     <Radio.Button
          //       data-aspm-click="c304258.d308764"
          //       data-aspm-desc="集群拓扑-表格切换"
          //       data-aspm-param={``}
          //       data-aspm-expo
          //       value="LIST"
          //     >
          //       <UnorderedListOutlined />
          //     </Radio.Button>
          //     <Radio.Button
          //       data-aspm-click="c304258.d308761"
          //       data-aspm-desc="集群拓扑-图切换"
          //       data-aspm-param={``}
          //       data-aspm-expo
          //       value="TOPO"
          //     >
          //       <ApartmentOutlined />
          //     </Radio.Button>
          //   </Radio.Group>
          // }
          // data-aspm-expo
          // data-aspm-param={``}
        >
          {type === 'LIST' ? (
            <ZoneList ref={zoneListRef} serverList={serverList} clusterData={clusterData} />
          ) : (
            <Topo clusterData={clusterData} serverList={serverList} />
          )}
        </MyCard>
      </div>
    );
  }
);

export default ZoneListOrTopo;
