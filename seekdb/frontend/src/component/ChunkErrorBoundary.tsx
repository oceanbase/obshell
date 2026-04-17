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

import React from 'react';

interface State {
  hasError: boolean;
}

/**
 * 专门用于捕获 webpack 懒加载 chunk 版本不匹配错误。
 *
 * 常见场景：升级部署 obshell 后，浏览器缓存了旧版本的 umi.js（主 bundle），
 * 但服务器上的异步 chunk 已更新，导致模块 ID 不一致，React 渲染时报
 * "TypeError: Cannot read properties of undefined (reading 'call')"。
 *
 * 检测到此类错误后，使用 sessionStorage 防止刷新死循环，仅自动刷新一次。
 */
class ChunkErrorBoundary extends React.Component<React.PropsWithChildren, State> {
  state: State = { hasError: false };

  static isChunkMismatchError(error: unknown): boolean {
    if (!(error instanceof Error)) return false;
    return (
      error.name === 'ChunkLoadError' ||
      /loading chunk/i.test(error.message) ||
      /loading css chunk/i.test(error.message) ||
      /Cannot read properties of undefined.*call/i.test(error.message)
    );
  }

  componentDidCatch(error: Error) {
    if (
      ChunkErrorBoundary.isChunkMismatchError(error) &&
      !sessionStorage.getItem('__chunk_reload__')
    ) {
      sessionStorage.setItem('__chunk_reload__', '1');
      window.location.reload();
    } else {
      this.setState({ hasError: true });
    }
  }

  render() {
    if (this.state.hasError) {
      return null;
    }
    return this.props.children;
  }
}

export default ChunkErrorBoundary;
