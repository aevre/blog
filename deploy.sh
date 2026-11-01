#!/usr/bin/env bash
# 一键部署：检测 podman/docker compose，生成后端与 nginx 配置并拉起全部服务。
# 用法：
#   ./deploy.sh            容器模式（rootless / rootful 由运行方式自动判定）
#   sudo ./deploy.sh manual 原生模式（不使用 docker/podman，系统 nginx + 系统 certbot）
set -euo pipefail

# 自行部署 打包产物
# 通过 `./deploy.sh self`或交互选择触发：构建并打包产物
deploy_self() {
  local ROOT
  ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  cd "$ROOT"

  if [ ! -f .env ]; then
    echo "ERROR: .env 不存在，请先 cp .env.example .env 并填写。" >&2
    exit 1
  fi
  set -a
  # shellcheck source=/dev/null
  source <(sed 's/\r$//' .env)
  set +a

  echo "===== 自行部署：构建并打包产物 ====="
  echo "工作目录：$ROOT"
  echo "目标：NGINX_SERVER_NAME=${NGINX_SERVER_NAME:-<未设置>}，RUSTFS_ENDPOINT=${RUSTFS_ENDPOINT:-<未设置>}"

  # 构建后端二进制
  echo "[1/4] 构建后端二进制（CGO_ENABLED=0 go build）→ backend/bin/blog-server"
  (cd backend && CGO_ENABLED=0 go build -o bin/blog-server ./cmd/server) ||
    {
      echo "ERROR: 后端构建失败" >&2
      exit 1
    }
  echo "      ✓ 后端二进制构建完成"

  # 构建前端产物
  echo "[2/4] 构建前端产物（pnpm install && pnpm build）→ frontend/.output"
  (cd frontend && pnpm install && pnpm build) ||
    {
      echo "ERROR: 前端构建失败" >&2
      exit 1
    }
  echo "      ✓ 前端产物构建完成"

  # 生成后端配置
  mkdir -p backend/bin
  # shellcheck disable=SC2016
  local BACKEND_VARS='$APP_NAME $APP_DOMAIN $DB_HOST $DB_PORT $DB_USER $DB_PASSWORD $JWT_SECRET $EMAIL_HOST $EMAIL_PORT $EMAIL_USERNAME $EMAIL_PASSWORD $EMAIL_FROM $MODEL_API_KEY $RUSTFS_ACCESS_KEY_ID $RUSTFS_SECRET_ACCESS_KEY $RUSTFS_ENDPOINT'
  envsubst "$BACKEND_VARS" <config/backend_config.toml >backend/bin/config.toml
  echo "[3/4] 生成后端配置：backend/bin/config.toml（已按 .env 变量填充）"

  # 4. 打包
  local BUNDLE
  BUNDLE="blog-selfhost-$(date +%Y%m%d).tar.gz"
  echo "[4/4] 打包产物 → $BUNDLE"
  echo "      包含：backend/bin/{blog-server,config.toml}、config/backend_config.toml、deploy/nginx/*.template、.env.example、frontend/"
  tar -czf "$BUNDLE" \
    --transform 's,^\./,frontend/,' \
    backend/bin/blog-server \
    backend/bin/config.toml \
    config/backend_config.toml \
    deploy/nginx .env.example \
    -C frontend/.output . 2>/dev/null &&
    echo "      ✓ 已打包：$BUNDLE（$(du -h "$BUNDLE" | cut -f1)）"
  echo ""
  echo "把 $BUNDLE 传到目标机解压后自行部署即可。"
  echo "详细步骤见仓库 docs/self-hosting.md。"
}

# 入口：让用户选择容器部署或自行部署
choose_mode() {
  local mode="${1:-}"
  if [ -z "$mode" ]; then
    if [ -t 0 ]; then
      printf '请选择部署方式：\n  1) 容器部署\n  2) 自行部署\n输入 [1/2]（默认 1）：'
      read -r ans
      case "${ans:-}" in
      2) mode=self ;;
      *) mode=container ;;
      esac
    else
      mode=container
    fi
  fi
  case "$mode" in
  container) : ;;
  self | package)
    deploy_self
    exit 0
    ;;
  *)
    echo "用法：./deploy.sh [container|self]（container=容器部署，self=自行部署打包，省略则交互选择）" >&2
    exit 1
    ;;
  esac
}
choose_mode "${1:-}"

if command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1; then
  COMPOSE=(podman compose)
elif command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
else
  echo "未检测到 podman 或 docker（需支持 compose 子命令），请先安装。" >&2
  exit 1
fi

if ! command -v envsubst >/dev/null 2>&1; then
  echo "缺少 envsubst 命令，请先安装 gettext（Debian/Ubuntu: apt install gettext；Fedora: dnf install gettext）。" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

if [ ! -f .env ]; then
  if [ ! -f .env.example ]; then
    echo ".env.example 缺失，无法生成 .env" >&2
    exit 1
  fi
  cp .env.example .env
  echo "已根据 .env.example 生成 .env，请按需修改其中的域名与密钥后重新运行 ./deploy.sh"
  exit 0
fi

# 载入 .env 供 envsubst 使用
set -a
# 去 CRLF，避免 \r 残留使端口等整数字段解析失败
# shellcheck source=/dev/null
source <(sed 's/\r$//' .env)
set +a

HTTP_PORT="${HTTP_PORT:-80}"
HTTPS_PORT="${HTTPS_PORT:-443}"

# 是否 rootless 由运行方式自动判断：普通用户 Podman 为 rootless；
# sudo 运行、或 docker root daemon 均为 rootful。无需通过环境变量区分。
# rootless 无法绑 1024 以下端口 → 回落 8080/8443；否则直接绑 80/443，无需手动转发。
ROOTLESS="false"
if [ "${COMPOSE[0]}" = "podman" ]; then
  if podman info --format '{{.Host.Security.Rootless}}' 2>/dev/null | grep -qi true; then
    ROOTLESS="true"
  fi
fi

if [ "$ROOTLESS" = "true" ] && [ "$HTTP_PORT" = "80" ]; then
  HTTP_PORT=8080
  HTTPS_PORT=8443
  echo "检测到 rootless（普通用户 Podman），已将 80/443 自动改为 8080/8443（详见 README 如何开放真实 80/443）。"
else
  echo "rootful（sudo 运行，或 docker root daemon）：直接绑定 80/443，无需手动端口转发。"
fi
export HTTP_PORT HTTPS_PORT

# 证书目录预创建
# 避免 rootful 下 docker/podman daemon 把缺失的 bind 源目录建成 root:root，
# 导致 certbot（降权到部署用户）写入失败、ACME 静默失败、TLS 回退 HTTP。
# 仅 root 运行时才把属主 chown 回部署用户；
# 非 root + docker daemon 场景目录本就由当前用户创建，无需也无权 chown。
ensure_cert_dirs() {
  mkdir -p "$ROOT/deploy/letsencrypt/etc/live" \
    "$ROOT/deploy/letsencrypt/etc/archive" \
    "$ROOT/deploy/letsencrypt/etc/renewal" \
    "$ROOT/deploy/acme"
  if [ "$(id -u)" = "0" ]; then
    local du dg
    du="$(stat -c '%U' "$ROOT")"
    dg="$(stat -c '%G' "$ROOT")"
    chown -R "$du:$dg" "$ROOT/deploy/letsencrypt" "$ROOT/deploy/acme"
  fi
}
ensure_cert_dirs

# 变量 proxy_pass + resolver 实现上游动态解析，容器重建换 IP 也不 502；
# Docker 内置 DNS 为 127.0.0.11，Podman rootless 取网桥网关。
if [ "${COMPOSE[0]}" = "podman" ]; then
  NETNAME="blog_blog_net"
  podman network exists "$NETNAME" 2>/dev/null || podman network create "$NETNAME" >/dev/null 2>&1 || true
  NGINX_RESOLVER="$(podman network inspect "$NETNAME" -f '{{range .Subnets}}{{.Gateway}}{{end}}' 2>/dev/null | awk '{print $1}')"
  [ -z "$NGINX_RESOLVER" ] && NGINX_RESOLVER="10.89.0.1"
else
  NGINX_RESOLVER="127.0.0.11"
fi
export NGINX_RESOLVER

RUSTFS_ENABLED=""
if [ -z "${RUSTFS_ENDPOINT:-}" ]; then
  echo "ERROR: RUSTFS_ENDPOINT 未配置。前端需用 RustFS 承载图片等静态资源，该项为必填，" >&2
  echo "       请设为公网 https:// 地址（如 https://oos.immortel.top/）。" >&2
  exit 1
fi
RUSTFS_HOST="$(printf '%s' "$RUSTFS_ENDPOINT" | sed -E 's#^[a-zA-Z]+://##; s#/.*$##; s#:[0-9]+$##')"
if [ -n "$RUSTFS_HOST" ] && [ "$RUSTFS_HOST" != "localhost" ] && [ "$RUSTFS_HOST" != "127.0.0.1" ] && printf '%s' "$RUSTFS_HOST" | grep -q '\.'; then
  RUSTFS_SERVER_NAME="$RUSTFS_HOST"
  RUSTFS_ADMIN_SERVER_NAME="admin.${RUSTFS_HOST}"
  RUSTFS_ENABLED="1"
  export RUSTFS_SERVER_NAME RUSTFS_ADMIN_SERVER_NAME
else
  echo "ERROR: RUSTFS_ENDPOINT=$RUSTFS_ENDPOINT 不是公网域名，无法在 Nginx 暴露供前端访问。" >&2
  echo "       前端需通过公网 URL 加载图片，请设为公网 https:// 地址（如 https://oos.immortel.top/）。" >&2
  exit 1
fi

mkdir -p deploy/runtime/backend deploy/runtime/nginx
# shellcheck disable=SC2016
BACKEND_VARS='$APP_NAME $APP_DOMAIN $DB_HOST $DB_PORT $DB_USER $DB_PASSWORD $JWT_SECRET $EMAIL_HOST $EMAIL_PORT $EMAIL_USERNAME $EMAIL_PASSWORD $EMAIL_FROM $MODEL_API_KEY $RUSTFS_ACCESS_KEY_ID $RUSTFS_SECRET_ACCESS_KEY $RUSTFS_ENDPOINT'
envsubst "$BACKEND_VARS" <config/backend_config.toml >deploy/runtime/backend/config.toml
echo "已生成后端配置：deploy/runtime/backend/config.toml"

# RustFS 证书缺失时跳过其反代块，保证 blog 站点仍可启动。
gen_nginx_conf() {
  local mode="$1"
  envsubst "\${NGINX_SERVER_NAME} \${NGINX_RESOLVER}" <"deploy/nginx/${mode}.conf.template" >deploy/runtime/nginx/default.conf
  if [ -n "${RUSTFS_ENABLED:-}" ]; then
    if [ "$mode" = "https" ] &&
      { [ ! -f "deploy/letsencrypt/etc/live/${RUSTFS_SERVER_NAME}/fullchain.pem" ] ||
        [ ! -f "deploy/letsencrypt/etc/live/${RUSTFS_SERVER_NAME}/privkey.pem" ]; }; then
      echo "WARN: RustFS 证书 deploy/letsencrypt/etc/live/${RUSTFS_SERVER_NAME}/ 缺失，跳过 RustFS 反代配置（请先签发证书）。" >&2
    else
      envsubst "\${RUSTFS_SERVER_NAME} \${RUSTFS_ADMIN_SERVER_NAME} \${NGINX_RESOLVER}" \
        <"deploy/nginx/rustfs.${mode}.conf.template" >>deploy/runtime/nginx/default.conf
    fi
  fi
}

# 写入解析后的实际端口，避免 python 版 podman-compose 不继承 shell 变量而忽略 rootless 回退值。
mkdir -p deploy/runtime
cat >deploy/runtime/compose.ports.yml <<EOF
services:
  nginx:
    ports:
      - "${HTTP_PORT}:80"
      - "${HTTPS_PORT}:443"
EOF
BASE_FILES=(-f compose.yml -f deploy/runtime/compose.ports.yml)

cert_days_left_for() {
  local domain="$1"
  local cert="deploy/letsencrypt/etc/live/${domain}/fullchain.pem"
  [ -f "$cert" ] || {
    echo -1
    return
  }
  local enddate
  enddate=$(openssl x509 -in "$cert" -noout -enddate 2>/dev/null | cut -d= -f2) || {
    echo -1
    return
  }
  local end_ts now_ts
  end_ts=$(date -d "$enddate" +%s 2>/dev/null) || {
    echo -1
    return
  }
  now_ts=$(date +%s)
  echo $(((end_ts - now_ts) / 86400))
}

# 签发/续期单个证书；首个参数为证书主域名，其后为额外 SAN 域名。
ensure_cert() {
  local domain="$1"
  shift
  local live="deploy/letsencrypt/etc/live/$domain"
  local san_args=()
  local s
  for s in "$@"; do san_args+=(-d "$s"); done
  if [ -f "$live/fullchain.pem" ]; then
    echo "证书已存在，项目模式续期：./deploy/certbot.sh renew"
    ./deploy/certbot.sh renew
  else
    echo "未找到证书，自动签发 $domain ${san_args[*]:-}：./deploy/certbot.sh certonly ..."
    if [ ${#san_args[@]} -gt 0 ]; then
      ./deploy/certbot.sh certonly -d "$domain" "${san_args[@]}"
    else
      ./deploy/certbot.sh certonly -d "$domain"
    fi
  fi
}

# 等待 80 端口可达：Let's Encrypt 仅用 80 做 HTTP-01，不可达则签发必失败。
wait_acme_ready() {
  for _ in $(seq 1 30); do
    if (exec 3<>/dev/tcp/127.0.0.1/80) 2>/dev/null; then
      echo "ACME 挑战端口 80 已可达"
      return 0
    fi
    sleep 1
  done
  return 1
}

nginx_cert_up() {
  "${COMPOSE[@]}" "${BASE_FILES[@]}" up -d nginx
}
nginx_cert_rm() {
  if [ "${COMPOSE[0]}" = "podman" ]; then
    podman rm -f blog_nginx >/dev/null 2>&1 || true
  else
    docker rm -f blog_nginx >/dev/null 2>&1 || true
  fi
}

NGINX_TLS="${NGINX_TLS:-http}"
NGINX_TLS_REQ="$NGINX_TLS"
NGINX_TLS_AUTO="${NGINX_TLS_AUTO:-on}"
RENEW_DAYS="${RENEW_DAYS:-30}"

# 证书签发/续期先于正式部署；blog 与 RustFS 独立判断，缺失才签发、临期才续期，互不牵连。
if [ "$NGINX_TLS_REQ" = "https" ] && [ "$NGINX_TLS_AUTO" = "on" ]; then
  issue_list=""
  renew_needed=""
  if command -v openssl >/dev/null 2>&1; then
    blog_days=$(cert_days_left_for "$NGINX_SERVER_NAME")
    if [ "$blog_days" -lt 0 ]; then
      issue_list="$issue_list $NGINX_SERVER_NAME"
    elif [ "$blog_days" -le "$RENEW_DAYS" ]; then
      renew_needed="1"
    fi
    if [ -n "${RUSTFS_ENABLED:-}" ]; then
      rustfs_days=$(cert_days_left_for "$RUSTFS_SERVER_NAME")
      if [ "$rustfs_days" -lt 0 ]; then
        issue_list="$issue_list $RUSTFS_SERVER_NAME"
      elif [ "$rustfs_days" -le "$RENEW_DAYS" ]; then
        renew_needed="1"
      fi
    fi
  else
    issue_list="$issue_list $NGINX_SERVER_NAME"
    [ -n "${RUSTFS_ENABLED:-}" ] && issue_list="$issue_list $RUSTFS_SERVER_NAME"
  fi

  if [ -n "$issue_list" ]; then
    echo "存在缺失证书（${issue_list}），先以 HTTP 启动 nginx 提供 ACME 挑战，随后签发..."
    gen_nginx_conf http
    nginx_cert_up
    if wait_acme_ready; then
      for d in $issue_list; do
        if [ -e "deploy/letsencrypt/etc/live/$d" ] || [ -e "deploy/letsencrypt/etc/archive/$d" ] || [ -e "deploy/letsencrypt/etc/renewal/$d.conf" ]; then
          echo "ERROR: $d 存在不完整或冲突的 Certbot 状态；为保护证书未自动删除，请检查 deploy/letsencrypt/etc。" >&2
          exit 1
        fi
        if [ "$d" = "$NGINX_SERVER_NAME" ]; then
          ensure_cert "$NGINX_SERVER_NAME"
        else
          ensure_cert "$RUSTFS_SERVER_NAME" "$RUSTFS_ADMIN_SERVER_NAME"
        fi
      done
      if [ -n "$renew_needed" ]; then
        ./deploy/certbot.sh renew
      fi
    else
      echo "ERROR: 80 端口不可达，ACME HTTP-01 挑战无法完成，证书无法签发。" >&2
      echo "       请确认 rootless 下已设 sysctl net.ipv4.ip_unprivileged_port_start=0，" >&2
      echo "       或宿主已有 80→${HTTP_PORT} 的反向代理；并确认域名已解析到本机且 80 对公网开放。" >&2
    fi
    nginx_cert_rm
  elif [ -n "$renew_needed" ]; then
    echo "证书临近过期，执行续期（renew 仅续期已到窗口的证书）..."
    ./deploy/certbot.sh renew
  else
    echo "证书有效，跳过签发/续期。"
  fi
fi

# blog 证书最终未拿到时回退 HTTP。
if [ "$NGINX_TLS" = "https" ]; then
  CERT_DIR="deploy/letsencrypt/etc/live/${NGINX_SERVER_NAME}"
  if [ ! -f "$CERT_DIR/fullchain.pem" ] || [ ! -f "$CERT_DIR/privkey.pem" ]; then
    echo "WARN: blog 证书 $CERT_DIR 缺失，回退到 HTTP 模板（请检查 80 端口可达性与 certbot 安装）。" >&2
    NGINX_TLS="http"
  fi
fi

gen_nginx_conf "$NGINX_TLS"
if [ "$NGINX_TLS" = "https" ]; then
  echo "已生成 HTTPS nginx 配置"
else
  echo "已生成 HTTP nginx 配置"
fi

FINAL_FILES=("${BASE_FILES[@]}")
if [ "$NGINX_TLS" = "https" ]; then
  FINAL_FILES+=(-f compose.https.yml)
  echo "HTTPS 模式：确保 ACME 挑战目录存在..."
  mkdir -p deploy/acme 2>/dev/null ||
    echo "警告：无法创建 deploy/acme，certbot 续期时将无法写入挑战文件（详见 README）。"
fi

# podman-compose 1.0.6 对 depends_on 条件（service_healthy / service_completed_successfully）处理不可靠：
# 并打印大量 “no container / invalid dependency” 噪声。故显式按依赖顺序拉起并轮询就绪，与 remote-deploy.sh 一致。
ENGINE="${COMPOSE[0]}"

# 过滤 podman-compose 1.0.6 噪声（merged JSON / recreating / no container 等），仅留真实报错与最终状态。
pod_compose_filter() {
  awk '
    /^>>>> Executing external compose provider/ { next }
    /^podman-compose version:/ { next }
    /^using podman version:/ { next }
    /^[][]podman, (ps|volume|network|--version)/ { next }
    /^[[:space:]]*\*\* merged:/ { in_json=1; depth=0; next }
    in_json {
      depth += gsub(/{/, "{") - gsub(/}/, "}")
      if (depth <= 0) in_json=0
      next
    }
    /^recreating:/ { next }
    /^\*\* (excluding|skipping):/ { next }
    /^exit code:/ { next }
    /^podman (stop|rm|start|volume|network|inspect|create|run|ps) / { next }
    /^Error: no (container with|such container)/ { next }
    /^Error: container .* has dependent containers/ { next }
    /^Error: creating container storage: the container name .* is already in use/ { next }
    { print }
  '
}
pod_compose_out_filter() {
  awk '/^[0-9a-f]{12,64}$/ { next } /^blog_[a-z_]+$/ { next } { print }'
}
pod_compose() {
  "${COMPOSE[@]}" "${FINAL_FILES[@]}" "$@" \
    > >(pod_compose_out_filter) \
    2> >(pod_compose_filter >&2)
}

dump_container_logs() {
  local c="$1"
  echo "------ $c 日志（最后 200 行）------" >&2
  "$ENGINE" logs --tail 200 "$c" >&2 2>/dev/null || echo "(无法读取 $c 日志)" >&2
  echo "------------------------------------" >&2
}

wait_postgres() {
  for _ in $(seq 1 60); do
    if "$ENGINE" exec blog_postgres pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" >/dev/null 2>&1; then
      echo "postgres 就绪"
      return 0
    fi
    sleep 2
  done
  echo "ERROR: 等待 postgres 就绪超时（60×2s）" >&2
  return 1
}

# 先删应用层容器，规避 podman-compose 重建时 name already in use / 残留 Created 容器。
"$ENGINE" rm -f blog_nginx blog_frontend blog_backend >/dev/null 2>&1 || true

echo "构建镜像..."
pod_compose build || true
echo "启动 postgres / redis"
pod_compose up -d postgres redis || true
wait_postgres || exit 1
echo "启动 backend / frontend"
pod_compose up -d backend frontend || true
echo "启动 nginx"
pod_compose up -d nginx || true
# reload 应用刚生成的 resolver 配置，避免后端重建换 IP 后 nginx 缓存旧地址导致 502。
pod_compose exec nginx nginx -s reload >/dev/null 2>&1 || true

echo ""
echo "部署完成！"
echo "前端访问：http://${NGINX_SERVER_NAME:-localhost}"
echo "后端 API：http://${NGINX_SERVER_NAME:-localhost}/api/v1/"
echo "RustFS API：https://${RUSTFS_SERVER_NAME}"
echo "RustFS 控制台：https://${RUSTFS_ADMIN_SERVER_NAME}"
