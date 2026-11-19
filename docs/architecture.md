# Architecture

## Overview

AviCosystem Backend uses a modular Go application structure.

The current application separates startup, infrastructure, middleware, domain modules, and HTTP server configuration.

## Application Layers

### cmd/api

Application entry point.

Responsible for initializing dependencies and starting the HTTP application.

### internal/config

Application configuration.

### internal/database

PostgreSQL connection and database integration.

### internal/cache

Redis connection and cache integration.

### internal/middleware

Cross-cutting HTTP middleware, including security, request handling, rate limiting, and proxy-related controls.

### internal/server

Fiber application initialization, global middleware registration, health endpoints, readiness checks, and route registration.

### internal/modules

Business domain boundaries.

Each module owns functionality related to its domain.

## Request Lifecycle

```text
HTTP Request
     |
     v
Fiber Application
     |
     v
Global Middleware
     |
     v
Route Handler
     |
     v
Domain Service
     |
     v
Repository
     |
     v
PostgreSQL / Redis
```

The diagram describes the intended separation of responsibilities. Not every module necessarily implements every layer yet.

## Current Authentication Routes

- POST /api/v1/auth/register
- POST /api/v1/auth/login

## Infrastructure

- PostgreSQL: persistent relational data
- Redis: rate limiting and cache infrastructure
- JWT: access token generation and validation
