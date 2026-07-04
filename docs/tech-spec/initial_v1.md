# Initial Tech Spec v1

Status: draft

## Goal

Build the first backend version of a multi-tenant feature flag service in Go that:

- enforces strict tenant isolation,
- lets a tenant manage its own flags,
- supports bulk per-user flag override updates,
- evaluates whether a flag is active for a user inside that tenant.

## V1 Scope

Included in v1:

- tenant-aware authentication,
- per-tenant flag CRUD,
- bulk per-user flag override updates,
- single-flag evaluation for a user through a GET endpoint,
- MySQL persistence,
- explicit layered backend structure.

Not included in v1:

- rule engines,
- percentage rollouts,
- segment targeting,
- multi-environment support,
- cross-tenant analytics,
- advanced user profiles,
- batch evaluation APIs.

## Core App Functionality

### 1. Tenant Authentication and Resolution

Each tenant will be provisioned with:

- a tenant record,
- an `app_id`,
- a tenant-specific JWT signing secret.

Request flow:

1. The client creates a JWT using the tenant secret.
2. The client sends the JWT in `Authorization: Bearer <jwt>`.
3. The client sends the tenant app in `X-App-ID: <app-id>`.
4. Authentication middleware loads the tenant/app record using `X-App-ID`.
5. The middleware validates the JWT using that tenant secret.
6. After validation, the middleware confirms the JWT `app_id` claim matches the resolved app.
7. The middleware attaches tenant context to the request.

Recommended JWT claims:

- `iat`
- `exp`

Decision:

- Tenant context is header-based in v1.
- We do not pass tenant identifiers in URL paths.

Reasoning:

- It keeps tenant context explicit.
- It avoids mixing tenant identity into every route shape.
- It gives the server a concrete lookup key for the tenant secret before JWT validation.

### 2. Flag Ownership

`tenant` and `flag` are separate entities.

Every flag belongs to exactly one tenant.

Every flag has:

- a stable key,
- optional description,
- a required `default_enabled` boolean.

In v1, this `default_enabled` field is the only default rule supported.

### 3. User Representation

In v1, user info is a simple string field named `user_id`.

Rules for `user_id` in v1:

- opaque to the service,
- tenant-scoped,
- case-sensitive,
- required,
- trimmed,
- max length 255 characters.

The service does not parse emails, UUIDs, or composite identifiers in v1.
The client owns the identifier format.

### 4. Per-User Flag Overrides

The service will support bulk per-user override writes for a flag.

In v1, an override means:

- a user is explicitly attached to a flag inside one tenant,
- the attachment stores an `enabled` boolean for that flag-user pair,
- the association is stored per tenant and per flag,
- the same `user_id` string in another tenant remains unrelated.

To keep v1 simple, the first model is:

- if a user has an explicit override for a flag, that override wins,
- otherwise the result falls back to `default_enabled`.

This avoids introducing a rule language while still supporting targeted rollout.

### 5. Evaluation Behavior

Evaluation inputs:

- authenticated tenant context,
- flag key,
- user id.

Evaluation output:

- a JSON response body with `success` and `result` or `status`.

Evaluation order:

1. resolve tenant from authenticated request,
2. find the tenant-owned flag by key,
3. check whether the tenant-owned user has an explicit override for that flag,
4. return that override value if present,
5. otherwise return the flag `default_enabled` value.

## Data Model

### tenants

- `id`
- `name`
- `app_id` unique
- `jwt_secret`
- `created_at`
- `updated_at`

### flags

- `id`
- `tenant_id`
- `key`
- `description`
- `default_enabled`
- `created_at`
- `updated_at`

Constraints:

- unique (`tenant_id`, `key`)

### flag_user_overrides

- `id`
- `tenant_id`
- `flag_id`
- `user_id`
- `enabled`
- `created_at`
- `updated_at`

Constraints:

- unique (`tenant_id`, `flag_id`, `user_id`)

Notes:

- `flag_user_overrides` is intentionally simple in v1.
- We do not create a full `users` table unless a later requirement needs richer user metadata.

## API Direction

The service should expose both management and evaluation capabilities in v1.
Implementation order should still start with the management flow because evaluation depends on that data existing.

Suggested priority:

1. authentication middleware and tenant context plumbing,
2. create flag,
3. list and fetch flags,
4. update flag,
5. bulk set per-user overrides for a flag,
6. evaluate a flag for a user,
7. soft archive flag.

## Initial API Shape

### Create Flag

`POST /v1/flags`

Request body:

```json
{
  "key": "new_dashboard",
  "description": "Enable the new dashboard experience",
  "default_enabled": false
}
```

### List Flags

`GET /v1/flags`

### Get Flag

`GET /v1/flags/{flagKey}`

### Update Flag

`PATCH /v1/flags/{flagKey}`

Request body:

```json
{
  "description": "Enable the new dashboard experience for selected users",
  "default_enabled": true
}
```

### Archive Flag

`DELETE /v1/flags/{flagKey}`

API behavior for v1:

- the flag becomes unavailable for future evaluation.
- the record is soft archived rather than hard deleted.

Reasoning:

- soft archive keeps safer key history and reduces accidental data loss.

### Bulk Set Per-User Overrides for a Flag

`POST /v1/flags/{flagKey}/users:bulk-set`

Request body:

```json
{
  "overrides": [
    {
      "user_id": "user_123",
      "enabled": true
    },
    {
      "user_id": "user_456",
      "enabled": false
    }
  ]
}
```

Expected behavior:

- deduplicate repeated `user_id` entries in the request,
- last value wins if the same `user_id` appears multiple times in one request,
- existing rows should be updated when the user already has an override,
- perform writes only inside the authenticated tenant.

### Evaluate Flag for a User

`GET /eval?flag=<flagKey>&user=<userID>`

Example:

```text
GET /eval?flag=new_dashboard&user=user_123
```

Response:

```json
{
  "success": true,
  "result": "on"
}
```

## Request Headers

Required headers:

- `Authorization: Bearer <jwt>`
- `X-App-ID: <app-id>`

Required JWT claim in v1:

- `app_id`

Optional but recommended JWT claims:

- `iat`
- `exp`

## Example curl Calls

### Create a flag

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

### Bulk set per-user overrides

```bash
curl -X POST http://localhost:8080/v1/flags/new_dashboard/users:bulk-set \
  -H "Authorization: Bearer $TENANT_JWT" \
  -H "X-App-ID: $APP_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "overrides": [
      {"user_id": "user_123", "enabled": true},
      {"user_id": "user_456", "enabled": false}
    ]
  }'
```

### Evaluate a flag

```bash
curl "http://localhost:8080/eval?flag=new_dashboard&user=user_123" \
  -H "Authorization: Bearer $TENANT_JWT" \
  -H "X-App-ID: $APP_ID"
```

## Response Conventions

Successful API responses should use JSON.

Example success envelope for evaluation:

```json
{
  "success": true,
  "result": "on"
}
```

Notes:

- `result` may be `on` or `off`.
- If we later want a more explicit field name, `status` can be treated as an acceptable equivalent in the API discussion phase, but v1 docs should standardize on `result`.

Error responses should also use JSON and must not expose raw SQL errors, stack traces, dependency names, secrets, or internal infrastructure details.

Recommended error envelope:

```json
{
  "success": false,
  "error": {
    "code": "flag_not_found",
    "message": "The requested flag was not found."
  }
}
```

Error response rules:

- `success` must be `false`.
- `error.code` should be stable and machine-friendly.
- `error.message` should be short, clear, and safe for clients to display or log.
- Internal diagnostics should stay in server logs only.

## API Error Responses

### Shared 4xx and 5xx Cases

Likely shared responses across authenticated APIs:

- `400 Bad Request` for malformed JSON, missing required fields, invalid query parameters, or invalid field values.
- `401 Unauthorized` for missing auth header, invalid JWT, expired JWT, or tenant/app auth mismatch.
- `404 Not Found` when a tenant-scoped flag does not exist.
- `409 Conflict` when creating a duplicate tenant-scoped flag key.
- `422 Unprocessable Entity` when the request shape is valid JSON but business validation fails.
- `500 Internal Server Error` for unexpected server failures.
- `503 Service Unavailable` when the service cannot reach a required dependency such as the database.

Client-facing messages should stay generic and safe. Examples:

- `400`: `"The request is invalid."`
- `401`: `"Authentication failed."`
- `404`: `"The requested flag was not found."`
- `409`: `"A flag with this key already exists."`
- `422`: `"The request could not be processed."`
- `500`: `"Something went wrong."`
- `503`: `"The service is temporarily unavailable."`

### Create Flag

`POST /v1/flags`

Expected responses:

- `201 Created` when the flag is created.
- `400 Bad Request` when required fields are missing or malformed.
- `401 Unauthorized` when authentication fails.
- `409 Conflict` when the flag key already exists for that tenant.
- `500 Internal Server Error` for unexpected server failures.
- `503 Service Unavailable` when persistence is unavailable.

### List Flags

`GET /v1/flags`

Expected responses:

- `200 OK` when flags are returned.
- `401 Unauthorized` when authentication fails.
- `500 Internal Server Error` for unexpected server failures.
- `503 Service Unavailable` when persistence is unavailable.

### Get Flag

`GET /v1/flags/{flagKey}`

Expected responses:

- `200 OK` when the flag is found.
- `401 Unauthorized` when authentication fails.
- `404 Not Found` when the tenant-scoped flag does not exist.
- `500 Internal Server Error` for unexpected server failures.
- `503 Service Unavailable` when persistence is unavailable.

### Update Flag

`PATCH /v1/flags/{flagKey}`

Expected responses:

- `200 OK` when the flag is updated.
- `400 Bad Request` when the body is malformed.
- `401 Unauthorized` when authentication fails.
- `404 Not Found` when the tenant-scoped flag does not exist.
- `422 Unprocessable Entity` when validation fails.
- `500 Internal Server Error` for unexpected server failures.
- `503 Service Unavailable` when persistence is unavailable.

### Archive Flag

`DELETE /v1/flags/{flagKey}`

Expected responses:

- `200 OK` when the flag is soft archived.
- `401 Unauthorized` when authentication fails.
- `404 Not Found` when the tenant-scoped flag does not exist.
- `500 Internal Server Error` for unexpected server failures.
- `503 Service Unavailable` when persistence is unavailable.

### Bulk Set Per-User Overrides for a Flag

`POST /v1/flags/{flagKey}/users:bulk-set`

Expected responses:

- `200 OK` when overrides are applied.
- `400 Bad Request` when the body is malformed.
- `401 Unauthorized` when authentication fails.
- `404 Not Found` when the tenant-scoped flag does not exist.
- `422 Unprocessable Entity` when override entries fail validation.
- `500 Internal Server Error` for unexpected server failures.
- `503 Service Unavailable` when persistence is unavailable.

### Evaluate Flag for a User

`GET /eval?flag=<flagKey>&user=<userID>`

Expected responses:

- `200 OK` with JSON result when evaluation succeeds.
- `400 Bad Request` when `flag` or `user` is missing or invalid.
- `401 Unauthorized` when authentication fails.
- `404 Not Found` when the tenant-scoped flag does not exist.
- `500 Internal Server Error` for unexpected server failures.
- `503 Service Unavailable` when persistence is unavailable.

## Backend Structure

Planned request path:

1. middleware (`authentication`, `monitoring`),
2. validation layer,
3. API handler,
4. service layer,
5. DAO/repository layer,
6. MySQL.

Guidelines:

- modules depend on interfaces, not implementations, wherever that keeps the design clean,
- dependency construction happens in main startup flow,
- startup fails immediately if config, DB, repositories, or services cannot be created,
- do not use `init()` functions for app wiring.

## Suggested Package Layout

```text
cmd/api/
internal/config/
internal/domain/
internal/http/middleware/
internal/http/validation/
internal/http/handlers/
internal/service/
internal/dao/
internal/mysql/
```

## Testing Direction

Backend work should follow TDD.

Test priorities:

1. authentication middleware behavior,
2. tenant isolation checks,
3. flag service CRUD behavior,
4. per-user override behavior,
5. evaluation behavior,
6. DAO integration against MySQL.

Prefer:

- table-driven tests for Go units,
- explicit tenant-isolation test cases in every layer that touches persisted data.

Planned commands to standardize in the repo:

- `make setup`
- `make test`
- `make run`
- `docker compose up -d`

## Deferred Decisions

These are intentionally left for later iterations:

- tenant provisioning API versus internal/admin-only provisioning,
- secret rotation flow,
- richer JWT claim requirements beyond `app_id`, `iat`, and `exp`,
- richer user metadata,
- rule-based evaluation beyond `default_enabled`.
