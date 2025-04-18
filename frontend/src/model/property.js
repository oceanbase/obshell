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
import { message } from '@oceanbase/design';
import * as PropertyService from '@/service/ocp-express/PropertyController';
import { DEFAULT_LIST_DATA } from '@/constant';

export const namespace = 'property';

const model = {
  namespace,
  state: {
    propertyListData: DEFAULT_LIST_DATA,
  },

  effects: {
    *getPropertyListData({ payload }, { call, put }) {
      const res = yield call(PropertyService.findNonFatalProperties, payload);
      if (res.successful) {
        yield put({
          type: 'update',
          payload: {
            propertyListData: res.data || DEFAULT_LIST_DATA,
          },
        });
      }
    },
    *editProperty({ payload, onSuccess }, { call, put }) {
      const res = yield call(PropertyService.updateProperty, payload);
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.src.model.property.ParameterUpdatedSuccessfully',
            defaultMessage: '参数更新成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
        yield put({
          type: 'getPropertyListData',
          payload: {},
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
