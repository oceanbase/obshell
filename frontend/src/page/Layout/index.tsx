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

import { getLocale, history, useSelector, useDispatch, setLocale } from 'umi';
import React, { useEffect } from 'react';
import { ConfigProvider, Spin, theme } from '@oceanbase/design';
import { ChartProvider } from '@oceanbase/charts';
import en_US from '@oceanbase/ui/es/locale/en-US';
import zh_CN from '@oceanbase/ui/es/locale/zh-CN';
import { ThemeProvider } from 'antd-style';
import { useRequest } from 'ahooks';
import * as v1Service from '@/service/obshell/v1';
import BlankLayout from './BlankLayout';
import ErrorBoundary from '@/component/ErrorBoundary';
import GlobalStyle from './GlobalStyle';
import { getEncryptLocalStorage, setEncryptLocalStorage } from '@/util';
import { getStatistics } from '@/service/obshell/ob';
import { telemetryReport } from '@/service/custom';
import * as obService from '@/service/obshell/ob';
import moment from 'moment';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children, location }) => {
  const { themeMode } = useSelector((state: DefaultRootState) => state.global);
  const { query } = location;
  const { mode } = query || {};
  const { uiMode } = query || {};
  const isPasswordFreeLogin = mode === 'passwordFreeLogin';

  const locale = getLocale();
  const localeMap = {
    'en-US': en_US,
    'zh-CN': zh_CN,
  };

  const dispatch = useDispatch();

  useEffect(() => {
    // 设置标签页的 title
    document.title = 'OB-Dashboard';
  }, []);

  useEffect(() => {
    dispatch({
      type: 'global/update',
      payload: {
        uiMode,
      },
    });
  }, []);

  // request and save publicKey to global state
  const { refresh, loading } = useRequest(v1Service.getSecret, {
    defaultParams: [{}],
    onSuccess: res => {
      if (res.successful) {
        const publicKey = res?.data?.public_key || '';
        setEncryptLocalStorage('publicKey', publicKey);
        dispatch({
          type: 'profile/update',
          payload: {
            publicKey,
          },
        });
      }
    },
  });

  // 用来触发非免密场景下空密码 401 跳到登录页
  useRequest(obService.getObInfo, {
    ready: !isPasswordFreeLogin,
  });

  const telemetryTime = getEncryptLocalStorage('telemetryTime');
  const isTelemetryOutdated = !telemetryTime || moment().diff(moment(telemetryTime), 'hours') >= 1;
  const isLogin = getEncryptLocalStorage('login') === 'true';

  useRequest(
    () =>
      getStatistics({
        HIDE_ERROR_MESSAGE: true,
      }),
    {
      ready:
        isTelemetryOutdated &&
        // 如果是免密登录场景，获取到 secret 后在进行 statistics 请求，避免 401 错误
        (isPasswordFreeLogin ? !loading : isLogin),
      // 登录状态变化后，重新发起请求，避免首次失败后不再发起请求
      refreshDeps: [isLogin],
      // 一小时 + 5秒 轮训一次。5s 是为了避免请求 telemetry 接口时 ，时间差(telemetryTime 判断)导致的请求失败
      pollingInterval: 1000 * 60 * 60 + 5000,
      onSuccess: res => {
        if (res.successful && res.data?.telemetryEnabled) {
          telemetryReport({
            content: res.data,
            component: 'obshell',
          }).then(res => {
            if (res?.code === 200) {
              setEncryptLocalStorage('telemetryTime', moment().format());
            }
          });
        }
      },
    }
  );

  useEffect(() => {
    if (location?.pathname === '/login') {
      refresh();
    }
  }, [location?.pathname]);

  useEffect(() => {
    // Example: 发送消息到子应用，要注意子应用是否加载完成
    // const iframe = document.getElementById('child-iframe');
    // iframe.contentWindow.postMessage(
    //   { locale: 'en-US'},
    //   '*' // 子应用的源（origin）
    // );
    console.log('obshell starting...');

    // 子应用作为 iframe 嵌入其他应用，通过 message 需要获取 locale 进行同步切换语言
    window.addEventListener('message', event => {
      // console.log('obshell receive message', event);
      if (event.data && event.data.locale) {
        setLocale(event.data.locale, true);

        event?.source?.postMessage({ changedLocale: true }, '*');
      }
    });
  }, []);

  // 登录页不需要 publicKey loading
  if (loading && location?.pathname !== '/login') {
    // publicKey 获取之后再加载内容，空密码场景下只需要 publicKey 即可登录（实现免密登录），无须获取密码
    return <Spin style={{ marginTop: '20vh' }}></Spin>;
  }

  return (
    <ConfigProvider
      navigate={history.push}
      locale={localeMap[locale] || zh_CN}
      form={{
        requiredMark: true,
      }}
      theme={{
        isDark: themeMode === 'dark',
        algorithm: themeMode === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
      }}
    >
      <ThemeProvider appearance={themeMode}>
        <ChartProvider theme={themeMode}>
          <GlobalStyle themeMode={themeMode} />
          <ErrorBoundary>
            <BlankLayout>{children}</BlankLayout>
          </ErrorBoundary>
        </ChartProvider>
      </ThemeProvider>
    </ConfigProvider>
  );
};

export default Layout;
