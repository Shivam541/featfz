# feat-client

Client-side helper area for quick API smoke tests and later showcase work.

## Generate a JWT

This repo includes a small zero-dependency Node script for generating an HS256 JWT that matches the phase-2 auth flow.

Run it with the default seeded tenant values:

```bash
node ./scripts/generate-jwt.mjs
```

Override values when needed:

```bash
APP_ID=app-acme \
JWT_SECRET=acme-secret \
JWT_SUBJECT=smoke-user \
JWT_EXPIRES_IN=3600 \
node ./scripts/generate-jwt.mjs
```

## Call the auth-check endpoint

```bash
TOKEN=$(node ./scripts/generate-jwt.mjs)

curl http://localhost:8080/v1/auth/check \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-App-ID: app-acme"
```
