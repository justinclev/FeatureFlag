# Feature Flags Platform

A microservices platform for managing and distributing feature flags, composed of three services backed by MongoDB.

```
┌─────────────────────────────────────────────────────┐
│  Browser                                            │
│  http://localhost:4200                              │
│         │                                           │
│    ┌────▼────────────┐                              │
│    │   dashboard     │  Angular 19 / nginx          │
│    └────┬────────────┘                              │
│         │ REST                                      │
│    ┌────▼────────────┐     ┌──────────────────────┐ │
│    │  dashboard-api  │     │    feature-api        │ │
│    │  FastAPI :8000  │     │    Go API     :8080   │ │
│    └────┬────────────┘     └──────────┬───────────┘ │
│         │                             │              │
│    ┌────▼─────────────────────────────▼───────────┐ │
│    │              MongoDB :27017                   │ │
│    └───────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

## Services

| Service         | URL                          | Description                      |
| --------------- | ---------------------------- | -------------------------------- |
| `dashboard`     | http://localhost:4200        | Angular management UI            |
| `dashboard-api` | http://localhost:8000/docs   | FastAPI management control plane |
| `feature-api`   | http://localhost:8080/health | Go distribution data plane       |
| `mongo`         | mongodb://localhost:27017    | MongoDB 7                        |
| `redis`         | redis://localhost:6379       | Redis 7 (flag read cache)        |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) ≥ 24 & Docker Compose ≥ 2.20
- [Node.js](https://nodejs.org/) 22 (for local dashboard dev)
- [Python](https://www.python.org/) 3.12 (for local dashboard-api dev)
- [Go](https://go.dev/) 1.23 (for local feature-api dev)

## Quick Start (Docker)

```bash
# 1. Clone and enter the repo
git clone <your-repo-url> && cd FeatureFlags

# 2. Build and start all services
make up

# 3. Open the dashboard
open http://localhost:4200
```

To stop everything:

```bash
make down
```

To stop and wipe all volumes (MongoDB data):

```bash
make clean
```

## Local Development

### dashboard (Angular)

```bash
cd dashboard
npm install
npm start          # http://localhost:4200
```

### dashboard-api (FastAPI)

```bash
cd dashboard-api
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
uvicorn app.main:app --reload --port 8000
# Swagger UI: http://localhost:8000/docs
```

### feature-api (Go)

```bash
cd feature-api
go mod tidy        # downloads dependencies and generates go.sum
go run ./...       # http://localhost:8080
```

> **Note:** Run `go mod tidy` at least once before `docker compose build` so that `go.sum` is committed.

## Environment Variables

### dashboard-api

| Variable          | Default                     | Description                     |
| ----------------- | --------------------------- | ------------------------------- |
| `MONGO_URI`       | `mongodb://localhost:27017` | MongoDB connection string       |
| `FEATURE_API_URL` | `http://localhost:8080`     | Internal URL of the feature-api |

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
- `.github/workflows/dashboard-api-ci.yml`
- `.github/workflows/feature-api-ci.yml`
