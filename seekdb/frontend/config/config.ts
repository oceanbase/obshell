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

import AntdMomentWebpackPlugin from '@ant-design/moment-webpack-plugin';
import { defineConfig } from '@umijs/max';
import routes from './routes';

const ocp = {
  // target: 'http://1.1.1.1',
};
export default defineConfig({
  routes,
  // 需要关闭 title 配置，避免与 useDocumentTitle 的逻辑冲突
  // 冲突的具体表现为: 切换路由时，父组件设置的文档标题会自动重置为配置的 title
  title: false,
  // UMI 4 中 favicon 改为 links 配置
  links: [{ rel: 'icon', href: '/assets/logo/ob_dashboard_favicon.svg' }],
  // 接口代理配置
  proxy: {
    '/api/v1': {
      ...ocp,
    },
    '/api/v2': {
      ...ocp,
    },
    services: {
      ...ocp,
    },
  },

  antd: {},
  dva: {},
  model: {}, // 启用简易数据流,支持 useModel
  locale: {
    default: 'zh-CN',
    // disable antd locale import in umi plugin-locale to avoid build error
    // issue: https://github.com/oceanbase/ocp-express/pull/22
    antd: false,
    title: false,
  },
  esbuildMinifyIIFE: true, // 修复 esbuild helpers 冲突问题
  codeSplitting: {
    jsStrategy: 'granularChunks',
  },
  // 开启运行时 publicPath
  runtimePublicPath: {},
  chainWebpack: (config, { env }) => {
    if (env === 'production') {
      config.optimization.delete('noEmitOnErrors');
      config.plugins.delete('optimize-css');

      // 因为删除原来适配webpack4的css压缩插件，css压缩可以用 mini-css-extract-plugin
      config.optimization.minimize(true);
      //  config.optimization.minimizer(`css-esbuildMinify`).use(CSSMinimizerWebpackPlugin);
    }
    // 添加 AntdMomentWebpackPlugin 插件
    config.plugin('antd-moment').use(AntdMomentWebpackPlugin, [
      {
        // 关闭 dayjs alias，避免 antd 以外的 dayjs 被 alias 成 moment
        disableDayjsAlias: true,
      },
    ]);
    // 静态资源的文件限制调整为 1GB，避免视频等大文件资源阻塞项目启动
    config.performance.maxAssetSize(1000000000);
    return config;
  },
});
