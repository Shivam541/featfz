Feature flag client showcase for the `featfz` monorepo.

## Run

```bash
npm run dev
```

The app listens on `http://127.0.0.1:3001`.

## Notes

- The client mints JWTs locally and proxies requests through local App Router routes under `/api/*`.
- Default demo values:
  - `app-acme`
  - `acme-secret`
  - `client-user`
- Build command:
  - `npm run build`
