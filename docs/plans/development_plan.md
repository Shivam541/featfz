# Development Plan

Status: draft

This document is the implementation plan for [initial_v1.md](/Users/shivam/projects/featfz/docs/tech-spec/initial_v1.md).
It should stay focused on delivery order, work slicing, and test gates.
Design details should be referenced from the tech spec instead of repeated here.

## Planning Rules

- Build backend-first.
- Keep each phase small enough to ship and review on its own.
- Follow TDD for backend code.
- Prefer table-driven tests for service, handler, and validation logic.
- Do not start a later phase until the current phase has passing tests and updated docs where needed.

## Delivery Shape

The work should move in narrow vertical slices:

1. make the app boot,
2. make the first protected route work,
3. add one flag flow end to end,
4. expand to the full v1 API surface,
5. harden with error handling, migrations, and test coverage.

This keeps tenant isolation and auth behavior verified early instead of leaving them until the end.

## Phase 0: Repo and Backend Setup

Goal:

- create the minimum backend project that can build, run tests, and connect to local dependencies.

Work:

- initialize the Go module and backend folder structure under `feat-manager/`,
- add `Makefile` targets for setup, run, test, and dependency helpers,
- add `docker compose` for MySQL,
- add config loading for local development,
- add a small `feat-manager/cmd/api` startup flow with dependency wiring,
- add a simple health endpoint for smoke testing,
- add a migration strategy choice and first migration scaffolding.

Outputs:

- runnable app skeleton,
- repeatable local database startup,
- first test command and first run command,
- folder structure aligned with the tech spec.

Tests and checks:

- `go test ./...` runs successfully,
- app starts with valid config,
- app fails fast with missing required config,
- health endpoint returns success.

Suggested commit slices:

1. module and folder skeleton,
2. Makefile and compose setup,
3. config and startup wiring,
4. health route and startup tests.

Exit criteria:

- a new developer can clone, start MySQL, run tests, and boot the API without manual setup steps outside the docs.

## Phase 1: Core App Skeleton

Goal:

- establish the shared internal structure before feature work starts.

Work:

- wire router, middleware chain, handlers, services, and DAO interfaces,
- define shared request context shape for authenticated tenant data,
- add common response helpers for success and error JSON,
- add structured logging and request tracing basics,
- add validation helpers for headers, path params, query params, and JSON bodies.

Outputs:

- reusable request pipeline,
- stable package boundaries,
- common response and validation patterns.

Tests and checks:

- middleware can be exercised in tests,
- shared error responses match the documented format,
- validation errors do not leak internals,
- handlers can be tested with mocked services.

Suggested commit slices:

1. router and middleware composition,
2. shared response helpers,
3. validation helpers,
4. interface definitions and test doubles.

Exit criteria:

- feature endpoints can be added without reworking app bootstrapping or response format later.

## Phase 2: Authentication and Tenant Context

Goal:

- make tenant-authenticated requests work end to end.

Work:

- implement `Authorization` and `X-App-ID` extraction,
- add tenant/app lookup DAO path,
- add JWT validation flow,
- enforce tenant context injection into downstream handlers,
- define auth failure mapping to safe client responses,
- add seed or fixture support for tenant/app test data.

Outputs:

- protected route middleware,
- tenant context available to handlers and services,
- safe auth error behavior.

Tests and checks:

- missing auth header,
- missing `X-App-ID`,
- invalid JWT,
- expired JWT,
- JWT `app_id` mismatch,
- valid JWT and tenant context success,
- cross-tenant access blocked before business logic runs.

Suggested commit slices:

1. tenant DAO and middleware contract,
2. JWT verification path,
3. request context plumbing,
4. auth middleware tests.

Exit criteria:

- at least one dummy protected endpoint succeeds only with valid tenant-scoped auth.

## Phase 3: Flag Domain and Persistence Base

Goal:

- build the minimal data layer needed for the first real feature flow.

Work:

- create migrations for tenants, flags, and flag-user override storage from the tech spec,
- implement DAO interfaces for flags and per-user overrides,
- decide repository method shapes before handlers are added widely,
- add archive-aware read/write behavior in the DAO layer,
- add DB test helpers for setup and cleanup.

Outputs:

- migrations checked into the repo,
- repository layer that can support create, read, update, archive, and override writes,
- DB-backed test fixtures.

Tests and checks:

- migration up works on a clean database,
- unique constraints behave as expected,
- archived flags are excluded from active reads,
- tenant scoping is enforced in every repository query,
- override upsert behavior works.

Suggested commit slices:

1. migrations,
2. flag repository,
3. override repository,
4. repository integration tests.

Exit criteria:

- the persistence layer is stable enough that service logic can be added without schema churn.

## Phase 4: First End-to-End Feature Slice

Goal:

- deliver one complete business flow before expanding to the full API set.

Recommended first slice:

- create flag.

Why first:

- it exercises auth, validation, handler, service, DAO, persistence, success responses, and error responses in one path,
- later read and update flows depend on this shape anyway.

Work:

- implement request validation,
- implement create service logic,
- implement create handler,
- map duplicate key and validation failures to documented responses,
- add one end-to-end API test path.

Tests and checks:

- create success,
- duplicate flag key in same tenant,
- same key in different tenants allowed,
- malformed payload,
- auth failure path.

Exit criteria:

- one real API path works cleanly from HTTP entry to database write with tests at each relevant layer.

## Phase 5: Flag Management API Completion

Goal:

- complete the rest of the flag management surface.

Work:

- add list flags,
- add get flag,
- add update flag,
- add soft archive flag,
- ensure archive behavior is reflected consistently in reads and updates.

Tests and checks:

- list returns tenant-owned active flags only,
- get returns tenant-owned active flag only,
- update changes allowed fields only,
- archive removes the flag from active reads,
- archived flag cannot be evaluated or updated as active,
- cross-tenant reads and writes are blocked.

Suggested commit slices:

1. list and get,
2. update,
3. archive,
4. read-path cleanup and tests.

Exit criteria:

- all documented flag CRUD behavior is implemented and covered by tests.

## Phase 6: Per-User Override Flow

Goal:

- implement targeted user-level flag behavior.

Work:

- add bulk override write endpoint,
- implement request normalization and deduplication,
- implement override upsert behavior,
- keep tenant isolation checks in service and DAO layers,
- return safe validation failures for malformed override entries.

Tests and checks:

- mixed `true` and `false` overrides in one request,
- duplicate `user_id` entries in one request with last value winning,
- override update on existing record,
- invalid user ids rejected,
- tenant isolation preserved.

Exit criteria:

- clients can apply user-specific flag states without touching evaluation yet.

## Phase 7: Evaluation Endpoint

Goal:

- implement the read path that clients will use at runtime.

Work:

- add `GET /eval`,
- validate `flag` and `user` query params,
- implement evaluation service logic using default flag state plus override precedence,
- return the documented JSON success envelope,
- map missing flag, auth failure, and dependency failures to safe error responses.

Tests and checks:

- evaluation returns default `on`,
- evaluation returns default `off`,
- explicit override `true` wins,
- explicit override `false` wins,
- missing flag returns not found,
- missing or invalid query params return bad request,
- cross-tenant data cannot influence result.

Exit criteria:

- runtime consumers can evaluate a single flag reliably through the documented API contract.

## Phase 8: Hardening and Developer Workflow

Goal:

- make the project dependable for daily development and future iterations.

Work:

- add test helpers and faster local dev commands where useful,
- add seed or fixture tooling for local manual testing,
- add migration commands to `Makefile`,
- document the test matrix and common dev flows,
- add a local hook only if it cleanly blocks pushes on failing tests,
- review logging to confirm internal errors are not leaked in responses.

Tests and checks:

- full test suite is stable,
- common setup and test commands are documented and working,
- push-blocking hook, if added, works predictably,
- logs keep diagnostics while responses stay safe.

Exit criteria:

- the repo is ready for regular feature work without repeated setup friction.

## Phase 9: Pre-MVP Review

Goal:

- confirm the implementation matches the intended v1 scope before adding new capabilities.

Review checklist:

- all endpoints in the tech spec exist,
- tenant isolation is verified in handlers, services, and DAO queries,
- soft archive behavior is consistent,
- evaluation response matches the documented JSON contract,
- error responses match the documented safe envelope,
- tests cover happy paths and key failure paths,
- docs reflect actual commands and actual API behavior.

Possible outputs:

- small cleanup tickets,
- missing test tickets,
- a short MVP-ready note,
- a deferred-work list for post-v1 improvements.

## Recommended Build Order Summary

1. repo setup,
2. app skeleton,
3. auth and tenant context,
4. migrations and repositories,
5. create flag,
6. remaining flag management,
7. user overrides,
8. evaluation,
9. hardening and review.

## What Not To Do Early

- do not start with all endpoints at once,
- do not add rollout rules before the simple override model is stable,
- do not add caching before tenant-safe behavior is proven,
- do not optimize abstractions before at least one full flag flow is working,
- do not leave integration tests until the end.

## Definition Of Done For MVP

The MVP is done when:

- the backend boots locally through documented commands,
- tenant-authenticated requests work,
- flag create, list, get, update, archive, bulk override, and eval all work,
- MySQL-backed persistence is in place,
- tests cover the main success and failure paths,
- docs describe how to run, test, and call the API.
