[English](self-hosting.md) · [中文](self-hosting.zh.md)

# Self-hosting (packaged artifacts + deployment steps)

For users who do not want the one-command automation and want full control over how each module runs, use the "self-hosting" mode:

```bash
./deploy.sh self
```

The script builds and packages the artifacts; it does not run any service and does not issue certificates automatically. How to run them is up to you on the target machine.

## What gets packaged

`blog-selfhost-<date>.tar.gz` contains:

- `backend/bin/blog-server`, `backend/bin/migration` — backend binaries (static build with `CGO_ENABLED=0`)
- `backend/bin/config.toml` — backend config already generated from the local `.env` (same level as the binary)
- `frontend/` — Nuxt SSR artifacts
- `config/backend_config.toml` — backend config template
- `deploy/nginx/*.template` — Nginx config templates
- `.env.example` — variable documentation

> Note: the package does not include the `.md` docs; the docs are the ones in this file and the other `docs/` files in the repo.

## Choosing where to build

`./deploy.sh self` needs to run on a machine with the full toolchain (see "Server build environment requirements" below). There are two options:

- **A. Build locally then upload (recommended)**: run `./deploy.sh self` on your dev machine / CI, then transfer the generated `blog-selfhost-*.tar.gz` to the target server and extract it. The target server does not need Go / pnpm (the build toolchain), but the frontend Nuxt SSR still needs Node.js to run; otherwise it only needs to be able to extract the tar and run the binaries / start containers (or use already-present services like Nginx, PostgreSQL, etc.).
- **B. Build directly on the server**: on the target server, clone the repo and run `./deploy.sh self` (or manually `go build` / `pnpm build`). In this case the target server must have the full toolchain listed under "Server build environment requirements" below.

The package produced by both methods is identical; the only difference is "which machine the build happens on." In most cases A is more convenient and safer.

## Server build environment requirements

If you choose to build on the target server, you need:

- **Go** `1.27` (see `backend/go.mod`; `deploy.sh self` uses `CGO_ENABLED=0 go build`, no C toolchain needed)
- **Node.js** `24` (see `frontend/Dockerfile`: `node:24`) and **pnpm** `11.15.1` (see `package.json`'s `packageManager`)
- **envsubst** (provided by `gettext`; used on the server side to generate `config/backend_config.toml` from `.env`)
- **git** (to clone the repo)
- Disk and time: the first `pnpm install` + `pnpm build` and `go build` pull dependencies and compile, so reserve enough space and a few minutes.

For method A, the target server does not need the above build toolchain (Go / pnpm / git / envsubst); it only needs to run the artifacts and their dependent services; the frontend Nuxt SSR still requires Node.js to run.

## Deployment steps (target machine)

Below are only the steps, not executed for you. How each module runs is your decision: containers (rootless / rootful, your choice), native (system packages / binaries / systemd), or existing instances (e.g. a managed PostgreSQL).

### 0. Target machine prerequisites

- PostgreSQL (create database `blog`, user, password), Redis
- Optional RustFS (object storage, for frontend images) or any S3-compatible store
- Nginx (reverse proxy / TLS termination), certbot (if you want HTTPS)

> **You must do the DNS resolution for the domain yourself**: the `.env` values `NGINX_SERVER_NAME` / `APP_DOMAIN` / `RUSTFS_ENDPOINT` in the package are only for Nginx / backend identification and forwarding config — they do not take effect automatically. Point the A/AAAA records to the target machine's public IP at your registrar (same for the RustFS domain); otherwise the site is unreachable from outside and certificates cannot be issued.

### 1. Place the artifacts

Extract the package to some directory on the target machine (e.g. `/srv/blog`). The backend points to `backend/bin/config.toml` via the `CONFIG_FILE` environment variable (same level as `blog-server`; or generate it yourself from `config/backend_config.toml`).

### 2. Backend

First `./backend/bin/migration`, then
  `CONFIG_FILE=/srv/blog/backend/bin/config.toml ./backend/bin/blog-server`

### 3. Frontend

Nuxt SSR runs with `node`; the target machine needs Node.js installed (same version as the build, see `frontend/Dockerfile`'s `node:24`).

`cd /srv/blog/frontend && NUXT_BACKEND_URL=http://127.0.0.1:8000 node server/index.mjs`

### 4. Database / cache / object storage

- Use existing instances: fill `DB_HOST`, `REDIS_HOST`, `RUSTFS_ENDPOINT` directly in `config.toml` / `.env`.
- Use containers: spin up separate postgres / redis / rustfs containers; the backend connects via service name or `127.0.0.1`.

### 5. Nginx reverse proxy + TLS

- The templates `deploy/nginx/http.conf.template`, `https.conf.template` already include the `location /.well-known/acme-challenge/` (ACME challenge) and 443 TLS.
- Change the upstream to the local `127.0.0.1:3000 / 127.0.0.1:8000` (the templates use container service names — replace them yourself), put the certificate at `/etc/letsencrypt/live/<domain>/`, write it into `/etc/nginx/conf.d/blog.conf`, then `nginx -t && systemctl reload nginx`.

### 6. Certificate (choose one of three)

- Automatic issuance: `CERT_MODE=system ./deploy/certbot.sh certonly -d <domain>` (system mode uses the host `/var/www/certbot`)
- Existing certificate: put `fullchain.pem` / `privkey.pem` at the corresponding paths
- With the script: `CERT_MODE=system ./deploy/certbot.sh certonly -d <domain>` (system `/etc/letsencrypt` mode)

### 7. Startup order

1. PostgreSQL / Redis / RustFS ready
2. Run migration
3. Start backend (8000), frontend (3000)
4. Start nginx (80/443); HTTPS needs a certificate first
