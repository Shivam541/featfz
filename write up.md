# Write Up

This file is append-only.

Do not rewrite earlier entries.
Add each new note at the end with enough context that a future AI or developer can understand what changed and why.

## Entry Format

- Date and time
- Prompt summary
- Decision points
- AI recommendation
- User response
- Resulting direction

## Initial Entry

### 2026-07-04

The project foundation was set for a multi-tenant feature flag service built in Go, with a lightweight React/Next.js client for API access and MySQL as the initial database. The backend is the primary component for this phase, while the frontend and client wrapper are supporting pieces to be built later.

The initial requirements emphasize strict tenant isolation, meaning feature flags and users must remain scoped to a single tenant with no cross-tenant sharing, lookup, or caching. The primary function of the service is to determine whether a feature flag is active for a specific user within a specific tenant.

The development process for the backend will follow TDD, with table-driven tests for Go. The workflow should remain small and feature-by-feature even though commits will go directly to `main`. All tests should run on every commit, and a local hook may be added to prevent broken pushes.

The repository should be fully documented for AI-assisted operation. That includes API descriptions, `curl` examples, test execution instructions, and a short client snippet showing how a consuming app would call the service. The project should use a `Makefile` for setup and execution, and `docker compose` for the MySQL dependency.

This entry also captures the documentation policy itself: any future prompt, design choice, tradeoff, or direction change should be appended here in refined form so the project history remains easy for AI to follow.

## Entry

### 2026-07-04 21:32:47 IST

Prompt summary:

The user asked to review `agents.md` and `docs/requirements/assign_init.md`, answer the listed open technical-design questions, and move forward by laying out an initial tech spec in a separate document that can be iterated later. The user also gave direct product and architecture guidance for the first version.

Decision points:

- Whether tenant context should live in URLs, headers, or both.
- What `user_id` should look like in v1.
- Whether feature flags should be rule-based immediately or start as simple booleans.
- Whether the API surface should focus only on CRUD first or include evaluation in the v1 design.
- How to structure the backend layers and dependency wiring.

AI recommendation:

The recommended v1 shape was to keep tenant context header-based, using `Authorization: Bearer <jwt>` plus `X-App-ID`, so the app can look up tenant configuration cleanly and avoid putting tenant identity into every route. The recommendation also kept `user_id` intentionally simple as an opaque string, defined flags as boolean records with a required `default_enabled` field, and used explicit per-flag user onboarding rather than a rule engine. For the API surface, the recommendation was to include both management and evaluation in the v1 design while still sequencing implementation from core app flow to the highest-priority CRUD endpoints first. The proposed backend structure followed middleware, validation, handlers, services, and DAO layers with interface-based dependencies and fail-fast startup wiring in `main`, with no `init()` functions.

User response:

The user set the initial direction explicitly: use tenant tokens and app ids, provide a separate secret per tenant, require clients to create JWTs with that secret, decode and validate them, then use the matched app id to scope flags and user data. The user also clarified that `tenant` and `flag` should be separate entities, every flag should have a default rule, bulk user onboarding to a flag is needed, `user_id` should remain a simple string, and feature flags should stay simple on/off records in v1. The user asked to lay out the basic app functionality first, then the specific APIs, and to prioritize the most important APIs.

Resulting direction:

An initial tech spec was added in a separate document to capture the first-pass design. `assign_init.md` was updated so the earlier open questions now read as current v1 decisions. `agents.md` was updated with the repo's current v1 design defaults so future AI work stays aligned. The main product direction is now: authenticated tenant resolution via headers, tenant-scoped flags and user onboarding, simple boolean evaluation semantics in v1, layered backend structure, interface-based dependencies, and fail-fast startup.

## Entry

### 2026-07-04 21:44:49 IST

Prompt summary:

The user requested a small follow-up documentation pass to correct a few v1 details in the spec and write-up.

Decision points:

- Whether a flag-user record only enables a user or can also explicitly disable one.
- Whether archive behavior should stay open or be fixed now.
- What the exact evaluation endpoint shape should be.
- Whether batch evaluation should exist in v1.
- Whether recommended JWT claims should be explicitly documented.

AI recommendation:

The recommended correction was to move from an enable-only onboarding model to explicit per-user flag overrides, where a flag-user record can store either `enabled=true` or `enabled=false`. The recommendation also fixed soft archive as the current v1 direction, removed batch evaluation from v1, and documented the evaluation endpoint as `GET /eval?flag=X&user=Y` returning `on` or `off`. The JWT guidance was also tightened so `app_id` remains required and `iat` plus `exp` are clearly recommended.

User response:

The user confirmed soft archive works for now, clarified that a user can be enabled or disabled explicitly at the flag level, specified that the evaluation endpoint should be `GET /eval?flag=X&user=Y` returning on or off, confirmed that optional JWT claims should be included, and explicitly said there should be no batch evaluation.

Resulting direction:

The docs now treat flag-user data as per-user override records instead of enable-only onboarding records. The chosen v1 archive behavior is soft archive. The evaluation API is now documented as a GET endpoint at `/eval` with `flag` and `user` query parameters and a plain `on` or `off` response. Batch evaluation is now explicitly out of scope for v1, and the JWT documentation now clearly includes recommended `iat` and `exp` claims.

## Entry

### 2026-07-04 21:52:01 IST

Prompt summary:

The user requested one more documentation update for the evaluation response contract and for API error handling guidance.

Decision points:

- Whether the evaluation endpoint should return plain text or JSON.
- Whether the response should include an explicit success boolean.
- How to document safe 4xx and 5xx responses without exposing internal system details.

AI recommendation:

The recommendation was to move the evaluation endpoint to a JSON success envelope and standardize the result as a simple `result` value with `on` or `off`. The recommendation also added a shared error response shape and per-endpoint HTTP status guidance so clients can integrate predictably while the server still avoids leaking stack traces, SQL errors, dependency names, or other internal details.

User response:

The user asked for the eval endpoint to return JSON, with the on or off value as `status` or `result`, plus one more parameter `success` as `true` or `false`. The user also asked to lay out possible 4xx and 5xx responses per API with clear error messages and no exposure of internal system details, and asked for this request to be recorded in `write up.md`.

Resulting direction:

The docs now define the eval endpoint as a JSON response with `success` and `result`, document a reusable JSON error envelope, and list likely 4xx and 5xx responses for the main APIs with safe client-facing messages. The decision history for this request was appended here as well.

## Entry

### 2026-07-04 21:58:26 IST

Prompt summary:

The user asked to move from the tech spec into a development plan and requested a phase-by-phase implementation breakdown for backend setup, service skeleton, step-by-step code delivery, and tests. The user explicitly asked that the plan should not duplicate design content already captured in the tech spec and should instead reference the spec for those details.

Decision points:

- How to break the backend work into phases without repeating the tech spec.
- How much implementation order detail to include versus design detail.
- How to structure the plan so it is useful for day-to-day development.

AI recommendation:

The recommended approach was to add a separate development-plan document focused on execution order, commit slices, test gates, and exit criteria. Instead of restating API and model details, the plan should point back to the tech spec and concentrate on how to deliver the system safely in narrow vertical slices.

User response:

The user asked for a development plan that breaks down the tech spec phase by phase, including backend setup, service skeleton, code step by step, and tests, and explicitly asked not to duplicate data from the tech spec.

Resulting direction:

A separate phased development plan document was added under `docs/plans/`. It focuses on backend-first execution, phase gates, recommended commit slices, test coverage expectations, and MVP completion criteria, while using the tech spec as the source of truth for design details.

## Entry

### 2026-07-04 22:07:58 IST

Prompt summary:

The user asked to review the repo guidance documents again, start implementation from `docs/plans/development_plan.md`, and begin with phase 0.

Decision points:

- How much of phase 0 should be real implementation versus placeholder scaffolding.
- What migration strategy to choose early so later schema work does not need another tool switch.
- How strict config loading should be during the first bootstrapping slice.

AI recommendation:

The recommendation was to treat phase 0 as a real vertical slice rather than a paper setup step: add the Go module and backend folder skeleton, wire startup through explicit dependency construction, require the core database config fields up front, expose a small health endpoint for smoke tests, and choose a migration workflow immediately so schema work in phase 3 can land without setup churn. For migrations, the recommendation was SQL-file migrations stored in-repo and executed through the `migrate/migrate` container so local development stays consistent with the Docker-based MySQL dependency.

User response:

The user asked to proceed from the repo documents and explicitly start with phase 0.

Resulting direction:

Phase 0 implementation started with a runnable Go API skeleton, env-based config loading, fail-fast MySQL bootstrap, a `/healthz` endpoint, `docker compose` MySQL setup, `Makefile` commands for common development flows, and an initial no-op migration scaffold to lock in the migration path early.

## Entry

### 2026-07-04 22:07:58 IST

Prompt summary:

The user clarified the product naming they want across the repo components.

Decision points:

- What the Go backend should be called in docs and runtime-facing labels.
- What names should be reserved for the UI and client showcase components before they are implemented.

AI recommendation:

The recommendation was to apply the naming immediately in the backend-facing docs and local runtime labels so future work does not mix placeholder repo naming with actual product naming.

User response:

The user asked to call the Go backend `feat-manager`, the UI `feat-ui`, and the client-side showcase `feat-client`.

Resulting direction:

The backend is now documented and labeled as `feat-manager`, while `feat-ui` and `feat-client` are reserved in the repo docs for the later UI and showcase work. Local MySQL defaults were also renamed to `feat_manager` for consistency with the backend component name.

## Entry

### 2026-07-04 22:07:58 IST

Prompt summary:

The user clarified that this repository should be a monorepo and asked not to keep the backend implementation in the repo root.

Decision points:

- Whether to leave phase 0 backend files in the root temporarily or move them immediately into an app directory.
- How to make room now for the future UI and client showcase apps without redoing the repo layout later.

AI recommendation:

The recommendation was to adopt the monorepo layout immediately by moving the backend bootstrap under `feat-manager/`, leaving the root for shared docs and app-level directories. This avoids another repo-wide move once `feat-ui` and `feat-client` are added.

User response:

The user said this will be a monorepo and asked not to put the whole backend code in the repo root.

Resulting direction:

The backend bootstrap now lives under `feat-manager/`, and the root repo layout is being treated as the shared monorepo home for `feat-manager`, `feat-ui`, and `feat-client`.

## Entry

### 2026-07-04 22:23:00 IST

Prompt summary:

The user asked to review `agents.md`, `docs/plans/development_plan.md`, and `docs/tech-spec/`, then begin implementation from phase 1 of the development plan.

Decision points:

- Whether to treat phase 1 as documentation-only scaffolding or land real reusable HTTP infrastructure immediately.
- How much structure to add now without overbuilding before authentication and flag endpoints exist.
- What minimal interfaces and shared contracts should exist so later feature slices do not need router or response rewrites.

AI recommendation:

The recommendation was to treat phase 1 as a real backend slice and add only the smallest durable primitives the later phases will rely on: a router dependency shape, middleware chaining, request-id and request logging basics, a shared authenticated-request context container, JSON response helpers, validation helpers for headers, path values, query params, and JSON bodies, plus service and repository interfaces that future auth and flag flows can depend on.

User response:

The user asked to start with phase 1 after reviewing the current repo documents and development plan.

Resulting direction:

Phase 1 was implemented in `feat-manager` by wiring `/healthz` through a reusable handler and router dependency path, adding request middleware for panic recovery, request ids, and structured request logging, introducing shared JSON response and validation helpers, defining the tenant/request context carrier for upcoming auth work, and adding service/repository interface contracts for later phases. The backend test suite passed with `go test ./...` after the changes, so the repo is ready to move into phase 2 without reworking the request pipeline.

## Entry

### 2026-07-04 22:49:00 IST

Prompt summary:

The user asked to review the repo guidance documents again, align to the development plan, and begin implementation for phase 2.

Decision points:

- Whether phase 2 should wait for the full phase 3 persistence schema or land the tenant-authentication slice now with only the tenant/app data it depends on.
- Whether JWT handling should pull in a third-party library immediately or stay explicit and local for the first auth slice.
- What the first protected endpoint should be before real flag handlers exist.

AI recommendation:

The recommendation was to keep phase 2 small and vertical: add only the tenant table and seed data needed for app-id lookup, implement a tenant-app DAO plus JWT verification, wire auth middleware into the existing request pipeline, and expose one protected probe endpoint so tenant context can be verified end to end before flag CRUD work starts. To keep the first slice explicit and dependency-light, the recommendation was to implement HS256 JWT verification directly in the backend instead of expanding the module surface area before the flag domain arrives.

User response:

The user asked to start with phase 2 after reviewing `agents.md`, `docs/plans/development_plan.md`, and `docs/tech-spec/`.

Resulting direction:

Phase 2 was implemented in `feat-manager` by adding a `tenants` migration and phase-2 seed data, a MySQL tenant-app lookup repository, JWT-based authentication service and middleware, a protected `GET /v1/auth/check` verification route, and table-driven tests for missing headers, invalid JWTs, expired JWTs, app-id mismatch, and successful tenant-context injection. The backend suite passed with `go test ./...`, using a repo-local Go build cache path so tests stay runnable in restricted environments too.

## Entry

### 2026-07-04 23:12:00 IST

Prompt summary:

The user asked for the exact `curl` command to hit the new authenticated endpoint and a quick script under `feat-client` to generate JWTs for testing.

Decision points:

- Whether to keep the JWT helper as documentation-only or add a runnable repo script.
- Whether the helper should introduce a package dependency or stay dependency-free.

AI recommendation:

The recommendation was to add a tiny runnable helper directly under `feat-client` so the repo contains the promised client-side snippet in executable form, while keeping it dependency-free by using Node's built-in `crypto` module for HS256 signing.

User response:

The user asked for the curl command and a quick JWT generation script in `feat-client`.

Resulting direction:

`feat-client` now contains a small `scripts/generate-jwt.mjs` helper that emits a tenant JWT using environment overrides for app id, secret, subject, and expiration. The `feat-client/README.md` file now documents both token generation and the matching `curl` call for `GET /v1/auth/check` against the local `feat-manager` service.

## Entry

### 2026-07-04 23:22:46 IST

Prompt summary:

The user asked to review `agents.md`, `docs/plans/development_plan.md`, and `docs/tech-spec/`, then begin implementation from phase 3.

Decision points:

- What the phase 3 schema should include for flags and flag-user overrides.
- How to enforce tenant scoping and archive-aware reads at the repository layer.
- How to add DB-backed test fixtures without breaking the normal unit test flow.

AI recommendation:

The recommended phase 3 slice was to add real MySQL migrations for `flags` and `flag_user_overrides`, keep the soft-archive model by using `archived_at`, and implement repository methods that always scope by tenant and ignore archived flags in active reads. For testing, the recommendation was to keep the integration tests self-contained with a `TEST_DB_DSN` gate, so the default `go test ./...` path stays green while the DB-backed cases can run against a local MySQL container after `make deps-up` and `make migrate-up`.

User response:

The user asked to start with phase 3 after reading the repo guidance and development plan.

Resulting direction:

Phase 3 is now in place in `feat-manager`: new migrations create `flags` and `flag_user_overrides`, repository implementations support create, list, find, update, archive, and bulk upsert flows, and DAO integration tests cover tenant scoping, unique constraints, archive visibility, and override upsert behavior. The backend test suite passes with `go test ./...`, and the README now documents the new persistence slice and its optional DB-backed test setup.

## Entry

### 2026-07-04 23:33:32 IST

Prompt summary:

The user asked to push the phase 3 change and then begin phase 4.

Decision points:

- How to get the phase 3 work onto `main` cleanly.
- Whether phase 4 should be the first end-to-end flag creation slice rather than jumping ahead to list/get/update.
- How to keep the create-flag contract aligned with the docs while preserving the earlier phase 3 repository behavior.

AI recommendation:

The recommendation was to create a single git commit for the phase 3 persistence work and push it to the remote before starting phase 4. For phase 4, the recommendation was to keep the scope narrow and deliver only `POST /v1/flags` end to end: add a flag creation service, wire it through startup and the router, validate the request body, map duplicate-key failures to a safe `409` response, and add a MySQL-backed integration test that proves the HTTP path actually writes a row.

User response:

The user asked to push the change and then move on to phase 4.

Resulting direction:

The phase 3 commit was pushed to `origin/main` as `1fef269`. Phase 4 is now implemented in `feat-manager` with authenticated `POST /v1/flags`, create-flag validation, duplicate-key handling, router/app wiring, and an integration test that creates a real flag record in MySQL. The backend suite still passes with `go test ./...`, and the backend README now documents the create-flag slice plus the DB-backed test setup.
