#!/bin/sh

set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
API_BASE="${API_BASE:-http://127.0.0.1:8080}"
APP_ID_ACME="${APP_ID_ACME:-app-acme}"
APP_ID_GLOBEX="${APP_ID_GLOBEX:-app-globex}"
JWT_SECRET_ACME="${JWT_SECRET_ACME:-acme-secret}"
JWT_SECRET_GLOBEX="${JWT_SECRET_GLOBEX:-globex-secret}"
USER_ID="${USER_ID:-user_123}"
FLAG_KEY="${SMOKE_FLAG_KEY:-phase8_eval_smoke_$(uuidgen | tr '[:upper:]' '[:lower:]' | tr -d '-')}"

ACME_JWT="$(APP_ID="$APP_ID_ACME" JWT_SECRET="$JWT_SECRET_ACME" JWT_SUBJECT="${JWT_SUBJECT:-smoke-user}" node "$ROOT_DIR/../feat-client/scripts/generate-jwt.mjs")"
GLOBEX_JWT="$(APP_ID="$APP_ID_GLOBEX" JWT_SECRET="$JWT_SECRET_GLOBEX" JWT_SUBJECT="${JWT_SUBJECT:-smoke-user}" node "$ROOT_DIR/../feat-client/scripts/generate-jwt.mjs")"

printf 'Using flag key: %s\n' "$FLAG_KEY"

curl -sS -X POST "$API_BASE/v1/flags" \
  -H "Authorization: Bearer $ACME_JWT" \
  -H "X-App-ID: $APP_ID_ACME" \
  -H "Content-Type: application/json" \
  -d "{\"key\":\"$FLAG_KEY\",\"description\":\"phase 8 smoke\",\"default_enabled\":false}"
printf '\n'

curl -sS -X POST "$API_BASE/v1/flags" \
  -H "Authorization: Bearer $GLOBEX_JWT" \
  -H "X-App-ID: $APP_ID_GLOBEX" \
  -H "Content-Type: application/json" \
  -d "{\"key\":\"$FLAG_KEY\",\"description\":\"phase 8 smoke\",\"default_enabled\":false}"
printf '\n'

curl -sS -X POST "$API_BASE/v1/flags/$FLAG_KEY/users/bulk-set" \
  -H "Authorization: Bearer $ACME_JWT" \
  -H "X-App-ID: $APP_ID_ACME" \
  -H "Content-Type: application/json" \
  -d "{\"overrides\":[{\"user_id\":\"$USER_ID\",\"enabled\":true}]}"
printf '\n'

printf 'acme eval: '
curl -sS "$API_BASE/eval?flag=$FLAG_KEY&user=$USER_ID" \
  -H "Authorization: Bearer $ACME_JWT" \
  -H "X-App-ID: $APP_ID_ACME"
printf '\n'

printf 'globex eval: '
curl -sS "$API_BASE/eval?flag=$FLAG_KEY&user=$USER_ID" \
  -H "Authorization: Bearer $GLOBEX_JWT" \
  -H "X-App-ID: $APP_ID_GLOBEX"
printf '\n'
