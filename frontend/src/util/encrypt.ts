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

import { JSEncrypt } from 'jsencrypt-ext';

export default function (input: string, publicKey: string) {
  // 如果公钥为空，则不加密，直接返回原始值
  if (!publicKey) {
    return input;
  }
  const encrypt = new JSEncrypt();
  encrypt.setPublicKey(publicKey);
  // 加密
  return encrypt.encrypt(input);
}
