import { useSearchParams } from '@umijs/max';
import { useMemo } from 'react';

/**
 * 获取单个查询参数
 * @param key 参数名
 * @returns 参数值（可能为 undefined）
 *
 * @example
 * const backUrl = useQueryParam('backUrl');
 * const status = useQueryParam('status');
 */
export function useQueryParam(key: string): string | undefined {
  const [searchParams] = useSearchParams();
  return useMemo(() => {
    return searchParams.get(key) || undefined;
  }, [searchParams, key]);
}

/**
 * 获取多个查询参数
 * @param keys 参数名数组或参数配置对象
 * @returns 参数对象
 *
 * @example
 * // 简单用法
 * const { status, name } = useQueryParams(['status', 'name']);
 *
 * // 带默认值
 * const { tab, taskTab } = useQueryParams({
 *   tab: 'detail',
 *   taskTab: ''
 * });
 */
export function useQueryParams<T extends string>(keys: T[]): Record<T, string | undefined>;
export function useQueryParams<T extends Record<string, string>>(
  config: T
): Record<keyof T, string>;
export function useQueryParams(
  keysOrConfig: string[] | Record<string, string>
): Record<string, string | undefined> | Record<string, string> {
  const [searchParams] = useSearchParams();

  return useMemo(() => {
    if (Array.isArray(keysOrConfig)) {
      // 数组形式：返回参数值，可能为 undefined
      return keysOrConfig.reduce((acc, key) => {
        acc[key] = searchParams.get(key) || undefined;
        return acc;
      }, {} as Record<string, string | undefined>);
    } else {
      // 对象形式：返回参数值，使用默认值
      return Object.keys(keysOrConfig).reduce((acc, key) => {
        const value = searchParams.get(key);
        acc[key] = value !== null ? value : keysOrConfig[key];
        return acc;
      }, {} as Record<string, string>);
    }
  }, [searchParams, keysOrConfig]);
}

/**
 * 获取所有查询参数作为对象
 * @returns 参数对象
 *
 * @example
 * const allParams = useAllQueryParams();
 * // { status: 'running', name: 'task1' }
 */
export function useAllQueryParams(): Record<string, string> {
  const [searchParams] = useSearchParams();

  return useMemo(() => {
    const params: Record<string, string> = {};
    searchParams.forEach((value, key) => {
      params[key] = value;
    });
    return params;
  }, [searchParams]);
}
