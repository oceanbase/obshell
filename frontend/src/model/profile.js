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
import * as ProfileService from '@/service/ocp-express/ProfileController';
import { DEFAULT_LIST_DATA } from '@/constant';
import { isURL, setEncryptLocalStorage } from '@/util';
import tracert from '@/util/tracert';

export const namespace = 'profile';

const model = {
  namespace,
  state: {
    password: '',
    publicKey: '',
    userData: {},
    credentialListData: DEFAULT_LIST_DATA,
  },

  effects: {
    *getUserData(_, { call, put }) {
      // const res = yield call(ProfileService.userInfo);
      // if (res.successful) {
      //   const userData = res.data || {};
      //   // 获取当前登录用户详情后，设置 tracert 的用户标识
      //   tracert.set({
      //     roleId: `${window.location.host}_${userData.id}`,
      //   });
      //   yield put({
      //     type: 'update',
      //     payload: {
      //       userData,
      //     },
      //   });
      //   yield put({
      //     type: 'auth/setAuthData',
      //   });
      // }
      // return res;
    },
    *modifyUserPassword({ payload }, { call }) {
      const res = yield call(ProfileService.changePassword, payload);
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.profile.PasswordModificationIsSuccessfulYou',
            defaultMessage: '密码修改成功，需要重新登录',
          })
        );

        const location = res.data && res.data.location;
        if (isURL(location)) {
          window.location.href = location;
        } else {
          history.push('/login');
        }
      }
    },

    *logout({}, { put }) {
      yield put({
        type: 'update',
        payload: {
          password: null,
        },
      });

      setEncryptLocalStorage('password', '');
      setEncryptLocalStorage('login', 'false');
      history.push('/login');
    },
  },

  reducers: {
    update(state, { payload }) {
      return { ...state, ...payload };
    },
  },
};

export default model;
