export function getFullPath(path: string) {
  const prefixPath = getPrefixPath();
  const newPath = `${prefixPath}${path}`;
  return newPath;
}

export function getPrefixPath() {
  const contextPath = getContextPath();
  return !contextPath || contextPath === '/' ? '' : contextPath;
}

export function getContextPath() {
  // 使用空字符串兜底，方便拼接
  return window.contextPath || '';
}
