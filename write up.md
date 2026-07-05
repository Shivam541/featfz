# Assignment Write-Up

## Project Summary

This assignment was delivered as a backend-first multi-tenant feature flag service in Go inside a monorepo with three planned components:

- `feat-manager` for the Go backend
- `feat-ui` for the management UI
- `feat-client` for the lightweight client-side showcase

The shipped v1 backend supports tenant-authenticated requests, flag CRUD, bulk per-user override updates, and single-flag evaluation through `GET /eval?flag=<flagKey>&user=<userID>`. MySQL is the initial database, local developer workflows are driven through `Makefile` plus `docker compose`, and the implementation was delivered in small phase-based slices with tests and documentation along the way.

This write-up is based on the current repo docs, `ai chat history.md`, `agents.md`, `docs/tech-spec/initial_v1.md`, `docs/plans/development_plan.md`, `feat-manager/README.md`, and the full git commit history.

## 1. What did I ask the AI to do, and what did I write or decide myself?

I used AI as the implementation engine, but I set the foundation and the guardrails first.

What I wrote or decided myself:

- I set the initial product direction and the core assignment requirements.
- I defined the rules in [agents.md](/Users/shivam/projects/featfz/agents.md) so the AI had strict boundaries around tenant isolation, TDD, table-driven tests, documentation expectations, monorepo layout, and workflow discipline.
- I finalized the main technical direction in [initial_v1.md](/Users/shivam/projects/featfz/docs/tech-spec/initial_v1.md), including header-based tenant context, JWT plus `X-App-ID`, simple boolean flags, per-user overrides, no batch evaluation in v1, and safe error responses.
- I asked for and then refined the phased execution plan in [development_plan.md](/Users/shivam/projects/featfz/docs/plans/development_plan.md), which became the roadmap the AI followed.
- I reviewed each phase, corrected direction when needed, and requested refactors or API changes when the implementation drifted from the desired shape.

What I asked the AI to do:

- Turn the initial requirements into a refined tech spec and then a non-duplicative development plan.
- Implement the backend phase by phase from that plan.
- Add tests, API docs, curl examples, seed data, migration helpers, smoke flows, hooks, and CI support.
- Perform API testing when asked, especially around auth, eval behavior, and tenant isolation.
- Refactor the codebase midstream when I changed the architectural direction toward GORM entities, grouped controllers, grouped services, and validator-based request validation.

In practice, once the initial foundation was solid and the requirements were finalized, I took a back seat and let the AI complete most of the implementation with minor interruptions in between. The work still needed human review after each phase, and a few important corrections and design changes were necessary during the build.

## 2. Where did I override, correct, or throw away the AI's output, and why?

The most important overrides were about keeping the project aligned with the assignment and preventing architectural drift.

1. I moved the project away from root-level backend code and enforced the monorepo shape.
The AI initially started phase 0 as a backend bootstrap, but I corrected the structure so the backend lived under `feat-manager/` rather than the repo root. This mattered because the assignment was not just about code compiling; it needed the right long-term layout for `feat-manager`, `feat-ui`, and `feat-client`.

2. I corrected the v1 product rules while the spec was still forming.
The early direction around user onboarding and evaluation needed tightening. I clarified that:

- `tenant` and `flag` are separate persisted entities
- flags are simple booleans in v1
- each flag must have `default_enabled`
- per-user entries are explicit overrides with `enabled=true` or `enabled=false`
- evaluation should be `GET /eval?flag=<flagKey>&user=<userID>`
- the response should be JSON with `success` and `result` or `status`
- batch evaluation is out of scope for v1

Those corrections were important because they simplified the system and kept the assignment focused.

3. I changed the architecture midstream.
After phase 4, I was not satisfied with the direction if it kept growing one handler at a time with too much endpoint-by-endpoint structure. I explicitly redirected the codebase toward:

- GORM-backed entity structs
- `go-playground/validator`
- grouped controllers
- grouped domain services such as `FlagService` and `EvalService`
- clearer dependency naming during startup wiring

This was not a cosmetic change. It threw away part of the earlier style direction and replaced it with a structure that would scale better as the API surface grew.

4. I corrected API shape details for consistency and usability.
One clear example was the bulk override endpoint. The AI had a colon-style action route first, but I asked for a more conventional REST-like shape, which became:

`POST /v1/flags/{flagKey}/users/bulk-set`

That change made the API easier to explain and more consistent with the rest of the backend.

5. I corrected tooling direction when it did not match the actual target.
The history also shows a CI correction: the first pass added GitLab CI thinking, but the actual target was GitHub Actions, so that direction was corrected to a root GitHub Actions workflow for the Go backend.

Overall, I did not need to micromanage every phase, but I did step in whenever the design, repo structure, or API shape needed a stronger human decision.

## 3. The biggest trade-offs I made, and the alternatives I considered

### Trade-off 1: Simple MySQL-backed v1 over a more ambitious infrastructure setup

I chose MySQL with `docker compose` for local development, and the project is documented in a way that can translate easily to a managed MySQL or RDS-style deployment later. The reason was simplicity, fast setup, and lower friction for an assignment project.

Alternative considered:

- a more production-heavy setup with broader infrastructure automation from day one
- another database choice such as PostgreSQL

Why I chose this:

- it was faster to stand up
- it kept the assignment focused on backend behavior, not platform work
- it matched the need for a quick, testable v1

The downside is that this keeps some production concerns, like deeper operational tuning and deployment ergonomics, out of the first version.

### Trade-off 2: Header-based tenant context plus per-tenant JWT secrets over URL-based tenancy or a centralized auth model

The final spec uses `Authorization: Bearer <jwt>` and `X-App-ID: <app-id>`, then resolves the tenant configuration from `X-App-ID`, validates the JWT with that tenant secret, and confirms the JWT `app_id` claim matches the resolved app.

Alternatives considered:

- tenant identifiers in route paths
- looser request scoping
- a more centralized or abstracted auth model earlier in the build

Why I chose this:

- it keeps tenant context explicit without putting tenant ids into every route
- it makes tenant isolation a first-class rule
- it fit the assignment goal of a clean v1 multi-tenant backend

The downside is that tenant secret management remains manual and operationally heavier than a more automated auth provisioning setup.

### Trade-off 3: Simple boolean flags and per-user overrides over a rule engine, batch eval, or richer rollout logic

The final v1 supports:

- `default_enabled`
- explicit per-user overrides
- single-flag evaluation

Alternatives considered:

- percentage rollouts
- segment-based or rule-based targeting
- batch evaluation APIs
- richer user models

Why I chose this:

- it kept the scope realistic for the assignment
- it reduced the chance of overengineering early
- it let the implementation stay strongly testable and phaseable

The downside is that the product is intentionally narrow. It is a good v1, but not yet a full-featured commercial flagging platform.

## 4. What is missing, or what would I do with another day?

If I had another day, I would focus on the highest-value gaps that remain after the MVP:

1. Add end-to-end tests.
The backend has strong unit, controller, router, and integration coverage, but I would add fuller E2E coverage around the real running app, seeded tenants, auth flow, override flow, and eval flow.

2. Improve the UI for flag management.
The backend is ahead of the rest of the monorepo. I would spend time improving `feat-ui` so flags, overrides, and evaluation flows can be managed more comfortably through a proper UI instead of mostly curl and backend-first tooling.

3. Remove the manual tenant secret dependency.
Right now the tenant secret model is simple and explicit, which was good for v1, but I would improve provisioning so secrets and tenant onboarding do not rely on manual setup steps.

4. Improve API performance with caching.
The current implementation prioritizes correctness and tenant isolation. With more time, I would add safe tenant-aware caching around repeated read paths without violating the no-cross-tenant-sharing rule.

5. Strengthen operational polish.
I would likely add better observability, production deployment notes, and more explicit failure dashboards or tracing for auth, DB, and eval behavior.

## Commit History Review

I reviewed the full project commit history, which currently contains 20 commits from `77cb019` through `a53529a`, so it is comfortably under the requested 30-commit threshold.

The history also shows that the work was split into small, reviewable units rather than one large dump:

1. repo setup and initial tech spec
2. development plan
3. backend bootstrap and phase 0
4. phase 1 core skeleton
5. phase 2 auth and tenant context
6. phase 3 persistence
7. phase 4 create-flag flow
8. architecture refactor toward GORM, validator, grouped controllers, and grouped services
9. phase 5 flag management
10. phase 6 bulk overrides
11. phase 7 evaluation
12. phase 8 smoke tooling, hooks, migration helpers, seed tooling, and response-safety checks
13. CI and phase 9 review notes

That phased history supports the main claim of this assignment write-up: the project was not built by dumping one prompt into an AI and accepting everything blindly. The human role was strongest in setting the initial rules, narrowing and correcting the spec, steering mid-course refactors, and deciding when the implementation was acceptable enough to let the AI continue autonomously.
