# Security Engineering

## Current Security Components

The backend includes security-related infrastructure such as:

- JWT access-token handling
- bcrypt password hashing
- Request rate limiting
- Security HTTP headers
- CORS configuration
- Request ID and request logging
- Request timeout
- Proxy trust configuration
- Account status validation

## Security Requirements

- Never store plaintext passwords.
- Never commit `.env` or secret credentials.
- Keep JWT secrets private.
- Validate incoming request data.
- Avoid exposing internal errors to clients.
- Apply least privilege to database credentials.
- Review dependencies regularly.
- Keep security-sensitive configuration explicit.

## Reporting

Security vulnerabilities should be reported privately.

See the repository's SECURITY.md file.
