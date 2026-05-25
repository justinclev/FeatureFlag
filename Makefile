.PHONY: build up down logs clean restart \
        dev-dashboard dev-dashboard-api dev-feature-api

# ── Docker Compose ──────────────────────────────────────────────────────────

build:
	docker compose build

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --remove-orphans

restart: down up

# ── Local dev shortcuts ──────────────────────────────────────────────────────

dev-dashboard:
	cd dashboard && npm start

dev-dashboard-api:
	cd dashboard-api && uvicorn app.main:app --reload --port 8000

dev-feature-api:
	cd feature-api && go run ./...

# ── Data & Testing ──────────────────────────────────────────────────────────

seed-flags:
	go run scripts/seed_flags.go

test-flags:
	go run scripts/test_flags.go

test-all: seed-flags test-flags
