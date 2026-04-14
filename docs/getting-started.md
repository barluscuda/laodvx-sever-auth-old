# Getting Started

This guide gets `laodvx-server-auth` running on a local machine with the least amount of setup.

## What You Need

- Go `1.26+`
- Docker

## How Routing Works

The server identifies tenants from the `X-Tenant-Id` request header, which must contain the tenant server's UUID.

- Tenant API routes (`/api/*`) require `X-Tenant-Id: <tenant-id>`
- System routes (`/system/api/*`) must be called without `X-Tenant-Id`

In production, nginx extracts the `server_id` from the subdomain and sets this header before forwarding. For local development and direct API calls (Postman, curl), pass the header manually.

Examples:

- User login: `POST http://127.0.0.1:3220/api/auth/login` with `X-Tenant-Id: <tenant-id>`
- System admin login: `POST http://127.0.0.1:3220/system/api/auth/login` (no `X-Tenant-Id`)

## 1. Create `.env`

Run:

```bash
make config
```

The wizard reads `.env.example.json` and writes a local `.env`.

Recommended setup:

```text
Choose setup profile
  1. Local app + Docker infra (Recommended)
```

That keeps the Go app on your host and runs PostgreSQL and Redis in Docker.
With that profile the generated app config stays on `localhost` for `DB_HOST` and `REDIS_HOST`, while Docker host-port settings are written separately for the infra containers.

The wizard now also:

- explains each setting while prompting
- auto-fills Docker service names only for the `Full Docker` profile
- corrects old Docker hostnames like `postgres` and `redis` back to `localhost` when you switch to profile 1
- defaults `DB_SSLMODE` to `disable` for local and Docker development profiles
- generates a strong `DEV_API_KEY` automatically when running outside `release` mode
- writes login-protection settings like `LOGIN_MAX_ATTEMPTS` and `LOGIN_LOCKOUT_SECONDS`

If `.env` already exists:

- `make config` opens update mode automatically when `.env` already exists
- `make config-update` updates values interactively
- `make config-reset` overwrites the file with fresh defaults
- legacy aliases `make env`, `make env-update`, and `make env-reset` still work

## 2. Start Infrastructure

```bash
make docker-services
```

This starts:

- PostgreSQL
- Redis

## 3. Start The Server

For local development:

```bash
make run
```

For full Docker:

```bash
make docker-stack
```

## 4. Check That It Works

Main server:

```bash
curl http://127.0.0.1:3220/.well-known/jwks.json
```

Dev server, non-release modes only:

```bash
curl -H "X-Dev-API-Key: <your-dev-key>" http://127.0.0.1:3221/api/status
```

## Common Workflow

1. Create a system admin on the dev server.
2. Log in as that system admin on the main server.
3. Create a tenant server.
4. Register a user under that tenant server — a 6-digit OTP is sent to the email address.
5. Verify the OTP with `POST /api/auth/verify-email`.
6. Log in as a tenant user or tenant admin.

The Postman collection in [api-docs/laodvx-server-auth.postman_collection.json](/home/mrbarlus/projects/laodvx-projects/laodvx-server-auth/api-docs/laodvx-server-auth.postman_collection.json) is the fastest way to walk through that flow.

## Useful Commands

| Command | What it does |
|---|---|
| `make config` | Create or update `.env` interactively |
| `make config-update` | Update an existing `.env` |
| `make config-reset` | Replace `.env` with new defaults |
| `make run` | Run the Go server locally |
| `make build-bin` | Build `./dist/server` |
| `make run-bin` | Run the existing `./dist/server` binary without rebuilding |
| `make docker-services` | Start PostgreSQL and Redis only |
| `make docker-stack` | Build and start everything in Docker |
| `make docker-start` | Start existing Docker services |
| `make docker-stop` | Stop Docker services |
| `make docker-tail` | Tail container logs |

## If Something Fails

- If user routes return `err_invalid_server_id`, the `X-Tenant-Id` header is missing or not a valid UUID.
- If tenant routes return `err_not_found`, the UUID in `X-Tenant-Id` does not match any registered server.
- If the dev server returns `err_missing_auth`, your `X-Dev-API-Key` header is missing or wrong.
- If PostgreSQL on `localhost` refuses TLS, set `DB_SSLMODE=disable`.
- If the app cannot start, check PostgreSQL, Redis, and required JWT key settings in `.env`.
