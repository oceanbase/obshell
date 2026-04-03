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

import { PUBLIC_KEY } from '@/constant/login';
import * as v1Service from '@/service/obshell/v1';
import { handleLoginSuccess } from '@/util/login';
import { setEncryptLocalStorage } from '@/util';
import { Spin } from '@oceanbase/design';
import { history, useSearchParams } from '@umijs/max';
import React, { useEffect, useState } from 'react';

const SSOLoginPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [done, setDone] = useState(false);

  useEffect(() => {
    const token = searchParams.get('token');
    if (!token) {
      history.push('/login');
      return;
    }
    // 1. 先拉取公钥，供 request 拦截器构造 X-OCS-Header（ssotoken + Keys）时使用
    v1Service
      .getSecret({ HIDE_ERROR_MESSAGE: true })
      .then(secretRes => {
        if (!secretRes?.successful || !secretRes?.data?.public_key) {
          history.push('/login');
          return;
        }
        setEncryptLocalStorage(PUBLIC_KEY, secretRes.data.public_key);
        // 2. 发送 /api/v1/sso/exchange：拦截器会用 ssotoken + Keys 构造加密 Header，响应 data 为 AES 加密的 session_id
        return v1Service.ssoExchange(token, { HIDE_ERROR_MESSAGE: true });
      })
      .then(res => {
        if (res?.successful && typeof res?.data === 'string') {
          // 3. 用拦截器已写入的 LOGIN_SECRET_KEY / LOGIN_IV 解密得到 session_id 并写入
          handleLoginSuccess(res.data);
          setEncryptLocalStorage('login', 'true');
          setEncryptLocalStorage('username', 'root');
          history.push('/overview');
        } else {
          history.push('/login');
        }
      })
      .catch(() => {
        history.push('/login');
      })
      .finally(() => {
        setDone(true);
      });
  }, [searchParams]);

  if (!done) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" tip="正在跳转..." />
      </div>
    );
  }
  return null;
};

export default SSOLoginPage;
