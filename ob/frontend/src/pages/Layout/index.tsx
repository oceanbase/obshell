import { PUBLIC_KEY } from '@/constant/login';
import { useQueryParam } from '@/hook/useQueryParams';
import { telemetryReport } from '@/service/custom';
import * as obService from '@/service/obshell/ob';
import { getStatistics } from '@/service/obshell/ob';
import * as v1Service from '@/service/obshell/v1';
import { getEncryptLocalStorage, setEncryptLocalStorage } from '@/util';
import { handleLoginSuccess } from '@/util/login';
import { ChartProvider } from '@oceanbase/charts';
import { ConfigProvider, Spin, theme } from '@oceanbase/design';
import en_US from '@oceanbase/ui/es/locale/en-US';
import zh_CN from '@oceanbase/ui/es/locale/zh-CN';
import { getLocale, history, Outlet, useDispatch, useLocation, useSelector } from '@umijs/max';
import { useRequest } from 'ahooks';
import moment from 'moment';
import React, { useEffect } from 'react';
import BlankLayout from './BlankLayout';

const Layout: React.FC = () => {
  const location = useLocation();
  const mode = useQueryParam('mode');
  const uiMode = useQueryParam('uiMode');
  const isPasswordFreeLogin = mode === 'passwordFreeLogin';

  const { themeMode } = useSelector((state: DefaultRootState) => state.global);

  const locale = getLocale();
  const localeMap = {
    'en-US': en_US,
    'zh-CN': zh_CN,
  };

  const dispatch = useDispatch();

  useEffect(() => {
    // 设置标签页的 title
    document.title = 'obshell Dashboard';
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
        setEncryptLocalStorage(PUBLIC_KEY, publicKey);
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

  // 免密登录场景下，获取到 secret 后在手动调用 v1Service.login 接口进行登录
  const { loading: loginLoading } = useRequest(v1Service.login, {
    ready: isPasswordFreeLogin && !loading,
    defaultParams: [{}],
    onSuccess: res => {
      if (res.successful) {
        handleLoginSuccess(res.data);
      }
    },
  });

  useRequest(
    () =>
      getStatistics({
        HIDE_ERROR_MESSAGE: true,
      }),
    {
      ready:
        isTelemetryOutdated &&
        // 如果是免密登录场景，获取到 secret 后在进行 statistics 请求，避免 401 错误
        (isPasswordFreeLogin ? !loading && !loginLoading : isLogin),
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
      <ChartProvider theme={themeMode}>
        <BlankLayout>
          <Outlet />
        </BlankLayout>
      </ChartProvider>
    </ConfigProvider>
  );
};

export default Layout;
