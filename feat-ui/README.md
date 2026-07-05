Feature flag admin UI for the `featfz` monorepo.

## Routes

- `/` module launcher
- `/flags` list, view, and evaluate flags
- `/manage` create, update, and archive flags

## Run

```bash
npm run dev
```

The app listens on `http://127.0.0.1:3003`.

## Notes

- App credentials stay hidden behind the `App details` disclosure unless expanded.
- The UI proxies API requests through local App Router routes under `/api/*`.
- Default demo values:
  - `app-acme`
  - `acme-secret`
  - `dashboard-user`
- Build command:
  - `npm run build`
