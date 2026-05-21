# Per-app manifest contract

Kickdesk loads **operational** config (repo path, ports, commands, workflows) from per-app manifests published under `~/.config/<app-id>/`. The operator registry (`~/.config/kickdesk/config.json`) lists **which app ids** exist and optional display labels only.

**App teams own manifests.** Kickdesk ships one schema sample: [examples/manifest.sample.json](../examples/manifest.sample.json). Real manifests are **published** by each app (see [Cross-subscriber setup](#cross-subscriber-setup-recommended)).

**Kickdesk does not read subscriber repos for manifests.** Registry entry → load `~/.config/<app-id>/manifest.json` (and optional `migrate-status`). Repo `.kickdesk.json` is **not** used; if the published manifest is missing, Kickdesk fails with an actionable error.

## Manifest locations (discovery order)

1. Explicit path in registry: `apps.<id>.manifest` (Kickdesk development / overrides only)
2. Environment: `KICKDESK_MANIFEST` (debug / one-off)
3. **`~/.config/<id>/manifest.json`** — subscriber default (required for thin registry)
4. **Legacy:** full app block inline in `config.json` (v0 monolithic)

Repo paths `.kickdesk.json` and `.kickdesk/manifest.json` were removed in v0.6; publish to `~/.config/<id>/` instead.

## Schema

See [examples/manifest.sample.json](../examples/manifest.sample.json). Required fields:

| Field                      | Required    | Notes                                                    |
| -------------------------- | ----------- | -------------------------------------------------------- |
| `id`                       | yes         | Must match registry key in `config.json`                 |
| `path`                     | yes         | Checkout path for running commands (`~` or absolute)     |
| `ports`                    | yes         | Role + port list for status table                        |
| `primary_port`             | recommended | Running/stopped detection                                |
| `commands`                 | yes         | At least `up`, `down`, `build`                           |
| `workflows.start` / `stop` | yes         | Ordered command keys                                     |
| `status.migrate`           | no          | File path; one line: `ok`, `pending:N`, or `unavailable` |

Publish must write `path` (typically the repo root used when publishing). If `status.migrate` is set and the file exists, kickdesk reads it when the db port is up. Otherwise it runs `commands.migrate-status` when defined.

## Publishing (app repo responsibility)

Each app repo should provide a publish step, for example:

```json
"kickdesk:publish-manifest": "tsx scripts/publish-kickdesk-manifest.ts"
```

The script writes **`~/.config/<app-id>/manifest.json`** including `path`. Kickdesk only **loads** that file; it does not invoke publish.

Suggested publish output:

- Generate manifest from repo truth (ports, package.json scripts, compose) and set `path` to the repo checkout.
- Refresh `~/.config/<app-id>/migrate-status` via `commands.migrate-status` (or a dedicated script) after migrate, db-up, or pulling migrations.

**Do not** commit `.kickdesk.json` in the app repo as the standard subscriber flow.

**Breaking change (v0.6):** subscriber publish scripts must emit `path` in the published manifest. Registry `path` is no longer used for thin registry entries.

## Registry (`config.json`) — kickdesk operator only

```json
{
  "apps": {
    "my-app": {
      "label": "My App"
    }
  }
}
```

Only `label` per app (empty `{}` is fine). Optional `manifest` override for development. **No** `path`, ports, or commands in the registry.

After registering an app id, run that app’s publish command once (and again when ports, path, or workflows change). Kickdesk loads `~/.config/<id>/manifest.json`.

## Profiles (`<name>.json`) — kickdesk operator only

```json
{
  "app_order": ["my-app", "other-app"],
  "labels": { "my-app": "My App" },
  "tmux": { "enabled": true, "top": 2, "session": "kickdesk" }
}
```

Launch: `kickdesk -c` (picker) or plain `kickdesk` (uses `last.cnfg`).

## Cross-subscriber setup (recommended)

**Agents:** start with [subscriber-setup-for-robots.md](subscriber-setup-for-robots.md) (one-sheet checklist). This section remains the normative contract alongside that doc; overlap is intentional.

A **reference subscriber** exists in the wild (first full implementation); mirror the same **roles** in other app repos. Adapt `app-id`, ports, and commands to your stack.

### Roles

| Layer                 | Owner                                     | Purpose                                                                      |
| --------------------- | ----------------------------------------- | ---------------------------------------------------------------------------- |
| **Kickdesk**          | KD repo                                   | Schema, discovery rules, this doc, `manifest.sample.json`                    |
| **Registry**          | Operator `~/.config/kickdesk/config.json` | Which app ids exist; optional `label` only                                   |
| **Registration**      | App repo                                  | Machine-readable subscriber marker + publish pointers (not operational data) |
| **Manifest source**   | App repo                                  | Single source for port/command **values** (code or generated)                |
| **Published runtime** | `~/.config/<app-id>/`                     | What Kickdesk loads: `manifest.json` (with `path`), `migrate-status`         |

### App repo layout (reference pattern)

```
<app-repo>/
  kickdesk.registration.json    # appId, configDir, KD spec paths, publish commands
  scripts/
    kickdesk-manifest.ts        # buildKickdeskManifest() — values + path
    publish-kickdesk-manifest.ts # writes ~/.config/<appId>/manifest.json
    migrate-status.ts            # stdout contract + writes ~/.config/<appId>/migrate-status
  package.json                   # kickdesk:publish-manifest, db:migrate:status
```

Optional: agent guide in `docs/`, Cursor rule, `AGENTS.md` section pointing at `kickdesk.registration.json`.

### `kickdesk.registration.json` (registration, not manifest)

Tool-agnostic marker for humans and agents. Example shape (strict JSON):

```json
{
  "appId": "my-app",
  "configDir": "~/.config/my-app",
  "kickdeskSpec": {
    "root": "../kickdesk",
    "rootEnv": "KICKDESK_ROOT",
    "readonlyFiles": [
      "docs/subscriber-setup-for-robots.md",
      "examples/manifest.sample.json",
      "docs/MANIFEST.md",
      "docs/DEV_PORTS.md"
    ]
  },
  "publish": {
    "manifest": "pnpm kickdesk:publish-manifest",
    "migrateStatus": "pnpm db:migrate:status"
  }
}
```

- **`kickdeskSpec.root`** — default relative path to the Kickdesk repo when opened as a sibling checkout.
- **`kickdeskSpec.rootEnv`** — if this environment variable is set, use its value instead of `root`.
- **`kickdeskSpec.readonlyFiles`** — KD contract docs only; agents read these for schema, not for port values.
- **`publish`** — commands the app team runs; Kickdesk never invokes them.

### Publish and migrate-status

| Command (example)              | Writes                              |
| ------------------------------ | ----------------------------------- |
| `pnpm kickdesk:publish-manifest` | `~/.config/<app-id>/manifest.json`  |
| `pnpm db:migrate:status`         | `~/.config/<app-id>/migrate-status` |

Manifest must set `status.migrate` to that file path. One line: `ok`, `pending:N`, or `unavailable`.

**When to re-run publish:** checkout path, dev port, compose host port, workflow keys, or command strings change. **When to refresh migrate-status:** after db up, migrate, or pulling new migrations.

### Subscriber + operator checklist

| Step | Who      | Action                                                                                                                                                                                                    |
| ---- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | App team | Pick `app-id` (registry key) and port family for your stack.                                                                                                                                              |
| 2    | App team | Add `kickdesk.registration.json` (see shape above). Set `configDir` to `~/.config/<app-id>`.                                                                                                              |
| 3    | App team | Add manifest **values** source (`scripts/kickdesk-manifest.ts` or equivalent) — ports from repo truth, not copied from `manifest.sample.json`. Include `path` in published output.                         |
| 4    | App team | Add publish script → writes `~/.config/<app-id>/manifest.json` only. Wire `kickdesk:publish-manifest` in `package.json` (or Makefile).                                                                   |
| 5    | App team | Add or extend migrate-status: stdout one line `ok` \| `pending:N` \| `unavailable`; write `~/.config/<app-id>/migrate-status`. Wire in `commands` and registration `publish.migrateStatus`.               |
| 6    | App team | Set manifest `status.migrate` to `~/.config/<app-id>/migrate-status`.                                                                                                                                     |
| 7    | App team | Add agent/IDE hooks (recommended): `AGENTS.md` section, optional app guide, IDE rule (see [IDE and agent access](#ide-and-agent-access-subscriber-repos)).                                                |
| 8    | Operator | Add app id to `~/.config/kickdesk/config.json`: `{ "label": "..." }` or `{}` — no path, ports, or commands.                                                                                                |
| 9    | Operator | In app repo: run publish, then migrate-status (with DB up if applicable).                                                                                                                                 |
| 10   | Operator | `kickdesk config validate --app <app-id>` — manifest path should be `~/.config/<app-id>/manifest.json`.                                                                                                 |

### IDE and agent access (subscriber repos)

Kickdesk does not configure IDEs. Each **subscriber app repo** documents how agents should interact with Kickdesk without editing the Kickdesk repo.

#### What agents may read (Kickdesk contract only)

Resolve `kickdeskSpec.root` using `kickdeskSpec.rootEnv` when set, else `root`, then read paths in `kickdeskSpec.readonlyFiles`:

| File                                   | Purpose                                        |
| -------------------------------------- | ---------------------------------------------- |
| `docs/subscriber-setup-for-robots.md`  | **Start here** — executable setup checklist    |
| `examples/manifest.sample.json`        | Field shapes only — **not** your app’s ports   |
| `docs/MANIFEST.md`                     | Discovery, schema, operator profiles (this doc) |
| `docs/DEV_PORTS.md`                    | Org port-family reference — human picks family |

**Do not** treat as required agent reads: Kickdesk Go source, `testdata/`, operator `~/.config/kickdesk/config.json`, other apps’ `~/.config/<id>/` trees.

#### What agents must use for your app’s values

| Source                                         | Purpose                                                                                  |
| ---------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `scripts/kickdesk-manifest.ts` (or equivalent) | Ports, workflows, `commands`, `path`                                                     |
| Repo truth                                     | `vite.config.ts`, `docker-compose.yml`, `package.json`, Makefile — keep manifest in sync |
| `kickdesk.registration.json`                   | `appId`, `configDir`, publish commands                                                   |

Never copy `manifest.sample.json` verbatim (`id: "my-app"` is a template).

#### Registration vs IDE rules

| File                         | Contains                                                 | Tool-specific?    |
| ---------------------------- | -------------------------------------------------------- | ----------------- |
| `kickdesk.registration.json` | `appId`, `configDir`, KD spec pointers, publish commands | **No** — portable |
| `AGENTS.md` (section)        | Read registration first; publish/migrate triggers        | Humans + agents   |
| `.cursor/rules/kickdesk.mdc` | Same behavior for matched globs                          | **Cursor only**   |

**`kickdesk.registration.json` does not grant filesystem access.** Operators may use a multi-root workspace or set `KICKDESK_ROOT` so agents can open `MANIFEST.md` when only the subscriber repo is open.

#### Agent triggers (re-run publish / migrate-status)

| Event                                                                 | Run                                  |
| --------------------------------------------------------------------- | ------------------------------------ |
| Dev port, compose host port, workflow keys, command strings, or `path` | `publish.manifest` from registration |
| DB up, migrate, or pull with new migrations                           | `publish.migrateStatus`              |

#### Cursor (recommended)

Example rule body (replace `<app-id>`):

```markdown
# Kickdesk subscriber

Read `kickdesk.registration.json` first. Publish to `~/.config/<app-id>/` only — do not add `.kickdesk.json` in this repo.

When changing dev ports, compose, package scripts, checkout path, or Kickdesk integration:

1. Read Kickdesk spec files from registration (`root` / `rootEnv` + `readonlyFiles`).
2. Update manifest values source — not `manifest.sample.json` verbatim.
3. Run publish from registration after port/script/workflow/path changes.
4. Run migrate-status from registration after db up / migrate / migration pull.
```

## Validation

```bash
kickdesk config validate
kickdesk config validate --app <app-id>
```

Reports manifest path per app. Subscribers should show `~/.config/<id>/manifest.json`.

## Kickdesk development fixtures

Generic fixtures (not for app teams to copy) live under [testdata/manifests/](../testdata/manifests/) with paths under [testdata/fixtures/](../testdata/fixtures/). From repo root:

```bash
KICKDESK_CONFIG=testdata/config.json kickdesk config validate
```

Operator examples: [examples/config.json](../examples/config.json), [examples/profiles/](../examples/profiles/), [examples/manifest.sample.json](../examples/manifest.sample.json). Profile `app_order` entries must match registry app ids (examples use `fixture-*` ids aligned with testdata).
