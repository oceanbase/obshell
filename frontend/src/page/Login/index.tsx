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
import * as obService from '@/service/obshell/ob';
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
  const { themeMode } = useSelector((state: DefaultRootState) => state.global);

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
        if (callback) {
          handleCallback();
        } else {
          history.push('/overview');
        }
      } else {
        message.error(
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
    getObInfo();
  };

  const logoUrl = '/assets/logo/ob_dashboard_simple_logo_en.svg';

  return (
    <Login
      logo={logoUrl}
      bgImage={
        themeMode === 'dark'
          ? '/assets/login/background_img_dark.svg'
          : '/assets/login/background_img.svg'
      }
      title={
        <div>
          Welcome to <div>OB-Dashboard !</div>
        </div>
      }
      description="Let's start a happy journey"
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
