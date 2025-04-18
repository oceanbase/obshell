/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

import { formatMessage } from '@/util/intl';
import { history, useDispatch, useSelector } from 'umi';
import React from 'react';
import { message } from '@oceanbase/design';
import { Login } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import * as obService from '@/service/obshell/ob';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { isEnglish } from '@/util';

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
  const { publicKey } = useSelector((state: DefaultRootState) => state.profile);

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
        message.error(res.error?.message);
      }
    },
  });

  const login = (values: Values) => {
    const { username, password } = values;
    // Save password to global state
    dispatch({
      type: 'profile/update',
      payload: {
        password,
      },
    });

    localStorage.setItem('password', password);
    getObInfo();
  };

  const logoUrl = isEnglish()
    ? themeMode === 'dark'
      ? '/assets/logo/ocp_express_logo_en_dark.svg'
      : '/assets/logo/ocp_express_logo_en.svg'
    : themeMode === 'dark'
    ? '/assets/logo/ocp_express_logo_zh_dark.svg'
    : '/assets/logo/ocp_express_logo_zh.svg';

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
          Welcome to <div>OCP Express !</div>
        </div>
      }
      description="Let's start a happy journey"
      showLocale={true}
      locales={['zh-CN', 'en-US']}
      loginProps={{
        loading: loading,
        onFinish: login,
      }}
    />
  );
};

export default LoginPage;
