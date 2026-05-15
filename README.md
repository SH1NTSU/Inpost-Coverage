# InPost Scraper

Minimal scraper-only scaffold for collecting InPost point status snapshots.

## Run locally

1. `make docker-up`
2. `make migrate-up`
3. `make run-scraper`

The scraper performs one scrape immediately at startup, then continues on
`SCRAPER_INTERVAL` (default `5m`). If you want to let it collect overnight,
start `make run-scraper` and leave that process running.

## Data model

- `points` stores the latest static metadata for each InPost point.
- `availability_snapshots` stores only `point_id`, `captured_at`, and `status`.

## Configuration

Use `.env.example` as the reference for the required environment variables.

## Deploy (Vercel + API + database)

### What belongs on Vercel

- **React/Vite UI (`web/`)** — yes. Vercel is ideal for the static SPA and CDN.
- **PostgreSQL** — not as a process you run on Vercel. Use a **managed Postgres** (Neon, Supabase, Vercel Postgres, Railway, Render, RDS, etc.) and set `DATABASE_URL` on the API host.
- **Go HTTP API (`cmd/api`)** — not as a long-lived server on Vercel’s default model. Vercel is built around serverless/edge, not a always-on Chi server. Run the API on **Fly.io**, **Railway**, **Render**, **Google Cloud Run**, **DigitalOcean App Platform**, or similar, using `Dockerfile.api` or `go run ./cmd/api` with your env vars.
- **Scraper (`cmd/scraper`)** — same as the API: a separate worker/container or cron job on one of those platforms, not Vercel.

### Vercel (frontend)

1. In the Vercel project, set **Root Directory** to `web` (this repo is Go + web; only `web` should build on Vercel).
2. Under **Environment Variables** (for *Production* and *Preview* if you use previews), set:
The UI calls your API using `VITE_API_BASE_URL` plus paths such as `/api/v1/...`. Leaving this unset in production sends `/api` to Vercel itself and will not work.
3. `web/vercel.json` adds a SPA fallback so React Router paths resolve to `index.html`.

### Go API (any container-friendly host)

1. Build with `Dockerfile.api` at the repo root, or build the binary with `go build -o api ./cmd/api`.
2. Set `DATABASE_URL`, `CORS_ALLOWED_ORIGINS` (include your `https://*.vercel.app` and custom domain), optional `REDIS_ADDR`, and the rest from `.env.example`.
3. Run migrations against that database (`go run ./cmd/migrate up` with the same `DATABASE_URL`).

### CORS

`CORS_ALLOWED_ORIGINS` must list every browser origin that talks to the API (comma-separated), e.g. `http://localhost:5173` and `https://your-project.vercel.app`.
