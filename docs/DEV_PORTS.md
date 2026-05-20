# KAM dev port families

Canonical local port map for Kick Asset Management repos on one machine. Each **app manifest** should use the matching family; see [MANIFEST.md](MANIFEST.md) and [examples/manifest.sample.json](../examples/manifest.sample.json). Operator registry: [examples/config.json](../examples/config.json).

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
| 70xx | kickagent | publish (contract server) | **7099** | `KICKAGENT_PUBLISH_PORT`, `KICKAGENT_MANIFEST_URL` — **publicweb Phase 2** |
| 70xx | kickagent | standalone (dev HTTP) | **7098** | `PORT` in `standalone.ts` — kickagent-only; **do not** point PW here |
| 80xx | merch-api | FastAPI | **8000** | `API_PORT` |
| 80xx | merch-api | gateway | **8001** | `PORT` in `gateway/src/server.js` |
| 80xx | merch-api | Postgres | **8043** | `POSTGRES_PORT`, `DATABASE_URL` |

**Cross-family link:** publicweb loads kickagent from `http://127.0.0.1:7099/manifest.json` — intentional, not a collision.

## kickagent (70xx) — two HTTP servers

kickdesk `primary_port` and `url` use **7099** (publish / contract). Status also tracks **7098** (standalone).

| Port | Role | Start (typical) | Who uses it |
|------|------|-----------------|-------------|
| **7099** | Publish server — `manifest.json` + ESM bundle | `pnpm serve-publish` / kickdesk `up` | **publicweb**, Phase 2 integration |
| **7098** | Standalone dev HTTP — `/hello`, `/health` | `pnpm standalone` | Kickagent dev smoke tests only |

**kickdesk startup order:** `build` → `publish` (bundle + `manifest.json` + sha256 for `--base-url http://127.0.0.1:7099`) → `up` (serve `publish/` on 7099). Skipping `publish` after a rebuild leaves PW loading a stale hash or wrong `moduleUrl`.

- **`pnpm start`** in kickagent is the **CLI**, not standalone or `serve-publish`.
- Hosts must **not** aim publicweb at 7098 for the plugin contract; use 7099.

**Canonical detail in the kickagent repo:** `~/projects/kickagent/docs/ports.md` (`docs/ports.md` in that checkout).

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

Pick an unused decade (e.g. **52xx**, **81xx**), document it here and in the app's published manifest, and use `+0` / `+1` / `+43` for app / gateway / Postgres.
