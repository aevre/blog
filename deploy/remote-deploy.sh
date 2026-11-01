#!/usr/bin/env bash
# 服务器侧部署脚本（供 CI actions 调用），逻辑与 deploy.sh 一致。
# 用法：deploy/remote-deploy.sh <backend|frontend|all>
set -euo pipefail

TARGET="${1:-all}"
cd "$HOME/projects/blog"

if [[ "$TARGET" == "backend" || "$TARGET" == "all" ]] && [ -f backend.tar ]; then
  echo "Loading backend image..."
  podman load -i backend.tar \
    2> >(awk '/^(Getting image source signatures|Copying blob |Copying config |Writing manifest to image destination)/ { print; next } { print > "/dev/stderr" }') && rm -f backend.tar
fi
if [[ "$TARGET" == "frontend" || "$TARGET" == "all" ]] && [ -f frontend.tar ]; then
  echo "Loading frontend image..."
  podman load -i frontend.tar \
    2> >(awk '/^(Getting image source signatures|Copying blob |Copying config |Writing manifest to image destination)/ { print; next } { print > "/dev/stderr" }') && rm -f frontend.tar
fi

if [ ! -s .env ]; then
  echo "ERROR: $PWD/.env 不存在或为空。请先在仓库 Settings → Secrets 配置 SERVER_ENV，" >&2
  echo "由 setup-env action 通过 SSH 写入服务器；或手动 cp .env.example .env 后填写。" >&2
  exit 1
fi
set -a
# .env 可能带 CRLF，尾部 \r 会让整数端口等字段解析失败
# shellcheck source=/dev/null
source <(sed 's/\r$//' .env)
set +a

if [ -z "${DB_PORT:-}" ]; then
  echo "WARN: .env 缺少 DB_PORT，使用默认 5432（请补全 SERVER_ENV 后重新生成 .env）" >&2
  DB_PORT=5432
  export DB_PORT
fi
if [ -z "${EMAIL_PORT:-}" ]; then
  echo "WARN: .env 缺少 EMAIL_PORT，使用默认 587（请补全 SERVER_ENV 后重新生成 .env）" >&2
  EMAIL_PORT=587
  export EMAIL_PORT
fi

# Let's Encrypt HTTP-01 只在 80 端口验证，故 80 必须可达才能自动签发。
# rootless 默认无法绑定 <1024，优先 80:80，仅在无法绑定 80 时回落 8080/8443。
HTTP_PORT="${HTTP_PORT:-80}"
HTTPS_PORT="${HTTPS_PORT:-443}"

# 本脚本是仓库的远程 CI 部署脚本 rootless，用户 fork 后亦可在自己的 GitHub Actions 中复用。
# 端口逻辑：检测到 rootless Podman 时把 80/443 回落 8080/8443；以 root 运行时则直接绑 80/443。
ROOTLESS="false"
if command -v podman >/dev/null 2>&1; then
  if podman info --format '{{.Host.Security.Rootless}}' 2>/dev/null | grep -qi true; then
    ROOTLESS="true"
  fi
fi

# rootless 能否绑 80 取决于 net.ipv4.ip_unprivileged_port_start <= 80
port_80_bindable() {
  [ "$ROOTLESS" = "false" ] && return 0
  local start
  start=$(sysctl -n net.ipv4.ip_unprivileged_port_start 2>/dev/null || echo 1024)
  [ "${start:-1024}" -le 80 ] 2>/dev/null
}

if [ "$ROOTLESS" = "true" ] && [ "$HTTP_PORT" = "80" ] && ! port_80_bindable; then
  HTTP_PORT=8080
  HTTPS_PORT=8443
  echo "检测到 rootless Podman 且 80 端口不可绑定（net.ipv4.ip_unprivileged_port_start>80）。"
  echo "已回落到 ${HTTP_PORT}/${HTTPS_PORT}。注意：HTTPS 自动签发需要 80 端口可达，请满足其一："
  echo "  1) 宿主执行：sysctl -w net.ipv4.ip_unprivileged_port_start=0（并写入 /etc/sysctl.d/ 持久化）；"
  echo "  2) 或在宿主另起监听 80/443 的反向代理转发到本机 ${HTTP_PORT}/${HTTPS_PORT}。"
fi
export HTTP_PORT HTTPS_PORT

# 端口覆盖文件写字面量端口并用 -f 挂上，因为 podman-compose 会优先读项目目录 .env 覆盖 shell 里 export 的回退值，
# 只有 -f 覆盖文件里的字面量才生效。
mkdir -p deploy/runtime
cat >deploy/runtime/compose.ports.yml <<EOF
services:
  nginx:
    ports:
      - "${HTTP_PORT}:80"
      - "${HTTPS_PORT}:443"
EOF
COMPOSE_FILES=(-f compose.prod.yml -f deploy/runtime/compose.ports.yml)
export HTTP_PORT HTTPS_PORT
echo "部署参数: target=$TARGET http=$HTTP_PORT https=$HTTPS_PORT rootless=$ROOTLESS"

# podman-compose 1.0.6 会把 merged config、provider 横幅、recreating 自愈错误等数百行噪声刷到 stderr，
# 统一过滤，仅保留真实报错与最终状态，便于 CI 日志定位。
pod_compose_filter() {
  awk '
    /^[[:space:]]*>>>> Executing external compose provider/ { next }
    /^[[:space:]]*podman-compose version:/ { next }
    /^[[:space:]]*using podman version:/ { next }
    /^\[.podman., .(ps|volume|network|--version)/ { next }
    # merged config 是一整段 JSON，用花括号深度判断，直到回到顶层 } 才结束，
    # 避免把内部嵌套的 } 误判为结束。
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
  podman compose "${COMPOSE_FILES[@]}" "$@" \
    > >(pod_compose_out_filter) \
    2> >(pod_compose_filter >&2)
}

# RustFS 证书缺失时跳过其反代块，保证 blog 站点仍可启动。
render_nginx_conf() {
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

# 容器异常退出时打印其日志，避免只看到退出码 1 而看不到真实报错
dump_container_logs() {
  local c="$1"
  echo "------ $c 日志（最后 200 行）------" >&2
  podman logs --tail 200 "$c" >&2 2>/dev/null || echo "(无法读取 $c 日志)" >&2
  echo "------------------------------------" >&2
}

NETNAME="blog_blog_net"
podman network exists "$NETNAME" 2>/dev/null || podman network create "$NETNAME" >/dev/null 2>&1 || true
NGINX_RESOLVER="$(podman network inspect "$NETNAME" -f '{{range .Subnets}}{{.Gateway}}{{end}}' 2>/dev/null | awk '{print $1}')"
[ -z "$NGINX_RESOLVER" ] && NGINX_RESOLVER="10.89.0.1"
export NGINX_RESOLVER

# RustFS 必填：前端用它承载图片等静态资源。
# 仅当 endpoint 为公网域名时启用容器内反代；服务名仅供后端内网访问。
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
echo "渲染后端配置 → deploy/runtime/backend/config.toml"
envsubst "$BACKEND_VARS" <config/backend_config.toml >deploy/runtime/backend/config.toml

# 过滤 Certbot 的正常日志路径提示；真正的错误仍保留在 stderr。
certbot_run() {
  "$@" 2> >(sed '/^[[:space:]]*Saving debug log to /d' >&2)
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
    certbot_run ./deploy/certbot.sh renew
  else
    echo "未找到证书，自动签发 $domain ${san_args[*]:-}：./deploy/certbot.sh certonly ..."
    if [ ${#san_args[@]} -gt 0 ]; then
      certbot_run ./deploy/certbot.sh certonly -d "$domain" "${san_args[@]}"
    else
      certbot_run ./deploy/certbot.sh certonly -d "$domain"
    fi
  fi
}

# 等待 ACME 挑战端口 80 可达，Let's Encrypt 只在 80 做 HTTP-01，不可达则签发必然失败。
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

NGINX_TLS="${NGINX_TLS:-http}"
NGINX_TLS_REQ="$NGINX_TLS"
NGINX_TLS_AUTO="${NGINX_TLS_AUTO:-on}"
RENEW_DAYS="${RENEW_DAYS:-30}"

# 读取指定域名证书剩余有效天数，缺失或无法解析返回 -1
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

# 证书自动签发/续期先于正式部署：blog 与 RustFS 各自独立判断。
# 仅对确实缺失/无效的证书做首次签发，仅对临近过期的证书续期，互不牵连。
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
    render_nginx_conf http
    pod_compose up -d nginx
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
        certbot_run ./deploy/certbot.sh renew
      fi
    else
      echo "ERROR: 80 端口不可达，ACME HTTP-01 挑战无法完成，证书无法签发。" >&2
      echo "       请确认 rootless 下已设 sysctl net.ipv4.ip_unprivileged_port_start=0，" >&2
      echo "       或宿主已有 80→${HTTP_PORT} 的反向代理；并确认域名已解析到本机且 80 对公网开放。" >&2
    fi
    podman rm -f blog_nginx >/dev/null 2>&1 || true
  elif [ -n "$renew_needed" ]; then
    echo "证书临近过期，执行续期（renew 仅续期已到窗口的证书）..."
    certbot_run ./deploy/certbot.sh renew
  else
    echo "证书有效，跳过签发/续期。"
  fi
fi

if [ "$NGINX_TLS_REQ" = "https" ]; then
  CERT_DIR="deploy/letsencrypt/etc/live/${NGINX_SERVER_NAME}"
  if [ ! -f "$CERT_DIR/fullchain.pem" ] || [ ! -f "$CERT_DIR/privkey.pem" ]; then
    if [ "$NGINX_TLS_AUTO" = "on" ]; then
      echo "WARN: 自动签发未产出证书，回退到 HTTP 模板。" >&2
      echo "WARN: 请检查 80 端口可达性：rootless 需 sysctl net.ipv4.ip_unprivileged_port_start=0，" >&2
      echo "WARN: 或宿主已有 80→${HTTP_PORT} 的反向代理；并确认 certbot 已安装、域名已解析到本机。" >&2
    else
      cat >&2 <<EOF
WARN: NGINX_TLS=https 但证书 $CERT_DIR 不存在，临时回退到 HTTP 模板。
      请用 deploy/certbot.sh 以部署用户身份签发/导入证书（证书归部署用户所有，容器可直接读取）：
        ./deploy/certbot.sh certonly -d ${NGINX_SERVER_NAME}
        # 或把宿主 /etc/letsencrypt 中已有的证书一次性导入项目目录：
        sudo ./deploy/certbot.sh import ${NGINX_SERVER_NAME}
EOF
    fi
    NGINX_TLS="http"
  fi
fi
render_nginx_conf "$NGINX_TLS"
if [ "$NGINX_TLS" = "https" ]; then
  echo "已生成 HTTPS nginx 配置"
else
  echo "已生成 HTTP nginx 配置"
fi

wait_postgres() {
  for _ in $(seq 1 60); do
    if podman exec blog_postgres pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" >/dev/null 2>&1; then
      echo "postgres 就绪"
      return 0
    fi
    sleep 2
  done
  echo "ERROR: 等待 postgres 就绪超时（60×2s）" >&2
  return 1
}

# 显式删容器再 up，规避 podman-compose --force-recreate 在依赖容器存在时 rm 失败、旧容器残留、run 报 name already in use。
case "$TARGET" in
backend)
  # 旧版 frontend 可能仍依赖 backend 而阻塞其删除，一并删 frontend 自愈一次
  RF=0
  if ! podman rm -f blog_backend >/dev/null 2>&1; then
    podman rm -f blog_frontend >/dev/null 2>&1 || true
    podman rm -f blog_backend >/dev/null 2>&1 || true
    RF=1
  fi
  podman rm -f blog_nginx >/dev/null 2>&1 || true
  # 先起基础设施并跑迁移，规避 podman-compose 1.0.6 的 healthy 竞态
  echo "启动 postgres / redis"
  pod_compose up -d postgres redis
  wait_postgres || exit 1
  # 迁移完成后再起 backend，podman-compose 1.0.6 会把 backend 留在 Created 不启动
  echo "启动 backend"
  pod_compose up -d backend
  [ "$RF" = "1" ] && {
    echo "启动 frontend（依赖自愈）"
    pod_compose up -d frontend
  } || true
  ;;
frontend)
  podman rm -f blog_nginx blog_frontend >/dev/null 2>&1 || true
  echo "启动 backend / frontend"
  pod_compose up -d backend || true
  pod_compose up -d frontend
  ;;
all)
  podman rm -f blog_nginx blog_frontend blog_backend >/dev/null 2>&1 || true
  echo "启动 postgres / redis"
  pod_compose up -d postgres redis
  wait_postgres || exit 1
  echo "启动 backend / frontend"
  pod_compose up -d backend frontend
  ;;
esac

# 启动任何停在 Created 的容器作为兜底
for c in blog_backend blog_frontend blog_nginx; do
  st=$(podman inspect -f '{{.State.Status}}' "$c" 2>/dev/null || true)
  if [ "$st" = "created" ]; then
    podman start "$c" >/dev/null 2>&1 || true
  fi
done

echo "拉起反向代理 nginx 并 reload"
pod_compose up -d nginx
pod_compose exec nginx nginx -s reload >/dev/null 2>&1

echo "Deployed: $TARGET"
