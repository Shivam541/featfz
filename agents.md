# Agents

## Repo Context

- Build a feature flag service in Go.
- This repo is a monorepo.
- Component names:
  - `feat-manager` for the Go backend,
  - `feat-ui` for the UI,
  - `feat-client` for the client-side showcase.
- Backend implementation files should live under `feat-manager/`, not in the repo root.
- A lightweight React/Next.js client will be used to call the backend APIs.
- MySQL is the initial database, managed through `docker compose`.
- The backend is the main component for this phase.

## Current V1 Design Defaults

- Tenant context should be header-based in v1, not URL-based.
- Each request should carry `Authorization: Bearer <jwt>` and `X-App-ID: <app-id>`.
- The server should use `X-App-ID` to load tenant/app configuration, validate the JWT with that tenant secret, and then confirm the JWT `app_id` claim matches the resolved app.
- `tenant` and `flag` are separate persisted entities.
- User identity in v1 is an opaque tenant-scoped string, not a rich user profile object.
- Feature flags are simple boolean records in v1. Each flag must store a required `default_enabled` value and may have explicit per-user override records with `enabled=true` or `enabled=false`.
- Public API scope for v1 includes flag CRUD, bulk per-user flag override updates, and evaluation APIs. Implementation order should still begin with core app behavior and the highest-priority endpoints first.
- Evaluation in v1 should use `GET /eval?flag=<flagKey>&user=<userID>` and return JSON with `success` plus `result` or `status` set to `on` or `off`.
- Code structure should stay layered: middleware (`authentication`, `monitoring`), request validation, handlers, services, and DAO/repository code.
- Modules should depend on interfaces instead of concrete implementations wherever practical.
- Dependencies must be constructed and injected during startup. If any dependency fails to initialize, startup should fail fast. Do not use `init()` functions.
- JWTs must include `app_id` and should also include `iat` and `exp`.
- No batch evaluation API is needed in v1.
- APIs should document clear 4xx and 5xx responses with safe client-facing messages and without exposing internal system details.

## Refactoring Direction

- Prefer entity structs mapped with GORM for persisted models.
- Prefer `go-playground/validator` for request validation instead of custom hand-written validation helpers where practical.
- Group related business operations into domain services such as `FlagService` and `EvalService` instead of creating one service type per endpoint.
- Use controller structs that can depend on multiple services and expose multiple handler methods, rather than creating a separate handler struct for every route action.
- Use meaningful dependency field names during wiring so startup code stays easy to read and debug.
- Keep the codebase simple, but allow slightly more structure when it reduces duplication or makes the API surface scale more cleanly.

## Coding Rules

- Use TDD for backend work.
- Prefer table-driven tests for Go code.
- Keep changes small, focused, and feature-by-feature.
- Treat tenant isolation as a hard requirement.
- Never allow cross-tenant reads, writes, caching, or shared state.
- Keep implementation simple and explicit, but favor grouped controllers, services, and entities when they reduce duplication or endpoint-specific boilerplate.

## Workflow Rules

- Commits go directly to `main`.
- Even though commits go to `main`, changes should still be split into small reviewable units.
- All tests should run on every commit.
- If a local hook is added, it should block pushes when tests fail.
- Use a `Makefile` for dependency setup, backend runs, test runs, and later frontend or client runs.
- Use `docker compose` for the MySQL dependency instead of ad hoc local setup.

## Documentation Rules

- The root `write up.md` file is append-only.
- Never rewrite or delete earlier entries in `write up.md`.
- Use `write up.md` to record:
  - User prompts and refinements.
  - Major decisions and tradeoffs.
  - When two paths were considered and how one was chosen.
  - What the AI suggested.
  - How the user responded, corrected, or changed direction.
- Every important decision should be documented so the repo can be resumed from context later.

## AI Automation Rules

- Document API descriptions, example `curl` calls, and test execution steps.
- Keep enough context in the repo that AI can operate the project autonomously.
- Record the final client snippet in the docs once the API shape is decided.
