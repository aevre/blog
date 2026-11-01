[English](self-hosting.md) · [中文](self-hosting.zh.md)

# 自行部署（打包产物 + 部署步骤）

不想用一键自动化、想完全自己掌控各模块运行方式的用户，用“自行部署”模式：

```bash
./deploy.sh self
```

脚本构建并打包产物，不运行任何服务、也不自动签发证书。
运行方式由你在目标机上自行决定。

## 打包内容

`blog-selfhost-<日期>.tar.gz` 包含：

- `backend/bin/blog-server`、`backend/bin/migration` —— 后端二进制（CGO_ENABLED=0 静态构建）
- `backend/bin/config.toml` —— 已按本机 `.env` 生成好的后端配置（与二进制同级）
- `frontend/` —— Nuxt SSR 产物
- `config/backend_config.toml` —— 后端配置模板
- `deploy/nginx/*.template` —— Nginx 配置模板
- `.env.example` —— 变量说明

> 注意：打包产物不含 `.md` 文档，文档以本文件及本站其他 `docs/` 为准。

## 构建位置选择

`./deploy.sh self` 需要在有完整工具链的机器上运行（见下“服务器构建环境需求”）。有两种选择：

- **A. 本地构建后上传（推荐）**：在自己的开发机 / CI 上跑 `./deploy.sh self`，把生成的
  `blog-selfhost-*.tar.gz` 传到目标服务器解压即可。目标服务器无需 Go / pnpm（构建工具链），
  但前端 Nuxt SSR 仍要 Node.js 运行；其余只要能解压 tar 并运行二进制 / 起容器（或装好 Nginx、PostgreSQL 等既有服务）。
- **B. 直接在服务器构建**：在目标服务器上 clone 仓库并跑 `./deploy.sh self`（或手动 `go build` / `pnpm build`）。
  此时目标服务器必须装齐下方“服务器构建环境需求”的全部工具链。

两种方式产出的包一致；区别只在“构建发生在哪台机器”。多数情况选 A 更省事、也更安全。

## 服务器构建环境需求

若选择在目标服务器上构建，需具备：

- **Go** `1.27`（见 `backend/go.mod`；`deploy.sh self` 用 `CGO_ENABLED=0 go build`，无需 C 工具链）
- **Node.js** `24`（见 `frontend/Dockerfile`：`node:24`）与 **pnpm** `11.15.1`（见 `package.json` 的 `packageManager`）
- **envsubst**（随 `gettext` 提供；用于在服务器侧按 `.env` 生成 `config/backend_config.toml`）
- **git**（clone 仓库用）
- 磁盘与时间：首次 `pnpm install` + `pnpm build` 与 `go build` 会拉取依赖并编译，请预留足够空间与几分钟。

方式 A 的目标服务器不需要以上构建工具链（Go / pnpm / git / envsubst），仅需能运行产物与依赖服务；前端 Nuxt SSR 仍需 Node.js 运行。

## 部署步骤（目标机）

下面只给步骤，不替你执行。各模块怎么跑由你决定：容器（rootless / rootful 自选）、原生（系统包 / 二进制 / systemd）、
或用已有实例（如托管 PostgreSQL）。

### 0. 目标机前置

- PostgreSQL（建库 `blog`、用户、密码）、Redis
- 可选 RustFS（对象存储，前端图片用）或任意 S3 兼容
- Nginx（反代 / TLS 终结）、certbot（要 HTTPS 时）

> **域名需要你自行做 DNS 解析**：包里 `.env` 的 `NGINX_SERVER_NAME` / `APP_DOMAIN` / `RUSTFS_ENDPOINT` 只是给 Nginx / 后端当识别与转发配置用，不会自动生效。请在域名服务商把 A/AAAA 记录指向目标机公网 IP（RustFS 域名同理），否则外部无法访问、证书也无法签发。

### 1. 放置产物

解压本包到目标机某目录（如 `/srv/blog`）。后端通过环境变量 `CONFIG_FILE` 指向
`backend/bin/config.toml`（与 `blog-server` 同级；或自行用 `config/backend_config.toml` 生成）。

### 2. 后端

先 `./backend/bin/migration`再
  `CONFIG_FILE=/srv/blog/backend/bin/config.toml ./backend/bin/blog-server`

### 3. 前端

Nuxt SSR 以 `node` 运行，目标机需装 Node.js（版本同构建，见 `frontend/Dockerfile` 的 `node:24`）。

`cd /srv/blog/frontend && NUXT_BACKEND_URL=http://127.0.0.1:8000 node server/index.mjs`

### 4. 数据库 / 缓存 / 对象存储

- 用已有实例：直接在 `config.toml` / `.env` 填 `DB_HOST`、`REDIS_HOST`、`RUSTFS_ENDPOINT`。
- 用容器：另起 postgres / redis / rustfs 容器，后端通过服务名或 `127.0.0.1` 连接。

### 5. Nginx 反代 + TLS

- 模板 `deploy/nginx/http.conf.template`、`https.conf.template` 已含
  `location /.well-known/acme-challenge/`（ACME 挑战）与 443 TLS。
- 把 upstream 改成本机 `127.0.0.1:3000 / 127.0.0.1:8000`（模板里是容器服务名，自行替换），
  证书放 `/etc/letsencrypt/live/<域名>/`，写进 `/etc/nginx/conf.d/blog.conf` 后
  `nginx -t && systemctl reload nginx`。

### 6. 证书（三选一）

- 自动签发：`CERT_MODE=system ./deploy/certbot.sh certonly -d <域名>`（系统模式使用宿主机 `/var/www/certbot`）
- 已有证书：把 `fullchain.pem` / `privkey.pem` 放到对应路径
- 用脚本：`CERT_MODE=system ./deploy/certbot.sh certonly -d <域名>`（系统 `/etc/letsencrypt` 模式）

### 7. 启动顺序

1. PostgreSQL / Redis / RustFS 就绪
2. 跑 migration
3. 起 backend（8000）、frontend（3000）
4. 起 nginx（80/443）；HTTPS 需先有证书
