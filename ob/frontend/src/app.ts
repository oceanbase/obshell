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
import './tailwind.css';

export const dva = {
  config: {
    // 在 onError 函数当中做的处理，需要同步到 UseRequestProvider（全局搜索）当中进行相同处理。使用 useRequest 抛的错暂时不能被 onError 函数捕获到。
    onError(err: ErrorEvent, dispatch) {
      // 用来捕获全局错误，避免页面崩溃
      if (err.preventDefault) {
        err.preventDefault();
      }
      message.error(
        formatMessage(
          {
            id: 'ocp-express.src.app.DvaErrorErrmessage',
            defaultMessage: 'dva error：{errMessage}',
          },
          { errMessage: err.message }
        ),
        3
      );
      // 3010 为新版接口找不到密码箱连接的错误码
      if (err.message === '3010') {
        dispatch({
          type: 'global/update',
          payload: {
            showCredentialModal: true,
            credentialErrorData: err?.data || {},
          },
        });
      } else if (err.message === '301000') {
        // 301000 为旧版接口找不到密码箱连接的错误码
        dispatch({
          type: 'global/update',
          payload: {
            showCredentialModal: true,
          },
        });
      }
    },
  },
};
