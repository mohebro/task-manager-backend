# Task Manager — Backend

Team task manager for your team. Go, Gin, PostgreSQL, Docker.

PR 1 is scaffold only: the server starts and answers `/health`.
Auth comes in PR 3, after the users migration in PR 2.
Tasks, projects and comments come after that, one step at a time.

## Run locally

```bash
go run ./cmd/server
# or
docker compose up --build
```

## Endpoints

- `GET /health` — `{"status":"ok"}`

## Checks

```bash
gofmt -l . && go vet ./... && go test ./...
```
