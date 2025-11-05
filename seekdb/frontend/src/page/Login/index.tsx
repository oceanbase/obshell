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
import { history, useDispatch, useSelector } from 'umi';
import React from 'react';
import { message } from '@oceanbase/design';
import { Login } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import * as obService from '@/service/obshell/observer';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { isEnglish, setEncryptLocalStorage } from '@/util';

interface LoginPageProps {
  location: {
    query: {
      callback?: string;
    };
  };
}

interface Values {
  username: string;
  password: string;
}

const LoginPage: React.FC<LoginPageProps> = ({
  location: {
    query: { callback },
  },
}) => {
  const dispatch = useDispatch();

  useDocumentTitle(formatMessage({ id: 'ocp-express.page.Login.Login', defaultMessage: '登录' }));

  const handleCallback = () => {
    if (callback) {
      history.push(callback.startsWith('/login') ? '/' : callback);
    } else {
      history.push('/');
    }
  };

  // use obService.getObInfo to simulate login
  const { run: getObInfo, loading } = useRequest(obService.getObInfo, {
    manual: true,
    defaultParams: [{}],
    onSuccess: res => {
      if (res.successful) {
        // 先更新 global 状态的 obInfoData
        dispatch({
          type: 'global/update',
          payload: {
            obInfoData: res.data || {},
          },
        });

        // 然后再进行跳转
        if (callback) {
          handleCallback();
        } else {
          history.push('/instance');
        }
      } else {
        message.error(
          res.error?.message ||
            formatMessage({
              id: 'ocp-v2.src.util.request.AuthenticationFailedAccessDenied',
              defaultMessage: '密码错误',
            })
        );
      }
    },
  });

  const login = (values: Values) => {
    const { username, password = '' } = values;
    // Save password to global state
    dispatch({
      type: 'profile/update',
      payload: {
        password,
      },
    });

    setEncryptLocalStorage('password', password);
    setEncryptLocalStorage('username', username);
    setEncryptLocalStorage('login', 'true');
    getObInfo();
  };

  const logoUrl = '/assets/logo/obshell_dashboard_login_logo.svg';

  return (
    <Login
      logo={logoUrl}
      bgImage={'/assets/login/background_img.svg'}
      showLocale={true}
      locales={['zh-CN', 'en-US']}
      loginProps={{
        loading: loading,
        onFinish: login,
        initialValues: {
          username: 'root',
        },
        passwordOptional: true,
        componentProps: {
          username: {
            disabled: true,
          },
        },
      }}
    />
  );
};

export default LoginPage;
