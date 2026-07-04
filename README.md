# featfz

Monorepo for the feature flag platform.

Apps in this repo:

- `feat-manager/` for the Go backend
- `feat-ui/` for the UI
- `feat-client/` for the client-side showcase

## Phase 0 status

The repo now includes the `feat-manager/` bootstrap needed to start local development:

- Go API entrypoint in `feat-manager/cmd/api`
- config loading from env
- MySQL connection bootstrap with fail-fast startup
- `GET /healthz` smoke-test endpoint
- Docker Compose for MySQL
- migration scaffold under `feat-manager/db/migrations`
- Make targets for setup, run, test, build, and migrations inside `feat-manager/`
- placeholder folders for `feat-ui` and `feat-client`

## Quick start

1. `cd feat-manager`
2. `make setup`
3. `make deps-up`
4. `make test`
5. `make run`

The API listens on `HTTP_ADDR`, which defaults to `:8080`.

## Health check

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"success":true,"status":"ok"}
```

## Migration strategy

Phase 0 chooses SQL-file migrations stored in `feat-manager/db/migrations/` and executed through the `migrate/migrate` container:

- `make migrate-up`
- `make migrate-down`

The first migration is intentionally a no-op scaffold so the migration workflow is ready before the flag schema lands in phase 3.
