# Task Manager — Backend

Team task manager for your team. Go, Gin, PostgreSQL, Docker.

PR 1 is scaffold only: the server starts and answers `/health`.
PR 2 adds the users migration and the User model.
Auth comes in PR 3. Tasks, projects and comments come after that,
one step at a time.

## Run locally

```bash
go run ./cmd/server
# or
docker compose up --build
```

## Migrations

Migrations are always given as an UP/DOWN pair. Apply them with `psql`:

```bash
# start Postgres first
docker compose up -d db
# up
psql "$DATABASE_URL" -f migrations/001_create_users.up.sql
# down (drops the users table, restores the previous schema)
psql "$DATABASE_URL" -f migrations/001_create_users.down.sql
```

## Endpoints

- `GET /health` — `{"status":"ok"}`

## Checks

```bash
gofmt -l . && go vet ./... && go test ./...
```
