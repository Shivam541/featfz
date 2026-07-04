# feat-manager

Go backend for the feature flag monorepo.

## Commands

1. `make setup`
2. `make deps-up`
3. `make migrate-up`
4. `make seed-phase2`
5. `make test`
6. `make run`

`make test`, `make run`, and `make build` use a repo-local Go build cache under `.cache/go-build`.

## Phase 2 auth slice

Phase 2 adds tenant-authenticated request handling before the flag APIs land.

- Required headers:
  - `Authorization: Bearer <jwt>`
  - `X-App-ID: <app-id>`
- JWT behavior:
  - `app_id` is required.
  - `exp` is enforced when present.
  - `iat` can already be included by clients now.
- Protected verification route:
  - `GET /v1/auth/check`

Sample response:

```json
{
  "success": true,
  "data": {
    "tenant_id": 1,
    "app_id": "app-acme",
    "subject": "user_123"
  }
}
```

## Seed data

Phase 2 includes a tenant seed file at `db/seeds/phase2_tenants.sql`.

The default seed creates:

- `app-acme` with secret `acme-secret`
- `app-globex` with secret `globex-secret`

Load it with:

1. `make deps-up`
2. `make migrate-up`
3. `make seed-phase2`

## Example curl

```bash
curl http://localhost:8080/v1/auth/check \
  -H "Authorization: Bearer $TENANT_JWT" \
  -H "X-App-ID: app-acme"
```

## Phase 3 persistence slice

Phase 3 adds the flag and flag-user override tables plus the repository layer that will back the flag APIs.

- Migrations:
  - `db/migrations/000003_phase3_flags.up.sql`
  - `db/migrations/000003_phase3_flags.down.sql`
- Repositories:
  - `internal/dao/flag_repository.go`
  - `internal/dao/flag_user_override_repository.go`
- DB-backed integration tests:
  - set `TEST_DB_DSN` to a MySQL DSN,
  - run `make deps-up`,
  - run `make migrate-up`,
  - run `go test ./...`.

Example DSN:

```bash
export TEST_DB_DSN='feat_manager:feat_manager@tcp(127.0.0.1:3306)/feat_manager?parseTime=true&multiStatements=true'
```

## Test command

```bash
make test
```
