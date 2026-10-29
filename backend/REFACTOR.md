# 后端重构说明

现有 `/api/v1` 方法、路径、请求 DTO、响应 DTO、Cookie 名称与属性保持兼容。新增 `/health/live` 与 `/health/ready`。

## 目录

- `cmd/server`：组合 Fx 模块、HTTP 生命周期。
- `internal/account`：认证、用户、邮件；按职责包含 handler、service、repository、mapper。
- `internal/content`：文章、分类关联、标签关联、RSS、详情缓存和浏览量。
- `internal/link`：友链申请和检查。
- `internal/setup`：初始化事务与站点配置。
- `internal/summary`：摘要会话与 SSE。
- `internal/api`：统一路由、错误响应、健康检查。
- 各功能的 HTTP 请求与响应 DTO 直接放在对应 `handler` 包，避免跨功能的 DTO 层。
- `internal/platform`：cache、config、database、logging、middleware、storage。
- `internal/shared`：contextx、errx、httpx、jwt、txmgr、validatorx；`httpx` 仅包含通用响应信封、分页包装与输出辅助。
- `internal/authz`：权限规则与所有权检查。
- `internal/scheduler`：定时任务与停止等待。
- `ent`：现有 schema 和生成代码；`entity`：现有领域模型。
- `migration`：原有嵌入式 goose migration，历史 SQL 未修改。

## 行为与运行限制

Redis 提供具体 Client，业务依赖按实际用途定义的小接口。缓存未命中返回 `cache.ErrMiss`，与连接故障分开。浏览量批量 GETDEL 允许部分 key 不存在。

刷新会话按用户和 session ID 隔离，仅存 SHA-256 token 哈希。PostgreSQL 事务与会话行锁原子轮换，旧 token 重用撤销对应 session。退出只撤销当前 session。旧版本 Redis 会话不迁移，部署后用户须重新登录。Access Token 仍有效至过期，没有增加 JWT 黑名单。刷新 token 使用带随机 jti 的签名 JWT，持久会话与 token 的 SHA-256 哈希、expires_at 保存在 PostgreSQL；过期时间与 JWT exp 一致。migration 00004 撤销旧不透明 token 会话并为旧哈希设置立即过期，升级后须重新登录；历史 migration 未改动。历史哈希保留至对应 JWT 过期，用于重用检测。scheduler 启动时及每小时清理，每轮最多删除 500 条过期哈希和 500 个已无 token 的失效会话，单轮超时 10 秒，使用索引及 SKIP LOCKED。过期 JWT 在数据库查询前被拒绝，不再触发所属会话撤销；持续活跃会话也会回收过期历史。若过期量超过每小时 500 条，积压会留待后续批次，应根据实际负载调整批量与频率。

公开文章详情 key 为 `blog:post:detail:<id>:<updated_at纳秒>`，TTL 一分钟。每次先从数据库查询发布状态与版本，再读取缓存；Redis 不可用时读取数据库。更新事务提交后、删除成功后清理旧版本 key。并发请求可能返回其开始读取时的版本；旧值回填到旧 key 后不会被新版本读取，最多保留一分钟。数据库不可用时不从缓存绕过发布状态检查。

登录按直连 IP 限制每分钟 10 次，Lua 计数并设置窗口过期。仅使用直接来源地址，不信任转发头。Redis 故障时放行并记录错误，因此故障期间限流失效；共享出口用户共用额度。被限流返回 HTTP 429，使用现有错误信封。

`/health/live` 仅检查进程存活；`/health/ready` 在两秒内依次检查 PostgreSQL、Redis，任一失败返回 503，因为验证码和运行能力仍依赖 Redis。

配置通过构造函数传入，不再使用全局配置单例。保留文件变更触发进程退出的现有行为。PostgreSQL DSN 改为标准 URL 编码，修复空密码吞掉 dbname 与特殊字符解析错误。

本轮初次集成验证受旧 DSN 缺陷影响，连接到了本机默认数据库 `immortal`，并应用了 migration 00001–00003。已精确删除本轮创建的两条测试用户和一条测试文章；该库中的 migration 表及结构仍保留。修复后在明确指定的临时库重跑全部 PostgreSQL 集成验证并清理临时库。

## 验证

```sh
go test ./...
go vet ./...
go test -race ./internal/... ./cmd/...
TEST_REDIS_ADDR=127.0.0.1:6379 go test -race ./internal/platform/cache -v
createdb blog_backend_test
TEST_POSTGRES_DB=blog_backend_test go test -p 1 ./internal/account/repository ./internal/content/repository -v
dropdb blog_backend_test
```

PostgreSQL 集成测试使用本机 Unix socket 和当前系统用户，仅指定专用临时数据库；会执行 migration 并写入测试数据。Redis 测试只使用唯一前缀并清理自己的 key，也支持 Unix socket 路径。

已覆盖 Fx 依赖图、现有路由、登录 JSON/Cookie、会话轮换/注销、真实 PostgreSQL 并发轮换、Redis 并发限流、缓存命中和草稿保护、Redis 故障回退、限流故障策略、真实 PostgreSQL migration/所有权/文章版本/事务回滚。未执行浏览器端到端、SMTP、SSE 上游或生产环境验证。
