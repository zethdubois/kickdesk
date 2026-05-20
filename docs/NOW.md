# Kickdesk — NOW (v0)

Everything in this doc is in scope for the first usable version. If it's not here, don't build it yet.

## Goal

One CLI cockpit for **multiple apps** during active development:

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

Kickdesk is app-agnostic. Each app publishes its own ports and commands; the operator registry only lists **app ids** (and optional labels).

| Port decade | Typical role        | Example offsets                          |
| ----------- | ------------------- | ---------------------------------------- |
| **50xx**    | Web app (Vite/Kit)  | app +0, Postgres host +43                |
| **70xx**    | Agent / tooling HTTP| publish +99, standalone +98 (if used)    |
| **80xx**    | API stack           | API +0, gateway +1, Postgres host +43    |

Org-specific port tables (if you use them): [DEV_PORTS.md](DEV_PORTS.md) — optional reference, not part of the Kickdesk subscriber contract.

Optional later in config: `depends_on` so `up --all` starts in order.

## Configuration (v0.6 modular)

**Cockpit:** `~/.config/kickdesk/config.json` — app ids and optional labels only (`default_profile` optional).

**Per-app manifest:** published to `~/.config/<id>/manifest.json` — includes **`path`** (checkout), ports, workflows, commands. See [MANIFEST.md](MANIFEST.md). Kickdesk does **not** load repo `.kickdesk.json`; missing publish fails with a clear error.

**Profiles:** `~/.config/kickdesk/<name>.json` — `app_order`, labels, tmux layout. **`last.cnfg`** stores the last profile. `kickdesk` loads `last.cnfg`; `kickdesk -c` opens the picker.

**Legacy:** monolithic `config.json` with full `apps.<id>` inline still works ([examples/config.legacy.json](../examples/config.legacy.json)).

**Setup:**

```bash
cp examples/config.json ~/.config/kickdesk/config.json
cp examples/profiles/two-up.json ~/.config/kickdesk/two-up.json
# Each app: publish manifest to ~/.config/<app-id>/manifest.json (includes path)
kickdesk config validate
```

Examples: [examples/config.json](../examples/config.json), [examples/manifest.sample.json](../examples/manifest.sample.json). Profile examples use `fixture-*` app ids; replace with your registry ids.

## Migrate status

**File (preferred when published):** `status.migrate` in manifest → one-line file (`ok`, `pending:N`, `unavailable`).

**Command (fallback):** `migrate-status` in `commands` — kickdesk runs it when the **db** port is up. Stdout contract (one line):

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
| `kickdesk` | Menu using profile name in `~/.config/kickdesk/last.cnfg` |
| `kickdesk -c` | Profile picker (hotkeys): **0** config editor (placeholder), **1–N** each `*.json` except `config.json` |
| `kickdesk -t N` | **Deprecated** — use `-c` profile with `tmux.top`; still works with a warning |
| `kickdesk status` | Table: app, ports, migrate pending?, git branch, dirty?, URL |
| `kickdesk run <app> <key>` | Run `commands.<key>` in app repo (non-interactive) |
| `kickdesk config validate` | Paths, ports, workflows, required commands, manifest paths |

**Menu keys (no Enter required):**

| Context | Keys |
|---------|------|
| Hub | `1`–`3` select app, `r` refresh, `q` quit |
| App selected | `Space` next step, `Enter` all remaining, `1`–`N` run one step, `b`/`Esc` back, `c` catalog, `r` refresh, `q` quit |

After each step, **any key** continues. After a full procedure, **any key** returns to the menu. Without a tmux profile, blocking commands (`up`, dev servers) open in a new GUI terminal when possible (`KICKDESK_TERMINAL` or auto-detect). With `-c` + `tmux.enabled`, they run in the named tmux pane (`kickdesk-<app-id>`, etc.). The `migrate` workflow step is omitted when **MIGRATE** is `ok` or `n/a`. Selected app row is highlighted in the hub table.

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

- Run shell commands in each app's manifest `path`
- Batch `up`/`down`/`build --all` may run concurrently unless `depends_on` is added
- Clear errors: which app, which command, stderr snippet
- Fast: no heavy polling; port check + git status should feel instant

## Definition of done (v0)

- [x] Config file documents apps (registry + published manifests)
- [x] `status` is the daily driver
- [ ] `up` / `down` / `build` work per app and with `--all` (via menu/`run` today)
- [x] `run` works for at least one custom command per app (e.g. `test`)
- [x] `config validate` catches missing paths and port clashes
- [ ] Used for a full day of multi-app work without reaching for ad-hoc scripts

When that's true, pull ideas from [ROADMAP.md](ROADMAP.md) one at a time.
