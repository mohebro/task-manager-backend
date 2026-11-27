# Task Manager — Backend

Team task manager for your team. Go, Gin, PostgreSQL, Docker.

PR 1 is scaffold only: the server starts and answers `/health`.
PR 2 adds the users migration and the User model.
PR 3 adds auth: register, login and me with JWT.
PR 4 adds Swagger docs for the auth endpoints.
PR 5 adds structured logging and handler/middleware/config tests.
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

## API docs (Swagger)

After `docker compose up`, open http://localhost:8080/swagger/index.html —
all three auth endpoints are there, and `/auth/me` can be tried with the
Authorize button (type `Bearer <token>`).

Docs are generated from the annotations on `main` and the auth handlers.
After changing an annotation, regenerate — never edit `docs/` by hand:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
swag init -g cmd/server/main.go
```

Trying the secured endpoint (`GET /auth/me`): it takes no body or
query input — the user id comes from the token, so it correctly shows
no parameters. The token goes in through the **Authorize** button at
the top of the page, not through the endpoint itself:

1. Try `POST /auth/register` (or `/auth/login`) and copy the token.
2. Click **Authorize** at the top, enter `Bearer <token>` — the word
   `Bearer`, a space, then the token. Just the token alone gives 401,
   because this auth type sends the value verbatim, without adding
   `Bearer` for you.
3. Try `GET /auth/me` — it sends `Authorization: Bearer <token>`.

The `swag` CLI and the `swag` library must be the same version (v1.16.4):
a newer CLI generates `docs.go` the older library cannot compile.

Annotations are read at generation time, so they cannot use `.env`
variables: `@host localhost:8080` is baked into `docs/` when you run
`swag init`. The runtime override is `SWAGGER_HOST` — when set, `main`
writes it into `docs.SwaggerInfo.Host` before the server starts, so the
UI calls the right host on a server. Leave it empty locally.

## Logging

Logs are one JSON line per request, to stdout:

```json
{"time":"...","level":"INFO","msg":"http request","request_id":"...","method":"POST","path":"/auth/login","status":200,"duration_ms":12}
```

Every request gets an id (`X-Request-ID` response header) so a client
report can be matched to a log line. A 500 is always logged server-side
with the full error and the request id; the client sees the exact error
only when `APP_ENV=dev`, otherwise a generic message plus the request
id. Passwords and tokens are never logged. `LOG_LEVEL` is `debug`,
`info` (default), `warn` or `error`.

Watch the logs while testing:

```bash
docker compose logs -f app
```

## Endpoints

- `GET /health` — `{"status":"ok"}`

## Checks

```bash
gofmt -l . && go vet ./... && go test ./...
```
