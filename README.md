# user-service

Authentication and user profile management for Snaply.

## Environment Variables

| Variable                   | Default                                                                    | Description                         |
|----------------------------|----------------------------------------------------------------------------|-------------------------------------|
| `DATABASE_URL`             | `postgres://snaply:snaply_secret@localhost:5432/users?sslmode=disable`    | PostgreSQL connection string        |
| `JWT_SECRET`               | `dev_secret_change_in_production`                                          | Shared HS256 signing key            |
| `JWT_ACCESS_TOKEN_MINUTES` | `15`                                                                       | Access token lifetime (minutes)     |
| `JWT_REFRESH_TOKEN_DAYS`   | `7`                                                                        | Refresh token lifetime (days)       |
| `SERVER_PORT`              | `8081`                                                                     | HTTP listen port                    |

## Endpoints

| Method | Path                        | Auth     | Description                          |
|--------|-----------------------------|----------|--------------------------------------|
| POST   | /api/v1/auth/register       | —        | Create account                       |
| POST   | /api/v1/auth/login          | —        | Get access + refresh token           |
| POST   | /api/v1/auth/refresh        | —        | Rotate refresh token                 |
| POST   | /api/v1/auth/logout         | —        | Revoke refresh token                 |
| GET    | /api/v1/users/:username     | —        | Public profile                       |
| PUT    | /api/v1/users/me            | X-User-ID header | Update own profile          |
| GET    | /api/v1/users/search?q=     | —        | Search users (cursor paginated)      |
| GET    | /health                     | —        | Health check                         |

## How to Run Locally

**Prerequisites:** Go 1.22, Docker, [golang-migrate CLI](https://github.com/golang-migrate/migrate)

```bash
# 1. Start a local Postgres
docker run -d --name snaply-pg -e POSTGRES_USER=snaply -e POSTGRES_PASSWORD=snaply_secret -e POSTGRES_DB=users -p 5432:5432 postgres:16-alpine

# 2. Run migrations
make migrate-up

# 3. Start the service
make run
```

Or run everything via Docker Compose:

```bash
docker compose up --build
```

## Testing

```bash
make test
```
