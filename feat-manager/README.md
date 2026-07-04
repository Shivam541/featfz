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

## Phase 4 create-flag slice

Phase 4 adds the first full management flow: creating a tenant-scoped flag through authenticated HTTP and persisting it to MySQL.

- Route:
  - `POST /v1/flags`
- Request body:
  - `key`
  - `description`
  - `default_enabled`
- Expected responses:
  - `201 Created` on success
  - `400 Bad Request` for malformed or incomplete JSON
  - `409 Conflict` when the tenant already owns that key
- Integration test:
  - set `TEST_DB_DSN`,
  - run `make deps-up`,
  - run `make migrate-up`,
  - run `go test ./...`

Example curl:

```bash
curl -X POST http://localhost:8080/v1/flags \
  -H "Authorization: Bearer $TENANT_JWT" \
  -H "X-App-ID: $APP_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "new_dashboard",
    "description": "Enable the new dashboard experience",
    "default_enabled": false
  }'
```

## Phase 5 flag-management slice

Phase 5 completes the flag-management surface with tenant-scoped list, get, update, and archive routes.

- Routes:
  - `GET /v1/flags`
  - `GET /v1/flags/{flagKey}`
  - `PATCH /v1/flags/{flagKey}`
  - `DELETE /v1/flags/{flagKey}`
- Behavior:
  - reads only return active flags for the authenticated tenant,
  - `PATCH` accepts `description` and/or `default_enabled`,
  - archived flags are hidden from list/get/update and return `404` on read paths,
  - archive is soft, not hard delete.
- Expected responses:
  - `200 OK` on list, get, update, and archive success,
  - `400 Bad Request` for invalid path or body input,
  - `404 Not Found` when the tenant flag does not exist or is already archived.
- Verification:
  - run `go test ./...`
  - if you want DB-backed coverage, set `TEST_DB_DSN`, run `make deps-up`, run `make migrate-up`, and then run `go test ./...`

Example curl calls:

JWTs for these requests can be generated with the Node helper in `feat-client/scripts/generate-jwt.mjs`.
It defaults to `JWT_SECRET=acme-secret`, `APP_ID=app-acme`, and `JWT_SUBJECT=smoke-user`, so you can run:

```bash
TOKEN=$(node ../feat-client/scripts/generate-jwt.mjs)
```

Set `JWT_SUBJECT` if you want a different subject, then reuse `$TOKEN` in the curls below.

```bash
curl http://localhost:8080/v1/flags \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-App-ID: $APP_ID"
```

```bash
curl http://localhost:8080/v1/flags/new_dashboard \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-App-ID: $APP_ID"
```

```bash
curl -X PATCH http://localhost:8080/v1/flags/new_dashboard \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-App-ID: $APP_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated rollout",
    "default_enabled": true
  }'
```

```bash
curl -X DELETE http://localhost:8080/v1/flags/new_dashboard \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-App-ID: $APP_ID"
```

## Phase 6 per-user override slice

Phase 6 adds the bulk per-user override write flow for a flag.

- Route:
  - `POST /v1/flags/{flagKey}/users/bulk-set`
- Behavior:
  - repeated `user_id` entries in one request are deduped,
  - last value wins when the same `user_id` appears more than once,
  - writes stay tenant-scoped,
  - invalid override entries return `422 Unprocessable Entity`.
- Expected responses:
  - `200 OK` when overrides are applied,
  - `400 Bad Request` when the body is malformed,
  - `404 Not Found` when the tenant flag does not exist,
  - `422 Unprocessable Entity` for invalid override entries.

Example curl:

```bash
curl -X POST http://localhost:8080/v1/flags/new_dashboard/users/bulk-set \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-App-ID: $APP_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "overrides": [
      {"user_id": "user_123", "enabled": true},
      {"user_id": "user_456", "enabled": false},
      {"user_id": "user_123", "enabled": false}
    ]
  }'
```

## Test command

```bash
make test
```
