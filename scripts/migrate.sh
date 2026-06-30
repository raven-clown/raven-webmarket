#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="${DB_NAME:-raven_webmarket}"
MYSQL=(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER")
if [ -n "$DB_PASSWORD" ]; then
  MYSQL+=(-p"$DB_PASSWORD")
fi
"${MYSQL[@]}" -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
"${MYSQL[@]}" "$DB_NAME" < "$ROOT/database/migrations/001_init.sql"
echo "Migration complete: $DB_NAME"
