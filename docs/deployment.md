[English](deployment.md) · [中文](deployment.zh.md)

# Container Deployment Guide (Podman / Docker Compose)

The entire stack — object storage, database, cache, backend, frontend, and reverse proxy — is consolidated into a single `compose.yml`. After cloning, prepare one `.env` and you can launch with one command. Both Podman and Docker work.

rootless / rootful is auto-detected from how you run the script; it does not depend on an environment variable. Plain-user Podman is rootless (80/443 automatically fall back to 8080/8443); running with `sudo`, or via the Docker root daemon, is rootful (binds 80/443 directly, no manual port forwarding needed).

> To fully control how each module runs and skip the container one-command automation, see [Self-hosting](self-hosting.md).

> To have CI build and push images, see [GitHub Actions auto-deploy](ci.md).

## 0. Choose a deployment mode

| Mode | Trigger | Ports | Certificate location | Best for |
| --- | --- | --- | --- | --- |
| Container (rootless) | `./deploy.sh` (plain-user Podman) | Auto fallback 8080/8443 | In-project `deploy/letsencrypt` (owned by the deploy user) | Not wanting root, accepting high ports |
| Container (rootful) | `sudo ./deploy.sh` (or Docker root daemon) | Binds 80/443 directly, no forwarding | In-project `deploy/letsencrypt` (owned by the deploy user) | Wanting real 80/443, no port forwarding |
| Self-hosting (package + steps) | `./deploy.sh self` | Decided by you | Decided by you | Fully controlling how each module runs; script only packages + gives steps |

- rootless / rootful is auto-detected from the running identity (`id` is root or not, whether Podman is rootless, whether the Docker daemon is root); it is not distinguished by an environment variable.
- The two container modes share the same compose logic and certificate directory.

## 1. Prerequisites

- Podman (with `podman compose`) or Docker (with the `docker compose` plugin)
- `envsubst` (usually provided by `gettext`)
- Ports `80` / `443` open on the server (plus `9001` for the RustFS console, as needed)

## 2. Prepare configuration

```bash
cp .env.example .env
# Edit .env; at least change the following:
#   APP_DOMAIN          Full public URL used by the backend (with scheme), e.g. https://blog.example.com
#   NGINX_SERVER_NAME   Nginx server_name (bare domain), e.g. blog.example.com
#   POSTGRES_PASSWORD / JWT_SECRET / RUSTFS_*   Various secrets
```

The meaning of each variable in `.env` is documented in the comments inside `.env.example`.

> The domain is not just something you fill into `.env`: `NGINX_SERVER_NAME` / `APP_DOMAIN` / `RUSTFS_ENDPOINT` are only used by the script to generate Nginx's `server_name` and the backend config; they do not route traffic automatically. To reach the site from the public internet via that domain and let Let's Encrypt complete ACME validation, you must **point the DNS A/AAAA records to the server's public IP at your registrar yourself** (same for the RustFS domain). Until DNS takes effect, certificate issuance fails and the browser cannot reach the site.

## 3. One-command deploy

```bash
./deploy.sh              # Container (rootless / rootful auto-detected)
sudo ./deploy.sh         # rootful: binds 80/443 directly, no port forwarding
```

The script automatically:

1. If there is no `.env`, generate one from `.env.example` (fill it in, then re-run).
2. Generate the backend config `deploy/runtime/backend/config.toml` from `config/backend_config.toml` via `envsubst`.
3. Generate the Nginx config `deploy/runtime/nginx/default.conf` based on `NGINX_TLS` — the HTTP template by default, switching to the HTTPS template when set to `https`.
4. Bring services up in dependency order (postgres/redis → migration → backend/frontend → nginx).
   - The `migration` service runs `ent`'s `Schema.Create` to initialize the schema once PostgreSQL is ready (one-shot; exit code 0 means success).
   - `backend` waits for the migration to finish via `depends_on: migration: service_completed_successfully`, so you do not need to create tables manually.

> Why not just rely on compose's `depends_on` conditions and `up` everything at once?
> `podman-compose` 1.0.6's support for `service_healthy` / `service_completed_successfully` is unreliable — it can leave `backend` stuck in `Created` without starting, or let `migration` exit before PostgreSQL is ready.
> `deploy.sh` instead brings services up explicitly in dependency order and polls for readiness, avoiding the problem.

After deployment:

- Frontend: `http://<NGINX_SERVER_NAME>`
- Backend API: `http://<NGINX_SERVER_NAME>/api/v1/`
- RustFS console: `http://<server-IP>:9001`

Common management commands:

```bash
podman compose ps                 # View status
podman compose logs -f backend    # View backend logs
podman compose down               # Stop and remove containers
./deploy.sh                       # Redeploy (re-run after updating code; it rebuilds automatically)
```

## On Rootless / Rootful and the 80/443 ports

`deploy.sh` (and the CI `deploy/remote-deploy.sh`) distinguishes rootless / rootful by how it is run — no environment variable needed:

- **rootful (recommended for those wanting real 80/443)**: run with `sudo`, or go through the Docker root daemon. The script binds 80/443 directly, no port forwarding. Note the certificate directory ownership under rootful: before compose mounts it for the first time, the script pre-creates `deploy/letsencrypt` and `deploy/acme` as the deploy user, and when running as root it `chown`s them back to the deploy user. This prevents the docker/podman daemon from creating the missing bind-source directory as `root:root`, which would make certbot writes fail.

  ```bash
  sudo ./deploy.sh
  ```

  Keep `HTTP_PORT=80` / `HTTPS_PORT=443` in `.env`. The script detects rootful and binds 80/443 directly.

- **rootless**: a plain-user Podman cannot map container ports to host ports below `1024`, so using the default `80/443` directly fails to start. There are three ways to handle this:

  1. Run rootful (see above, `sudo ./deploy.sh`).
  2. Open up unprivileged ports:

     ```bash
     echo 'net.ipv4.ip_unprivileged_port_start=0' | sudo tee /etc/sysctl.d/99-podman-port.conf
     sudo sysctl --system
     ```

     After this, a plain `./deploy.sh` can also bind 80/443, at the cost of letting all users bind low ports. `deploy.sh` and `deploy/remote-deploy.sh` auto-detect this value: as long as `net.ipv4.ip_unprivileged_port_start<=80`, they keep using `80:80` / `443:443` (no `.env` change needed), and the ACME HTTP-01 challenge can hit port 80 directly for issuance.

  3. Stay rootless + high ports, then forward on the host:

     Set `.env` to `HTTP_PORT=8080` / `HTTPS_PORT=8443` (under rootless the script also auto-falls back to these two ports), then forward `80→8080` and `443→8443` on the host. This is essentially TCP port forwarding (stream) — **do not terminate TLS on the upstream reverse proxy**; the certificate lives on the in-container nginx, so let both the ACME HTTP-01 challenge on 80 and access on 443 go straight to the container nginx.

     **Method a: socat (simplest, use as-is)**

     ```bash
     sudo socat TCP-LISTEN:80,fork,reuseaddr TCP:127.0.0.1:8080 &
     sudo socat TCP-LISTEN:443,fork,reuseaddr TCP:127.0.0.1:8443 &
     ```

     Note `&` only lasts for the current session and is lost after reboot. For a simple persistent run, add `nohup ... &`; for boot-time auto-start see Method c.

     **Method b: host Nginx doing TCP passthrough (stream, boot auto-start, more stable)**

     Method b relies on Nginx's stream module for layer-4 TCP forwarding, not the usual `http {}` reverse proxy. Below we assume your host `nginx.conf` already includes the top-level `include /etc/nginx/modules-enabled/*.conf;` to load dynamic modules, and a top-level `stream { include /etc/nginx/stream/*.conf; }` block as the layer-4 forwarding block. If your nginx.conf does not yet have a top-level `stream {}` block, see the "When there is no top-level stream block" supplement at the end.

     **Step 1: install the stream module (Ubuntu / Debian)**

     ```bash
     sudo apt-get update && sudo apt-get install -y libnginx-mod-stream
     ```

     After install it works by default: the package auto-creates a symlink in `/etc/nginx/modules-enabled/` to enable the module, and together with your existing top-level `stream { include /etc/nginx/stream/*.conf; }` no further `nginx.conf` change is needed. Verify:

     ```bash
     ls -l /etc/nginx/modules-enabled/ | grep -i stream   # should show a mod-stream.conf symlink
     nginx -V 2>&1 | grep -o with-stream                  # dynamic module also visible (or statically compiled)
     ```

     In the rare case the symlink is not created automatically, add one manually:

     ```bash
     sudo ln -s /usr/share/nginx/modules-available/mod-stream.conf /etc/nginx/modules-enabled/mod-stream.conf
     ```

     For RHEL / CentOS / Rocky use: `sudo yum install -y nginx-mod-stream`.

     **Step 2: create the stream forwarding config**

     Since the outer `stream {}` is already in `nginx.conf`, **do not write the outer `stream {}` here** — just create `/etc/nginx/stream/stream.conf` (any `*.conf` filename works) in the include directory:

     ```nginx
     # /etc/nginx/stream/stream.conf
     server { listen 80;  proxy_pass 127.0.0.1:8080; }
     server { listen 443; proxy_pass 127.0.0.1:8443; }
     ```

     **Step 3: validate and reload**

     ```bash
     sudo nginx -t            # must be "syntax is ok" / "test is successful"
     sudo systemctl reload nginx
     ```

     Prerequisite: the host has Nginx installed and running (only for forwarding — it does not host sites or hold certificates). To verify the forwarding, see the hint below Method c — if the public internet can reach `http://domain:80/.well-known/acme-challenge/` and gets 404, port 80 is already passing through to the container nginx.

     When there is no top-level `stream {}` block, e.g. the default Debian/Ubuntu `nginx.conf`: for a dynamic module, add `load_module modules/ngx_stream_module.so;` at the very top of `nginx.conf`, before `events {}`; add `include /etc/nginx/stream.conf;` outside `http {}`, and create that top-level file with `stream { ... }`. Note `stream {}` is a top-level directive and cannot go inside `conf.d/` (which is inside `http {}` and would report `"stream" directive is not allowed here`).

     **Method c: systemd persistent socat (boot auto-start, crash self-heal)**

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

     To verify the forwarding works: `curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:80/.well-known/acme-challenge/` returning `404` means port 80 is forwarded to the container nginx and serves the challenge directory. The public internet must also be able to reach the domain on `:80`. Only then does the deploy script's "ACME challenge port 80 is reachable" check pass and certificates auto-issue.

If running rootless, and you have neither opened low ports nor set up a host reverse proxy, `deploy.sh` and `deploy/remote-deploy.sh` automatically change `80/443` to `8080/8443` so it can start, and print in the terminal how to open the real ports. In that case port 80 is not inside the container, the ACME HTTP-01 challenge only validates port 80, and it fails and falls back to HTTP. For HTTPS to work, you must satisfy one of Method 2 or Method 3 above.

## 4. Enable HTTPS

By default the site is served over HTTP; you can switch to HTTPS. Certificates are issued by Let's Encrypt using the ACME HTTP-01 challenge. certbot writes the validation file in the challenge directory, Let's Encrypt accesses `http://domain/.well-known/acme-challenge/...` to confirm domain ownership, and only issues after validation passes. HTTP-01 always validates port 80 only. The in-container Nginx already has `location /.well-known/acme-challenge/` configured to serve that directory.

### Automatic issuance

Set `NGINX_TLS` to `https` in `.env`, then run the deploy script as usual — no manual step-by-step needed:

```bash
./deploy.sh               # Container one-command deploy (rootless / rootful auto-detected)
sudo ./deploy.sh          # rootful
bash deploy/remote-deploy.sh all    # CI server-side deploy
```

Before generating the final config, the deploy script first brings up Nginx with the HTTP template to serve the ACME challenge, then uses certbot to issue or renew the certificate, and finally regenerates the config with the HTTPS template and starts it. This avoids the deadlock of "Nginx won't start without a certificate, and you can't get a certificate without Nginx running." `NGINX_TLS_AUTO` is on by default; set it to `off` to stop the script from auto-issuing, and prepare the certificate yourself in advance with `deploy/certbot.sh`.

Each domain's certificate status is judged independently — blog and RustFS are separate:

- Missing or expired: fresh issuance with certonly.
- Remaining validity no more than `RENEW_DAYS` (default 30 days): renew.
- Remaining validity greater than 30 days: skip.

Automatic issuance requires port 80 reachable from the public internet and certbot installed. If port 80 is unreachable, the script errors out and falls back to the HTTP template, serving the site over HTTP. For opening port 80 under rootless, see the Rootless section above.

Equivalent manual issue / renew commands:

```bash
./deploy/certbot.sh certonly -d blog.example.com     # first issuance
./deploy/certbot.sh renew                            # renew, using the in-project renewal config
sudo ./deploy/certbot.sh import blog.example.com     # import from host /etc/letsencrypt
```

`deploy/certbot.sh` runs Certbot as the deploy user by default; certificates are stored in the project directory `deploy/letsencrypt/etc/{live,archive,renewal}/`, owned by the deploy user, so the Nginx container can read them directly. When no Webroot is given, the wrapper uses the absolute host path `deploy/acme`, which is mounted into the container as `/var/www/certbot`; project mode rejects the container path as a host Webroot. When no email is given, the script reads `EMAIL_FROM` or `CERTBOT_EMAIL` from `.env` and automatically adds `--agree-tos --no-eff-email --non-interactive`.

`certbot.sh` supports the `CERT_MODE` environment variable: `project` (default, in-project directory, for container mode) and `system` (`/etc/letsencrypt`, owned by root and managed by the system certbot, for manually issuing certificates with the system certbot after self-hosting). The `import` subcommand only makes sense in `project` mode.

### DNS-01 challenge

When rootless is paired with high ports (e.g. 8080), Let's Encrypt reaches port 80 and the challenge fails. In that case you can forward the host's port 80 to 8080, or use the DNS-01 challenge (certbot's DNS plugin), which needs no inbound port 80 and is best suited for rootless:

```bash
./deploy/certbot.sh certonly --dns-cloudflare -d blog.example.com
```

### Recovering from abnormal certificate state

During automatic issuance, if `live/<domain>` is missing but `archive/<domain>` still exists, or `renewal/<domain>.conf` is empty, that is a leftover broken config from a mid-issuance failure — Certbot thinks the certificate already exists and refuses to issue. The deploy script stops without deleting these files, protecting historical archives and private keys. After making a backup and checking the state, clean it manually and redeploy:

```bash
rm -rf deploy/letsencrypt/etc/live/<domain> deploy/letsencrypt/etc/archive/<domain>
rm -f  deploy/letsencrypt/etc/renewal/<domain>.conf
```

## 5. RustFS reverse proxy (required — frontend serves images and other static assets)

The frontend serves images and other static assets through RustFS, so `RUSTFS_ENDPOINT` is required and must be a publicly reachable `https://` address. After setting it in `.env`, the deploy script automatically:

- Derives the API host (e.g. `oos.immortel.top`) and the console host (auto-prefixed with `admin.`, e.g. `admin.oos.immortel.top`) from that address's domain.
- Adds the corresponding reverse proxy in the in-container Nginx: API → `rustfs:9000`, console → `rustfs:9001` (`client_max_body_size 10G`).
- Requests SAN certificates for these two domains during issuance/renewal (independent from the blog site).

Note: after enabling, the host Nginx should not separately terminate these two domains with an `http {}` block (it conflicts with the second layer's stream passthrough) — just keep the TCP passthrough of 80/443 (or 8080/8443) to the container Nginx; the old host `oos.conf` can be removed. `RUSTFS_ENDPOINT` still serves as the address the backend uses to connect to object storage (frontend and backend both go through this public domain).

`deploy/letsencrypt/`, `deploy/acme/`, `deploy/runtime/` are generated by the deploy script at runtime and are already git-ignored — do not commit them.
