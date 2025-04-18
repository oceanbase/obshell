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
import React, { useState, useRef } from 'react';
import { FullscreenBox } from '@oceanbase/ui';
import type { FullscreenBoxProps } from '@oceanbase/ui/es/FullscreenBox';
import type { Graph } from '@antv/g6';
import GraphToolbar from '@/component/GraphToolbar';
import useStyles from './index.style';

interface FullScreenProps extends FullscreenBoxProps {
  // 左上角标题
  title?: React.ReactNode;
  // 左上角描述
  description?: React.ReactNode;
  graph?: Graph;
  onReload?: () => void;
  // 需要自定义 header 的类型定义，因为设置了默认的 extra，所以传入的 header 必须是对象
  header?: {
    title?: React.ReactNode;
    extra?: React.ReactNode;
  };
  className?: string;
  style?: React.CSSProperties;
}

const FullScreen: React.FC<FullScreenProps> = ({
  title,
  description,
  graph,
  onReload,
  header,
  onChange,
  children,
  className,
  style,
  ...restProps
}) => {
  const { styles } = useStyles();
  const [fullscreen, setFullscreen] = useState(false);
  const boxRef = useRef();

  function toggleFullscreen() {
    if (boxRef.current) {
      boxRef.current.changeFullscreen(!fullscreen);
    }
  }

  function handleFullscreenChange(fs) {
    setFullscreen(fs);
    if (onChange) {
      onChange(fs);
    }
  }

  return (
    <FullscreenBox
      ref={boxRef}
      defaultMode="viewport"
      // 全屏状态下才展示 header
      header={
        fullscreen && {
          extra: graph && (
            <GraphToolbar mode="embed" graph={graph} showFullscreen={false} onReload={onReload} />
          ),

          ...header,
        }
      }
      onChange={handleFullscreenChange}
      className={`${styles.container} ${className}`}
      style={{
        ...style,
        // 集群和租户拓扑图:
        // 全屏状态下，paddingTop 设为 0，正常展示
        // 非全屏状态下，paddingTop 设为 50，避免和全局 Header 出现重叠
        paddingTop: fullscreen ? 0 : 50,
      }}
      {...restProps}
    >
      {/* 左上角全屏入口: 未设置 title 和 description 才展示 */}
      {/* {!fullscreen && !(title || description) && (
        <div className={styles.bubble}>
          <Tooltip
            title={formatMessage({
              id: 'ocp-express.component.FullScreen.EnterFullScreen',
              defaultMessage: '进入全屏',
            })}
          >
            <FullscreenOutlined onClick={toggleFullscreen} className={styles.icon} />
          </Tooltip>
        </div>
      )} */}

      {/* 左上角标题和描述: 设置了 title 或 description 才展示 */}
      {!fullscreen && (title || description) && (
        <div className={styles.titleBar}>
          <span className={styles.title}>
            {formatMessage(
              {
                id: 'ocp-express.component.FullScreen.ProtectedModeTitle',
                defaultMessage: '保护模式：{title}',
              },
              { title },
            )}
          </span>
          <span className={styles.description}>{description}</span>
        </div>
      )}

      {!fullscreen && graph && (
        <GraphToolbar
          mode="fixed"
          graph={graph}
          showFullscreen={true}
          onFullscreen={() => {
            toggleFullscreen();
          }}
          onReload={onReload}
          className={styles.toolbar}
        />
      )}

      {children}
    </FullscreenBox>
  );
};

export default FullScreen;
