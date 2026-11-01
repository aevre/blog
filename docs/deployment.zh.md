[English](deployment.md) · [中文](deployment.zh.md)

# 容器部署指南（Podman / Docker Compose）

整套服务（对象存储、数据库、缓存、后端、前端、反向代理）都已收敛到一个 `compose.yml` 中，
clone 后只需准备一份 `.env` 即可一键启动，Podman 与 Docker 均可。

rootless / rootful 由运行方式自动判断，不依赖环境变量：普通用户 Podman 为 rootless（80/443 自动回落 8080/8443）；
`sudo` 运行、或 docker root daemon 为 rootful（直接绑 80/443，无需手动端口转发）。

> 想完全自己掌控各模块运行时、不要使用容器一键自动化部署，见 [自行部署](self-hosting.zh.md)。

> 想用 CI 构建镜像下发，见 [GitHub Actions 自动部署](ci.zh.md)。

## 0. 选择部署方式

| 方式 | 触发 | 端口 | 证书存放 | 适用 |
| --- | --- | --- | --- | --- |
| 容器（rootless） | `./deploy.sh`（普通用户 Podman） | 自动回落 8080/8443 | 项目内 `deploy/letsencrypt`（部署用户拥有） | 不想用 root、接受高位端口 |
| 容器（rootful） | `sudo ./deploy.sh`（或 docker root daemon） | 直接绑 80/443，无需转发 | 项目内 `deploy/letsencrypt`（部署用户拥有） | 想要真 80/443、不想搞端口转发 |
| 自行部署（打包+步骤） | `./deploy.sh self` | 由用户自行决定 | 由用户自行决定 | 完全自己掌控各模块运行方式，脚本只打包 + 给步骤 |

- rootless / rootful 由运行身份自动判定（`id` 是否为 root、podman 是否 rootless、docker daemon 是否为 root），不靠环境变量区分。
- 容器两方式共用同一套 compose 逻辑与证书目录。

## 1. 前置要求

- 已安装 Podman（含 `podman compose`）或 Docker（含 `docker compose` 插件）
- `envsubst`（通常随 `gettext` 提供）
- 服务器已开放 `80` / `443` 端口（以及按需的 `9001` 用于 RustFS 控制台）

## 2. 准备配置

```bash
cp .env.example .env
# 编辑 .env，至少修改以下项：
#   APP_DOMAIN       后端使用的完整公开地址（含协议），如 https://blog.example.com
#   NGINX_SERVER_NAME  Nginx 的 server_name（裸域名），如 blog.example.com
#   POSTGRES_PASSWORD / JWT_SECRET / RUSTFS_*  各类密钥
```

`.env` 中的变量含义见 `.env.example` 内的注释。

> 域名不只是填在 `.env` 里：`NGINX_SERVER_NAME` / `APP_DOMAIN` / `RUSTFS_ENDPOINT` 仅用于脚本生成 Nginx 的 `server_name` 与后端配置，并不会把流量自动路由过来。要让外网通过该域名访问、并让 Let's Encrypt 完成 ACME 验证，你必须**自行在域名服务商处把 DNS A/AAAA 记录解析到服务器公网 IP**（RustFS 域名同理）。DNS 未生效前，证书签发会失败、浏览器也无法访问。

## 3. 一键部署

```bash
./deploy.sh              # 容器（rootless / rootful 由运行方式自动判定）
sudo ./deploy.sh         # rootful：直接绑 80/443，无需端口转发
```

脚本会自动：

1. 若没有 `.env` 则基于 `.env.example` 生成一份（生成后请先填写再重跑）
2. 通过 `envsubst` 由 `config/backend_config.toml` 生成后端配置 `deploy/runtime/backend/config.toml`
3. 根据 `NGINX_TLS` 生成 Nginx 配置 `deploy/runtime/nginx/default.conf`，默认使用 HTTP 模板，设为 `https` 时切换为 HTTPS 模板
4. 按依赖顺序拉起服务（postgres/redis → migration → backend/frontend → nginx）
   - 其中的 `migration` 服务会在 PostgreSQL 就绪后自动执行 `ent` 的
     `Schema.Create` 初始化表结构（一次性，退出码 0 即成功），
     `backend` 通过 `depends_on: migration: service_completed_successfully`
     等待迁移完成后再启动，因此无需手动建表。

> 为什么不复用 compose 的 `depends_on` 条件一把 `up` 到底？
> `podman-compose` 1.0.6 对 `service_healthy` / `service_completed_successfully` 支持不可靠，
> 会把 `backend` 留在 `Created` 不启动、或让 `migration` 在 PG 未就绪时退出。
> `deploy.sh` 改为显式按依赖顺序拉起并轮询就绪，规避该问题。

部署完成后：

- 前端：`http://<NGINX_SERVER_NAME>`
- 后端 API：`http://<NGINX_SERVER_NAME>/api/v1/`
- RustFS 控制台：`http://<服务器IP>:9001`

常用管理命令：

```bash
podman compose ps                 # 查看状态
podman compose logs -f backend    # 查看后端日志
podman compose down               # 停止并移除容器
./deploy.sh                       # 重新部署（更新代码后重跑即可，会自动重建）
```

## 关于 Rootless / Rootful 与 80/443 端口

`deploy.sh`（及 CI 的 `deploy/remote-deploy.sh`）按运行方式区分 rootless / rootful，无需设置环境变量：

- **rootful（推荐想要真 80/443 的用户）**：以 `sudo` 运行，或走 docker root daemon。脚本直接绑定 80/443，无需端口转发。注意 rootful 下证书目录属主：脚本会在 compose 首次挂载前以部署用户预创建 `deploy/letsencrypt` 与 `deploy/acme`，并在 root 运行时把属主 `chown` 回部署用户，避免 docker/podman daemon 把缺失的 bind 源目录建成 `root:root` 导致 certbot 写入失败。

  ```bash
  sudo ./deploy.sh
  ```

  保持 `.env` 中 `HTTP_PORT=80` / `HTTPS_PORT=443` 即可。脚本会判定为 rootful 并直接绑 80/443。

- **rootless**：普通用户 Podman 无法把容器端口映射到宿主机 `1024` 以下的端口，直接用默认 `80/443` 会启动失败。有三种处理方式：

  1. 以 rootful 运行（见上，`sudo ./deploy.sh`）
  2. 放开非特权端口

     ```bash
     echo 'net.ipv4.ip_unprivileged_port_start=0' | sudo tee /etc/sysctl.d/99-podman-port.conf
     sudo sysctl --system
     ```

     之后普通 `./deploy.sh` 也能绑定 80/443，代价是所有用户都能绑定低位端口。
     `deploy.sh` 与 `deploy/remote-deploy.sh` 会自动检测该值：只要 `net.ipv4.ip_unprivileged_port_start<=80`，
     就继续使用 `80:80` / `443:443`（无需改 `.env`），ACME HTTP-01 挑战直接走 80 端口即可签发。

  3. 保持 rootless + 高位端口，再用宿主机反代 / 转发

     将 `.env` 改为 `HTTP_PORT=8080` / `HTTPS_PORT=8443`（rootless 下脚本也会自动回落到这两个端口），
     然后在宿主机把 `80→8080`、`443→8443` 转发。本质是 TCP 端口转发（stream），
     **不要在上游反代上终止 TLS**——证书在容器内的 nginx 上，让 80 的 ACME HTTP-01 挑战与 443 的访问都直达容器 nginx。

     **方式 a：socat（最简单，随装随用）**

     ```bash
     sudo socat TCP-LISTEN:80,fork,reuseaddr TCP:127.0.0.1:8080 &
     sudo socat TCP-LISTEN:443,fork,reuseaddr TCP:127.0.0.1:8443 &
     ```

     注意 `&` 只在当前会话有效，重启后会丢失。简单常驻可加 `nohup ... &`；要开机自启见方式 c。

     **方式 b：宿主 Nginx 做 TCP 透传（stream，开机自启、更稳）**

     方式 b 依赖 Nginx 的 stream 模块做四层 TCP 转发，不是普通的 `http {}` 反代。下面假设你的宿主
     `nginx.conf` 已包含顶层 `include /etc/nginx/modules-enabled/*.conf;` 以加载动态模块，以及顶层
     `stream { include /etc/nginx/stream/*.conf; }` 作为四层转发块。若你的 nginx.conf 还没有顶层
     `stream {}` 块，见末尾「没有顶层 stream 块时」的补充。

     **步骤 1：装 stream 模块（Ubuntu / Debian）**

     ```bash
     sudo apt-get update && sudo apt-get install -y libnginx-mod-stream
     ```

     装完默认即可用：该包会自动在 `/etc/nginx/modules-enabled/` 下建立软链启用模块，
     配合你已有的顶层 `stream { include /etc/nginx/stream/*.conf; }`，无需再改 `nginx.conf`。
     确认一下：

     ```bash
     ls -l /etc/nginx/modules-enabled/ | grep -i stream   # 应有 mod-stream.conf 软链
     nginx -V 2>&1 | grep -o with-stream                  # 动态模块也可见（或静态编译）
     ```

     极少数情况没自动建链，手动补一条：

     ```bash
     sudo ln -s /usr/share/nginx/modules-available/mod-stream.conf /etc/nginx/modules-enabled/mod-stream.conf
     ```

     RHEL / CentOS / Rocky 用：`sudo yum install -y nginx-mod-stream`。

     **步骤 2：创建 stream 转发配置**

     因为外层 `stream {}` 已经在 `nginx.conf` 里了，**这里不要再写 `stream {}` 外层**，
     直接在 include 目录里新建 `/etc/nginx/stream/stream.conf`（文件名任意 `*.conf` 都行）：

     ```nginx
     # /etc/nginx/stream/stream.conf
     server { listen 80;  proxy_pass 127.0.0.1:8080; }
     server { listen 443; proxy_pass 127.0.0.1:8443; }
     ```

     **步骤 3：校验并重载**

     ```bash
     sudo nginx -t            # 必须 syntax is ok / test is successful
     sudo systemctl reload nginx
     ```

     前提：宿主已装并运行 Nginx（仅用于转发，不托管站点、不持有证书）。验证转发见方式 c 下方的提示——
     公网能访问 `http://域名:80/.well-known/acme-challenge/` 且返回 404 即说明 80 已透传到容器 nginx。

     没有顶层 `stream {}` 块时，例如 Debian/Ubuntu 默认 `nginx.conf`：若是动态模块，在 `nginx.conf`
     最顶部、`events {}` 之前加 `load_module modules/ngx_stream_module.so;`；在 `http {}` 之外加一行
     `include /etc/nginx/stream.conf;`，并新建该顶层文件写 `stream { ... }`。注意 `stream {}` 是顶层指令，
     不能放进 `conf.d/`，那里在 `http {}` 内部，会报 `"stream" directive is not allowed here`。

     **方式 c：systemd 持久化 socat（开机自启、崩溃自愈）**

     ```ini
     # /etc/systemd/system/blog-port-forward.service
     [Unit]
     Description=Forward 80/443 to rootless podman 8080/8443
     After=network-online.target
     Wants=network-online.target

     [Service]
     Type=simple
     ExecStart=/bin/sh -c 'socat TCP-LISTEN:80,fork,reuseaddr TCP:127.0.0.1:8080 & socat TCP-LISTEN:443,fork,reuseaddr TCP:127.0.0.1:8443 & wait'
     Restart=always

     [Install]
     WantedBy=multi-user.target
     ```

     ```bash
     sudo systemctl enable --now blog-port-forward
     ```

     验证转发是否生效：`curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:80/.well-known/acme-challenge/`
     返回 `404` 即说明 80 已转发到容器 nginx 并提供挑战目录。公网还需能访问到域名 `:80`。满足后，部署脚本里的
     “ACME 挑战端口 80 已可达”检查才会通过，证书才能自动签发。

若以 rootless 运行，且既未放开低位端口也没有宿主反代，`deploy.sh` 与 `deploy/remote-deploy.sh` 会自动把
`80/443` 改为 `8080/8443` 以保证能启动，并在终端提示如何开放真实端口。此时 80 端口不在容器内，ACME HTTP-01
挑战只验证 80 端口，会失败并回退到 HTTP。要让 HTTPS 生效，需满足上文的方式 2 或方式 3 之一。

## 4. 启用 HTTPS

默认以 HTTP 提供服务，可切换到 HTTPS。证书由 Let's Encrypt 签发，使用 ACME HTTP-01 挑战。certbot 在挑战目录写入校验文件，Let's Encrypt 通过 `http://域名/.well-known/acme-challenge/...` 访问并确认域名归属，验证通过后才签发。HTTP-01 固定只验证 80 端口。容器内 Nginx 已配置 `location /.well-known/acme-challenge/` 提供该目录。

### 自动签发

把 `.env` 中 `NGINX_TLS` 设为 `https`，然后照常运行部署脚本即可，无需手动分步操作：

```bash
./deploy.sh               # 容器一键部署（rootless / rootful 自动判定）
sudo ./deploy.sh          # rootful
bash deploy/remote-deploy.sh all    # CI 服务器侧部署
```

部署脚本在生成最终配置之前，先以 HTTP 模板拉起 Nginx 提供 ACME 挑战，再以 certbot 签发或续期证书，最后用 HTTPS 模板重新生成配置并启动。这一步避开了“没有证书 Nginx 起不来、起不来就拿不到证书”的死锁。`NGINX_TLS_AUTO` 默认开启，设为 `off` 时脚本不再自动签发，改由你预先用 `deploy/certbot.sh` 准备好证书。

每个域名独立判断证书状态，blog 与 RustFS 各算各的：

- 缺失或已过期：全新签发，使用 certonly
- 剩余有效期不超过 `RENEW_DAYS`，默认 30 天：续期，使用 renew
- 剩余有效期大于 30 天：跳过

自动签发需要 80 端口对公网可达且已安装 certbot。若 80 端口不可达，脚本报错并回退到 HTTP 模板，站点以 HTTP 提供。rootless 下开放 80 端口的方法见上文 Rootless 章节。

手动签发与续期的等价命令：

```bash
./deploy/certbot.sh certonly -d blog.example.com     # 首签
./deploy/certbot.sh renew                            # 续期，沿用项目内 renewal 配置
sudo ./deploy/certbot.sh import blog.example.com     # 从宿主 /etc/letsencrypt 导入
```

`deploy/certbot.sh` 默认以部署用户身份运行项目模式 certbot，证书存放在项目目录 `deploy/letsencrypt/etc/{live,archive,renewal}/`，归部署用户所有，nginx 容器可直接读取。未显式指定 Webroot 时，脚本使用宿主机绝对路径 `deploy/acme`，该目录映射到容器内的 `/var/www/certbot`；项目模式拒绝把容器路径作为宿主 Webroot。未指定邮箱时，脚本读取 `.env` 的 `EMAIL_FROM` 或 `CERTBOT_EMAIL`，并自动带上 `--agree-tos --no-eff-email --non-interactive`。

`certbot.sh` 支持 `CERT_MODE` 环境变量：`project`（默认，项目内目录，容器模式用）与 `system`（`/etc/letsencrypt`，root 拥有、由系统 certbot 管理，供自行部署后在目标机以系统 certbot 手动签发证书使用）。`import` 子命令仅 `project` 模式有意义。

### DNS-01 挑战

rootless 配合高位端口（如 8080）时，Let's Encrypt 访问的是 80 端口，挑战会失败。此时可把宿主机 80 转发到 8080，或用 DNS-01 挑战（certbot 的 DNS 插件），无需入站 80 端口，最适合 rootless：

```bash
./deploy/certbot.sh certonly --dns-cloudflare -d blog.example.com
```

### 证书状态异常的恢复

自动签发时发现 `live/<域名>` 缺失但 `archive/<域名>` 仍在，或 `renewal/<域名>.conf` 为空，即签发中途失败留下的残缺配置。部署脚本不会自动删除这些状态，而是停止并提示检查，以保护历史 archive 和私钥。确认备份后需要手工清理时，再删除下列路径并重新部署：

```bash
rm -rf deploy/letsencrypt/etc/live/<域名> deploy/letsencrypt/etc/archive/<域名>
rm -f  deploy/letsencrypt/etc/renewal/<域名>.conf
```

## 5. RustFS 反代（前端显示图片等静态资源所需，必配）

前端通过 RustFS 承载图片等静态资源，因此 `RUSTFS_ENDPOINT` 为必填项，且必须是公网可达的
`https://` 地址。在 `.env` 中设置后，部署脚本会自动：

- 由该地址的域名推导 API 主机（如 `oos.immortel.top`）与控制台主机（自动加 `admin.` 前缀，如
  `admin.oos.immortel.top`）；
- 在容器内的 Nginx 增加对应反代：API → `rustfs:9000`，控制台 → `rustfs:9001`（`client_max_body_size 10G`）；
- 证书签发/续期时一并申请这两个域名的 SAN 证书（与 blog 站点各自独立）。

注意：启用后宿主 Nginx 不要再单独以 `http {}` 块终结这两个域名（会与第二层 stream 透传冲突），
只需保持 80/443（或 8080/8443）的 TCP 透传到容器 Nginx 即可；原来的宿主 `oos.conf` 可移除。
`RUSTFS_ENDPOINT` 仍作为后端连接对象存储的地址（前后端统一走该公网域名）。


`deploy/letsencrypt/`、`deploy/acme/`、`deploy/runtime/` 由部署脚本在运行时生成，已被 `.gitignore` 忽略，请勿提交。
