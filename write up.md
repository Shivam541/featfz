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
