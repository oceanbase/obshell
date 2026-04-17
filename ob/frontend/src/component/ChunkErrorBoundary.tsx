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
 * Catches webpack lazy-load chunk mismatch errors that occur after an upgrade
 * when the browser has cached a stale umi.js bundle whose module IDs no longer
 * match the newly deployed async chunks, resulting in:
 * "TypeError: Cannot read properties of undefined (reading 'call')"
 *
 * On detection, the page is reloaded once automatically. sessionStorage is used
 * to prevent an infinite reload loop.
 */
class ChunkErrorBoundary extends React.Component<React.PropsWithChildren<object>, State> {
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
