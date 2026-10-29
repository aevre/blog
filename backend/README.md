# Blog Server

A blog backend service written in Go. It serves published posts and links, handles
JWT-based user authentication, generates RSS feeds, and streams LLM-generated post
summaries over Server-Sent Events.

The service is built around [Echo v5](https://github.com/labstack/echo) for HTTP,
[Uber Fx](https://github.com/uber-go/fx) for dependency injection,
[Ent](https://entgo.io/) as the ORM on PostgreSQL, and Redis for caching and
counters. Scheduled background jobs handle link health checks and view-count
persistence.

## Features

- **Authentication & authorization**
  - Email + password registration and login with JWT access tokens.
  - Refresh tokens are signed JWTs stored only as SHA-256 hashes with their expiry in PostgreSQL and returned only as `HttpOnly` cookies.
  - Refresh-token rotation is transactional; reuse revokes the affected session, while logout revokes only the current session.
  - Captcha codes sent by email (SMTP) before registration.
  - Login is limited to 10 attempts per minute per directly connected IP; Redis failures are logged and fail open.
  - RBAC with two roles (`admin`, `reader`). Only `admin` can create/update/delete posts and links; `reader` can only read.
- **Posts**
  - Public list / detail / metadata endpoints.
  - Admin CRUD with tags and categories (many-to-many).
  - View counts are incremented in Redis on read and flushed to PostgreSQL hourly.
  - `read_time_minutes` is derived automatically from content length.
- **Links (友链)**
  - Public list of enabled links.
  - Public "apply for a link" endpoint that creates a disabled link pending review.
  - Link availability is checked hourly by a scheduler job (HTTPS `GET`, status `normal`/`abnormal`).
- **RSS**
  - Paginated RSS 2.0 feed with `atom:link` (self/next/prev/first/last) navigation.
  - Complete archive feed (all published posts).
- **LLM summarization**
  - `POST /api/v1/model/summarize` opens a session; `GET /api/v1/model/summarize/:sessionId` streams the summary as SSE (`data:` lines, `event: done`, `event: error`).
  - Calls GitHub Models / Azure AI inference (`models.inference.ai.azure.com`) with the configured API key.
- **Infrastructure**
  - Structured Zap logging (console in dev, JSON in production).
  - TOML configuration with environment-variable overrides and hot-reload (file change triggers a graceful restart).
  - Graceful HTTP shutdown with a configurable timeout.
  - Liveness and PostgreSQL/Redis readiness endpoints.

## Tech Stack

| Concern            | Library / Tool                                   |
| ------------------ | ----------------------------------------------- |
| Web framework      | Echo v5                                         |
| Dependency inject. | Uber Fx                                         |
| ORM                | Ent v0.14                                       |
| Database           | PostgreSQL (`lib/pq`)                            |
| Cache / counters   | Redis (`redis/go-redis`)                        |
| Object storage     | AWS S3 SDK v2 (S3-compatible; implementation present but not wired into the running server — see below) |
| Logging            | Uber Zap                                        |
| Config             | Viper (TOML) + `fsnotify` hot-reload            |
| JWT                | `golang-jwt/jwt/v5`                             |
| Validation         | `go-playground/validator/v10`                   |
| Email              | `gopkg.in/gomail.v2`                            |

## Requirements

- Go 1.27+ (declared `go 1.27.0` in `go.mod`)
- PostgreSQL
- Redis

## Configuration

Configuration is loaded from a TOML file (default `config.toml` in the working
directory) via Viper. Values can be overridden by environment variables using
the `APP` prefix with `.` replaced by `_` (e.g. `APP_SERVER_PORT`, `APP_JWT_SECRET`,
`APP_DATABASE_PASSWORD`). In production (`app.environment = "production"`) a JWT
secret and a database password are required.

There is **no committed example file** (the repo's `.gitignore` excludes
`config.toml`). Copy the values below into a `config.toml` at the project root.

```toml
[app]
name = "my-app"
version = "1.0.1"
environment = "development"      # "development" | "production"
debug = true
domain = "http://192.168.31.75:3000"   # used to build RSS item URLs
cors_origins = ["*"]             # currently not consumed by any middleware

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
level = "info"            # zap level; "debug" in development
format = ""               # "json" forces JSON output (otherwise dev console)
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

[rustfs]                  # S3-compatible storage (see "Status / Not yet exposed")
region = "us-east-1"
access_key_id = "rustfsadmin"
secret_access_key = "rustfsadmin"
endpoint = "http://localhost:9000/"
```

### Hot reload

When the config file changes on disk, the `config` module triggers an Fx
graceful shutdown (`fx.Shutdowner`). A process manager or container restart
policy must restart the application to load the new configuration.

## Running

```bash
# Download dependencies
go mod download

# Run the application (loads config.toml, runs Goose migrations,
# starts the HTTP server and scheduled jobs)
go run ./cmd/server

# Build a static binary
go build -o blog-server ./cmd/server
./blog-server
```

### Docker

`Dockerfile` builds the `blog-server` binary with CGO disabled and exposes port
`8000`. Mount or bake in a `config.toml`. The application runs embedded Goose
migrations before accepting HTTP traffic.

## API

Application routes are mounted under `/api/v1`. Protected routes require an
`Authorization: Bearer <accessToken>` header. Health endpoints are exposed as
`GET /health/live` and `GET /health/ready` outside the API prefix. Responses use
a common envelope:

```json
{ "code": 0, "msg": "success", "data": { } }
```

Errors return the same envelope with a non-zero `code` and an HTTP status mapped
from it (e.g. `401` for unauthorized/forbidden, `400` for invalid params,
`404` for not found, `409` for conflicts, `500` for internal errors). Messages
are in Chinese for client display.

### Auth

| Method | Path                         | Auth | Description                                  |
| ------ | ---------------------------- | ---- | -------------------------------------------- |
| POST   | `/api/v1/auth/captcha`       | —    | Send a captcha email.                        |
| POST   | `/api/v1/auth/register`      | —    | Register with email + captcha.               |
| POST   | `/api/v1/auth/login`         | —    | Login, returns access token; sets refresh cookie. |
| POST   | `/api/v1/auth/logout`        | —    | Clears the refresh-token cookie.             |
| POST   | `/api/v1/auth/refresh`       | cookie | Rotates access + refresh tokens from cookie. |

Request bodies:

```jsonc
// POST /api/v1/auth/captcha
{ "email": "user@example.com", "type": "Register" }   // type ∈ {Register, PasswordReset, ChangeEmail}

// POST /api/v1/auth/register
{ "email": "user@example.com", "password": "secret123", "passwordConfirm": "secret123", "captcha": "AB12CD" }

// POST /api/v1/auth/login
{ "email": "user@example.com", "password": "secret123" }
```

`login` / `register` responses include `accessToken`, `uuid`, `username`,
`role`, and `avatar`. The refresh JWT is set as an `HttpOnly` cookie
and is not returned in the JSON body. PostgreSQL stores its SHA-256 hash and exact JWT expiry.
Rotation retains used hashes until expiry to detect reuse and revoke the session.
Expired JWTs are rejected before session lookup. Cleanup runs at startup and every hour,
deleting up to 500 expired token records and 500 empty expired or revoked sessions per run,
with a 10-second timeout. Indexed batches use `SKIP LOCKED` for multiple instances.
Expired records may remain until a later batch; they cannot be used to refresh.
Migration 00004 revokes previous opaque-token sessions, so users must sign in again.
Reusing an already expired JWT does not revoke its former session.

### Posts

| Method | Path                              | Auth            | Description                          |
| ------ | --------------------------------- | --------------- | ------------------------------------ |
| GET    | `/api/v1/posts`                   | —               | Paginated published posts (`page`, `pageSize`). |
| GET    | `/api/v1/posts/meta`              | —               | Post id + updated_at (for sitemaps). |
| GET    | `/api/v1/posts/:id`              | —               | Single published post (increments view count). |
| POST   | `/api/v1/posts`                  | admin           | Create a post.                       |
| GET    | `/api/v1/admin/posts`            | admin           | Admin list (filters: `status`, `keyword`). |
| GET    | `/api/v1/admin/posts/:id`        | admin           | Admin post detail.                   |
| PUT    | `/api/v1/admin/posts/:id`        | admin or owner  | Update a post.                       |
| DELETE | `/api/v1/admin/posts/:id`        | admin or owner  | Soft-delete a post.                  |

`POST /api/v1/posts` body (UserID is taken from the JWT, never from the client):

```jsonc
{
  "title": "My post",
  "summary": "Optional summary",
  "cover": "https://...",          // optional
  "content": "Post body...",
  "status": "published",           // draft | published | archived
  "categoryIDs": [1, 2],           // optional, existing category ids
  "tags": [3, 4]                   // optional, existing tag ids
}
```

`status` values: `draft`, `published`, `archived`. Post tags and categories are
referenced by **existing** ids; there are no API endpoints to create tags,
categories, or users.

### Links

| Method | Path                               | Auth | Description                              |
| ------ | ---------------------------------- | ---- | ---------------------------------------- |
| GET    | `/api/v1/links`                    | —    | List enabled links (public).             |
| POST   | `/api/v1/links/apply-link`         | —    | Submit a link application (created disabled). |

`POST /api/v1/links/apply-link` body:

```jsonc
{ "name": "Example", "url": "https://example.com", "description": "…", "avatar": "https://…" }
```

Link status (`normal`/`abnormal`) is computed only by the hourly scheduler job;
there is no HTTP endpoint to approve, edit, or delete links.

### RSS

| Method | Path                       | Auth | Description                                  |
| ------ | -------------------------- | ---- | -------------------------------------------- |
| GET    | `/api/v1/rss`              | —    | RSS 2.0 feed; optional `?page=N` (page size 10). |
| GET    | `/api/v1/rss/complete`     | —    | RSS 2.0 feed with all published posts.       |

### LLM summarization

| Method | Path                                      | Auth | Description                              |
| ------ | ----------------------------------------- | ---- | ---------------------------------------- |
| POST   | `/api/v1/model/summarize`                 | —    | Create a summary session, returns `sessionId`. |
| GET    | `/api/v1/model/summarize/:sessionId`      | —    | SSE stream of the generated summary.     |

`POST /api/v1/model/summarize` body: `{ "content": "…article text…" }`.

The stream emits `data: <text>\n\n` lines, terminates with `event: done\ndata: [DONE]`,
and reports failures via `event: error\ndata: <message>`.

## Project Structure

```
cmd/server/             # Application composition, migrations, HTTP and job lifecycle
internal/
  account/              # Auth, users, mail, account models
  content/              # Posts, categories, tags, RSS and content models
  link/                 # Links and link models
  setup/                # Site initialization and site models
  summary/              # Summary sessions and SSE
  api/                  # Routes, errors, health checks
  platform/             # Cache, config, database, logging, middleware, storage
  shared/               # Context, errors, HTTP envelopes, JWT, random values, transactions, validation
  authz/                # Authorization rules
  scheduler/            # Background jobs
ent/                    # Schemas and generated code
migration/              # Embedded Goose SQL migrations and startup module
templates/              # Email templates
```

See [REFACTOR.md](REFACTOR.md) for behavior, deployment limits, and integration test commands.

### Request flow

1. `cmd/server/main.go` builds the Fx container. Embedded Goose migrations run
   against PostgreSQL before Echo begins accepting requests; scheduled jobs start
   with the application. Redis availability is reported by the readiness endpoint.
2. Global middleware provides request logging, panic recovery, a 10 MiB body
   limit, and atomic Redis-backed login limiting.
3. Routes are registered under `/api/v1`. Protected routes attach
   `middleware.AuthMiddleware`, which parses the Bearer token, resolves the user
   into `contextx.User`, and stores it on the request context.
4. Each feature handler owns its request/response DTOs. Shared HTTP envelopes and
   pagination types live in `internal/shared/httpx`. Domain errors are rendered by
   the central `api.ErrorHandler`.
5. Services use feature-local repositories (Ent) and the transaction manager for
   writes. Posts and links authorize through `authz.Authorizer`.

### Scheduled jobs

Started on app start, cancelled on stop:

- **View-count flush** — every hour, reads `blog:post:view_count:*` keys from
  Redis (`SCAN` + `GETDEL`) and applies bulk increments to PostgreSQL.
- **Link status check** — every hour, `GET`s each HTTPS link and updates its
  status to `normal`/`abnormal` when it changes.

## Status / Not yet exposed

The following exist in the codebase but are **not** reachable over HTTP:

- **S3 object storage**: `storage.S3Storage` implements `Upload`/`Download`/
  `Delete`/`Copy`/`Exists`, but it is not registered in the Fx graph and no
  handler uses it. There is **no `/api/upload` endpoint**.
- **Links admin CRUD**: no endpoints to create/update/delete/approve links;
  link lifecycle is partially driven by the scheduler.
- **Tags, categories, users**: no endpoints to create or manage them; posts only
  reference pre-existing ids.
- **Comments**: a `Comment` table is defined in the Ent schema, but there is no
  repository, service, or handler for it.
- **CORS**: `app.cors_origins` is defined in config but no CORS middleware is
  registered, so cross-origin requests are not handled by the server.
- `service.PostService.GetPostsWithContent` is a stub that always returns
  `(nil, nil)`.
