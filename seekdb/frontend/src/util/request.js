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
import { getDvaApp, getLocale, history } from '@umijs/max';
import CryptoJS from 'crypto-js';
import { Base64 } from 'js-base64';
/**
 * request 网络请求工具
 * 提供诸如参数序列号, 缓存, 超时, 字符编码处理, 错误处理等常用功能,
 */
import Cookies from 'js-cookie';
import { extend } from 'umi-request';
// import tracert from '@/util/tracert';
import { getEncryptLocalStorage, setEncryptLocalStorage } from '@/util';
import { aesEncrypt } from '@/util/aes';
import encrypt from '@/util/encrypt';
import { cloneDeep } from 'lodash';
import queryString from 'query-string';

const statusCodeMessage = {
  400: formatMessage({
    id: 'ocp-express.src.util.request.TheErrorOccurredInThe',
    defaultMessage: '发出的请求有错误，服务器没有进行新建或修改数据的操作。',
  }),

  401: formatMessage({
    id: 'ocp-express.src.util.request.TheUserIsNotLogged',
    defaultMessage: '用户未登录，或者登录使用的用户名和密码错误。',
  }),

  403: formatMessage({
    id: 'ocp-express.src.util.request.YouDoNotHaveThe',
    defaultMessage: '没有权限进行对应操作，请联系管理员。',
  }),

  404: formatMessage({
    id: 'ocp-express.src.util.request.TheRequestIsForA',
    defaultMessage: '发出的请求针对的是不存在的记录，服务器没有进行操作。',
  }),

  405: formatMessage({
    id: 'ocp-express.src.util.request.TheRequestMethodCannotBe',
    defaultMessage: '请求方法不能被用于请求相应的资源，或者请求路径不正确。',
  }),
  406: formatMessage({
    id: 'ocp-express.src.util.request.TheRequestFormatIsNot',
    defaultMessage: '请求的格式不可得。',
  }),

  410: formatMessage({
    id: 'ocp-express.src.util.request.TheRequestedResourceIsPermanently',
    defaultMessage: '请求的资源被永久删除，且不会再得到的。',
  }),

  422: formatMessage({
    id: 'ocp-express.src.util.request.AValidationErrorOccursWhen',
    defaultMessage: '当创建一个对象时，发生一个验证错误。',
  }),

  500: formatMessage({
    id: 'ocp-express.src.util.request.AnErrorOccurredOnThe',
    defaultMessage: '服务器发生错误，请检查服务器。',
  }),

  502: formatMessage({
    id: 'ocp-express.src.util.request.GatewayError',
    defaultMessage: '网关错误。',
  }),
  503: formatMessage({
    id: 'ocp-express.src.util.request.TheServiceIsUnavailableAnd',
    defaultMessage: '服务不可用，服务器暂时过载或维护。',
  }),

  504: formatMessage({
    id: 'ocp-express.src.util.request.TheGatewayTimedOut',
    defaultMessage: '网关超时。',
  }),
};
const messageMap = {
  'Unexpected error: failed to start maintenance: cluster is under maintenance': formatMessage({
    id: 'ocp-v2.src.util.request.ClusterInOMState',
    defaultMessage: '集群处于运维状态中',
  }),

  'Unexpected error: failed to start maintenance: agent is under maintenance': formatMessage({
    id: 'ocp-v2.src.util.request.TheCurrentNodeIsIn',
    defaultMessage: '当前节点处于运维状态中',
  }),

  'Known error: Tenant has been locked.': formatMessage({
    id: 'ocp-v2.src.util.request.TenantIsLocked',
    defaultMessage: '租户已被锁定',
  }),
  'Verification failed: access denied': formatMessage({
    id: 'ocp-v2.src.util.request.AuthenticationFailedAccessDenied',
    defaultMessage: '验证失败：访问被拒绝',
  }),
};

const getMsg = (errorMessage = '') => {
  if (messageMap[errorMessage]) {
    return messageMap[errorMessage];
  }

  if (errorMessage?.includes('observer process not exist')) {
    return formatMessage({
      id: 'ocp-v2.src.util.request.ObserverProcessDoesNotExist',
      defaultMessage: 'observer 进程不存在',
    });
  }

  if (errorMessage?.includes('oceanbase db is nil')) {
    return formatMessage({
      id: 'ocp-v2.src.util.request.OceanbaseDbIsNil',
      defaultMessage: '未获得集群连接',
    });
  }

  const reg = /failed to start maintenance: (.*?) is under maintenance/;
  const [, name] = errorMessage?.match(reg) || [];

  if (name) {
    return name
      ? formatMessage(
          {
            id: 'ocp-v2.src.util.request.NameIsInOperationAnd',
            defaultMessage: '{name} 处于运维状态中',
          },
          { name: name }
        )
      : undefined;
  }
};

/**
 * 异常处理程序
 * response 为浏览器的 Response 对象，而 data 才是后端实际返回的响应数据
 */
const errorHandler = ({ request, response, data }) => {
  // 因为 errorHandler 不处在组件中，所以只能通过获取 dva 实例来修改 model 中的值
  const dvaApp = getDvaApp();
  const dispatch = dvaApp?._store?.dispatch;

  console.error(response, 'error response');
  const {
    options: {
      HIDE_ERROR_MESSAGE,
      // 请求失败时，需要隐藏错误信息的错误码列表，根据返回的错误码是否在列表中
      // 来决定是否隐藏全局的错误提示，常用于定制化处理特定请求错误的场景
      HIDE_ERROR_MESSAGE_CODE_LIST,
      // 出现 403、404 等错误码时，是否重定向到错误页面
      SHOULD_ERROR_PAGE = true,
    } = {},
  } = request || {};

  // 当遇到 500 错误时，response 为 null，因此需要做容错处理，避免前端页面崩溃
  const { status } = response || {};
  // 401 状态为未登录情况，不展示接口错误信息，直接跳转登录页，因此需要单独处理
  if (status === 401) {
    setEncryptLocalStorage('login', 'false');

    // 未登录状态，清空 tracert 的用户标识
    // tracert.set({
    //   roleId: null,
    // });
    if (window.location.pathname !== '/login') {
      // 将当前的 pathname 和 search 记录在 state 中，以便登录后能跳转到之前访问的页面
      // 如果当前已经是登录页了，则没必要跳转
      history.push({
        pathname: '/login',
        query: {
          callback: `${window.location.pathname}${window.location.search}`,
        },
      });
    }
  } else {
    const { error = {} } = data || {};
    const { code, message: errorMessage } = error;
    // 错误展示一定要在 throw err 之前执行，否则抛错之后就无法展示了
    // 优先展示后端返回的错误信息，如果没有，则根据 status 进行展示
    const msg = errorMessage ? getMsg(errorMessage) || errorMessage : statusCodeMessage[status];

    // 是否隐藏错误信息
    const hideErrorMessage =
      HIDE_ERROR_MESSAGE ||
      (HIDE_ERROR_MESSAGE_CODE_LIST && HIDE_ERROR_MESSAGE_CODE_LIST.includes(code));
    // 是否展示错误信息
    const showErrorMessage = !hideErrorMessage;
    // 有对应的错误信息才进行展示，避免遇到 204 等状态码(退出登录) 时，报一个空错误
    if (msg && code !== 60017 && code !== 60018 && showErrorMessage) {
      message.error(msg, 3);
    }
    // 403 状态为无权限情况，跳转到 403 页面
    if (status === 403 && SHOULD_ERROR_PAGE) {
      history.push('/error/403');
    } else if (status === 404 && SHOULD_ERROR_PAGE) {
      if (code === 60018) {
        // 60018 租户密码不存在，提示录入租户管理员密码密码。
        // return dispatch({
        //   type: 'global/update',
        //   payload: {
        //     showTenantAdminPasswordModal: true,
        //     tenantAdminPasswordErrorData: {
        //       type: 'ADD',
        //       errorMessage: msg,
        //       ...(error?.target || {}),
        //     },
        //   },
        // });
      } else {
        history.push('/error/404');
      }
    } else if (status === 500) {
      // if (code === 60017) {
      //   return dispatch({
      //     type: 'global/update',
      //     payload: {
      //       showTenantAdminPasswordModal: true,
      //       tenantAdminPasswordErrorData: {
      //         type: 'EDIT',
      //         errorMessage: msg,
      //         ...(error?.target || {}),
      //       },
      //     },
      //   });
      // }
    }
  } // 一定要返回 data，否则就会在错误处理这一层断掉，后续无法获取响应的数据
  return data;
};
/**
 * 配置request请求时的默认参数
 */
const request = extend({
  errorHandler, // 默认错误处理
  credentials: 'include', // 默认请求是否带上cookie
});
function base64Encode(key, iv) {
  if (!key || !iv) return '';
  try {
    // 将 key 和 iv 合并，并转为 WordArray
    const combined = cloneDeep(key).concat(iv);
    // 将 WordArray 转为 Base64
    const base64String = CryptoJS.enc.Base64.stringify(combined);

    return base64String; // 返回 Base64 编码的字符串
  } catch (error) {
    console.error(error, 'error');
  }
}

// 测试环境不进行加密
const isTestEnv = process.env.ENV === 'tester';

request.interceptors.request.use((url, options) => {
  // 由于 one-api 接口生成能力不足，需要对日志下载的接口手动设置 responseType 和 Accept 字段
  if (
    // 任务的日志下载
    url.includes('downloadDiagnosis') ||
    // 日志查询的日志下载
    url.includes('logs/download')
  ) {
    options.responseType = 'arrayBuffer';
    options.headers.Accept = '*/*';
  }

  const dvaApp = getDvaApp();
  const { profile } = dvaApp?._store?.getState?.() || {};
  const { password: profilePassword, publicKey: profilePublicKey } = profile || {};

  // 兼容页面刷新
  const password = profilePassword || getEncryptLocalStorage('password');
  const publicKey = profilePublicKey || getEncryptLocalStorage('publicKey');
  const isFormData = options.data instanceof FormData;

  /* OBShell 接口混合加密: https://www.oceanbase.com/docs/common-oceanbase-database-cn-1000000002016169 */
  const { encryptedContent, secretKey, iv } = aesEncrypt(isFormData ? undefined : options.data);

  const keys = base64Encode(secretKey, iv);

  const X_OCS_File_SHA256 = encrypt(options.sha256sum, publicKey);

  const ocsHeader = isTestEnv
    ? JSON.stringify({
        // 集群 root@sys 密码: 从登录成功后保存的全局状态中获取
        Auth: password,
        // Auth 的有效时间戳: 最近一次请求的 30 分钟之内有效
        Ts: `${Math.round(new Date().getTime() / 1000, 1000) + 30 * 60}`,
        // 请求的 URI
        Uri: `${url}${
          Object.keys(options.params).length > 0
            ? `?${queryString.stringify(options.params, { sort: false })}`
            : ''
        }`,
      })
    : Base64.decode(
        Base64.encode(
          encrypt(
            JSON.stringify({
              // 集群 root@sys 密码: 从登录成功后保存的全局状态中获取
              Auth: password,
              // Auth 的有效时间戳: 最近一次请求的 30 分钟之内有效
              Ts: `${Math.round(new Date().getTime() / 1000, 1000) + 30 * 60}`,
              // 请求的 URI
              Uri: `${url}${
                Object.keys(options.params).length > 0
                  ? `?${queryString.stringify(options.params, { sort: false })}`
                  : ''
              }`,

              // 加密 HTTP 请求的 Body (Body 使用 data 参数)时，使用的 AES 加密算法的 key 和 IV
              ...(options.data ? { Keys: keys } : {}),
            }),
            publicKey
          )
        )
      );

  // url list to skip auth validation
  const skipUrlList = ['/api/v1/secret'];

  // 对 GET 请求，将 data 转换为 params
  if (options.method === 'GET' && options.data) {
    const queryString = new URLSearchParams(options.data).toString();
    const newUrl = queryString ? `${url}?${queryString}` : url;

    return {
      url: newUrl,
      options: {
        ...options,
        data: undefined, // 清除 data，避免冲突
      },
    };
  }

  return {
    url,
    options: {
      ...options,
      data: isTestEnv ? options.data : isFormData ? options.data : encryptedContent,
      headers: {
        ...(options.headers || {}),
        // 因为 token 在每次请求中可能都会变化，因此不能在新建 request 实例时直接配置 headers (因为如果页面不刷新，headers 就不变)
        // 而是需要每次请求都单独设置，以保证总是传递最新的 CSRF Token
        'X-XSRF-TOKEN': Cookies.get('XSRF-TOKEN'),
        // umi 3.x 的 locale 插件初始化可能会在 request 插件之后，导致 getLocale 可能为空
        // 为了避免前端页面崩溃，这里对 getLocale 判断处理
        'Accept-Language': getLocale ? getLocale() : 'zh-CN',
        ...(skipUrlList.includes(url)
          ? {}
          : {
              'X-OCS-Header': ocsHeader,
              ...(isTestEnv ? {} : isFormData ? { 'X-OCS-File-SHA256': X_OCS_File_SHA256 } : {}),
            }),
      },
    },
  };
});

export default request;
