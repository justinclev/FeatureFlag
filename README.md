# Feature Flags Platform

A microservices platform for managing and distributing feature flags, composed of two services backed by MongoDB.

```
┌─────────────────────────────────────────────────────┐
│  Browser                                            │
│  http://localhost:4200                              │
│         │                                           │
│    ┌────▼────────────┐                              │
│    │   dashboard     │  Angular 19 / nginx          │
│    └────┬────────────┘                              │
│         │ REST                                      │
│    ┌────▼────────────┐                              │
│    │    feature-api  │                              │
│    │    Go API :8081 ├────────┐                     │
│    └──────────┬──────┘        │                     │
│               │               │                     │
│    ┌──────────▼───────┐  ┌────▼─────────────┐       │
│    │  MongoDB :27018  │  │   Redis :6380    │       │
│    └──────────────────┘  └──────────────────┘       │
└─────────────────────────────────────────────────────┘
```

## Services

| Service         | URL                          | Port (Host) | Description                      |
| --------------- | ---------------------------- | ----------- | -------------------------------- |
| `dashboard`     | http://localhost:4200        | 4200        | Angular management UI            |
| `feature-api`   | http://localhost:8081/health | 8081        | Go distribution data plane       |
| `mongo`         | localhost:27018              | 27018       | MongoDB 7 (Data Store)           |
| `redis`         | localhost:6380               | 6380        | Redis 7 (Flag Read Cache)        |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) ≥ 24 & Docker Compose ≥ 2.20
- [Node.js](https://nodejs.org/) 22 (for local dashboard dev)
- [Go](https://go.dev/) 1.23 (for local feature-api dev)

## Quick Start (Docker)

```bash
# 1. Clone and enter the repo
git clone <your-repo-url> && cd FeatureFlags

# 2. Build and start all services
docker compose up --build

# 3. Open the dashboard
open http://localhost:4200
```

> **Note:** The `feature-api` requires an `X-API-KEY` header for all requests except `/health`. The default key in Docker Compose is `test-api-key`.

To stop everything:

```bash
docker compose down
```

## Local Development

### dashboard (Angular)

```bash
cd dashboard
npm install
npm start          # http://localhost:4200
```

### feature-api (Go)

```bash
cd feature-api
go mod tidy        # downloads dependencies and generates go.sum
# Set required environment variables
export API_KEY=local-dev-key
go run ./...       # http://localhost:8080
```

> **Note:** Run `go mod tidy` at least once before `docker compose build` so that `go.sum` is committed.

## Environment Variables

### feature-api

| Variable                | Default                     | Description                               |
| ----------------------- | --------------------------- | ----------------------------------------- |
| `API_KEY`               | _(required)_                | Auth key for `X-API-KEY` header           |
| `PORT`                  | `8080`                      | HTTP listen port                          |
| `MONGO_URI`             | `mongodb://localhost:27017` | MongoDB connection string                 |
| `MONGO_DB_NAME`         | `feature_flags`             | MongoDB database name                     |
| `MONGO_COLLECTION_NAME` | `flags`                     | MongoDB collection name                   |
| `REDIS_ADDR`            | `localhost:6379`            | Redis address                             |
| `REDIS_PASSWORD`        | _(empty)_                   | Redis password (optional)                 |
| `REDIS_CACHE_PREFIX`    | `flags:id:`                 | Redis key prefix                          |
| `CACHE_TTL_SECONDS`     | `30`                        | Redis cache TTL                           |
| `LOG_LEVEL`             | `info`                      | Logging level (debug, info, warn, error)  |
| `REQUEST_TIMEOUT_MS`    | `5000`                      | Internal request timeout                  |
| `CORS_ALLOWED_ORIGIN`   | `http://localhost:4200`     | Allowed CORS origin                       |

## CI/CD

Each service has its own GitHub Actions workflow triggered on pushes to `main` affecting that service's directory:

- `.github/workflows/dashboard-ci.yml`
- `.github/workflows/feature-api-ci.yml`
