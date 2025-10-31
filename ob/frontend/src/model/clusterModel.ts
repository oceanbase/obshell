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

import { useSetState, useRequest } from 'ahooks';
import * as ObClusterController from '@/service/ocp-express/ObClusterController';

export default () => {
  const [state, setState] = useSetState({
    clusterData: {} as API.ClusterInfo,
    clusterDataLoading: false,
  });

  const { run: getClusterData } = useRequest(ObClusterController.getClusterInfo, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        setState({
          clusterData: res.data,
        });
      }
      setState({
        clusterDataLoading: false,
      });
    },
  });

  return {
    getClusterData: (...prams: Parameters<typeof getClusterData>) => {
      setState({
        clusterDataLoading: true,
      });
      return getClusterData(...prams);
    },
    update: setState,
    ...state,
  };
};
