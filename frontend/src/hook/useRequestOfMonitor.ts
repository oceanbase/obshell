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

import { useRef, useCallback } from 'react';
import type { BaseOptions, CombineService } from '@ahooksjs/use-request/es/types';
import { useRequest } from 'ahooks';

function useLockFn<P extends any[] = any[], V extends any = any>(fn: (...args: P) => Promise<V>) {
  const lockRef = useRef(false);
  return [
    useCallback(
      async (...args: P) => {
        if (lockRef.current) return;
        lockRef.current = true;
        try {
          const ret = await fn(...args);
          lockRef.current = false;
          return ret;
        } catch (e) {
          lockRef.current = false;
          throw e;
        }
      },
      [fn]
    ),
    (value: boolean) => {
      lockRef.current = value;
    },
  ];
}

function useRequestOfMonitor<R, P extends any[]>(
  service: CombineService<R, P>,
  options?: BaseOptions<R, P> & { isRealtime?: boolean }
) {
  const { isRealtime, ...restOptions } = options || {};
  const requestResult = useRequest(service, restOptions);

  // 使用竞态锁来防止监控请求堆积
  const [run, setLock] = useLockFn(requestResult.run);

  return {
    ...requestResult,
    run: (...params: P) => {
      // 实时模式下，接口进入轮询模式，采用竞态锁来防止接口堆积
      if (isRealtime) {
        run(...params);
      } else {
        setLock(false);
        requestResult.run(...params);
      }
    },
  };
}

export default useRequestOfMonitor;
