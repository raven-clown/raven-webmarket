# Raven Webmarket — Production Deployment Guide

This guide covers running Raven Webmarket **without keeping a CMD/terminal window open**. Services run as background daemons (Linux systemd) or Windows Services (NSSM).

For development-only quick start, see [README.md](./README.md).

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Variables](#environment-variables)
3. [Linux Production (systemd)](#linux-production-systemd)
4. [Windows Production (Windows Service)](#windows-production-windows-service)
5. [Docker Production (Recommended)](#docker-production-recommended)
6. [Kubernetes & HPA Autoscale](#kubernetes--hpa-autoscale)
7. [Cloudflare Security Setup](#cloudflare-security-setup)
8. [Admin Roles (Separate from Player Login)](#admin-roles-separate-from-player-login)
9. [Frontend Features](#frontend-features)
10. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Linux (Ubuntu/Debian)

| Package | Install |
|---------|---------|
| Go 1.22+ | `sudo apt install golang-go` or [go.dev/dl](https://go.dev/dl/) |
| Node.js 20+ | `curl -fsSL https://deb.nodesource.com/setup_20.x \| sudo -E bash - && sudo apt install -y nodejs` |
| MariaDB 11+ | `sudo apt install mariadb-server` or Docker |
| Redis 7+ | `sudo apt install redis-server` or Docker |
| MinIO | Docker (`deploy/docker-compose.yml`) |
| mysql client | `sudo apt install mariadb-client` |
| Git | `sudo apt install git` |

### Windows Server / Windows 10+

| Package | Install |
|---------|---------|
| Go 1.22+ | [go.dev/dl](https://go.dev/dl/) |
| Node.js 20 LTS | [nodejs.org](https://nodejs.org/) |
| MariaDB | Installer or Docker Desktop |
| Redis | Docker Desktop or Memurai |
| MinIO | Docker Desktop |
| MySQL CLI | Included with MariaDB installer |
| Git | Git for Windows |

### Optional (All Platforms)

| Tool | Purpose |
|------|---------|
| Docker + Docker Compose | Easiest production stack, no manual service setup |
| Cloudflare | CDN, WAF, DDoS protection, SSL |
| Nginx / Caddy | Reverse proxy in front of API + frontend |
| Certbot | Let's Encrypt SSL (if not using Cloudflare Full SSL) |

---

## Environment Variables

Copy and edit `.env`:

```bash
cp .env.example .env
```

### Critical Production Settings

| Variable | Description |
|----------|-------------|
| `APP_ENV` | Set to `production` |
| `SESSION_SECRET` / `JWT_SECRET` | Long random strings (never use defaults) |
| `DISCORD_CLIENT_ID` / `DISCORD_CLIENT_SECRET` | Discord OAuth app |
| `DISCORD_REDIRECT_URI` | Must match Discord app (e.g. `https://api.yourdomain.com/api/v1/auth/callback`) |
| `FRONTEND_URL` | Public shop URL (e.g. `https://shop.yourdomain.com`) |
| `API_BASE_URL` | Public API URL (e.g. `https://api.yourdomain.com`) |
| `CORS_ORIGINS` | Comma-separated allowed origins |
| `TRUST_CLOUDFLARE` | Set `true` when behind Cloudflare (uses `CF-Connecting-IP`) |
| `TRUSTED_PROXIES` | CIDR list of your reverse proxy (default includes private ranges) |
| `PAYMENT_WEBHOOK_SECRET` | HMAC secret for payment webhooks (`X-Webhook-Signature` header) — **required in production** |
| `FIVEM_WEBHOOK_SECRET` | Must match FiveM resource secret |
| `FIVEM_API_KEY` | Bearer token for `/api/v1/game/mailbox/*` (set if those routes are reachable from internet) |
| `NEXT_PUBLIC_API_URL` | Public API URL for browser |
| `NEXT_PUBLIC_APP_URL` | Public frontend URL |
| `MONGO_ENABLED` | `true` to connect optional MongoDB side store |
| `MONGO_URI` | Full URI (overrides host/port/user if set) |
| `MONGO_HOST` / `MONGO_PORT` | Default `127.0.0.1:27017` |
| `MONGO_USER` / `MONGO_PASSWORD` | MongoDB credentials |
| `MONGO_DB_NAME` | Database name (default `raven_webmarket`) |

---

## Linux Production (systemd)

### Automated Install

```bash
sudo bash scripts/install-linux.sh
sudo nano /opt/raven-webmarket/.env
bash /opt/raven-webmarket/scripts/migrate.sh
sudo systemctl start raven-api raven-frontend
```

### Manual Steps

```bash
# 1. Build API binary
cd backend && go build -o raven-api ./cmd/server

# 2. Build frontend (standalone)
cd frontend
npm ci && npm run build
# Copy .next/standalone output to deployment path

# 3. Install systemd units
sudo cp deploy/systemd/*.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now raven-api raven-frontend
```

### Service Management

```bash
sudo systemctl status raven-api
sudo systemctl restart raven-frontend
journalctl -u raven-api -f
```

Services restart automatically on crash or reboot. **No terminal window needed.**

---

## Windows Production (Windows Service)

Run PowerShell **as Administrator**:

```powershell
.\scripts\install-windows.ps1
notepad C:\RavenWebmarket\.env
cd C:\RavenWebmarket
.\scripts\migrate.ps1
```

Services install via NSSM and run hidden in the background.

### Management

- Open `services.msc` → find **RavenWebmarketAPI** and **RavenWebmarketFrontend**
- Logs: `C:\RavenWebmarket\logs\`
- Restart: `Restart-Service RavenWebmarketAPI`

---

## Docker Production (Recommended)

Best option for 24/7 uptime without manual process management:

```bash
cp .env.example .env
# Edit .env for production URLs and secrets

docker compose -f deploy/docker-compose.yml up -d --build
bash scripts/migrate.sh
```

Docker Compose sets `restart: always` — containers auto-restart on failure or host reboot.

| Service | Port |
|---------|------|
| Frontend | 3000 |
| API | 8080 |
| MariaDB | 3306 |
| Redis | 6379 |
| MinIO | 9000 / 9001 |

Put Nginx or Cloudflare Tunnel in front for HTTPS.

### Optional MongoDB

MongoDB is **not required** for the shop. Enable when you want a document store for future features (analytics, event archive, etc.).

**Start MongoDB with Docker:**

```bash
docker compose -f deploy/docker-compose.yml --profile mongo up -d mongodb
```

**Enable in `.env`:**

```env
MONGO_ENABLED=true
MONGO_HOST=127.0.0.1
MONGO_PORT=27017
MONGO_USER=raven
MONGO_PASSWORD=changeme
MONGO_DB_NAME=raven_webmarket
```

Or use a single URI:

```env
MONGO_ENABLED=true
MONGO_URI=mongodb://raven:changeme@127.0.0.1:27017/raven_webmarket
```

Check status in Admin → **Health & Monitoring** (`mongodb: up` or `disabled`).

---

## Kubernetes & HPA Autoscale

For production clusters with automatic pod scaling. **Owner baseline:** Intel i9 14th Gen, 64 GB RAM, GTX 770 (GPU unused by shop stack).

### Owner hardware profile

| Component | Spec |
|-----------|------|
| CPU | Intel Core i9 14th Gen (~24 cores / 32 threads) |
| RAM | 64 GB |
| GPU | NVIDIA GeForce GTX 770 (not required for API/frontend) |

HPA and pod resources in this repo are tuned for this host. See [README.md — Reference Hardware Profile](./README.md#reference-hardware-profile-autoscale-baseline) to scale down/up for other machines.

### Prerequisites

| Requirement | Notes |
|-------------|-------|
| Kubernetes 1.25+ | EKS, GKE, AKS, k3s, etc. |
| `kubectl` configured | Must point to your cluster |
| **metrics-server** | Required for HPA — without it pods will not scale |
| Container images | Build/push `raven-webmarket/api` and `raven-webmarket/frontend` to your registry |
| MariaDB, Redis, MinIO | In-cluster or external (set `DB_*`, `REDIS_*` in secret) |

Install metrics-server if missing:

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl get apiservice v1beta1.metrics.k8s.io
```

### One-Command Deploy

**Linux / macOS / Git Bash:**

```bash
cp .env.example .env
# Edit .env for production
bash scripts/migrate.sh
bash scripts/k8s-apply.sh
```

**Windows PowerShell:**

```powershell
Copy-Item .env.example .env
.\scripts\migrate.ps1
.\scripts\k8s-apply.ps1
```

### Manual Deploy

```bash
kubectl apply -k deploy/kubernetes/
kubectl create secret generic raven-env --from-env-file=.env -n raven-webmarket --dry-run=client -o yaml | kubectl apply -f -
```

### Manifest Layout

```
deploy/kubernetes/
├── kustomization.yaml      # Apply all resources
├── namespace.yaml          # raven-webmarket namespace
├── api-deployment.yaml     # API Deployment + Service
├── frontend-deployment.yaml
├── hpa-api.yaml            # HPA — API (min 3, max 16, 55% CPU / 65% RAM)
├── hpa-frontend.yaml       # HPA — Frontend (min 2, max 12, 55% CPU / 65% RAM)
└── servicemonitor.yaml     # Prometheus scrape /metrics
```

### HPA Defaults (i9-14 / 64 GB profile)

| Target | Min Pods | Max Pods | CPU Target | Memory Target | Pod CPU limit | Pod RAM limit |
|--------|----------|----------|------------|---------------|---------------|---------------|
| API | 3 | 16 | 55% | 65% | 1500m | 1 GiB |
| Frontend | 2 | 12 | 55% | 65% | 1000m | 768 MiB |

After clone/migrate, run migration `004_autoscale_i9_profile.sql` (via `scripts/migrate.sh`) to sync admin UI defaults.

### Tune Autoscale from Admin UI

1. Login as **dev_admin** → `/admin/autoscale`
2. Adjust min/max replicas and CPU/memory targets
3. Click **Save** → copy generated YAML
4. Apply to cluster:

```bash
kubectl apply -f deploy/kubernetes/hpa-api.yaml
kubectl apply -f deploy/kubernetes/hpa-frontend.yaml
```

Or paste YAML from admin preview into the files above, then apply.

### Verify HPA

```bash
kubectl get hpa -n raven-webmarket
kubectl describe hpa raven-api-hpa -n raven-webmarket
kubectl top pods -n raven-webmarket
kubectl get hpa -n raven-webmarket -w
```

### Build & Push Images (before deploy)

```bash
docker build -f deploy/Dockerfile.api -t your-registry/raven-webmarket/api:latest .
docker build -f deploy/Dockerfile.frontend -t your-registry/raven-webmarket/frontend:latest .
docker push your-registry/raven-webmarket/api:latest
docker push your-registry/raven-webmarket/frontend:latest
```

Update image names in `deploy/kubernetes/*-deployment.yaml` to match your registry.

### Ingress (optional)

Expose services with your ingress controller or Cloudflare Tunnel pointing to `raven-frontend` and `raven-api` services in namespace `raven-webmarket`.

---

## Cloudflare Security Setup

### 1. DNS & SSL

- Point `shop.yourdomain.com` → frontend (port 3000 or reverse proxy)
- Point `api.yourdomain.com` → API (port 8080)
- SSL/TLS mode: **Full (strict)** with origin certificate or Let's Encrypt

### 2. Enable in `.env`

```env
TRUST_CLOUDFLARE=true
TRUSTED_PROXIES=173.245.48.0/20,103.21.244.0/22,103.22.200.0/22,103.31.4.0/22,141.101.64.0/18,108.162.192.0/18,190.93.240.0/20,188.114.96.0/20,197.234.240.0/22,198.41.128.0/17,162.158.0.0/15,104.16.0.0/13,104.24.0.0/14,172.64.0.0/13,131.0.72.0/22
```

Or use simplified private-range trust if Cloudflare Tunnel connects directly:

```env
TRUST_CLOUDFLARE=true
```

### 3. Cloudflare Dashboard (WAF)

Enable these under **Security → WAF**:

| Rule | Purpose |
|------|---------|
| OWASP Core Ruleset | Block common web attacks |
| Rate limiting on `/api/v1/auth/*` | Prevent OAuth brute force |
| Rate limiting on `/api/v1/orders/checkout` | Anti-spam purchase |
| Rate limiting on `/api/v1/redeem` | Anti-spam redeem |
| Bot Fight Mode | Block automated abuse |

### 4. Built-in API Protections

The Go API already includes:

- Redis rate limiting (per real IP via Cloudflare)
- Redis distributed locks on checkout, redeem, payment webhooks
- `SELECT FOR UPDATE` on stock and balances
- Security headers (`X-Content-Type-Options`, `X-Frame-Options`, etc.)
- Payment webhook HMAC verification (`X-Webhook-Signature` when `PAYMENT_WEBHOOK_SECRET` is set)
- Payment webhook **blocked in production** if `PAYMENT_WEBHOOK_SECRET` is empty
- Game mailbox endpoints require `Authorization: Bearer <FIVEM_API_KEY>` or `X-Webhook-Secret: <FIVEM_WEBHOOK_SECRET>`
- Admin JWT separate from player JWT (`raven_admin_token` vs `raven_token`)

### 5. Recommended Cloudflare Page Rules

- Cache Level: Bypass for `api.yourdomain.com/*`
- Cache Level: Standard for `shop.yourdomain.com` static assets

---

## Admin Roles (Separate from Player Login)

**Important:** Shop admin access is **NOT** tied to Discord player login or `DISCORD_ADMIN_IDS` for the backoffice.

### Role System

| Role | DB Value | Access |
|------|----------|--------|
| **Admin** | `admin` | CMS, users, KPI, audit logs, activity, purchase logs, monitoring view |
| **Developer Admin** | `dev_admin` | Everything admin has + security, autoscale, cache reset, admin account management |

### Default Accounts (change passwords immediately)

After migration `002_admin_rbac.sql`:

| Username | Default Password | Role |
|----------|------------------|------|
| `devadmin` | `ChangeMeDev123!` | dev_admin |
| `admin` | Set via dev_admin → Security page | admin |

Also shown on the login page at `/admin/login`.

### Login

1. Open `https://shop.yourdomain.com/admin/login`
2. Enter username + password (NOT Discord OAuth)
3. Admin JWT stored in `raven_admin_token` cookie

### Discord ID on Admin Accounts

When creating admin accounts (dev_admin only), you can optionally set a `discord_id` field for audit log attribution. This does **not** grant admin access via Discord login — it is metadata only.

### Player vs Admin

| | Player (Shop) | Admin (Backoffice) |
|--|---------------|-------------------|
| Login | Discord OAuth | Username + password |
| Token cookie | `raven_token` | `raven_admin_token` |
| Requires ESX DB | Yes | No |
| Role source | ESX `users.discord_id` | `admin_accounts.role` |

---

## Frontend Features

| Feature | Path | Description |
|---------|------|-------------|
| Shop | `/shop` | Product catalog with filters |
| Announcements | `/announcements` | Official notices (EN/TH) |
| Daily Updates | `/news` | Patch notes and changelog |
| Forum | `/forum` | Community threads (login required to post) |
| Cookie Consent | All pages | GDPR-style banner, stores preference |
| Language | Navbar EN/TH | Default English, persists in localStorage |
| Ads | Homepage sidebar | Managed via CMS (`post_type=ad`) |

CMS content managed at `/admin/cms` → **Announcements / Ads / Updates** panel.

---

## Troubleshooting

### API returns wrong client IP behind Cloudflare

Set `TRUST_CLOUDFLARE=true` in `.env` and restart API.

### Payment webhook rejected

Ensure gateway sends header `X-Webhook-Signature` = HMAC-SHA256 hex of raw body using `PAYMENT_WEBHOOK_SECRET`.

### Forum posts fail

Player must be logged in via Discord OAuth with a linked ESX account.

### Services won't start (Linux)

```bash
journalctl -u raven-api -n 50
# Check DB/Redis connectivity in .env
```

### Windows service stops immediately

Check `C:\RavenWebmarket\logs\RavenWebmarketAPI-error.log`

### HPA not scaling

```bash
kubectl get apiservice v1beta1.metrics.k8s.io
kubectl describe hpa raven-api-hpa -n raven-webmarket
```

Install metrics-server if missing (see DEPLOYMENT.md → Kubernetes section).

### Frontend empty announcements

Run migration 003: `bash scripts/migrate.sh` or `.\scripts\migrate.ps1`

---

## Reverse Proxy Example (Nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name shop.yourdomain.com;
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

See also: [README.md](./README.md) | [WEBSITE.MD](./WEBSITE.MD) | [PROGRESS.md](./PROGRESS.md)
