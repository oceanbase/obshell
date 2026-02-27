import { getFullPath } from '@/util/global';
import type { TabPaneProps } from '@oceanbase/design/es/tabs';
import type { MenuItem } from '@oceanbase/ui/es/BasicLayout';
import { history, useLocation } from '@umijs/max';
import { flatten, isArray } from 'lodash';
import { match } from 'path-to-regexp';
import { useEffect } from 'react';
import useSeekdbInfo from './usSeekdbInfo';

// 根据根路由或 Tab 的权限进行重定向
// 若当前路由为根路由，则重定向到第一个有权限的子路由
// 若当前路由为非根路由，并且没有访问权限，则重定向到 403 页面
// 使用本钩子时，routes 中不可配置 404 页面兜底
const useRedirectRoute = (
  // 匹配根路由 rootPath
  // 内部会自动包一层 getFullPath()，外部传入无需处理
  rootPath: string,
  // 菜单列表
  menus: (MenuItem | (TabPaneProps & { accessible: boolean; link: string }))[]
) => {
  const { pathname } = useLocation() || {};
  const { isSeekdbInfoDataLoaded } = useSeekdbInfo();

  const matchFn = match(getFullPath(rootPath));
  const isRootPath = matchFn(pathname);

  // 将有权限的菜单展开并筛选出来
  const fullMenus = flatten(
    menus.map(item => [item, ...(isArray(item?.children) ? item?.children : [])])
  );
  const accessibleMenus = fullMenus.filter(item => item.accessible);
  // 若当前路由为根路由，则重定向到第一个有权限的子路由
  useEffect(() => {
    if (!isRootPath || !isSeekdbInfoDataLoaded) return;

    const firstAccessiblePath = accessibleMenus && accessibleMenus[0]?.link;

    // seekdbInfoData 已加载，进行正常的权限判断
    if (!firstAccessiblePath) {
      history.push('/error/403');
      return;
    }

    if (firstAccessiblePath !== pathname) {
      history.replace(firstAccessiblePath);
    }
  }, [pathname, isRootPath, isSeekdbInfoDataLoaded, accessibleMenus]);

  // 若当前路由为非根路由，并且没有访问权限，则重定向到 403 页面
  useEffect(() => {
    if (isRootPath || !fullMenus?.length || !isSeekdbInfoDataLoaded) return;

    const relatedMenus = fullMenus.filter(menu => pathname.startsWith(menu.link));
    const sortMenus = relatedMenus.sort((a, b) => b?.link.length - a?.link.length);
    if (relatedMenus?.length && !sortMenus[0]?.accessible) history.push('/error/403');
  }, [pathname, isRootPath, isSeekdbInfoDataLoaded]);
};
export default useRedirectRoute;
