# user-service

Authentication and user profile management for Snaply. Go 1.22, port 8081, own Postgres DB `users`.

## Layout

`cmd/main.go` → `internal/handler` (chi routes, `routes.go`) → `internal/service` (`auth_service.go`, `user_service.go`) → `internal/repository` (`user.go`, sqlx + pgx/v5) → `internal/model`

## Endpoints

| Method | Path                    | Auth              | Description                     |
|--------|-------------------------|-------------------|----------------------------------|
| POST   | /api/v1/auth/register   | —                 | Create account                   |
| POST   | /api/v1/auth/login      | —                 | Get access + refresh token       |
| POST   | /api/v1/auth/refresh    | —                 | Rotate refresh token             |
| POST   | /api/v1/auth/logout     | —                 | Revoke refresh token             |
| GET    | /api/v1/users/search    | —                 | Cursor-paginated user search     |
| PUT    | /api/v1/users/me        | X-User-ID header  | Update own profile               |
| GET    | /api/v1/users/{username}| —                 | Public profile                   |
| GET    | /health                 | —                 | Health check                     |

## Conventions

- JWT: HS256, shared `JWT_SECRET` with `api-gateway`. Access token 15 min, refresh token 7 days (`JWT_ACCESS_TOKEN_MINUTES` / `JWT_REFRESH_TOKEN_DAYS`).
- Migrations via `golang-migrate`, files in `migrations/`. Run with `make migrate-up`.
- Downstream trust: this service does not re-validate JWTs from `api-gateway` requests where `X-User-ID` is already set — the gateway is the sole JWT boundary.
- Tests use a hand-rolled `mock_repository.go` in `internal/service` (no mockgen).

## Running

```bash
docker run -d --name snaply-pg -e POSTGRES_USER=snaply -e POSTGRES_PASSWORD=snaply_secret -e POSTGRES_DB=users -p 5432:5432 postgres:16-alpine
make migrate-up
make run    # or: docker compose up --build
make test
```

Default `DATABASE_URL`: `postgres://snaply:snaply_secret@localhost:5432/users?sslmode=disable`
