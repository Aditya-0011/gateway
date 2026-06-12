# API Gateway Service

A high-performance HTTP gateway written in Go using the Fiber framework.

[![Go Version](https://img.shields.io/badge/Go->=1.25.3-00add8?style=flat-square&logo=go)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-Framework-244c5a?style=flat-square)](https://gofiber.io/)
[![Redis](https://img.shields.io/badge/Database-Redis-DC382D?style=flat-square&logo=redis)](https://redis.io/)

## Overview

The Gateway Service acts as the public-facing entry point for the microservices infrastructure. Built on top of the exceptionally fast Fiber framework, it translates incoming HTTP/REST requests and forwards them to internal gRPC microservices (`auth`, `manager`). It provides a centralized layer for security, compression, CORS, and rate-limiting.

## Architecture & Tech Stack

- **Framework**: `github.com/gofiber/fiber/v3` for high-throughput HTTP routing.
- **Middlewares**: Integrated Helmet (security headers), Compress (payload optimization), CORS, and Rate Limiter.
- **Cache/Storage**: Connects to Redis to back the sliding-window rate limiter.
- **gRPC Clients**: Maintains persistent connections to internal services using compiled Protocol Buffer contracts.
- **Logging**: Idiomatic structured logging via `log/slog`.

### Project Structure

```text
.
├── db/            # Database and cache connection logic (e.g., Redis)
├── grpc/          # Setup and clients for communicating with backend gRPC services
├── middlewares/   # Custom Fiber HTTP interceptors
├── routes/        # HTTP route definitions mapping to internal clients
├── schema/        # Data validation and structuring schemas
├── utils/         # Helper utilities
├── main.go        # Gateway entrypoint and middleware orchestration
└── go.mod         # Dependency management
```

## Features

- ⚡ **High-Performance Routing**: Exposes a lightning-fast HTTP API built on top of Fiber.
- 🚦 **Rate Limiting**: Sliding-window rate limiter backed by Redis to prevent abuse.
- 🛡️ **Robust Security Middlewares**: Built-in HTTP security headers (Helmet), payload compression, and strict CORS policies.
- 🔐 **Authentication & Authorization**: JWT/Token validation and request authorization before forwarding to internal services.
- ✅ **Request Validation**: Custom validation middleware to reject malformed REST payloads early.
- 🔄 **gRPC Proxying**: Seamlessly translates REST payloads to tightly-coupled internal gRPC contracts.
- 🚦 **Graceful Shutdown**: Safely drains active HTTP requests and closes gRPC client connections upon OS interrupts.

## API Summary

The gateway implements a unified REST interface that orchestrates interactions across all internal microservices:
- **Auth Routes**: Login, registration, and API key management endpoints (proxied to `auth`).
- **Portfolio Routes**: Endpoints for managing user profiles, experiences, projects, and technologies (proxied to `manager`).
- **Public vs Protected**: Segregates public read endpoints from secured write endpoints using authentication middlewares.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.25.3 or higher
- A running [Redis](https://redis.io/download/) server (for rate limiting)
- Running instances of internal gRPC services (auth, manager)
- The `common` contracts module

### Configuration

The service relies on environment variables for configuration. You should export these directly in your shell environment.

| Variable | Description | Default | Required |
| :--- | :--- | :--- | :---: |
| `REDIS_URL` | Connection string for Redis | - | **Yes** |
| `DEVELOPMENT` | Set to `"true"` to enable looser defaults (like default CORS and IP proxying) | - | No |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed origins | - | Yes (if not in dev) |
| `PORT` | The port on which the HTTP gateway will listen | `3000` | No |
| `AUTH_SERVICE_URL` | Internal gRPC address of the Auth Service | - | **Yes** |
| `MANAGER_SERVICE_URL` | Internal gRPC address of the Manager Service | - | **Yes** |

### Polyrepo Local Setup

This project uses a polyrepo architecture. Services like `auth`, `manager`, and `gateway` rely on the `common` repository. To run this locally without dependency errors, clone all repositories side-by-side into the same parent directory so that relative paths resolve correctly.

Example local setup:
```text
infrastructure/
├── common/
├── auth/
├── manager/
└── gateway/
```

### Running the Service

For standard execution, use the Go CLI:

```bash
go run main.go
```

#### Hot Reloading for Local Development

To watch for file changes and automatically rebuild and restart the server, you can use [air](https://github.com/cosmtrek/air). Since `air` is written purely in Go, it compiles to a single binary and runs natively on Windows, macOS, and Linux without any OS-specific shell scripts.

```bash
go install github.com/air-verse/air@latest
air
```

## CI/CD & Production Deployment

The service is fully containerized and leverages GitHub Actions (`.github/workflows/main.yml`) for automated builds and deployment.

### 1. Server Initialization
Before the CI/CD pipeline can deploy the application for the first time, you must initialize the host environment on your VM.

Create the shared network that the containers will use to communicate (if you haven't already created it for other services):
```bash
podman network create infra-network
```

Create the environment variable file that the container will read on startup:
```bash
touch ~/gateway-production.conf
# Edit the file to include your REDIS_URL and internal gRPC URLs:
# REDIS_URL=redis://localhost:6379
# AUTH_SERVICE_URL=auth-service:7295
# MANAGER_SERVICE_URL=manager-service:7296
# CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

### 2. Automated Build (GHCR)
The workflow uses Docker Buildx with QEMU to cross-compile a `linux/arm64` container image. 
Because the project depends on private GitHub modules (`common`), the workflow injects a `PAT_TOKEN` secret during the build stage to securely download the private contracts without leaving credentials in the final image. The resulting image is pushed to the GitHub Container Registry (`ghcr.io/aditya-0011/gateway`).

### 3. Production Execution (Podman)
In production, the service is deployed to a remote Linux VM using **Podman**. The CI pipeline automatically SSHs into the server, pulls the latest GHCR image, and seamlessly swaps the running container.

The container is orchestrated with the following constraints:
- Attached to the shared internal network (`--network infra-network`), allowing it to communicate with internal gRPC services using their container names.
- Exposes port `3000` to the host VM to accept incoming public traffic (`-p 3000:3000`).
- Reads configuration purely from the host-level environment file (`--env-file ~/gateway-production.conf`).
- Configured to restart automatically (`--restart unless-stopped`).
