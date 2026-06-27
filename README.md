# Gateway service

A high-performance HTTP perimeter gateway written in Go using the Fiber framework.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.26.4-00add8?style=flat-square&logo=go)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-Framework-244c5a?style=flat-square)](https://gofiber.io/)
[![Redis](https://img.shields.io/badge/Database-Redis-DC382D?style=flat-square&logo=redis)](https://redis.io/)

## Overview

The gateway service acts as the public-facing RESTful perimeter for the platform. Built on the Fiber framework, it aggregates incoming HTTP requests from the frontend UIs and forwards them to the internal gRPC microservices (`auth`, `manager`). It provides a centralized layer for security, CORS handling, and sliding-window rate-limiting.

## Architecture

This section explains the technologies and physical layout of the gateway.

- **Framework**: `github.com/gofiber/fiber/v3` for HTTP routing
- **Middlewares**: Integrated payload compression, CORS, and rate limiting
- **Storage**: Connects to Redis to back the sliding-window rate limiter
- **gRPC clients**: Maintains persistent connections to internal services
- **Logging**: Idiomatic structured logging via `log/slog`

### Project structure

- `db/`: Database and cache connection logic
- `grpc/`: Setup and clients for communicating with backend gRPC services
- `middlewares/`: Custom Fiber HTTP interceptors
- `routes/`: HTTP route definitions mapping to internal clients
- `schema/`: Data validation and structuring schemas
- `utils/`: Helper utilities
- `main.go`: Gateway entrypoint and middleware orchestration
- `go.mod`: Dependency management

## Features

This section outlines the capabilities of the gateway.

- **High-performance routing**: Exposes an HTTP API built on top of Fiber.
- **Rate limiting**: Uses a sliding-window rate limiter backed by Redis.
- **Security middlewares**: Uses strict CORS policies.
- **Perimeter security**: Secures routes using cookie-based sessions (backed by Redis) for the admin console, and API key validation middleware for server-to-server communication (e.g., the Next.js frontend portfolio backend).
- **Request validation**: Rejects malformed REST payloads early using custom middleware.
- **gRPC proxying**: Translates REST payloads to internal gRPC contracts.
- **Graceful shutdown**: Drains active HTTP requests and closes gRPC client connections upon OS interrupts.

## API summary

The gateway implements a unified REST interface that orchestrates interactions for the platform:

- **Auth routes**: Login, registration, and API key management endpoints
- **Portfolio routes**: Endpoints for managing user profiles, experiences, projects, messages, and technologies
- **Public vs protected**: Segregates public read endpoints from secured write endpoints

## Getting started

This section explains how to run the gateway locally.

### Prerequisites

- [Go](https://golang.org/dl/) 1.26.4 or higher
- A running [Redis](https://redis.io/download/) server
- Running instances of internal gRPC services (`auth`, `manager`)

### Configuration

Export these variables directly in your shell environment:

| Variable | Description | Required |
| :--- | :--- | :---: |
| `REDIS_URL` | Connection string for Redis | **Yes** |
| `DEVELOPMENT` | Set to `"true"` to enable looser defaults | No |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed origins | Yes (if not in dev) |
| `PORT` | The port on which the HTTP gateway will listen (Default: `3000`) | No |
| `AUTH_SERVICE_URL` | Internal gRPC address of the auth service | **Yes** |
| `MANAGER_SERVICE_URL` | Internal gRPC address of the manager service | **Yes** |

### Running the service

Run the service using the Go CLI:

```bash
go run main.go
```

To watch for file changes and automatically rebuild and restart the server, use [air](https://github.com/cosmtrek/air):

```bash
go install github.com/air-verse/air@latest
air
```

## Deployment

The service is containerized and leverages GitHub Actions (`.github/workflows/main.yml`) for automated builds and deployment.
