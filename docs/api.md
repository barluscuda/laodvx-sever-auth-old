# API Guide

This is a practical guide to the available endpoints. If you want ready-to-run requests, import the Postman collection in [api-docs/laodvx-server-auth.postman_collection.json](/home/mrbarlus/projects/laodvx-projects/laodvx-server-auth/api-docs/laodvx-server-auth.postman_collection.json).

## Before You Call Anything

There are two host styles:

- tenant routes: `{server_id}.auth.localhost`
- system routes: `auth.localhost`

Examples:

```text
http://127.0.0.1:3220/api/auth/login
Host: <server_id>.auth.localhost
```

```text
http://127.0.0.1:3220/system/api/auth/login
Host: auth.localhost
```

Protected routes use:

```text
Authorization: Bearer <access_token>
```

Dev routes use:

```text
X-Dev-API-Key: <DEV_API_KEY>
```

## Public Main-Server Endpoints

### `GET /.well-known/jwks.json`

Returns the public key set for verifying access tokens.

### `GET /api/ping`

Tenant-scoped health check.

Call it on a tenant host:

```text
Host: <server_id>.auth.localhost
```

## Tenant User Endpoints

These routes must be called on a tenant host.

### `POST /api/user`

Start registration for a new user account. Validates the email and password, stores a pending registration in cache, and sends a 6-digit OTP to the email address. The user is **not** written to the database until the OTP is verified.

Request:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Success: `202 Accepted` (no body)

### `POST /api/auth/verify-email`

Complete registration by submitting the OTP received by email.

Request:

```json
{
  "email": "user@example.com",
  "otp": "123456"
}
```

Success: `204 No Content`

The OTP is a 6-digit code valid for 10 minutes by default (`EMAIL_OTP_EXPIRY_SECONDS`). If the OTP is wrong or expired the response is `422 err_verification_token_invalid`.

### `POST /api/user/resend-verification`

Resend the verification OTP for a pending registration. Always returns `202` regardless of whether the email is registered, to avoid email enumeration.

Request:

```json
{
  "email": "user@example.com"
}
```

Success: `202 Accepted` (no body)

### `POST /api/auth/login`

Log in as a tenant user or tenant admin.

Request:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Success response:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<jwt>"
}
```

### `POST /api/auth/refresh`

Exchange a refresh token for a new token pair.

Request:

```json
{
  "refresh_token": "<jwt>"
}
```

### `GET /api/user/me`

Returns the currently logged-in tenant user.

Allowed roles:

- `user`
- `tenant_admin`

## Tenant Admin Endpoints

These routes also use the tenant host, but require a `tenant_admin` token.

### `GET /api/admin/user`

Returns all users in the current tenant.

### `GET /api/admin/user?email=user@example.com`

Find a user by email inside the current tenant.

### `GET /api/admin/user/:id`

Find a user by UUID inside the current tenant.

## System Admin Authentication

These routes use the base domain host, not a tenant host.

### `POST /system/api/auth/login`

Request:

```json
{
  "username": "admin",
  "password": "secret"
}
```

Success response:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<jwt>"
}
```

### `POST /system/api/auth/refresh`

Request:

```json
{
  "refresh_token": "<jwt>"
}
```

## System Admin Server Management

These routes require a `system_admin` access token.

### `POST /system/api/server`

Create a tenant server.

```json
{
  "name": "my-server"
}
```

### `GET /system/api/server`

List all tenant servers.

### `GET /system/api/server/:id`

Get one tenant server by UUID.

### `PUT /system/api/server/:id`

Rename a tenant server.

```json
{
  "name": "new-server-name"
}
```

### `DELETE /system/api/server/:id`

Delete a tenant server.

Success: `204 No Content`

## System Admin User Lookup

These routes require a `system_admin` access token.

### `GET /system/api/user`

Returns all users across all tenant servers.

### `GET /system/api/user?server_id=<uuid>`

Returns all users inside one tenant server.

### `GET /system/api/user?email=user@example.com`

Returns all accounts with the same email across tenant servers.

### `GET /system/api/user?server_id=<uuid>&email=user@example.com`

Returns one user by email inside one tenant server.

### `GET /system/api/user/:id`

Returns one user by UUID globally.

## Dev Server Endpoints

The dev server runs only when `SERVER_MODE != release`.

Base URL by default:

```text
http://127.0.0.1:3221
```

Every request requires:

```text
X-Dev-API-Key: <DEV_API_KEY>
```

### `GET /api/status`

Simple status check for the dev server.

### `POST /api/system/admin`

Create a system admin account during development.

```json
{
  "username": "admin",
  "password": "password123"
}
```

### `GET /api/system/admin`

List all system admin accounts.

### `GET /api/system/admin/:id`

Get a system admin by UUID.

### `DELETE /api/system/admin/:id`

Delete a system admin.

### `GET /debug/pprof/*`

Available only when `ENABLE_PPROF=true`, and still protected by `X-Dev-API-Key`.
