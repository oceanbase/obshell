module.exports = {
  mode: 'jit', // 开启 JIT 即时编译模式，通过扫描模板文件按需生成 CSS
  // Tailwind v2 使用 purge 配置（v3 使用 content）
  purge: [
    './src/**/*.{js,jsx,ts,tsx}',
    './src/**/*.html',
    './src/**/*.less',
  ],
  darkMode: false, // or 'media' or 'class'
  theme: {
    extend: {
      maxWidth: {
        '200': '200px',
      },
      textColor: {
        'colorText': '#132039',           // 主要文本色
        'colorTextSecondary': '#5C6B85',  // 次要文本色
        'colorTextTertiary': '#8592AD',   // 三级文本色
        'colorTextQuaternary': '#c1cbe0', // 四级文本色（禁用状态）
      },
      borderColor: {
        'colorBorder': '#CDD5E4',
        'colorBorderSecondary': '#E2E8F3',
      },
      backgroundColor: {
        'colorBgContainer': '#FFFFFF',
        'colorBgLayout': '#f3f6fc',
      },
    },
  },
  variants: {
    extend: {},
  },
  plugins: [],
  // 重要：与 Ant Design 样式兼容，避免冲突
  corePlugins: {
    preflight: false, // 禁用 Tailwind 的默认样式，避免与 Ant Design 冲突
    textColor: true,  // 确保文本颜色功能启用
    borderStyle: true, // 确保边框样式功能启用
    backgroundColor: true, // 确保背景颜色功能启用
  },
  // 在 Tailwind v2 中启用 important 选项
  important: true,
}
