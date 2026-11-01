#!/usr/bin/env bash
# certbot wrapper. Project mode is the default and keeps all Certbot state in
# deploy/letsencrypt; system mode uses the host Certbot directories.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DEPLOY_USER="$(stat -c '%U' "$PROJECT_DIR")"
DEPLOY_GROUP="$(stat -c '%G' "$PROJECT_DIR")"
CERT_MODE="${CERT_MODE:-project}"

case "$CERT_MODE" in
project)
  CERT_BASE="$PROJECT_DIR/deploy/letsencrypt"
  CERT_CONFIG="$CERT_BASE/etc"
  CERT_WORK="$CERT_BASE/work"
  CERT_LOGS="$CERT_BASE/logs"
  ACME_WEBROOT="$PROJECT_DIR/deploy/acme"
  CERTBOT_ARGS=(--config-dir "$CERT_CONFIG" --work-dir "$CERT_WORK" --logs-dir "$CERT_LOGS")
  ;;
system)
  CERT_BASE="/var/lib/letsencrypt"
  CERT_CONFIG="/etc/letsencrypt"
  CERT_WORK="/var/lib/letsencrypt"
  CERT_LOGS="/var/log/letsencrypt"
  ACME_WEBROOT="/var/www/certbot"
  CERTBOT_ARGS=()
  ;;
*)
  echo "ERROR: CERT_MODE 必须是 project 或 system" >&2
  exit 2
  ;;
esac

CERTBOT_EMAIL="${CERTBOT_EMAIL:-${EMAIL_FROM:-}}"

has_arg() {
  local arg
  for arg in "${USER_ARGS[@]}"; do
    [ "$arg" = "$1" ] && return 0
  done
  return 1
}

has_option_with_value() {
  local i arg
  for ((i = 0; i < ${#USER_ARGS[@]}; i++)); do
    arg="${USER_ARGS[i]}"
    case "$arg" in
    "$1"|"$2") return 0 ;;
    "$1"=*|"$2"=*) return 0 ;;
    esac
  done
  return 1
}

reject_storage_overrides() {
  local arg
  for arg in "${USER_ARGS[@]}"; do
    case "$arg" in
    --config-dir|--config-dir=*|--work-dir|--work-dir=*|--logs-dir|--logs-dir=*)
      echo "ERROR: 请不要覆盖 CERT_MODE=$CERT_MODE 的 Certbot 存储路径；wrapper 已固定项目/系统目录。" >&2
      exit 2
      ;;
    esac
  done
}

validate_webroot() {
  local i arg path=""
  for ((i = 0; i < ${#USER_ARGS[@]}; i++)); do
    arg="${USER_ARGS[i]}"
    case "$arg" in
    -w|--webroot-path)
      path="${USER_ARGS[i + 1]:-}"
      ;;
    -w=*|--webroot-path=*)
      path="${arg#*=}"
      ;;
    esac
    [ -n "$path" ] && break
  done
  [ -n "$path" ] || return 0
  case "$path" in
  /*) ;;
  *) echo "ERROR: Webroot 必须是宿主机绝对路径：$path" >&2; exit 2 ;;
  esac
  if [ "$CERT_MODE" = project ] && [ "$path" = /var/www/certbot ]; then
    echo "ERROR: project 模式不能使用容器内路径 /var/www/certbot；请使用 $ACME_WEBROOT 或显式传入宿主目录。" >&2
    exit 2
  fi
  [ -d "$path" ] || {
    echo "ERROR: Webroot 目录不存在：$path" >&2
    exit 2
  }
}

certbot_as_deploy_user() {
  if [ "$CERT_MODE" = system ]; then
    exec certbot "${CERTBOT_ARGS[@]}" "${USER_ARGS[@]}"
  fi
  if [ "$(id -u)" = 0 ]; then
    command -v sudo >/dev/null 2>&1 || {
      echo "ERROR: root 执行项目模式需要 sudo，以降权到 $DEPLOY_USER。" >&2
      exit 1
    }
    local home
    home="$(getent passwd "$DEPLOY_USER" | cut -d: -f6)"
    [ -n "$home" ] || {
      echo "ERROR: 无法解析部署用户 $DEPLOY_USER 的 HOME。" >&2
      exit 1
    }
    exec sudo -u "$DEPLOY_USER" env HOME="$home" CERTBOT_EMAIL="$CERTBOT_EMAIL" \
      certbot "${CERTBOT_ARGS[@]}" "${USER_ARGS[@]}"
  fi
  exec certbot "${CERTBOT_ARGS[@]}" "${USER_ARGS[@]}"
}

backup_existing_state() {
  local backup
  backup="$CERT_BASE/import-backup-$(date +%Y%m%d%H%M%S)"
  mkdir -p "$backup"
  cp -a "$CERT_CONFIG" "$backup/etc"
  echo "已备份现有项目 Certbot 状态：$backup" >&2
}

ini_value() {
  awk -F '=' -v key="$1" '
    $1 ~ "^[[:space:]]*" key "[[:space:]]*$" { sub(/^[[:space:]]*/, "", $2); sub(/[[:space:]]*$/, "", $2); print $2; exit }
  ' "$2"
}

validate_lineage() {
  local root="$1" domain="$2"
  local cert="$root/live/$domain/cert.pem"
  local fullchain="$root/live/$domain/fullchain.pem"
  local key="$root/live/$domain/privkey.pem"
  local archive="$root/archive/$domain"
  [ -d "$archive" ] || { echo "ERROR: 导入结果缺少 archive/$domain" >&2; return 1; }
  for f in cert.pem chain.pem fullchain.pem privkey.pem; do
    [ -L "$root/live/$domain/$f" ] || { echo "ERROR: live/$domain/$f 不是相对 symlink" >&2; return 1; }
    [ -e "$root/live/$domain/$f" ] || { echo "ERROR: live/$domain/$f 是断链" >&2; return 1; }
    case "$(readlink "$root/live/$domain/$f")" in
      ../archive/*) ;;
      *) echo "ERROR: live/$domain/$f 未指向相对 archive 路径" >&2; return 1 ;;
    esac
  done
  openssl x509 -in "$cert" -noout >/dev/null 2>&1 || { echo "ERROR: 导入证书无效" >&2; return 1; }
  openssl x509 -in "$fullchain" -noout >/dev/null 2>&1 || { echo "ERROR: 导入 fullchain 无效" >&2; return 1; }
  openssl pkey -in "$key" -pubout -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1 >"$root/.key.pub.hash"
  openssl x509 -in "$cert" -pubkey -noout 2>/dev/null | openssl pkey -pubin -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1 >"$root/.cert.pub.hash"
  cmp -s "$root/.key.pub.hash" "$root/.cert.pub.hash" || { echo "ERROR: 证书与私钥不匹配" >&2; return 1; }
  rm -f "$root/.key.pub.hash" "$root/.cert.pub.hash"
}

rewrite_renewal() {
  local src="$1" dst="$2" domain="$3" account="$4"
  awk -v archive="$dst/archive/$domain" \
      -v cert="$dst/live/$domain/cert.pem" \
      -v key="$dst/live/$domain/privkey.pem" \
      -v chain="$dst/live/$domain/chain.pem" \
      -v fullchain="$dst/live/$domain/fullchain.pem" \
      -v webroot="$PROJECT_DIR/deploy/acme" \
      -v account="$account" '
    /^[[:space:]]*archive_dir[[:space:]]*=/ { print "archive_dir = " archive; next }
    /^[[:space:]]*cert[[:space:]]*=/ { print "cert = " cert; next }
    /^[[:space:]]*privkey[[:space:]]*=/ { print "privkey = " key; next }
    /^[[:space:]]*chain[[:space:]]*=/ { print "chain = " chain; next }
    /^[[:space:]]*fullchain[[:space:]]*=/ { print "fullchain = " fullchain; next }
    /^[[:space:]]*account[[:space:]]*=/ { print "account = " account; next }
    /^[[:space:]]*webroot_path[[:space:]]*=/ { print "webroot_path = " webroot; next }
    /^[[:space:]]*webroot_map[[:space:]]*=/ {
      line = $0
      gsub(/:[[:space:]]*["'"'][^"'"']*["'"']/, ": \"" webroot "\"", line)
      print line
      next
    }
    { print }
  ' "$src" >"$dst/renewal/$domain.conf"
}

import_certificate() {
  local domain="${USER_ARGS[1]:-}"
  [ -n "$domain" ] || { echo "ERROR: 用法: $0 import <域名>" >&2; exit 2; }
  [ "$CERT_MODE" = project ] || { echo "ERROR: CERT_MODE=system 时不能 import；系统证书已在 /etc/letsencrypt。" >&2; exit 2; }
  [ "$(id -u)" = 0 ] || { echo "ERROR: import 需要 root 读取系统账号并保留私钥权限，请使用 sudo。" >&2; exit 1; }

  local src="/etc/letsencrypt" src_live="$src/live/$domain" src_arch="$src/archive/$domain"
  local src_renewal="$src/renewal/$domain.conf" account staging account_dir account_rel
  [ -d "$src_live" ] || { echo "ERROR: 系统 lineage 不存在：$src_live" >&2; exit 1; }
  [ -d "$src_arch" ] || { echo "ERROR: 系统 archive 不存在：$src_arch" >&2; exit 1; }
  [ -f "$src_renewal" ] || { echo "ERROR: 缺少系统 renewal 配置：$src_renewal；无法安全确定账号和参数。" >&2; exit 1; }
  account="$(ini_value account "$src_renewal")"
  [ -n "$account" ] || { echo "ERROR: $src_renewal 缺少 account，拒绝猜测账号关系。" >&2; exit 1; }
  mapfile -t account_matches < <(find "$src/accounts" -mindepth 1 -type d -name "$account" -print)
  [ "${#account_matches[@]}" -eq 1 ] || {
    echo "ERROR: 系统账号 $account 未能唯一定位（匹配 ${#account_matches[@]} 个目录），拒绝生成不完整 renewal。" >&2
    exit 1
  }
  account_dir="${account_matches[0]}"
  account_rel="${account_dir#"$src/accounts/"}"

  mkdir -p "$CERT_CONFIG" "$CERT_BASE"
  staging="$(mktemp -d "$CERT_BASE/.import-$domain.XXXXXX")"
  trap 'rm -rf "$staging"' RETURN
  mkdir -p "$staging/etc/live" "$staging/etc/archive" "$staging/etc/renewal" "$staging/etc/accounts/$(dirname "$account_rel")"
  cp -a "$account_dir" "$staging/etc/accounts/$account_rel"
  [ -d "$src/renewal-hooks" ] && cp -a "$src/renewal-hooks" "$staging/etc/"
  for metadata in options-ssl-nginx.conf ssl-dhparams.pem; do
    [ -e "$src/$metadata" ] && cp -a "$src/$metadata" "$staging/etc/"
  done
  cp -a "$src_arch" "$staging/etc/archive/"
  cp -a "$src_live" "$staging/etc/live/"
  rewrite_renewal "$src_renewal" "$staging/etc" "$domain" "$account"
  validate_lineage "$staging/etc" "$domain"
  grep -q "^account[[:space:]]*=[[:space:]]*$account[[:space:]]*$" "$staging/etc/renewal/$domain.conf" || {
    echo "ERROR: 导入 renewal 的账号校验失败" >&2; exit 1;
  }
  if ! grep -q "^webroot_path[[:space:]]*=[[:space:]]*$PROJECT_DIR/deploy/acme[[:space:]]*$" "$staging/etc/renewal/$domain.conf" &&
    ! grep -q "webroot_map.*$PROJECT_DIR/deploy/acme" "$staging/etc/renewal/$domain.conf"; then
    echo "ERROR: 导入 renewal 的 Webroot 校验失败" >&2
    exit 1
  fi

  if [ -e "$CERT_CONFIG/live/$domain" ] || [ -e "$CERT_CONFIG/archive/$domain" ] || [ -e "$CERT_CONFIG/renewal/$domain.conf" ]; then
    backup_existing_state
    echo "ERROR: 项目中已存在 $domain 的数据；为避免覆盖，导入已停止。请检查备份后手工比较。" >&2
    exit 1
  fi
  if [ -e "$CERT_CONFIG/accounts/$account_rel" ] && ! diff -qr "$account_dir" "$CERT_CONFIG/accounts/$account_rel" >/dev/null; then
    backup_existing_state
    echo "ERROR: 项目中已存在不同内容的 ACME 账号 $account；为避免覆盖账号密钥，导入已停止。" >&2
    exit 1
  fi
  cp -a "$staging/etc/." "$CERT_CONFIG/"
  chown -R "$DEPLOY_USER:$DEPLOY_GROUP" "$CERT_BASE"
  echo "已安全导入 $domain，保留 ACME 账号 $account、历史 archive、相对 live symlink 与 renewal 参数。"
}

cmd="${1:-}"
USER_ARGS=("$@")
if [ "$cmd" = import ]; then
  import_certificate
  exit 0
fi
reject_storage_overrides

if [ "$cmd" = certonly ]; then
  challenge_method=0
  for arg in "${USER_ARGS[@]}"; do
    case "$arg" in
    --webroot|--webroot-path|--webroot-path=*|-w|--standalone|--manual|--dns-*|--http-01-port|--tls-sni-01-port|--tls-sni-01-address)
      challenge_method=1
      ;;
    esac
  done
  if [ "$challenge_method" = 0 ]; then
    USER_ARGS=(certonly --webroot -w "$ACME_WEBROOT" "${USER_ARGS[@]:1}")
  elif has_arg --webroot && ! has_option_with_value -w --webroot-path; then
    USER_ARGS=(certonly "${USER_ARGS[@]:1}" -w "$ACME_WEBROOT")
  fi
  if [ -n "$CERTBOT_EMAIL" ] && ! has_option_with_value --email -m; then
    USER_ARGS=("${USER_ARGS[0]}" --agree-tos --no-eff-email --non-interactive --email "$CERTBOT_EMAIL" "${USER_ARGS[@]:1}")
  elif ! has_arg --agree-tos; then
    USER_ARGS=("${USER_ARGS[0]}" --agree-tos --non-interactive "${USER_ARGS[@]:1}")
  fi
fi

if [ "$CERT_MODE" = system ] && [ "$(id -u)" != 0 ]; then
  echo "ERROR: CERT_MODE=system 需要 root；请使用 sudo，或使用默认的 project 模式。" >&2
  exit 1
fi
mkdir -p "$ACME_WEBROOT"
validate_webroot
if [ "$CERT_MODE" = project ]; then
  mkdir -p "$CERT_CONFIG" "$CERT_WORK" "$CERT_LOGS"
  chown -R "$DEPLOY_USER:$DEPLOY_GROUP" "$CERT_BASE" "$ACME_WEBROOT" 2>/dev/null || true
fi
certbot_as_deploy_user
