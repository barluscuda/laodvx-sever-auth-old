# laodvx-server-auth

REST API authentication server built with Go, Gin, PostgreSQL, and Redis.  
Supports multi-server tenancy — users are scoped to a registered server, so the same email can exist across different servers.

## Tech stack

- **Go 1.26** — Gin · GORM
- **PostgreSQL 17** — primary store
- **Redis 7** — caching + login-attempt tracking
- **Docker** — containerization
- **ES256 (P-256)** — JWT signing with auto-rotating key pairs
- **Subdomain-based routing** — `{server_id}.auth.localhost/api/*` for multi-tenancy

## Quick start

```bash
make config          # interactive .env wizard
make docker-services # start postgres + redis
make run             # run server locally
```

Server binds to `auth.localhost:3220` by default. Add to `/etc/hosts`:
```
127.0.0.1 auth.localhost
127.0.0.1 *.auth.localhost
```

See [Getting Started](docs/getting-started.md) for detailed setup instructions.

In non-release modes the app also starts a dev server on `SERVER_DEV_PORT`. All dev endpoints, including optional `pprof`, require `X-Dev-API-Key`.

## Documentation

| Doc | Description |
|---|---|
| [Getting Started](docs/getting-started.md) | Prerequisites, env wizard, run commands |
| [Configuration](docs/configuration.md) | All environment variables and defaults |
| [API Reference](docs/api.md) | All endpoints, request/response shapes |
| [Authentication](docs/authentication.md) | Token design, refresh flow, login security |
| [Architecture](docs/architecture.md) | Project structure, design decisions, layers |

## Postman collection

Import [`api-docs/laodvx-server-auth.postman_collection.json`](api-docs/laodvx-server-auth.postman_collection.json) into Postman.  
Set the `base_url`, `dev_url`, `server_id`, and `dev_api_key` collection variables to get started.

---

## Developer

**Barluscuda** — backend developer focused on building clean, secure server infrastructure in Go.  
`laodvx-server-auth` is part of the **laodvx** project — a modular auth backend designed for multi-tenant applications.
