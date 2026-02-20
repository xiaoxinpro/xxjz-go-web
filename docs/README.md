# 小歆记账（Go + Vue3 版）

基于原 ThinkPHP + MySQL 版本迁移而来，后端 Golang + 多数据库（默认 SQLite），前端 Vue3，API 保持与微信小程序兼容。

## 技术栈

- **后端**: Go 1.21+，Gin，database/sql（SQLite / MySQL / PostgreSQL）
- **前端**: Vue 3，Vite，Vue Router，Pinia
- **配置**: config.yaml 为主，环境变量覆盖
- **部署**: Docker 单容器，可选 docker-compose

## 目录结构

```
xxjz-go-web/
├── backend/           # Go 后端
│   ├── cmd/
│   │   ├── server/    # HTTP 服务入口
│   │   └── import-mysql/  # 旧 MySQL 导出导入工具
│   ├── internal/
│   │   ├── config/    # 配置加载
│   │   ├── handler/   # API 处理
│   │   ├── service/   # 业务逻辑
│   │   ├── repository/# 数据访问
│   │   ├── session/   # 会话
│   │   └── importsql/ # MySQL→SQLite 转换
│   ├── pkg/db/        # 数据库连接与迁移
│   └── migrations/    # SQL 迁移脚本
├── frontend/          # Vue3 前端
├── config.yaml        # 主配置文件
├── Dockerfile
├── Makefile
├── scripts/           # 构建脚本
└── docs/              # 文档
```

## 首次运行与初始化

- **后端启动时**：自动执行数据库迁移（建表），无需手动导入。
- **前端首次访问**：若检测到尚未初始化（数据库中无任何用户），会自动跳转到 **初始化页** `/init`，在此可：
  - **创建管理员账号**：填写用户名、密码、邮箱，创建第一个用户（即管理员），然后跳转登录页；
  - **或导入旧数据库**：上传旧版（ThinkPHP/MySQL）导出的 `xxjz.sql` 文件，系统会转换为当前数据库并导入，导入后跳转登录页。
- 初始化完成后，正常使用登录与业务功能。

## 本地运行

### 后端

```bash
# 从仓库根目录
cp config.yaml backend/  # 或设置 CONFIG=config.yaml
cd backend
go run ./cmd/server
# 默认 :8080，SQLite 数据库文件为 ./data/xxjz.db；首次运行自动建表
```

### 前端开发

```bash
cd frontend
npm install
npm run dev
# 开发服务器 :5173，代理 /api 与 /Home 到后端 8080
# 首次打开会进入初始化页，创建管理员或导入旧数据后再登录
```

### 命令行导入旧 MySQL 数据（可选）

若不想在初始化页上传，也可用 CLI 导入后再访问前端：

```bash
cd backend
# 确保 config 中 database.driver 为 sqlite
go run ./cmd/import-mysql ../old-proj-thinkphp/xxjz.sql
# 或: FILE=path/to/xxjz.sql go run ./cmd/import-mysql
```

## Docker

```bash
docker build -t xxjz-go-web:latest .
docker run -p 8080:8080 -v /path/to/data:/app/data xxjz-go-web:latest
```

或使用 Makefile：`make docker-build`

## 配置说明

- `config.yaml`: 数据库 driver（sqlite/mysql/postgres）、DSN、应用与业务参数。
- 环境变量可覆盖：`PORT`、`DB_DRIVER`、`DB_DSN`、`CONFIG`、`SESSION_SECRET` 等。

## 与旧版差异

- 数据库：默认 SQLite，可选 MySQL/PostgreSQL；表结构保持 xxjz_ 前缀。
- API 路径：推荐 `/api/*`，兼容 `/Home/Api/*`。
- 请求/响应格式与旧版一致（含 base64(data)、uid、data 等），便于微信小程序无缝切换。

详细接口见 [API.md](API.md)。
