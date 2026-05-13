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
