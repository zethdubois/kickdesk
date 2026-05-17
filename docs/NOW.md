# Kam-Suite — NOW (v0)

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

Fill in the three tools you care about. Example shape:

| App | Repo | Port(s) | URL |
|-----|------|---------|-----|
| _app-a_ | `~/projects/...` | | `http://localhost:...` |
| _app-b_ | | | |
| _app-c_ | | | |

Optional later in config: `depends_on` so `up --all` starts in order.

## Configuration

**Location:** `~/.config/kam-suite/config.json`

**Schema (v0):**

```json
{
  "apps": {
    "app-a": {
      "path": "~/projects/foo",
      "ports": [3000],
      "url": "http://localhost:3000",
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

**`up` vs `serve`:** Use only `up` in v0 (dev server = up). Add `serve` later if an app splits infra from dev server.

## Commands (v0 only)

| Command | Behavior |
|---------|----------|
| `kam-suite status` | Table: app, running?, port(s) in use, git branch, dirty?, URL |
| `kam-suite up [app\|--all]` | Run `commands.up`; detect port conflicts first |
| `kam-suite down [app\|--all]` | Run `commands.down` |
| `kam-suite build [app\|--all]` | Run `commands.build`; parallel for `--all` |
| `kam-suite run <app> <key>` | Run `commands.<key>` (operate layer) |
| `kam-suite config validate` | Paths exist; ports not double-booked across apps |

Optional nice-to-have in v0 if cheap: `kam-suite focus <app>` — same as status but visually emphasizes one app.

## Status output

For each app, answer:

1. Is something listening on its port(s)?
2. Can we attribute it to a process we care about? (best-effort; perfect PID tracking is not required v0)
3. Git: branch + clean/dirty
4. Clickable URL

Colored terminal output is fine; keep the layout stable so muscle memory works.

## Implementation notes

- Run shell commands in each app's `path`
- Batch `up`/`down`/`build --all` may run concurrently unless `depends_on` is added
- Clear errors: which app, which command, stderr snippet
- Fast: no heavy polling; port check + git status should feel instant

## Definition of done (v0)

- [ ] Config file documents three real apps
- [ ] `status` is the daily driver
- [ ] `up` / `down` / `build` work per app and with `--all`
- [ ] `run` works for at least one custom command per app (e.g. `test`)
- [ ] `config validate` catches missing paths and port clashes
- [ ] Used for a full day of multi-app work without reaching for ad-hoc scripts

When that's true, pull ideas from [ROADMAP.md](ROADMAP.md) one at a time.
