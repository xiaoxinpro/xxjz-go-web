# 小歆记账 Go + Vue3 版

由 [ThinkPHP + MySQL 版本](old-proj-thinkphp/) 迁移而来，后端 **Golang**，数据库默认 **SQLite**（可选 MySQL/PostgreSQL），前端 **Vue3**，API 与微信小程序兼容。

## 快速开始

- **配置**: 以 [config.yaml](config.yaml) 为主，环境变量可覆盖。
- **构建**: `make build` 或 `scripts/build.ps1` / `scripts/build.sh`，产物输出到 **build/**，若已存在会先删除再重建。
- **运行**: 进入 `build/` 执行 `./server`（Linux/macOS）或 `server.exe`（Windows），无需额外依赖，默认端口 8080。
- **本地开发与 Docker**: 见 [docs/README.md](docs/README.md)（后端/前端开发、导入旧数据、Docker）。
- **API 文档**: [docs/API.md](docs/API.md)。

## CI

- 提交前请确保通过 **单元测试**：`cd backend && go test ./...`
- **Docker 镜像** 由 GitHub Actions 在 push 到 main/master 或打 tag 时自动构建并推送到 GHCR。

## 目录

| 目录 | 说明 |
|------|------|
| [backend/](backend/) | Go 后端（Gin、多数据库、Session） |
| [frontend/](frontend/) | Vue3 + Vite 前端 |
| **build/** | 构建输出（生成后可在此目录直接运行，已加入 .gitignore） |
| [docs/](docs/) | 项目说明与 API 文档 |
| [scripts/](scripts/) | 构建脚本（build.ps1 / build.sh） |
| [.github/workflows/](.github/workflows/) | 测试与 Docker 构建流水线 |
| [.cursor/skills/](.cursor/skills/xxjz-go-vue/) | 项目代码规范 Skill |

## 旧项目

原 ThinkPHP 代码与数据库结构见 [old-proj-thinkphp/](old-proj-thinkphp/)。使用 `backend/cmd/import-mysql` 或接口 `POST /api/admin/import` 可将旧 MySQL 导出导入到新项目（目标为 SQLite 时）。
