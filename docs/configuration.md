# Configuration

This project is configured through environment variables. The easiest way to create them is:

```bash
make config
```

The wizard reads `.env.example.json` and writes `.env`.

Wizard behavior:

- `make config` creates `.env` when missing and switches to update mode when `.env` already exists.
- The recommended `Local app + Docker infra` profile keeps app connections on `localhost`.
- The `Full Docker` profile uses Docker service names like `postgres` and `redis`.
- If an older `.env` still has Docker service names, choosing profile 1 in update mode will switch those managed values back to `localhost`.
- Legacy aliases `make env`, `make env-update`, and `make env-reset` still work.

## Server Settings

| Variable | Meaning | Default |
|---|---|---|
| `SERVER_MODE` | App mode. `release` disables the dev server. | `release` |
| `SERVER_IP` | Bind address for the main server. | `127.0.0.1` |
| `SERVER_PORT` | Main server port. | `3220` |
| `SERVER_DOMAIN` | Base domain used to detect tenant subdomains. | `auth.localhost` |
| `SERVER_DEV_IP` | Bind address for the dev server. | `127.0.0.1` |
| `SERVER_DEV_PORT` | Dev server port. | `3221` |
| `SERVER_DEV_DOMAIN` | Dev server domain label. | `auth.localhost` |
| `DEV_API_KEY` | Required for every dev-server request. | required in non-release mode |
| `ENABLE_PPROF` | Enables `/debug/pprof/*` on the dev server. | `false` |

Notes:

- The dev server starts whenever `SERVER_MODE != release`.
- If `ENABLE_PPROF=true`, `pprof` is still protected by `X-Dev-API-Key`.
- The main server and dev server both use HTTP timeouts for basic slow-client protection.

## Database Settings

| Variable | Meaning | Default |
|---|---|---|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | PostgreSQL database name | `dx_auth` |
| `DB_SSLMODE` | PostgreSQL SSL mode | `require` |
| `DB_MAX_OPEN_CONNS` | Maximum open PostgreSQL connections | `25` |
| `DB_MAX_IDLE_CONNS` | Maximum idle PostgreSQL connections | `25` |
| `DB_CONN_MAX_LIFETIME_SECONDS` | Maximum PostgreSQL connection lifetime, in seconds | `3600` |
| `DB_CONN_MAX_IDLE_TIME_SECONDS` | Maximum PostgreSQL connection idle time, in seconds | `900` |

Notes:

- The service now configures the Go SQL pool explicitly instead of relying on driver defaults.
- If `DB_MAX_IDLE_CONNS` is greater than `DB_MAX_OPEN_CONNS`, it is capped to the open-connection limit.
- For local development, Docker Compose, and most `localhost` Postgres setups, use `DB_SSLMODE=disable`.
- Reserve `require` or `verify-full` for databases that are actually configured to accept TLS.

## Redis Settings

| Variable | Meaning | Default |
|---|---|---|
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | empty |
| `REDIS_CACHE_TTL` | Cache lifetime, in seconds | `3600` |

Redis is used for:

- caching users and servers
- tracking failed login attempts
- storing pending registrations (unverified users) until OTP verification completes

If Redis is unavailable during login, authentication fails instead of bypassing rate limiting.

## JWT Settings

| Variable | Meaning | Default |
|---|---|---|
| `JWT_ACCESS_KEYS_DIR` | Subdirectory inside `keys/` for access-token keys | `access` |
| `JWT_REFRESH_KEYS_DIR` | Subdirectory inside `keys/` for refresh-token keys | `refresh` |
| `JWT_ACCESS_EXPIRY` | Access-token lifetime, in seconds | `900` |
| `JWT_REFRESH_EXPIRY` | Refresh-token lifetime, in seconds | `2592000` |

Notes:

- Tokens are signed with `ES256`
- Access and refresh keys are separate
- Public access-token keys are exposed at `/.well-known/jwks.json`

## Login Protection Settings

| Variable | Meaning | Default |
|---|---|---|
| `LOGIN_MAX_ATTEMPTS` | Failed logins before lockout | `5` |
| `LOGIN_LOCKOUT_SECONDS` | Lockout period in seconds | `900` |

These values are included by the env wizard and written into `.env`.

## Email Settings

| Variable | Meaning | Default |
|---|---|---|
| `EMAIL_SMTP_HOST` | SMTP relay hostname. Leave empty to disable email sending (noop mode). | empty |
| `EMAIL_SMTP_PORT` | SMTP relay port. | `587` |
| `EMAIL_SMTP_USERNAME` | SMTP auth username. Leave empty for unauthenticated relay (e.g. MailHog). | empty |
| `EMAIL_SMTP_PASSWORD` | SMTP auth password. | empty |
| `EMAIL_FROM_ADDRESS` | Sender address for all transactional emails. | `no-reply@example.com` |
| `EMAIL_OTP_EXPIRY_SECONDS` | How long a registration OTP remains valid, in seconds. | `600` |

Notes:

- When `EMAIL_SMTP_HOST` is empty, the email sender runs in noop mode: OTP codes are accepted as-is but no actual email is sent. This is useful in development when paired with the dev server.
- For local testing, point `EMAIL_SMTP_HOST` at a MailHog or Mailpit container and leave `EMAIL_SMTP_USERNAME` empty.

## Docker Port Settings

These are only written when the env wizard is using Docker-related options.

| Variable | Meaning | Default |
|---|---|---|
| `DOCKER_IP` | Host IP for main app, DB, and Redis bindings | `127.0.0.1` |
| `DOCKER_DEV_IP` | Host IP for dev-server binding | `127.0.0.1` |
| `DOCKER_SERVER_PORT` | Host port for the main app container | `3220` |
| `DOCKER_SERVER_DEV_PORT` | Host port for the dev app container | `3221` |
| `DOCKER_DB_PORT` | Host port for PostgreSQL | `5432` |
| `DOCKER_REDIS_PORT` | Host port for Redis | `6379` |

## Recommended Local Setup

For most developers:

- `SERVER_MODE=debug`
- `SERVER_IP=127.0.0.1`
- `SERVER_PORT=3220`
- `SERVER_DOMAIN=auth.localhost`
- `SERVER_DEV_IP=127.0.0.1`
- `SERVER_DEV_PORT=3221`
- Docker for PostgreSQL and Redis

## Recommended Production Setup

- `SERVER_MODE=release`
- `SERVER_IP=0.0.0.0`
- `SERVER_DOMAIN=auth.your-domain.com`
- `SERVER_DEV_IP` unused because the dev server should not start
- strong database credentials
- TLS handled by your reverse proxy or ingress
- a private Redis instance, not internet-exposed
