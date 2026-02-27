# Cursor Rules 说明文档

本项目包含了 8 个专门为 OceanBase 前端项目定制的 Cursor Rules，帮助开发团队保持代码一致性和最佳实践。

## 规则文件列表

### 1. project-structure.mdc ⭐ (总是应用)

- **作用**: 项目整体结构指南
- **应用范围**: 所有文件 (alwaysApply: true)
- **内容**:
  - 核心目录结构说明
  - 文件命名约定
  - 技术栈概览

### 2. typescript-react.mdc

- **作用**: TypeScript 和 React 开发规范
- **应用范围**: `*.ts, *.tsx` 文件
- **内容**:
  - TypeScript 类型定义规范
  - React 组件开发规范
  - 代码质量要求

### 3. component-development.mdc

- **作用**: 组件开发规范和最佳实践
- **应用范围**: `src/component/**/*`, `src/page/**/*`
- **内容**:
  - 组件文件结构
  - 通用组件 vs 页面组件开发
  - 组件设计原则

### 4. api-service.mdc

- **作用**: API 服务和状态管理开发规范
- **应用范围**: `src/service/**/*`, `src/model/**/*`
- **内容**:
  - API 服务结构 (OBShell/OCP Express/自定义)
  - DVA Model 规范
  - 请求处理最佳实践

### 5. styling.mdc

- **作用**: 样式开发规范和主题系统
- **应用范围**: `*.less`, 组件和页面文件
- **内容**:
  - Less 样式开发规范
  - Less 变量使用规范
  - 响应式设计规范

### 6. internationalization.mdc

- **作用**: 国际化(i18n)开发规范
- **应用范围**: `src/locale/**/*`, 页面和组件文件
- **内容**:
  - 国际化文件结构
  - 文本键值命名规范
  - 动态文本处理

### 7. testing.mdc

- **作用**: 测试开发规范和最佳实践
- **应用范围**: `test/**/*`, `*.test.ts`, `*.test.tsx` 等测试文件
- **内容**:
  - Jest 测试框架使用
  - 组件测试和工具函数测试
  - Mock 和测试覆盖率要求

### 8. debugging.mdc

- **作用**: 调试日志开发规范
- **应用范围**: `*.ts, *.tsx` 文件
- **内容**:
  - console.log 标准格式规范
  - 调试最佳实践
  - 生产环境清理要求

## 如何使用规则

### 自动应用

- `project-structure.mdc` 会自动应用到所有对话

### 按文件类型应用

- 当你编辑特定类型的文件时，相应的规则会自动生效
- 例如：编辑 `.tsx` 文件时，`typescript-react.mdc` 和 `component-development.mdc` 会生效

### 手动引用规则

你可以在对话中主动提及特定规则：

```
@styling 请帮我优化这个组件的样式
@api-service 如何正确定义这个API接口？
@testing 为这个组件写单元测试
@debugging 规范这些调试日志
```

## 规则优势

✅ **一致性**: 确保团队代码风格统一
✅ **最佳实践**: 集成了 React + TypeScript + Ant Design 最佳实践
✅ **效率提升**: 减少代码 review 中的样式问题
✅ **知识传承**: 新团队成员快速了解项目规范
✅ **错误预防**: 提前避免常见的开发陷阱

## 规则维护

- 规则文件会随着项目演进而更新
- 如发现规则与实际需求不符，请及时更新对应的 `.mdc` 文件
- 建议定期 review 规则文件，确保其与项目当前状态同步

## 技术架构适配

这些规则专为以下技术栈设计：

- **前端框架**: React + TypeScript
- **状态管理**: DVA
- **UI 组件库**: Ant Design + Ant Design Pro
- **样式方案**: Less
- **测试框架**: Jest + Testing Library
- **国际化**: react-intl
- **构建工具**: Umi 4

---

_如有任何疑问或建议，请联系团队负责人进行规则调整。_
