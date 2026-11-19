# Development Guide

## Requirements

- Go 1.26
- PostgreSQL
- Redis
- Git

## Environment

Copy the example environment file:

```powershell
Copy-Item .env.example .env
```

Configure the required values locally.

Never commit the actual `.env` file.

## Dependencies

```powershell
go mod download
```

## Run

```powershell
go run ./cmd/api
```

## Tests

Run all tests:

```powershell
go test ./... -count=1
```

Run authentication tests:

```powershell
go test ./internal/modules/auth -v -count=1
```

## Static Analysis

```powershell
go vet ./...
```

## Build

```powershell
go build ./...
```

To produce a local executable:

```powershell
New-Item -ItemType Directory -Force .\bin
go build -o .\bin\api.exe .\cmd\api
```

The bin directory is excluded from Git.

## Formatting

```powershell
gofmt -w .\cmd .\internal
```
