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

import useDocumentTitle from '@/hook/useDocumentTitle';
import { useQueryParam } from '@/hook/useQueryParams';
import * as v1Service from '@/service/obshell/v1';
import { setEncryptLocalStorage } from '@/util';
import { formatMessage } from '@/util/intl';
import { handleLoginSuccess } from '@/util/login';
import { message } from '@oceanbase/design';
import { Login } from '@oceanbase/ui';
import { history, useDispatch } from '@umijs/max';
import { useRequest } from 'ahooks';
import React from 'react';

interface Values {
  username: string;
  password: string;
}

const LoginPage: React.FC = () => {
  const callback = useQueryParam('callback');
  const dispatch = useDispatch();

  useDocumentTitle(formatMessage({ id: 'ocp-express.page.Login.Login', defaultMessage: '登录' }));

  const handleCallback = () => {
    if (callback) {
      history.push(callback.startsWith('/login') ? '/' : callback);
    } else {
      history.push('/');
    }
  };

  const { run: runLogin, loading } = useRequest(v1Service.login, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        handleLoginSuccess(res.data);

        if (callback) {
          handleCallback();
        } else {
          history.push('/overview');
        }
      } else {
        message.error(
          res.error?.message ||
            formatMessage({
              id: 'ocp-v2.src.util.request.AuthenticationFailedAccessDenied',
              defaultMessage: '验证失败：访问被拒绝',
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

    setEncryptLocalStorage('username', username);
    setEncryptLocalStorage('login', 'true');
    runLogin();
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
