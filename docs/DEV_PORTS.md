# KAM dev port families

Canonical local port map for Kick Asset Management repos on one machine. **kickdesk** reads [examples/config.json](../examples/config.json); each app repo should match the same numbers in compose, `.env.example`, and framework config.

## Glance rule

**First two digits = product; last two digits = role.**

| Decade | Product | Repo |
|--------|---------|------|
| **50xx** | Web app (SvelteKit / Vite) | publicweb |
| **70xx** | Agent tooling (manifest, standalone) | kickagent |
| **80xx** | Merch API stack | merch-api |

| Offset | Role | Examples |
|--------|------|----------|
| `+0` | Primary HTTP | 5000, 8000 |
| `+1` | Secondary HTTP (gateway) | 8001 |
| `+43` | Postgres on **host** (container still 5432 inside) | 5043, 8043 |
| `+99` | Tooling HTTP in 70xx (historical) | 7099 publish |

## Allocation

| Family | App | Role | Port | Env / config |
|--------|-----|------|------|----------------|
| 50xx | publicweb | HTTP | **5000** | `vite.config.ts` `server.port`, `strictPort: true` |
| 50xx | publicweb | Postgres | **5043** | `DATABASE_URL_DEV`, `scripts/compose.sh` |
| 70xx | kickagent | manifest publish | **7099** | `KICKAGENT_PUBLISH_PORT`, `PUBLIC_KICKAGENT_MANIFEST_URL` |
| 70xx | kickagent | standalone API | **7098** | `PORT` in `standalone.ts` |
| 80xx | merch-api | FastAPI | **8000** | `API_PORT` |
| 80xx | merch-api | gateway | **8001** | `PORT` in `gateway/src/server.js` |
| 80xx | merch-api | Postgres | **8043** | `POSTGRES_PORT`, `DATABASE_URL` |

**Cross-family link:** publicweb loads kickagent from `http://127.0.0.1:7099/manifest.json` — intentional, not a collision.

## Ports we deliberately avoid on the host

| Port | Why |
|------|-----|
| 5173 | Vite default — every tutorial uses it |
| 3000 | CRA / many frontends |
| 5432 | Every Docker Postgres example |

Other projects on your laptop can keep using those; KAM repos use the table above.

## After changing Postgres host ports

Recreate the Docker container (host mapping is fixed at create time):

```bash
# publicweb
pnpm db:down && pnpm db:up

# merch-api
make local-db-down && make local-db-up
# or: make local-db-reset
```

Update your local `.env` if it still points at old ports (5433 or 5432).

## Adding a new repo

Pick an unused decade (e.g. **52xx**, **81xx**), document it here and in `examples/config.json`, and use `+0` / `+1` / `+43` for app / gateway / Postgres.
