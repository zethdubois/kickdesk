# KAM dev port families

> **Org-specific (Kick Asset Management).** Not required for Kickdesk subscribers. Port values belong in each app's **published** `~/.config/<app-id>/manifest.json` per [MANIFEST.md](MANIFEST.md). This file is an optional local reference for one organization's repos.

Canonical local port map when multiple KAM-style apps run on one machine.

## Glance rule

**First two digits = product decade; last two digits = role.**

| Decade | Product role              |
| ------ | ------------------------- |
| **50xx** | Web app (SvelteKit / Vite) |
| **70xx** | Agent / tooling HTTP       |
| **80xx** | API stack (API + gateway)  |

| Offset | Role | Examples |
|--------|------|----------|
| `+0` | Primary HTTP | 5000, 8000 |
| `+1` | Secondary HTTP (gateway) | 8001 |
| `+43` | Postgres on **host** (container still 5432 inside) | 5043, 8043 |
| `+99` | Tooling HTTP in 70xx (historical) | 7099 publish |

## Example allocation (this org)

| Family | Stack | Role | Port | Typical config source |
|--------|-------|------|------|------------------------|
| 50xx | Web | HTTP | **5000** | Vite `server.port` |
| 50xx | Web | Postgres | **5043** | compose / `DATABASE_URL` |
| 70xx | Agent | publish (contract server) | **7099** | env `*_PUBLISH_PORT` |
| 70xx | Agent | standalone (dev HTTP) | **7098** | dev server only |
| 80xx | API | FastAPI | **8000** | `API_PORT` |
| 80xx | API | gateway | **8001** | gateway `PORT` |
| 80xx | API | Postgres | **8043** | `POSTGRES_PORT` |

**Note:** HTTP `manifest.json` on port 7099 is a **runtime bundle contract** for one web+agent integration — not the same file as Kickdesk's `~/.config/<id>/manifest.json`.

## 70xx — two HTTP servers (agent stack)

| Port | Role | Typical use |
|------|------|-------------|
| **7099** | Publish / contract server | Integration consumers |
| **7098** | Standalone dev HTTP | Local smoke tests only |

Do not point integration consumers at the standalone port when the contract server is required.

## Ports we deliberately avoid on the host

| Port | Why |
|------|-----|
| 5173 | Vite default |
| 3000 | CRA / many frontends |
| 5432 | Docker Postgres default |

## After changing Postgres host ports

Recreate the Docker container (host mapping is fixed at create time). Update local `.env` if it still points at old ports.

## Adding a new app decade

Pick an unused decade (e.g. **52xx**, **81xx**), document it in your org notes and in the app's **published manifest**, and use `+0` / `+1` / `+43` for app / gateway / Postgres.
