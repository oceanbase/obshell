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
import React from 'react';
import { Password } from '@oceanbase/ui';
import type { PasswordProps } from '@oceanbase/ui/es/Password';

const OCPPassword: React.FC<PasswordProps> = props => {
  // 特殊字符支持 ~!@#%^&*_\-+=`|(){}[]:;',.?/
  const ocpPasswordRules = [
    {
      validate: (val?: string) => val?.length >= 8 && val?.length <= 32,
      message: formatMessage({
        id: 'ocp-express.src.component.Password.TheDescriptionMustBeTo',
        defaultMessage: '长度为 8~32 个字符',
      }),
    },

    {
      validate: (val?: string) => /^[0-9a-zA-Z~!@#%^&*_\-+=|(){}\[\]:;,.?/]+$/.test(val),
      message: formatMessage({
        id: 'ocp-express.src.component.Password.ItCanOnlyContainLetters',
        defaultMessage: '只能包含字母、数字和特殊字符（~!@#%^&*_-+=|(){}[]:;,.?/）',
      }),
    },

    {
      validate: (val?: string) =>
        /^(?=(.*[a-z]){2,})(?=(.*[A-Z]){2,})(?=(.*\d){2,})(?=(.*[~!@#%^&*_\-+=|(){}\[\]:;,.?/]){2,})[A-Za-z\d~!@#%^&*_\-+=|(){}\[\]:;,.?/]{2,}$/.test(
          val
        ),

      message: formatMessage({
        id: 'ocp-express.src.component.Password.ItMustContainAtLeast',
        defaultMessage: '大小写字母、数字和特殊字符都至少包含 2 个',
      }),
    },
  ];

  return <Password rules={ocpPasswordRules} {...props} />;
};

export default OCPPassword;
