---
name: xxjz-go-vue
description: 小歆记账项目 Go 后端与 Vue3 前端的代码风格与约定，用于统一后续输出。
---

# 小歆记账项目代码规范

## 适用范围

本 Skill 适用于本仓库内 **backend/**（Golang）与 **frontend/**（Vue3）的编写与修改。

## Go 后端

- **布局**: 业务逻辑在 `internal/service`，HTTP 处理在 `internal/handler`，数据访问在 `internal/repository`；配置在 `internal/config`，不硬编码。
- **命名**: 包名小写单词；导出函数/类型 PascalCase；接口名以 -er 结尾或表意（如 `SessionStore`）。
- **API 响应**: 与旧版 PHP 兼容，统一返回 JSON；未登录时 `uid: 0`，错误信息放在 `uname` 或 `data` 字符串中；支持 GET/POST 及 base64(data) 参数。
- **数据库**: 使用 `database/sql` 与配置中的 driver（sqlite/mysql/postgres）；表前缀 `xxjz_`；不在 handler 内直接写 SQL，经 repository 访问。
- **错误**: 业务错误返回 200 + JSON 中 ok/msg 或 uid/data；仅严重错误可 4xx/5xx。

## Vue3 前端

- **结构**: 页面在 `views/`，状态在 `stores/`（Pinia），路由在 `router/`；API 调用统一使用 `/api/*`，并设置 `credentials: 'include'`。
- **风格**: 组件使用 `<script setup>` 与 Composition API；样式使用 scoped；避免内联魔法数字，可抽成常量。
- **与后端**: 不单独定义一套接口格式，与现有 API 文档（docs/API.md）一致；登录态依赖 Cookie Session，不在前端存密码。

## 通用

- **配置**: 以项目根目录或 backend 下的 `config.yaml` 为准；环境变量仅作覆盖，不替代完整配置。
- **禁止**: 在代码中硬编码数据库连接串、密钥、管理员 UID；禁止提交含敏感信息的 config 覆盖版本。
