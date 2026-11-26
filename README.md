# Task Manager — Backend

Team task manager for your team. Go, Gin, PostgreSQL, Docker.

PR 1 is scaffold only: the server starts and answers `/health`.
PR 2 adds the users migration and the User model.
PR 3 adds auth: register, login and me with JWT.
Tasks, projects and comments come after that, one step at a time.

## Run locally

```bash
# load env (DATABASE_URL + JWT_SECRET are required now)
set -a; source .env; set +a
go run ./cmd/server
# or
docker compose up --build
```

## Migrations

Migrations are always given as an UP/DOWN pair. `docker compose up`
applies them automatically with the `migrate` service: it waits for
the db to be healthy, then runs `up`.

```bash
docker compose up -d
docker compose logs migrate --tail 20
# check the table
docker compose exec db psql -U taskmanager -d taskmanager -c "\d users"
```

To roll back manually with `psql` (drops the users table, restores
the previous schema):

```bash
psql "$DATABASE_URL" -f migrations/001_create_users.down.sql
```

## Auth

- `POST /auth/register` — `{"name","email","password"}` (password at least 8 chars) → 201 with `{"user","token"}`
- `POST /auth/login` — `{"email","password"}` → 200 with `{"user","token"}`; wrong email or password both give 401 without saying which
- `GET /auth/me` — needs `Authorization: Bearer <token>` → 200 with the user, no token needed in the body

Test with curl (after `docker compose up`):

```bash
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Mohsen","email":"mohsen@example.com","password":"password123"}'
# copy the token from the answer, then:
curl http://localhost:8080/auth/me \
  -H 'Authorization: Bearer <token>'
# and without a token it must be 401:
curl http://localhost:8080/auth/me
```

The server refuses to start without `JWT_SECRET`. In Docker Compose a
dev-only default is set; set a real `JWT_SECRET` in `.env` for anything
beyond local testing.

## Endpoints

- `GET /health` — `{"status":"ok"}`

## Checks

```bash
gofmt -l . && go vet ./... && go test ./...
```
