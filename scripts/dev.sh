#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
if [ ! -f .env ]; then
  cp .env.example .env
  echo "Created .env from .env.example — edit before production use."
fi
docker compose -f deploy/docker-compose.yml up -d mariadb redis minio
bash "$ROOT/scripts/migrate.sh"
(
  cd "$ROOT/backend"
  go run ./cmd/server
) &
BACKEND_PID=$!
(
  cd "$ROOT/frontend"
  npm run dev
) &
FRONTEND_PID=$!
trap 'kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true' EXIT
wait
