# Blog Server

一个使用 Go 编写的博客后端服务。它提供已发布文章与友链的读取、基于 JWT 的用户认证、
RSS 订阅生成，以及通过 Server-Sent Events 流式返回的 LLM 文章摘要。

服务基于 [Echo v5](https://github.com/labstack/echo) 提供 HTTP 能力，使用
[Uber Fx](https://github.com/uber-go/fx) 做依赖注入，使用 [Ent](https://entgo.io/) 作为
ORM 访问 PostgreSQL，并使用 Redis 做缓存与计数。后台定时任务负责友链健康检查和浏览量落库。

[English](./README.md)

## 功能特性

- **认证与授权**
  - 邮箱 + 密码注册与登录，访问令牌使用 JWT。
  - 刷新令牌使用签名 JWT，PostgreSQL 仅保存其 SHA-256 哈希和过期时间，并只通过 `HttpOnly` Cookie 返回。
  - 刷新令牌在事务内轮换；旧令牌重用会撤销对应会话，退出仅撤销当前会话。
  - 注册前通过邮件（SMTP）发送验证码。
  - 登录按直连 IP 每分钟最多尝试 10 次；Redis 故障时记录错误并放行。
  - 基于角色的访问控制（RBAC），包含两种角色（`admin`、`reader`）。仅 `admin` 可创建/更新/删除文章与友链；`reader` 仅可读取。
- **文章**
  - 公开的文章列表 / 详情 / 元数据接口。
  - 管理端 CRUD，支持标签与分类（多对多）。
  - 浏览量在读取时写入 Redis，并按小时批量落库到 PostgreSQL。
  - `read_time_minutes`（预计阅读时长）根据正文长度自动计算。
- **友链**
  - 公开的已启用友链列表。
  - 公开的“申请友链”接口，创建一条默认禁用的友链待审核。
  - 友链可用性由每小时运行的定时任务检查（HTTPS `GET`，状态为 `normal`/`abnormal`）。
- **RSS**
  - 分页的 RSS 2.0 feed，包含 `atom:link`（self/next/prev/first/last）导航。
  - 完整归档 feed（所有已发布文章）。
- **LLM 摘要**
  - `POST /api/v1/model/summarize` 创建会话；`GET /api/v1/model/summarize/:sessionId` 以 SSE 流式返回摘要（`data:` 行，`event: done`，`event: error`）。
  - 调用 GitHub Models / Azure AI inference（`models.inference.ai.azure.com`），使用配置文件中的 API Key。
- **基础设施**
  - 结构化 Zap 日志（开发环境控制台，生产环境 JSON）。
  - TOML 配置，支持环境变量覆盖与热更新（文件变更触发优雅重启）。
  - 可配置的优雅关闭超时。
  - 存活检查以及 PostgreSQL/Redis 就绪检查。

## 技术栈

| 关注点         | 库 / 工具                                                  |
| -------------- | --------------------------------------------------------- |
| Web 框架       | Echo v5                                                   |
| 依赖注入       | Uber Fx                                                   |
| ORM            | Ent v0.14                                                 |
| 数据库         | PostgreSQL（`lib/pq`）                                     |
| 缓存 / 计数    | Redis（`redis/go-redis`）                                  |
| 对象存储       | AWS S3 SDK v2（兼容 S3；代码已实现但未接入运行中的服务，见下文） |
| 日志           | Uber Zap                                                  |
| 配置           | Viper（TOML）+ `fsnotify` 热更新                          |
| JWT            | `golang-jwt/jwt/v5`                                       |
| 参数校验       | `go-playground/validator/v10`                             |
| 邮件           | `gopkg.in/gomail.v2`                                      |

## 环境要求

- Go 1.27+（`go.mod` 中声明为 `go 1.27.0`）
- PostgreSQL
- Redis

## 配置

配置通过 Viper 从 TOML 文件加载（默认读取工作目录下的 `config.toml`）。可通过环境变量覆盖，
使用 `APP` 前缀，`.` 替换为 `_`（例如 `APP_SERVER_PORT`、`APP_JWT_SECRET`、
`APP_DATABASE_PASSWORD`）。生产环境（`app.environment = "production"`）要求提供 JWT 密钥和数据库密码。

仓库中**没有提交示例配置文件**（`.gitignore` 排除了 `config.toml`）。请将以下配置复制到项目根目录的 `config.toml`。

```toml
[app]
name = "my-app"
version = "1.0.1"
environment = "development"      # "development" | "production"
debug = true
domain = "http://192.168.31.75:3000"   # 用于拼接 RSS 文章链接
cors_origins = ["*"]             # 当前未被任何中间件使用

[server]
host = "0.0.0.0"
port = 8000
read_timeout = "30s"
write_timeout = "30s"
idle_timeout = "60s"
max_header_bytes = 1048576
graceful_shutdown = "10s"

[database]
host = "localhost"
port = 5432
user = "immortal"
password = "immortal"
name = "blog"
ssl_mode = "disable"
max_open_conns = 25
max_idle_conns = 5
conn_max_lifetime = "1h"
conn_max_idle_time = "30m"
timeout = "5s"

[redis]
host = "localhost"
port = 6379
password = ""
db = 0
pool_size = 10
min_idle_conns = 2
dial_timeout = "5s"
read_timeout = "3s"
write_timeout = "3s"
pool_timeout = "0s"
idle_timeout = "0s"
idle_check_frequency = "0s"

[jwt]
secret = "change-me"
access_expiration = "14m"
refresh_expiration = "168h"
issuer = "my-app"

[log]
level = "info"            # zap 级别；开发环境可用 "debug"
format = ""               # "json" 强制 JSON 输出（否则为开发控制台格式）
file_path = ""
max_size = 100
max_backups = 3
max_age = 30
compress = true

[email]
host = "mail.example.com"
port = 587
username = "noreply@example.com"
password = "your_password"
from = "noreply@example.com"

[llm]
apikey = "your-github-models-or-azure-key"

[rustfs]                  # 兼容 S3 的对象存储（见“状态 / 尚未暴露”）
region = "us-east-1"
access_key_id = "rustfsadmin"
secret_access_key = "rustfsadmin"
endpoint = "http://localhost:9000/"
```

### 热更新

配置文件在磁盘上发生变化时，`config` 模块会触发 Fx 优雅关闭（`fx.Shutdowner`）。
需要由进程管理器或容器重启策略重新启动应用，新的配置才会生效。

## 运行

```bash
# 下载依赖
go mod download

# 启动应用（加载 config.toml、执行 Goose migration、
# 启动 HTTP 服务和定时任务）
go run ./cmd/server

# 构建静态二进制
go build -o blog-server ./cmd/server
./blog-server
```

### Docker

`Dockerfile` 会构建禁用 CGO 的 `blog-server` 二进制，并暴露 `8000` 端口。
挂载或内置一份 `config.toml`；应用会在接受 HTTP 请求前执行内嵌的 Goose migration。

## API

业务路由挂载在 `/api/v1` 下。受保护路由需要在 `Authorization: Bearer <accessToken>` 头中携带访问令牌。
健康检查不使用 API 前缀，分别为 `GET /health/live` 和 `GET /health/ready`。响应使用统一信封：

```json
{ "code": 0, "msg": "success", "data": { } }
```

错误响应使用相同信封，携带非零 `code`，并返回与之对应的 HTTP 状态码
（例如未授权/禁止为 `401`，参数错误为 `400`，未找到为 `404`，冲突为 `409`，内部错误为 `500`）。
面向客户端的 `msg` 为中文。

### 认证

| 方法 | 路径                          | 鉴权 | 说明                                     |
| ---- | ----------------------------- | ---- | ---------------------------------------- |
| POST | `/api/v1/auth/captcha`        | —    | 发送验证码邮件。                         |
| POST | `/api/v1/auth/register`       | —    | 使用邮箱 + 验证码注册。                  |
| POST | `/api/v1/auth/login`          | —    | 登录，返回访问令牌并设置刷新 Cookie。     |
| POST | `/api/v1/auth/logout`         | —    | 清除刷新令牌 Cookie。                    |
| POST | `/api/v1/auth/refresh`        | cookie | 使用 Cookie 中的刷新令牌轮换访问/刷新令牌。 |

请求体：

```jsonc
// POST /api/v1/auth/captcha
{ "email": "user@example.com", "type": "Register" }   // type ∈ {Register, PasswordReset, ChangeEmail}

// POST /api/v1/auth/register
{ "email": "user@example.com", "password": "secret123", "passwordConfirm": "secret123", "captcha": "AB12CD" }

// POST /api/v1/auth/login
{ "email": "user@example.com", "password": "secret123" }
```

`login` / `register` 的响应包含 `accessToken`、`uuid`、`username`、`role`、`avatar`；
刷新 JWT 通过 `HttpOnly` Cookie 下发，不会出现在 JSON 响应体中。PostgreSQL 仅保存其 SHA-256 哈希及与 JWT 一致的过期时间。
轮换后保留旧哈希到对应 JWT 过期，用于检测重用并撤销会话；过期 JWT 在查询会话前即被拒绝。
清理任务在启动时及之后每小时执行，每轮最多删除 500 条过期 token 记录，再删除最多 500 个已无 token 的过期或撤销会话，单轮超时 10 秒。
查询使用过期时间索引及 `SKIP LOCKED`，支持多实例执行。积压记录可能延后到后续批次删除，但不能用于刷新。
migration 00004 会撤销旧不透明 token 会话，升级后用户须重新登录。已经过期的 JWT 再次提交不会触发所属会话撤销。

### 文章

| 方法 | 路径                                | 鉴权             | 说明                                       |
| ---- | ----------------------------------- | ---------------- | ------------------------------------------ |
| GET  | `/api/v1/posts`                     | —                | 分页的已发布文章（`page`、`pageSize`）。    |
| GET  | `/api/v1/posts/meta`                | —                | 文章 id + updated_at（用于 sitemap）。      |
| GET  | `/api/v1/posts/:id`                 | —                | 单篇已发布文章（会递增浏览量）。            |
| POST | `/api/v1/posts`                     | admin            | 创建文章。                                 |
| GET  | `/api/v1/admin/posts`              | admin            | 管理端列表（过滤项：`status`、`keyword`）。 |
| GET  | `/api/v1/admin/posts/:id`          | admin            | 管理端文章详情。                           |
| PUT  | `/api/v1/admin/posts/:id`          | admin 或作者     | 更新文章。                                 |
| DELETE | `/api/v1/admin/posts/:id`         | admin 或作者     | 软删除文章。                               |

`POST /api/v1/posts` 请求体（UserID 取自 JWT，绝不接受客户端传入）：

```jsonc
{
  "title": "我的文章",
  "summary": "可选摘要",
  "cover": "https://...",          // 可选
  "content": "正文……",
  "status": "published",           // draft | published | archived
  "categoryIDs": [1, 2],           // 可选，已有分类 id
  "tags": [3, 4]                   // 可选，已有标签 id
}
```

`status` 取值：`draft`、`published`、`archived`。文章的标签与分类通过**已有** id 引用；
当前没有创建标签、分类或用户的 API 接口。

### 友链

| 方法 | 路径                                | 鉴权 | 说明                                       |
| ---- | ----------------------------------- | ---- | ------------------------------------------ |
| GET  | `/api/v1/links`                     | —    | 列出已启用的友链（公开）。                 |
| POST | `/api/v1/links/apply-link`          | —    | 提交友链申请（创建为禁用状态，待审核）。   |

`POST /api/v1/links/apply-link` 请求体：

```jsonc
{ "name": "示例", "url": "https://example.com", "description": "……", "avatar": "https://…" }
```

友链状态（`normal`/`abnormal`）仅由每小时运行的定时任务计算；没有用于审核、编辑或删除友链的 HTTP 接口。

### RSS

| 方法 | 路径                        | 鉴权 | 说明                                            |
| ---- | --------------------------- | ---- | ----------------------------------------------- |
| GET  | `/api/v1/rss`               | —    | RSS 2.0 feed；可选 `?page=N`（每页 10 条）。     |
| GET  | `/api/v1/rss/complete`      | —    | 包含所有已发布文章的 RSS 2.0 feed。              |

### LLM 摘要

| 方法 | 路径                                           | 鉴权 | 说明                              |
| ---- | ---------------------------------------------- | ---- | --------------------------------- |
| POST | `/api/v1/model/summarize`                      | —    | 创建摘要会话，返回 `sessionId`。   |
| GET  | `/api/v1/model/summarize/:sessionId`           | —    | 以 SSE 流式返回生成的摘要。        |

`POST /api/v1/model/summarize` 请求体：`{ "content": "……正文……" }`。

流会以 `data: <text>\n\n` 行推送内容，以 `event: done\ndata: [DONE]` 结束，
失败通过 `event: error\ndata: <message>` 上报。

## 项目结构

```
cmd/server/             # 应用装配、migration、HTTP 与后台任务生命周期
internal/
  account/              # 认证、用户、邮件和账户模型
  content/              # 文章、分类、标签、RSS 和内容模型
  link/                 # 友链及友链模型
  setup/                # 站点初始化和站点模型
  summary/              # Summary sessions and SSE
  api/                  # Routes, errors, health checks
  platform/             # Cache, config, database, logging, middleware, storage
  shared/               # 上下文、错误、HTTP 信封、JWT、随机值、事务和校验
  authz/                # Authorization rules
  scheduler/            # Background jobs
ent/                    # Schemas and generated code
migration/              # 内嵌 Goose SQL migration 与启动模块
templates/              # 邮件模板
```

行为变化、部署限制和集成测试命令见 [REFACTOR.md](REFACTOR.md)。

### 请求流程

1. `cmd/server/main.go` 构建 Fx 容器；内嵌 Goose migration 在 Echo 接受请求前作用于 PostgreSQL，定时任务随应用启动。Redis 可用性由就绪检查报告。
2. 全局中间件提供请求日志、panic 恢复、10 MiB 请求体限制和基于 Redis 原子计数的登录限流。
3. 路由注册在 `/api/v1` 下。受保护路由挂载 `middleware.AuthMiddleware`，
   解析 Bearer 令牌，将用户解析为 `contextx.User` 并存入请求上下文。
4. 各功能 handler 自己持有请求/响应 DTO；通用 HTTP 信封和分页类型位于 `internal/shared/httpx`。
   领域错误由统一的 `api.ErrorHandler` 渲染。
5. 服务层通过功能包内的 repository（Ent）与事务管理器执行写操作，文章与友链通过
   `authz.Authorizer` 进行授权。

### 定时任务

在应用启动时运行，停止时取消：

- **浏览量落库** —— 每小时读取 Redis 中的 `blog:post:view_count:*` 键（`SCAN` + `GETDEL`），并批量累加到 PostgreSQL。
- **友链状态检查** —— 每小时对每个 HTTPS 友链发起 `GET`，状态变化时将其更新为 `normal`/`abnormal`。

## 状态 / 尚未暴露

以下能力存在于代码库中，但**无法通过 HTTP 访问**：

- **S3 对象存储**：`storage.S3Storage` 实现了 `Upload`/`Download`/`Delete`/`Copy`/`Exists`，
  但并未注册到 Fx 图中，也没有任何处理器使用它。**不存在 `/api/upload` 接口**。
- **友链管理端 CRUD**：没有创建/更新/删除/审核友链的接口；友链生命周期部分由定时任务驱动。
- **标签、分类、用户**：没有创建或管理它们的接口；文章仅引用已存在的 id。
- **评论**：Ent schema 中定义了 `Comment` 表，但没有对应的 repository、service 或 handler。
- **CORS**：`app.cors_origins` 已在配置中定义，但没有注册任何 CORS 中间件，因此服务端不处理跨域请求。
- `service.PostService.GetPostsWithContent` 是一个桩方法，始终返回 `(nil, nil)`。
