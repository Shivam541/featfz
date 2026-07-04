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

## Test command

```bash
make test
```
