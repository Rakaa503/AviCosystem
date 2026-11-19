# AviCosystem Backend

Backend service for the AviCosystem platform.

**Build Now. Secure First.**

## Technology Stack

- Go 1.26
- Fiber v3
- PostgreSQL
- GORM
- Redis
- JWT
- bcrypt

## Architecture

The backend follows a modular architecture.

```text
cmd/
  api/

internal/
  cache/
  config/
  database/
  middleware/
  modules/
  server/
```

### Domain Modules

| Module | Responsibility |
|---|---|
| ai | AI-related domain |
| ask | Question and answer domain |
| auth | Authentication |
| chat | Communication domain |
| community | Community domain |
| learn | Learning domain |
| notification | Notification domain |
| schedule | Scheduling domain |
| stack | Technology stack domain |
| users | User management |

Module directories represent the domain boundaries. Their implementation status may differ.

## Current HTTP Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | /health | Application health |
| GET | /ready | Dependency readiness |
| POST | /api/v1/auth/register | User registration |
| POST | /api/v1/auth/login | User login |

## Getting Started

Requirements:

- Go 1.26 or compatible
- PostgreSQL
- Redis

Clone the repository:

```bash
git clone https://github.com/Rakaa503/AviCosystem.git
cd AviCosystem/backend
```

Prepare environment configuration:

```powershell
Copy-Item .env.example .env
```

Configure the environment variables in `.env` before starting the application.

Download dependencies:

```bash
go mod download
```

Run the API:

```bash
go run ./cmd/api
```

## Development Checks

```bash
go test ./... -count=1
go vet ./...
go build ./...
```

## Documentation

- [Architecture](docs/architecture.md)
- [API Reference](docs/api-reference.md)
- [Development Guide](docs/development.md)
- [Database Design](docs/database-design.md)
- [Security](docs/security.md)
- [Contributing](CONTRIBUTING.md)

## Security

Never commit `.env`, credentials, private keys, or production secrets.

See [SECURITY.md](SECURITY.md).

## License

License information will be added when the project license is established.
