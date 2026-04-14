# Authentication

This service supports three roles:

- `user` — standard tenant user
- `tenant_admin` — tenant-level admin (JWT claim value: `user_admin`)
- `system_admin` — global admin across all tenant servers

TenantUser and tenant admin accounts belong to a single tenant server. System admins are global.

## Basic Flow

Tenant user flow:

1. Register on `POST /{tenant-host}/api/user` — sends a 6-digit OTP to the email address; the account is not active yet
2. Verify email with `POST /{tenant-host}/api/auth/verify-email` — submit `email` + `otp`; the user is created in the database on success
3. Log in on `POST /{tenant-host}/api/auth/login`
4. Use `Authorization: Bearer <access_token>`
5. Refresh with `POST /api/auth/refresh`

Login before OTP verification returns `401 err_invalid_credentials` because the user record does not exist in the database yet. Resend the OTP at any time with `POST /{tenant-host}/api/user/resend-verification`.

System admin flow:

1. Log in on `/system/api/auth/login`
2. Use the returned access token for system admin routes
3. Refresh with `/system/api/auth/refresh`

## Access Tokens

Access tokens:

- are JWTs
- use `ES256`
- default to `15` minutes
- include the user ID, role, and tenant server ID when applicable

TenantUser and `tenant_admin` tokens are bound to the tenant server they were issued for. A token from one tenant cannot be used against another tenant host.

## Refresh Tokens

Refresh tokens:

- are separate JWTs
- default to `30` days
- are single-use
- are tracked in the database by token ID

When a refresh token is used:

1. the server validates the token
2. the server marks its token ID as used
3. the server issues a brand-new access token and refresh token

If the same refresh token is used twice, the second request is rejected.

## Login Protection

Login protection is built in:

- failed attempts are tracked in Redis
- after `5` failed attempts, the account is locked for `15` minutes by default
- successful login clears the failed-attempt counter
- login timing is equalized so “wrong password” and “unknown account” are harder to distinguish

If Redis is unavailable during login, the request fails with an internal error. The system does not skip throttling when its protection backend is down.

## Token Verification

Public signing keys are available here:

```text
GET /.well-known/jwks.json
```

Use that endpoint if another service needs to verify access tokens offline.

## Dev Server Authentication

The dev server does not use JWTs. It uses this header:

```text
X-Dev-API-Key: <DEV_API_KEY>
```

This applies to:

- `/api/status`
- `/api/system/admin/*`
- `/debug/pprof/*` when `ENABLE_PPROF=true`

## Common Authentication Errors

| Error code | Meaning |
|---|---|
| `err_invalid_credentials` | Wrong email/username or password |
| `err_account_locked` | Too many failed login attempts |
| `err_missing_auth` | Missing bearer token or dev API key |
| `err_token_invalid` | Malformed or invalid token |
| `err_token_used` | Refresh token was already used |
| `err_forbidden` | Valid token, but wrong role or wrong tenant |

See [error-codes.md](/home/mrbarlus/projects/laodvx-projects/laodvx-server-auth/docs/error-codes.md) for the full list.
