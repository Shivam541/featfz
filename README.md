# featfz

Monorepo for a multi-tenant feature flag service.

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker Desktop or Docker Engine with Compose

## Install

1. Install Go dependencies and backend tooling:
   ```bash
   cd feat-manager
   make setup
   ```
2. Install frontend dependencies:
   ```bash
   cd ../feat-ui && npm install
   cd ../feat-client && npm install
   ```

## Run

1. Start the database:
   ```bash
   cd feat-manager
   make deps-up
   ```
2. Run backend migrations and seed data:
   ```bash
   make migrate-up
   make seed-phase2
   ```
3. Start the backend API:
   ```bash
   make run
   ```
4. Start the admin UI on port `3003`:
   ```bash
   cd ../feat-ui
   npm run dev
   ```
5. Start the client showcase on port `3001`:
   ```bash
   cd ../feat-client
   npm run dev
   ```

## Test

Run backend tests:

```bash
cd feat-manager
make test
```

Run frontend checks:

```bash
cd feat-ui && npm run lint && npm run build
cd ../feat-client && npm run lint && npm run build
```

## Apps

- `feat-manager/` - Go backend API and MySQL persistence
- `feat-ui/` - admin UI for browsing, viewing, evaluating, and managing flags
- `feat-client/` - client-side showcase and API playground

## Current UI layout

- `feat-ui` home page is a launcher
- `feat-ui/flags` keeps list, view, and evaluate together
- `feat-ui/manage` keeps create, update, and archive together
- app credentials stay hidden inside an expandable `App details` disclosure

## Local ports

- Backend API: `http://127.0.0.1:8080`
- `feat-ui`: `http://127.0.0.1:3003`
- `feat-client`: `http://127.0.0.1:3001`

## Backend docs

- [feat-manager/README.md](feat-manager/README.md)
- [docs/tech-spec/initial_v1.md](docs/tech-spec/initial_v1.md)
- [docs/plans/development_plan.md](docs/plans/development_plan.md)

## Frontend entrypoints

- [feat-ui/src/app/page.tsx](feat-ui/src/app/page.tsx)
- [feat-ui/src/app/flags/page.tsx](feat-ui/src/app/flags/page.tsx)
- [feat-ui/src/app/manage/page.tsx](feat-ui/src/app/manage/page.tsx)
- [feat-client/src/app/page.tsx](feat-client/src/app/page.tsx)

## Notes

- Tenant context is header-based in v1.
- Requests should send `Authorization: Bearer <jwt>` and `X-App-ID: <app-id>`.
- JWTs include `app_id` and should include `iat` and `exp`.
- `user_id` is tenant-scoped and treated as an opaque string.
