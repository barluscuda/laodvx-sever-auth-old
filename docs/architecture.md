# Architecture

This project follows a simple layered design so HTTP code, business logic, and persistence stay separate.

## High-Level Shape

Request flow:

```text
HTTP Handler -> Service -> Repository -> PostgreSQL / Redis
```

Each layer depends on interfaces in `internal/ports`, not on concrete implementations.

## Project Layout

```text
cmd/server/                 app entrypoint
config/                     env loading and typed config
internal/adapters/database/ PostgreSQL connection
internal/adapters/http/     handlers, DTOs, middleware, routers
internal/adapters/redis/    Redis connection
internal/adapters/repository/ database and cache repositories
internal/domain/model/      core data models
internal/domain/service/    business logic
internal/pkg/token/         JWT claims and key handling
internal/ports/             service and repository interfaces
internal/server/            HTTP server setup and shutdown
docs/                       human docs
api-docs/                   Postman collection
```

## Main Concepts

### Tenant Servers

`Server` is a first-class model. Every user belongs to a server, and the active server is identified by the `X-Tenant-Id` request header, which must contain the tenant UUID.

In production, nginx extracts the `server_id` from the subdomain and sets this header before forwarding. For local development, pass the header directly.

### Roles

- `user`: standard tenant user
- `tenant_admin`: tenant-level admin (JWT claim value: `user_admin`)
- `system_admin`: global admin across all tenant servers

### Two HTTP Servers

The app can run two HTTP servers:

- main server: public API, user auth, system auth, tenant and system routes
- dev server: local bootstrap and diagnostics for non-release modes

## What PostgreSQL Stores

PostgreSQL stores:

- tenant servers
- users
- system admins
- refresh-token records

Refresh tokens are stored by token ID so they can be enforced as single-use.

## What Redis Stores

Redis stores:

- cached server lookups
- cached user lookups
- failed login counters
- pending registrations (unverified users waiting for OTP confirmation)

Pending registrations are stored with a TTL matching `EMAIL_OTP_EXPIRY_SECONDS`. The user record is only written to PostgreSQL after the OTP is verified.

If Redis is down during login, auth fails instead of silently disabling protection.

## Middleware Responsibilities

Main middleware responsibilities:

- extract tenant `server_id` from `X-Tenant-Id` header
- reject invalid tenant IDs early
- confirm the tenant server exists
- validate access tokens
- enforce role checks
- ensure tenant tokens match the tenant server

## Security Notes

- passwords are stored as bcrypt hashes
- access and refresh tokens use separate signing keys
- refresh tokens are single-use
- login attempts are rate-limited
- email verification uses a short-lived 6-digit OTP; users are not written to the database until verified
- resend-verification always returns `202` to avoid revealing whether an email is registered
- both HTTP servers use request timeouts
- dev endpoints and optional `pprof` require `X-Dev-API-Key`

## Why This Structure Helps

This layout makes it easier to:

- test business rules without HTTP concerns
- swap repository implementations without changing handlers
- keep auth logic in one place
- reason about tenant isolation
