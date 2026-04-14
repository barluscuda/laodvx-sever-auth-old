# Error Codes

Every non-success response uses the same JSON shape:

```json
{
  "error": "err_invalid_credentials",
  "message": "The email or password is incorrect."
}
```

Use `error` in frontend logic. Treat `message` as a readable fallback, not as a stable API contract.

## Request Errors

| Code | HTTP | Meaning |
|---|---|---|
| `err_invalid_server_id` | 400 | The tenant ID in the host name is missing or invalid |
| `err_invalid_id` | 400 | A path ID is not a valid UUID |
| `err_invalid_request` | 400 | The request body or query values are invalid |

## Authentication Errors

| Code | HTTP | Meaning |
|---|---|---|
| `err_missing_auth` | 401 | Missing bearer token or missing dev API key |
| `err_invalid_credentials` | 401 | Wrong login identifier or password |
| `err_account_locked` | 429 | Too many failed login attempts |
| `err_token_invalid` | 401 | Invalid, malformed, or wrong-type token |
| `err_token_expired` | 401 | Expired token |
| `err_token_used` | 401 | Refresh token was already used |
| `err_invalid_token_claims` | 401 | Token structure was valid but claims were unusable |

## Authorization Errors

| Code | HTTP | Meaning |
|---|---|---|
| `err_forbidden` | 403 | Valid auth, but wrong role or wrong tenant |

## Validation Errors

| Code | HTTP | Meaning |
|---|---|---|
| `err_verification_token_invalid` | 422 | OTP is missing, wrong, or expired |

## Resource Errors

| Code | HTTP | Meaning |
|---|---|---|
| `err_not_found` | 404 | The requested user, server, or admin was not found |
| `err_duplicate_email` | 409 | A user with that email already exists inside the tenant |

## Internal Errors

| Code | HTTP | Meaning |
|---|---|---|
| `err_internal` | 500 | Unexpected server-side failure |

## Frontend Tip

Good pattern:

1. read the `error` field
2. map it to your own UI copy
3. fall back to `message` if your app has no custom text for that code
