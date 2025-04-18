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
import { getLocale, history, useSelector, useDispatch } from 'umi';
import React, { useEffect } from 'react';
import { ConfigProvider, theme } from '@oceanbase/design';
import { ChartProvider } from '@oceanbase/charts';
import en_US from '@oceanbase/ui/es/locale/en-US';
import zh_CN from '@oceanbase/ui/es/locale/zh-CN';
import { ThemeProvider } from 'antd-style';
import { useRequest } from 'ahooks';
import * as v1Service from '@/service/obshell/v1';
import BlankLayout from './BlankLayout';
import ErrorBoundary from '@/component/ErrorBoundary';
import GlobalStyle from './GlobalStyle';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const { themeMode } = useSelector((state: DefaultRootState) => state.global);

  const locale = getLocale();
  const localeMap = {
    'en-US': en_US,
    'zh-CN': zh_CN,
  };

  const dispatch = useDispatch();

  useEffect(() => {
    // 设置标签页的 title
    document.title = formatMessage({
      id: 'ocp-express.config.title',
      defaultMessage: 'OceanBase 云平台',
    });
  }, []);

  // request and save publicKey to global state
  useRequest(v1Service.getSecret, {
    defaultParams: [{}],
    onSuccess: res => {
      if (res.successful) {
        const publicKey = res?.data?.public_key || '';
        localStorage.setItem('publicKey', publicKey);

        dispatch({
          type: 'profile/update',
          payload: {
            publicKey,
          },
        });
      }
    },
  });

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
