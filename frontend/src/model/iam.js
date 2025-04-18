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
import { history } from 'umi';
import { message } from '@oceanbase/design';
import * as CustomService from '@/service/custom';
import { isURL } from '@/util';
import tracert from '@/util/tracert';

export const namespace = 'iam';

const model = {
  namespace,
  state: {
    userListData: {
      page: {
        totalElements: 0,
      },

      contents: [],
    },

    userData: {},
    roleListData: {
      page: {
        totalElements: 0,
      },

      contents: [],
    },

    roleData: {},
    rateLimitResourceListData: {},
    rateLimitPolicyParam: {},
  },

  effects: {
    *login({ payload, onSuccess, onFail }, { call }) {
      const res = yield call(CustomService.login, payload);
      if (res.successful) {
        if (onSuccess) {
          onSuccess(res);
        }
      } else if (onFail) {
        onFail(res);
      }
    },
    *logout(_, { call, put }) {
      const res = yield call(CustomService.logout);
      if (res.successful) {
        // 退出登录后，清空 tracert 的用户标识
        tracert.set({
          roleId: null,
        });
        const location = res.data && res.data.location;
        if (isURL(location)) {
          window.location.href = location;
        } else if (window.location.pathname !== '/login') {
          // 将当前的 pathname 和 search 记录在 state 中，以便登录后能跳转到之前访问的页面
          // 如果当前已经是登录页了，则没必要跳转
          history.push({
            pathname: '/login',
            query: {
              callback: `${window.location.pathname}${window.location.search}`,
            },
          });
        }
        // 重置当前登录用户的信息
        yield put({
          type: 'profile/update',
          payload: {
            userData: {},
          },
        });
        // 重置当前登录用户的权限
        yield put({
          type: 'auth/reset',
        });
      }
    },
  },

  reducers: {
    update(state, { payload }) {
      return { ...state, ...payload };
    },
  },
};

export default model;
