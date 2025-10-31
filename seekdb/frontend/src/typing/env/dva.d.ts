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

/* eslint-disable */
import { History } from 'history';
import { Action, Dispatch, Reducer } from 'redux';
import { Effect } from 'redux-saga';
import {
  actionChannel,
  all,
  call,
  cancel,
  join,
  put,
  race,
  select,
  take,
} from 'redux-saga/effects';

export interface CredentialError {
  clusterName: string;
  tenantName: string;
  username: string;
  errorMessage: string;
}

interface DvaLoading {
  global: boolean;
  effects: { [key: string]: boolean | undefined };
  models: { [key: string]: boolean | undefined };
}

type EffectSaga = (action: DvaAction, effects: EffectsCommandMap) => Generator<Effect, any, any>;
type WatcherSaga = (effects: EffectsCommandMap) => Generator<Effect, any, any>;
interface EffectsMapObject {
  [key: string]:
    | EffectSaga
    | [WatcherSaga, { type: 'watcher' }]
    | [EffectSaga, { type: 'takeEvery' }]
    | [EffectSaga, { type: 'takeLatest' }]
    | [EffectSaga, { type: 'throttle'; ms: number }]
    | [EffectSaga, { type: 'poll'; delay: number }];
}
interface SubscriptionAPI {
  history: History;
  dispatch: Dispatch<any>;
}
type Subscription = (api: SubscriptionAPI, done: Function) => void;
interface SubscriptionsMapObject {
  [key: string]: Subscription;
}

interface EffectsCommandMap {
  all: typeof all;
  call: typeof call;
  put: typeof put;
  select: typeof select;
  take: typeof take;
  cancel: typeof cancel;
  join: typeof join;
  race: typeof race;
  actionChannel: typeof actionChannel;
}

export type ReducersMapObject<S, A extends Action = Action> = {
  [key: string]: Reducer<S, A>;
};

declare global {
  interface DvaEntireState {
    loading: DvaLoading;
    [key: string]: any;
  }

  interface DvaModel<T_State> {
    namespace: string;
    state: T_State;
    reducers?: ReducersMapObject<T_State, DvaAction>;
    effects?: EffectsMapObject;
    subscriptions?: SubscriptionsMapObject;
  }

  /**
   * Action which will trigger reducer/effect execution
   * @type P: Type of payload
   * @type C: Type of callback
   */
  interface DvaAction<P = any, C = any> {
    type: string;
    payload?: P;
    callback?: C;
    [key: string]: any;
  }

  type DvaEffectsCommandMap = EffectsCommandMap;
  /**
   * The return type of yield call
   */
  type CallReturn<T extends (...args: any) => any> = Unpacked<ReturnType<T>>;

  interface DefaultRootState {
    loading: DvaLoading;
    router: {
      location: {
        query?: {
          [key: string]: string;
        };
      };
    };
    global: {
      themeMode: 'light' | 'dark';
      publicKey: string;
      uiMode: string;
      // 后端返回的应用信息没有定义类型，需要手动指定
      appInfo: Global.AppInfo;
      systemInfo: API.SystemInfo;
      showCredentialModal: boolean;
      // 密码箱连接缺失或者连接失败对应的结构化信息
      credentialErrorData: CredentialError;
    };
    profile: {
      password?: string;
      publicKey?: string;
      userData: API.AuthenticatedUser;
    };
    iam: {
      userListData: {
        page?: {
          totalElements?: number;
        };
        contents?: API.User[];
      };
    };
    cluster: {
      clusterData: API.ClusterInfo;
    };
    tenant: {
      tenantData: API.TenantInfo;
      unitSpecList: API.UnitSpec[];
    };
    obproxy: {
      obproxyCluster: API.ObproxyCluster;
    };
    backup: {
      storageConfigList: API.ObBackupServiceStorageConfig[];
    };
    alarm: {
      channelListData?: {
        page?: {
          totalElements?: number;
        };
        contents?: API.OcpChannel[];
      };
    };
    log: {
      logData: any;
    };
    compute: {
      hostTypeListData: API.PaginatedResponsePaginatedData_HostType_;
      idcListData: API.PaginatedResponsePaginatedData_Idc_;
      regionListData: API.PaginatedResponsePaginatedData_Region_;
    };
    report: {
      collapsed: boolean;
      siderWidth: number;
    };
  }
}
