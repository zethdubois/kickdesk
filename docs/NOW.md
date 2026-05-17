# Kickdesk — NOW (v0)

Everything in this doc is in scope for the first usable version. If it's not here, don't build it yet.

## Goal

One CLI cockpit for **three apps** during active development:

- **Status** — running or not, ports, localhost URLs, git branch/dirty
- **Build** — same verb everywhere, even if the underlying command differs
- **Operate** — arbitrary named tasks per app (test, migrate, logs, …)

## Non-goals (v0)

- Web dashboard
- Plugin / hook systems
- HTTP health checks
- CPU/memory metrics
- CI/CD integration
- Team / multi-user features
- Environment profiles (staging vs dev)
- Cross-platform single-binary requirement (Linux-first is fine)

## App registry

| App | Family | Ports (role) | URL |
|-----|--------|--------------|-----|
| `publicweb` | 50xx | app 5000, db 5043 | http://localhost:5000 |
| `kickagent` | 70xx | publish 7099 (PW contract), standalone 7098 (KA dev only) | http://127.0.0.1:7099/manifest.json |
| `merch-api` | 80xx | api 8000, gateway 8001, db 8043 | http://localhost:8000/docs |

See [DEV_PORTS.md](DEV_PORTS.md) for the org port matrix. Kickagent port semantics: `kickagent/docs/ports.md` in that repo. Template: [examples/config.json](../examples/config.json) → copy to `~/.config/kickdesk/config.json`.

Optional later in config: `depends_on` so `up --all` starts in order.

## Configuration

**Location:** `~/.config/kickdesk/config.json`

**Schema (v0):**

```json
{
  "app_order": ["publicweb", "kickagent", "merch-api"],
  "apps": {
    "app-a": {
      "path": "~/projects/foo",
      "primary_port": 3000,
      "ports": [3000],
      "url": "http://localhost:3000",
      "workflows": {
        "start": ["db-up", "up"],
        "stop": ["stop-server", "down"]
      },
      "commands": {
        "up": "npm run dev",
        "down": "…",
        "build": "npm run build",
        "test": "npm test",
        "migrate": "…"
      }
    }
  }
}
```

Per app:

| Field | Required | Notes |
|-------|----------|-------|
| `path` | yes | Repo root; commands run here |
| `ports` | yes | Used for status + conflict check before `up` |
| `url` | recommended | Shown in `status` |
| `commands` | yes | At minimum `up`, `down`, `build`; any other key is operable via `run` |

**Optional `migrate-status` command:** If present, kickdesk runs it when the app's **db** role port is up and shows a **MIGRATE** column in `status` / the hub table. Stdout contract (one line):

| Output | Meaning |
|--------|---------|
| `ok` | DB reachable, schema matches repo migrations |
| `pending:N` | N migrations not yet applied (N > 0) |
| `unavailable` | Cannot determine (env, connection, etc.) |

Exit `0` for `ok` and `pending:N`; non-zero on unexpected failure (kickdesk shows `error`). Workflow is unchanged — run `migrate` when the column shows pending.

**`up` vs `serve`:** Use only `up` in v0 (dev server = up). Add `serve` later if an app splits infra from dev server.

## Commands (v0 only)

| Command | Behavior |
|---------|----------|
| `kickdesk` / `kickdesk menu` | Single hotkey screen: status table + optional workflow under selected app (1–3) |
| `kickdesk status` | Table: app, ports, migrate pending?, git branch, dirty?, URL |
| `kickdesk run <app> <key>` | Run `commands.<key>` in app repo (non-interactive) |
| `kickdesk config validate` | Paths, ports, workflows, required commands |

**Menu keys (no Enter required):**

| Context | Keys |
|---------|------|
| Hub | `1`–`3` select app, `r` refresh, `q` quit |
| App selected | `Space` next step, `Enter` all remaining, `1`–`N` run one step, `b`/`Esc` back, `c` catalog, `r` refresh, `q` quit |

After a full procedure or **Enter**-all, **Space** returns to the menu so logs stay visible. Blocking commands (`up`, dev servers) open in a new terminal when possible (`KICKDESK_TERMINAL` or auto-detect). The `migrate` workflow step is omitted when **MIGRATE** is `ok` or `n/a`; it stays when `pending:N` or `unavailable`.

Config adds `app_order`, `primary_port`, and per-app `workflows.start` / `workflows.stop` (ordered command keys).

Deferred: dedicated `kickdesk up|down|build` subcommands and `--all` batch.

Optional nice-to-have in v0 if cheap: `kickdesk focus <app>` — same as status but visually emphasizes one app.

## Status output

For each app, answer:

1. Is something listening on its port(s)?
2. Can we attribute it to a process we care about? (best-effort; perfect PID tracking is not required v0)
3. Git: branch + clean/dirty
4. Migrations: `ok`, `pending:N`, or `unavailable` (when `migrate-status` is configured and db port is up)
5. Clickable URL

Colored terminal output is fine; keep the layout stable so muscle memory works.

## Implementation notes

- Run shell commands in each app's `path`
- Batch `up`/`down`/`build --all` may run concurrently unless `depends_on` is added
- Clear errors: which app, which command, stderr snippet
- Fast: no heavy polling; port check + git status should feel instant

## Definition of done (v0)

- [x] Config file documents three real apps
- [x] `status` is the daily driver
- [ ] `up` / `down` / `build` work per app and with `--all` (via menu/`run` today)
- [x] `run` works for at least one custom command per app (e.g. `test`)
- [x] `config validate` catches missing paths and port clashes
- [ ] Used for a full day of multi-app work without reaching for ad-hoc scripts

When that's true, pull ideas from [ROADMAP.md](ROADMAP.md) one at a time.
