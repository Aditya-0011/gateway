# Gateway service

A high-performance HTTP perimeter gateway written in Go using the Fiber framework.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.26.4-00add8?style=flat-square&logo=go)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-Framework-244c5a?style=flat-square)](https://gofiber.io/)
[![Redis](https://img.shields.io/badge/Database-Redis-DC382D?style=flat-square&logo=redis)](https://redis.io/)

## Overview

The gateway service acts as the public-facing RESTful perimeter for the platform. Built on the Fiber framework, it aggregates incoming HTTP requests from the frontend UIs and forwards them to the internal gRPC microservices (`auth`, `manager`, `wallet`). It provides a centralized layer for security, CORS handling, and sliding-window rate-limiting.

## Architecture

This section explains the technologies and physical layout of the gateway.

- **Framework**: `github.com/gofiber/fiber/v3` for HTTP routing
- **Middlewares**: panic recovery, CORS, security headers, rate limiting, authentication with an Origin check for cookie sessions, and request validation
- **Storage**: Connects to Redis for cookie sessions (`session:{u<id>}:<sid>`, 15-minute idle timeout, 12-hour maximum lifetime), cached identities (`session:{u<id>}:authz`, 10 minutes), cached API keys (`apikey:<hash>`, 5 minutes), the set of users with cached entries (`authz:users`, 15 minutes, refreshed on every write), API key origins (`cors:key-origins`, 5 minutes), and the sliding-window rate limiter (`ratelimit:*`)
- **gRPC clients**: Maintains persistent connections to internal services
- **Logging**: Idiomatic structured logging via `log/slog`

### Project structure

- `db/`: Database and cache connection logic
- `grpc/`: Setup and clients for communicating with backend gRPC services
- `middlewares/`: Custom Fiber HTTP interceptors
- `routes/`: HTTP route definitions mapping to internal clients
- `schema/`: Data validation and structuring schemas
- `internal/`: Config loading, crypto helpers, and protobuf JSON encoding
- `main.go`: Gateway entrypoint and middleware orchestration
- `go.mod`: Dependency management

## Features

This section outlines the capabilities of the gateway.

- **High-performance routing**: Exposes an HTTP API built on top of Fiber.
- **Rate limiting**: Uses a sliding-window rate limiter backed by Redis.
  - Session and anonymous requests: 60 per minute per client IP.
  - API-key requests skip the IP bucket and count against their key owner instead: 600 per minute. A key that is not cached yet costs an auth-service lookup, so lookups are also capped at 30 per minute per client IP.
  - `/login` and `/reset-password`: 5 per minute per IP, plus 10 per 15 minutes per account (keyed by the SHA-256 of the lower-cased email). Other credential routes keep the per-IP limit.
  - `POST /message/add`: 5 per hour per visitor and 100 per hour per key owner.
  - `POST /chat/allow` lets the Portfolio check its chat limits: 30 questions per visitor and 300 per key owner, each per hour.
  - The visitor is the client IP (IPv6 grouped by /64). A portfolio server can pass the real visitor in `X-Visitor-IP`, but the gateway honours it only for a server-only key (one with no origins) on a request without an `Origin` header.
- **Security middlewares**: Uses strict CORS policies.
- **Perimeter security**: Cookie sessions (backed by Redis, `__Secure-session` scoped to `API_DOMAIN`) for the UIs, and per-user API keys for portfolios. An `X-API-KEY` header always takes the key path and never falls back to the cookie. Keys only reach Manager's portfolio routes (`/user/details`, `/technology/list`, `/project/list`, `/experience/list`, `/message/add` and `/chat/allow`); every other route requires a session. Disabling or deleting a user also drops their cached API key, so the key stops working at once. Ending a session deletes `session:{u<id>}:authz` only when it was the user's last live session, so other devices keep it.
- **Two CORS lanes**: `CORS_ALLOWED_ORIGINS` (first-party UIs) get credentialed CORS. Any other origin is considered only on the portfolio routes of the Manager host, without credentials, if some key lists it (per-key origins live in the database and are cached in the Redis set `cors:key-origins`, rebuilt from the auth service); the request is then rejected unless the presented key lists that origin. A third-party page therefore never receives a cookie-backed response.
- **Request validation**: Rejects malformed REST payloads early using custom middleware.
- **gRPC proxying**: Translates REST payloads to internal gRPC contracts.
- **Health check**: `GET /healthz` returns `200` when Redis answers a ping within one second, otherwise `503`. It answers only the exact path `/healthz` (other spellings such as `/healthz/` or `/HEALTHZ` get the normal 404), skips logging and rate limiting, and HAProxy blocks it from the internet.
- **Panic recovery**: A handler panic is logged with its stack trace and answered with `500`; the process keeps running.
- **Graceful shutdown**: Drains active HTTP requests and closes gRPC client connections upon OS interrupts.

## API summary

The gateway implements a unified REST interface that orchestrates interactions for the platform:

- **Auth routes**: Login, first login, logout, session check and revocation, password change, recovery codes, password reset, admin user and project management (`/admin/*`), and API key management (`/key`, `/rotate-key`, `/key/revoke`, `/key/origins`)
- **Portfolio routes**: Endpoints for managing user profiles, experiences, projects, messages, and technologies, plus the chatbot's model list (`/chat/models`) and visitor limit check (`/chat/allow`)
- **Wallet routes**: Endpoints for managing financial transactions and balances
- **API key vs session**: The six portfolio routes accept an API key or a session; everything else requires a session, except login and password reset

## Getting started

This section explains how to run the gateway locally.

### Prerequisites

- [Go](https://golang.org/dl/) 1.26.4 or higher
- A running [Redis](https://redis.io/download/) server
- Running instances of internal gRPC services (`auth`, `manager`, `wallet`)

### Configuration

Export these variables directly in your shell environment:

| Variable | Description | Required |
| :--- | :--- | :---: |
| `REDIS_URL` | Redis connection string (sessions, API key cache, rate limits) | **Yes** |
| `DOMAIN` | Apex domain | **Yes** (except dev) |
| `API_DOMAIN` | Parent of the API hosts, defaults to `api.<DOMAIN>`. Services are served on `<service>.<API_DOMAIN>` and the `__Secure-session` cookie is scoped to it, so UI and Portfolio hosts never receive it | No |
| `CORS_ALLOWED_ORIGINS` | Comma-separated `https://` origins of the UIs. Also the allowlist for the CSRF `Origin` check | **Yes** (except dev) |
| `AUTH_SERVICE_ADDR` | `host:port` of the auth service | **Yes** |
| `MANAGER_SERVICE_ADDR` | `host:port` of the manager service | **Yes** |
| `WALLET_SERVICE_ADDR` | `host:port` of the wallet service | **Yes** |
| `INTERNAL_AUTH_TOKEN` | Shared secret (32+ chars) sent to every service as a bearer token | **Yes** |
| `INTERNAL_TLS_CA` | PEM CA bundle; enables TLS to the services when set. The service certificates must then match the host in `*_SERVICE_ADDR`: with `<APP_IP>:<port>` they need `APP_IP` as an IP SAN, or set `INTERNAL_TLS_SERVER_NAME` | No |
| `INTERNAL_TLS_SERVER_NAME` | Name to verify the service certificates against instead of the address host (for example a DNS SAN shared by all services) | No |
| `DEVELOPMENT` | `true` serves all routes on any host, relaxes cookie `Secure` and uses localhost CORS defaults | No |
| `BIND_ADDR` | Address to listen on. Defaults to `127.0.0.1` in production (HAProxy runs on the same host) and to all interfaces in development | No |
| `PORT` | HTTP port (default `3000`) | No |

### Running the service

Run the service using the Go CLI:

```bash
go run main.go
```

To watch for file changes and automatically rebuild and restart the server, use [air](https://github.com/air-verse/air):

```bash
go install github.com/air-verse/air@latest
air
```

## Deployment

The service is containerized and leverages a manually triggered GitHub Actions workflow (`.github/workflows/main.yml`, `workflow_dispatch`) for builds and deployment with Podman to the edge box, where HAProxy terminates TLS in front of it.
