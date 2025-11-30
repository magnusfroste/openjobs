# OpenJobs API File & Directory Layout

This document summarizes the structure of the `openjobs-api` repository so contributors can quickly locate code, docs, and deployment assets without moving any existing files.

## Top-Level Inventory

| Path | Files* | Purpose |
| --- | --- | --- |
| `cmd/` | 13 | Service entrypoints (API + standalone plugins) |
| `internal/` | 6 | Core packages (API handlers, database, middleware, scheduler) |
| `pkg/` | 6 | Shared domain models and storage helpers |
| `connectors/` | 25 | Source-specific connectors (code + Dockerfiles + docs) |
| `docs/` | 8 | Documentation hub (architecture, deployment, guides) |
| `migrations/` | 10 | SQL migrations for Supabase/Postgres |
| `scripts/` | 2 | Operational scripts (connectivity tests, helpers) |
| `Dockerfile`, `docker-compose.plugins.yml`, `deploy-*.sh` | 3 | Container builds & orchestration |
| `*.md` in repo root | 38 | Ops notes, incident reports, and fix guides |

<sub>*File counts are from `python3` walk output on Nov 30, 2025 (excluding hidden files).</sub>

## Source Code Layout

- **`cmd/`**
  - `openjobs/` – Main API binary (port 8080); wires handlers, scheduler, and middleware.
  - `plugin-*/` – Individual connector servers (e.g., `plugin-indeed`, `plugin-jooble`). Each exposes `/health`, `/sync`, `/jobs`.
- **`internal/`**
  - `api/` – HTTP handlers (jobs CRUD, analytics, plugin status, sync logs, API key validation).
  - `database/` – Supabase/Postgres connection bootstrap.
  - `middleware/` – CORS + shared HTTP middleware.
  - `scheduler/` – Plugin registry, cron job orchestration, HTTP plugin mode toggle.
- **`pkg/`**
  - `models/` – `JobPost`, `SyncLog`, `PluginInfo`, and shared response structs.
  - `storage/` – Database accessors used by API + connectors.

## Connectors

`connectors/` hosts one folder per external data source. Each folder typically contains:

- `connector.go` – Fetch/transform logic implementing `PluginConnector`.
- `README.md` – Source-specific notes, rate limits, environment variables.
- `Dockerfile` – Container entrypoint used by Easypanel and docker-compose.

Current production connectors (Nov 30, 2025): Arbetsförmedlingen, EURES (Adzuna), Remotive, RemoteOK, Indeed (HTTP), Indeed Chrome scraper, Jooble. Experimental directories (e.g., `jobbland/`) are empty placeholders that should stay untouched until ready.

## Documentation

- `docs/README.md` – Table of contents for architecture, deployment, connector guides, and migrations.
- `docs/architecture/` – System diagrams, microservices migration rationale.
- `docs/deployment/` – Containers overview + docker-compose guide.
- `docs/connectors/` & `docs/migrations/` – Reserved namespaces for future deep-dives.

The repo root also contains operational playbooks (e.g., `INCIDENT_REPORT_2025-11-28.md`, `EASYPANEL_*` guides). These stay at the top level so incidents can be tracked chronologically.

## Infrastructure & Deployment Assets

- `Dockerfile` – Builds the monolithic API container.
- `docker-compose.plugins.yml` – Spins up API + plugin microservices locally.
- `deploy-easypanel.sh`, `deploy-microservices.sh` – Helper scripts for Easypanel / multi-service deployments.
- `infrastructure.json` – Infra metadata consumed by Easypanel.

## Tests & Utilities

- `test-single-job-scrape.go` – Quick manual scraper test harness.
- `scripts/test_supabase_connection.sh` – Verifies Supabase credentials from CI/local shells.

---
Maintaining this structure keeps documentation discoverable while ensuring connectors and API services stay isolated. Update this document whenever new top-level directories or major doc collections are added.
