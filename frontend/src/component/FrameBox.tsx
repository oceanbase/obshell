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

interface Props {
  data: string;
  title: string;
}

/**
 * 使用 iframe 用于展示后端返回的文件内容，例如 ASH、AWR 报告文件
 */
const FrameBox: React.FC<Props> = ({ title, data }) => {
  return (
    <iframe
      title={title}
      src={window.URL.createObjectURL(new Blob([data], { type: 'text/html' }))}
      // 高度+宽度
      style={{ width: '100%', border: '0px', height: '100vh' }}
      sandbox="allow-same-origin allow-scripts allow-popups allow-forms"
      scrolling="auto"
    />
  );
};

export default FrameBox;
