# API Reference

Base path:

```text
/api/v1
```

## Health

### GET /health

Checks whether the HTTP application responds.

Example response:

```json
{
  "status": "ok",
  "service": "AviCosystem"
}
```

## Readiness

### GET /ready

Checks PostgreSQL and Redis connectivity.

Successful response:

```json
{
  "status": "ready",
  "dependencies": {
    "postgresql": "connected",
    "redis": "connected"
  }
}
```

A dependency failure returns HTTP 503.

## Authentication

### POST /api/v1/auth/register

Creates a user account.

Request:

```json
{
  "name": "Example User",
  "email": "user@example.com",
  "password": "StrongPassword123!"
}
```

Successful response: HTTP 201.

Possible status codes:

- 201: account created
- 400: invalid registration input
- 409: email already registered
- 500: internal registration failure

### POST /api/v1/auth/login

Authenticates a user.

Request:

```json
{
  "email": "user@example.com",
  "password": "StrongPassword123!"
}
```

Successful response: HTTP 200.

Response fields:

- user
- access_token
- token_type
- expires_in

Possible status codes:

- 200: authentication successful
- 400: malformed request body
- 401: invalid credentials
- 403: account unavailable
- 500: internal login failure

## Authentication Notes

Access tokens are JWT tokens.

The configured access-token lifetime is controlled by application configuration.

Do not expose passwords, password hashes, or JWT secrets in API responses or logs.
