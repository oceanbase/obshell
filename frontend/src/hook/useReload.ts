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

import { useState } from 'react';
import { useUpdateEffect } from 'ahooks';

const useReload = (initialLoading: boolean, callback?: () => void): [boolean, () => void] => {
  const [loading, setLoading] = useState(initialLoading);
  const reload = () => {
    setLoading(true);
    setTimeout(() => {
      setLoading(false);
    }, 100);
  };
  useUpdateEffect(() => {
    // 加载完成后执行回调函数
    if (!loading) {
      if (callback) {
        callback();
      }
    }
  }, [loading]);
  return [loading, reload];
};

export default useReload;
