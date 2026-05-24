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
│    │    Go API :8081 │                              │
│    └──────────┬──────┘                              │
│               │                                     │
│    ┌──────────▼───────────────────────────┐         │
│    │              MongoDB :27018          │         │
│    └──────────────────────────────────────┘         │
└─────────────────────────────────────────────────────┘
```

## Services

| Service         | URL                          | Description                      |
| --------------- | ---------------------------- | -------------------------------- |
| `dashboard`     | http://localhost:4200        | Angular management UI            |
| `feature-api`   | http://localhost:8081/health | Go distribution data plane       |
| `mongo`         | mongodb://localhost:27018    | MongoDB 7                        |
| `redis`         | redis://localhost:6380       | Redis 7 (flag read cache)        |

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
go run ./...       # http://localhost:8080
```

> **Note:** Run `go mod tidy` at least once before `docker compose build` so that `go.sum` is committed.

## Environment Variables

### feature-api

| Variable         | Default                     | Description               |
| ---------------- | --------------------------- | ------------------------- |
| `MONGO_URI`      | `mongodb://localhost:27017` | MongoDB connection string |
| `REDIS_ADDR`     | `localhost:6379`            | Redis address             |
| `REDIS_PASSWORD` | _(empty)_                   | Redis password (optional) |
| `PORT`           | `8080`                      | HTTP listen port          |
| `MONGO_DB_NAME`  | `feature_flags`             | MongoDB database name     |

## CI/CD

Each service has its own GitHub Actions workflow triggered on pushes to `main` affecting that service's directory:

- `.github/workflows/dashboard-ci.yml`
- `.github/workflows/feature-api-ci.yml`
