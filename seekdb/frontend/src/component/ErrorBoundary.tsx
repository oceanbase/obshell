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

import Empty from '@/component/Empty';
import { formatMessage } from '@/util/intl';
import { Alert, Button } from '@oceanbase/design';
import { history } from '@umijs/max';
import React from 'react';

interface ErrorBoundaryProps {
  message?: React.ReactNode;
  description?: React.ReactNode;
}

interface ErrorBoundaryState {
  error?: Error | null;
  info: {
    componentStack?: string;
  };
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state = {
    error: undefined,
    info: {
      componentStack: '',
    },
  };

  componentDidCatch(error: Error | null, info: object) {
    this.setState({ error, info });
  }

  render() {
    const { message, description, children } = this.props;
    const { error, info } = this.state;
    const componentStack = info && info.componentStack ? info.componentStack : null;
    const errorMessage = typeof message === 'undefined' ? (error || '').toString() : message;
    const errorDescription = typeof description === 'undefined' ? componentStack : description;
    if (error) {
      return (
        <Empty
          spm={formatMessage({
            id: 'ocp-v2.src.component.ErrorBoundary.ErrorHandling',
            defaultMessage: '错误处理',
          })}
          mode="page"
          image="/assets/common/crash.svg"
          title={formatMessage({
            id: 'ocp-express.src.component.ErrorBoundary.PageExecutionError',
            defaultMessage: '页面执行出错',
          })}
          data-aspm-expo
          data-aspm-param={``}
        >
          <Button
            spm={formatMessage({
              id: 'ocp-v2.src.component.ErrorBoundary.ErrorHandlingReturnToHome',
              defaultMessage: '错误处理-返回首页',
            })}
            type="primary"
            onClick={() => {
              history.push('/');
              // 刷新页面，以便重新加载应用、避免停留在当前的 ErrorBoundary 页面
              window.location.reload();
            }}
            data-aspm-expo
            data-aspm-param={``}
          >
            {formatMessage({
              id: 'ocp-express.src.component.ErrorBoundary.ReturnToHomePage',
              defaultMessage: '返回首页',
            })}
          </Button>
          {/* 展示报错详情，方便定位问题原因 */}
          <Alert
            spm={formatMessage({
              id: 'ocp-v2.src.component.ErrorBoundary.ErrorHandlingErrorDetails',
              defaultMessage: '错误处理-错误详情',
            })}
            type="error"
            showIcon={true}
            message={errorMessage}
            description={errorDescription}
            style={{
              marginTop: 24,
              overflow: 'auto',
              maxHeight: '50vh',
              // 为了避免被 Empty 的水平居中样式影响，需要设置 textAlign
              textAlign: 'left',
            }}
            data-aspm-expo
            data-aspm-param={``}
          />
        </Empty>
      );
    }
    return children;
  }
}

export default ErrorBoundary;
