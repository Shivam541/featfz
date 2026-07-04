# feat-manager

Go backend for the feature flag monorepo.

## Commands

1. `make setup`
2. `make deps-up`
3. `make migrate-up`
4. `make seed-phase2`
5. `make test`
6. `make run`
7. `make smoke-eval`
8. `make hooks-install`
9. `make migrate-status`
10. `make migrate-create NAME=<name>`
11. `make seed-phase8`

`make test`, `make run`, and `make build` use a repo-local Go build cache under `.cache/go-build`.
`make smoke-eval` expects the API to be running locally on `127.0.0.1:8080`.
`make hooks-install` sets `core.hooksPath` to `.githooks` so the local pre-push hook blocks pushes when tests fail.
`make migrate-status` checks the current migration state.
`make migrate-create NAME=<name>` scaffolds a new SQL migration under `db/migrations`.
`make seed-phase8` loads repeatable manual-testing flags and overrides for both tenants.

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

## Phase 7 evaluation slice

Phase 7 adds the runtime read path for a single flag and user.

- Route:
  - `GET /eval?flag=<flagKey>&user=<userID>`
- Behavior:
  - returns the flag override when one exists,
  - otherwise falls back to the flag default,
  - `flag` and `user` are required query parameters,
  - the response is `{"success": true, "result": "on"|"off"}`.
- Expected responses:
  - `200 OK` when evaluation succeeds,
  - `400 Bad Request` when the query is missing or malformed,
  - `404 Not Found` when the tenant flag does not exist,
  - `503 Service Unavailable` for unexpected dependency failures.

Example curl:

```bash
curl "http://localhost:8080/eval?flag=new_dashboard&user=user_123" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-App-ID: $APP_ID"
```

## Phase 8 hardening slice

Phase 8 starts with a local smoke workflow for the eval path.

- Command:
  - `make smoke-eval`
- Hook:
  - `make hooks-install`
- Migration helpers:
  - `make migrate-status`
  - `make migrate-create NAME=<name>`
- Manual seed:
  - `make seed-phase8`
- What it does:
  - generates JWTs for `app-acme` and `app-globex`,
  - creates the same smoke flag in both tenants,
  - applies a user override in `acme`,
  - prints the live eval results for both tenants.
- What the hook does:
  - runs `make test` before every push,
  - exits non-zero if tests fail,
  - blocks the push when the test suite is red.
- What the migration helpers do:
  - `make migrate-status` prints whether the local database is up to date,
  - `make migrate-create NAME=<name>` creates a new SQL migration file pair.
- What the manual seed does:
  - loads `phase8_manual_default_off` and `phase8_manual_default_on` for both tenants,
  - adds an `acme` override for `user_123` on the default-off flag,
  - gives you a repeatable DB state for manual eval testing.
- Requirements:
  - backend running on `127.0.0.1:8080`,
  - MySQL running with the seeded tenants.

To seed the manual fixtures after migrations:

```bash
make seed-phase8
```

You can also override the generated flag key:

```bash
SMOKE_FLAG_KEY=phase8_eval_smoke_custom make smoke-eval
```

## Test command

```bash
make test
```

## Test Matrix

- `make test`
  - runs the Go test suite with the repo-local build cache.
- `make smoke-eval`
  - runs a live eval smoke flow against `http://127.0.0.1:8080`.
- `make hooks-install`
  - installs the local pre-push hook that blocks pushes on failing tests.
- `make seed-phase8`
  - loads repeatable manual testing data for both tenants.
- `make migrate-status`
  - checks whether the database is at the expected migration state.

## Phase 9 Review

Phase 9 is the pre-MVP review pass.

- What was checked:
  - all documented v1 endpoints exist in the router,
  - tenant scope is enforced in middleware, services, and DAO queries,
  - archive behavior is soft and excluded from active reads,
  - evaluation falls back to the flag default when no override exists,
  - error responses stay on generic client-safe envelopes,
  - tests cover the main success and failure paths,
  - docs reflect the commands and API behavior we actually use.
- Review result:
  - no blocking MVP gaps were found in the implemented v1 surface.
- Deferred work:
  - keep the `.github/workflows/go.yml` file review separate if it is meant to be adopted,
  - revisit any post-v1 feature ideas only after the MVP is stable.

MVP-ready note:

The backend is ready for the documented v1 workflow: local boot, tenant-authenticated management APIs, user overrides, eval, and the supporting test and seed commands.
