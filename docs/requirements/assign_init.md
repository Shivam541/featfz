# assign_init

## Goal

Build a multi-tenant feature flag service that can determine whether a feature flag is active for a given user within a given tenant.

## Functional Requirements

- The system must support multiple tenants.
- Feature flags must be isolated per tenant.
- Users must be isolated per tenant.
- The active state of a feature flag must be evaluated using the combination of tenant, feature flag, and user.
- The same user identity in different tenants must never share feature state.
- No cross-tenant lookup, sharing, or leakage is allowed.
- The service should expose APIs that make tenant context explicit.

## Out of Scope For Now

- Advanced rollout strategies.
- Multi-database support.
- Cross-tenant analytics.
- Production-grade frontend polish.

## Current V1 Decisions

- Tenant context will be header-based in v1, not URL-based.
- Each request must send `Authorization: Bearer <jwt>` and `X-App-ID: <app-id>`.
- The server will resolve the tenant from `X-App-ID`, validate the JWT using that tenant secret, and confirm the JWT `app_id` claim matches the resolved app before continuing.
- JWTs must include `app_id` and should also include `iat` and `exp`.
- The user identifier in v1 will be a simple opaque string field named `user_id`.
- `user_id` is tenant-scoped, required, case-sensitive, and not parsed by the service beyond basic validation.
- Feature flags will be simple on/off records in v1.
- Each flag will store a required `default_enabled` value as its default rule.
- Per-user flag overrides are part of v1, including explicit `true` and `false` values at the flag level.
- The service should expose both flag management and evaluation APIs in v1, while implementation can still proceed in a staged order with core app flow and high-priority flag CRUD endpoints first.
- The evaluation endpoint in v1 will be `GET /eval?flag=<flagKey>&user=<userID>` and return JSON with `success` and `result` or `status` set to `on` or `off`.
- No batch evaluation API is needed in v1.
- APIs should define safe error responses for likely 4xx and 5xx cases without exposing internal implementation details.
