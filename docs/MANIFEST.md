# Per-app manifest contract

Kickdesk loads **operational** config (ports, commands, workflows) from per-app manifests. The cockpit file (`~/.config/kickdesk/config.json`) only lists which repos to include, optional labels, and profile/tmux layout.

**App teams own manifests.** Kickdesk ships one schema sample only: [examples/manifest.sample.json](../examples/manifest.sample.json). Real manifests are **published** by each app to `~/.config/<app-id>/` (see [Cross-subscriber setup](#cross-subscriber-setup-recommended)).

**Kickdesk does not need to read subscriber repos for manifests.** Given a thin registry entry (`path` + optional `label`), the normal path is: resolve `app-id` → load `~/.config/<app-id>/manifest.json` (and optional `migrate-status`). In-repo paths (`.kickdesk.json`, etc.) are **optional fallbacks** for teams that want repo-local copies; they are not part of the recommended subscriber contract.

## Manifest locations (discovery order)

1. Explicit path in registry: `apps.<id>.manifest`
2. Environment: `KICKDESK_MANIFEST` (debug / one-off)
3. **`~/.config/<id>/manifest.json`** — **recommended** published copy (subscriber default)
4. `<repo>/.kickdesk.json` — optional fallback only (not required for subscribers)
5. `<repo>/.kickdesk/manifest.json` — optional fallback only
6. **Legacy:** full app block inline in `config.json` (v0 monolithic; migrate operators to thin registry + published manifest)

For a registered app with `id` matching the registry key, Kickdesk should prefer **(3)** once the operator has run the app’s publish step. Repo paths **(4–5)** must not be documented as the primary or required integration path.

## Schema

See [examples/manifest.sample.json](../examples/manifest.sample.json) for a copy-paste template. Required fields:

| Field                      | Required    | Notes                                                    |
| -------------------------- | ----------- | -------------------------------------------------------- |
| `id`                       | yes         | Must match registry key in `config.json`                 |
| `ports`                    | yes         | Role + port list for status table                        |
| `primary_port`             | recommended | Running/stopped detection                                |
| `commands`                 | yes         | At least `up`, `down`, `build`                           |
| `workflows.start` / `stop` | yes         | Ordered command keys                                     |
| `status.migrate`           | no          | File path; one line: `ok`, `pending:N`, or `unavailable` |

If `status.migrate` is set and the file exists, kickdesk reads it when the db port is up. Otherwise it runs `commands.migrate-status` when defined.

## Publishing (app repo responsibility)

Each app repo should provide a publish step, for example:

```json
"kickdesk:publish-manifest": "tsx scripts/publish-kickdesk-manifest.ts"
```

The script writes **`~/.config/<app-id>/manifest.json` only**. Kickdesk only **loads** that file; it does not invoke publish.

Suggested publish output:

- Generate manifest from repo truth (ports, package.json scripts, compose).
- Refresh `~/.config/<app-id>/migrate-status` via `commands.migrate-status` (or a dedicated script) after migrate, db-up, or pulling migrations.

**Do not** document “commit `.kickdesk.json` in the repo” as the standard subscriber flow. Teams may add an in-repo copy for their own tooling; Kickdesk operators should rely on the published config dir.

## Registry (`config.json`) — kickdesk operator only

```json
{
  "apps": {
    "publicweb": {
      "path": "~/projects/publicweb",
      "label": "Public Web"
    }
  }
}
```

No per-app commands here — only `path`, optional `label`, optional explicit `manifest` path (override when not using `~/.config/<id>/manifest.json`).

After an app is registered, the operator runs that app’s publish command once (and again when ports/workflows change). Kickdesk then consumes everything under `~/.config/<id>/` per discovery **(3)** above.

## Profiles (`<name>.json`) — kickdesk operator only

```json
{
  "app_order": ["publicweb", "kickagent"],
  "labels": { "publicweb": "Public Web" },
  "tmux": { "enabled": true, "top": 2, "session": "kickdesk" }
}
```

Launch: `kickdesk -c` (picker) or plain `kickdesk` (uses `last.cnfg`).

## Cross-subscriber setup (recommended)

Reference implementation: **publicweb** (first full subscriber). Other apps (kickagent, merch-api, …) should mirror the same **roles**, not necessarily the same filenames, as long as publish targets `~/.config/<app-id>/`.

### Roles

| Layer                 | Owner                                     | Purpose                                                                      |
| --------------------- | ----------------------------------------- | ---------------------------------------------------------------------------- |
| **Kickdesk**          | KD repo                                   | Schema, discovery rules, this doc, `manifest.sample.json`                    |
| **Registry**          | Operator `~/.config/kickdesk/config.json` | Which repos exist (`path`, `label`); no ports/commands                       |
| **Registration**      | App repo                                  | Machine-readable “we are a subscriber” + pointers (not operational manifest) |
| **Manifest source**   | App repo                                  | Single source for port/command **values** (code or generated)                |
| **Published runtime** | `~/.config/<app-id>/`                     | What Kickdesk loads: `manifest.json`, `migrate-status`                       |

### App repo layout (publicweb pattern)

```
<app-repo>/
  kickdesk.registration.json    # appId, configDir, KD spec paths, publish commands
  scripts/
    kickdesk-manifest.ts        # buildKickdeskManifest() — values (ports, workflows, commands)
    publish-kickdesk-manifest.ts # writes ~/.config/<appId>/manifest.json
    migrate-status.ts            # stdout contract + writes ~/.config/<appId>/migrate-status
  package.json                   # kickdesk:publish-manifest, db:migrate:status
```

Optional (tool-specific, not required by Kickdesk): agent guide in `docs/`, Cursor rule, `AGENTS.md` section pointing at `kickdesk.registration.json`.

**No in-repo operational manifest is required.** publicweb intentionally does **not** commit `.kickdesk.json`; discovery uses the published config dir only.

### `kickdesk.registration.json` (registration, not manifest)

Tool-agnostic marker for humans and agents. Example shape:

```json
{
  "appId": "publicweb",
  "configDir": "~/.config/publicweb",
  "kickdeskSpec": {
    "root": "${KICKDESK_ROOT:-../kickdesk}",
    "readonlyFiles": ["docs/MANIFEST.md", "examples/manifest.sample.json"]
  },
  "publish": {
    "manifest": "pnpm kickdesk:publish-manifest",
    "migrateStatus": "pnpm db:migrate:status"
  }
}
```

- **`kickdeskSpec.readonlyFiles`** — KD contract docs only; agents read these for schema, not for port values.
- **`publish`** — commands the app team runs; Kickdesk never invokes them.

### Publish and migrate-status

| Command                                    | Writes                              |
| ------------------------------------------ | ----------------------------------- |
| `pnpm kickdesk:publish-manifest` (example) | `~/.config/<app-id>/manifest.json`  |
| `pnpm db:migrate:status` (example)         | `~/.config/<app-id>/migrate-status` |

Manifest must set `status.migrate` to that file path (e.g. `~/.config/publicweb/migrate-status`). One line: `ok`, `pending:N`, or `unavailable`.

**When to re-run publish:** dev port, compose host port, workflow keys, or command strings change. **When to refresh migrate-status:** after db up, migrate, or pulling new migrations.

### New app subscriber checklist (execute in order)

Use this when onboarding **any** app repo. Reference: **publicweb** (complete); adapt `app-id`, ports, and commands to your stack.

| Step | Who      | Action                                                                                                                                                                                                    |
| ---- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | App team | Pick `app-id` (registry key, e.g. `merch-api`) and port family per [DEV_PORTS.md](DEV_PORTS.md).                                                                                                          |
| 2    | App team | Add `kickdesk.registration.json` (see shape below). Set `configDir` to `~/.config/<app-id>`.                                                                                                              |
| 3    | App team | Add manifest **values** source: `scripts/kickdesk-manifest.ts` (or `.js`) with `buildKickdeskManifest()` — ports from repo truth (vite/compose/Makefile), not copied from `manifest.sample.json`.         |
| 4    | App team | Add `scripts/publish-kickdesk-manifest.ts` → writes `~/.config/<app-id>/manifest.json` only. Wire `kickdesk:publish-manifest` in `package.json` (or `Makefile` target).                                   |
| 5    | App team | Add or extend `migrate-status` script: stdout one line `ok` \| `pending:N` \| `unavailable`; also write `~/.config/<app-id>/migrate-status`. Wire in `commands` and registration `publish.migrateStatus`. |
| 6    | App team | Set manifest `status.migrate` to `~/.config/<app-id>/migrate-status`.                                                                                                                                     |
| 7    | App team | Add agent/IDE hooks (recommended): `AGENTS.md` section, optional `docs/guides/kickdesk-manifest.md`, IDE rule (see [IDE and agent access](#ide-and-agent-access-subscriber-repos)).                       |
| 8    | Operator | Add app to `~/.config/kickdesk/config.json`: `{ "path": "~/projects/<repo>", "label": "..." }` only — no ports/commands.                                                                                  |
| 9    | Operator | In app repo: run publish, then migrate-status (with DB up if applicable).                                                                                                                                 |
| 10   | Operator | `kickdesk config validate --app <app-id>` — manifest path should be `~/.config/<app-id>/manifest.json`.                                                                                                   |

**Do not** commit `.kickdesk.json` in the app repo unless you have a separate reason; Kickdesk should load the published config dir.

**Copy from publicweb** when bootstrapping: `kickdesk.registration.json`, `scripts/kickdesk-manifest.ts`, `scripts/publish-kickdesk-manifest.ts`, `scripts/migrate-status.ts` — rename `publicweb` → your `app-id` and replace ports/commands/workflows.

### Operator checklist (after app repo is wired)

1. Add app to `~/.config/kickdesk/config.json` with `path` (and optional `label`) only.
2. In the app repo: `pnpm kickdesk:publish-manifest` (or equivalent).
3. `pnpm db:migrate:status` (or equivalent) with DB up when checking migrate column.
4. `kickdesk config validate --app <id>` — should resolve `~/.config/<id>/manifest.json`.

### IDE and agent access (subscriber repos)

Kickdesk does not configure IDEs. Each **subscriber app repo** documents how agents (Cursor, Copilot, etc.) should interact with Kickdesk **without** editing the Kickdesk repo or reading operator config.

#### What agents may read (Kickdesk contract only)

Paths come from `kickdesk.registration.json` → `kickdeskSpec.readonlyFiles`, resolved under `kickdeskSpec.root` (default sibling `../kickdesk`, or `$KICKDESK_ROOT`):

| File                            | Purpose                                        |
| ------------------------------- | ---------------------------------------------- |
| `docs/MANIFEST.md`              | Discovery, schema, subscriber setup (this doc) |
| `examples/manifest.sample.json` | Field shapes only — **not** your app’s ports   |

**Do not** treat the following as required agent reads: Kickdesk Go source, `testdata/`, operator `~/.config/kickdesk/config.json`, other apps’ `~/.config/<id>/` trees.

#### What agents must use for your app’s values

| Source                                         | Purpose                                                                                  |
| ---------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `scripts/kickdesk-manifest.ts` (or equivalent) | Ports, workflows, `commands` strings                                                     |
| Repo truth                                     | `vite.config.ts`, `docker-compose.yml`, `package.json`, Makefile — keep manifest in sync |
| `kickdesk.registration.json`                   | `appId`, `configDir`, publish commands                                                   |

Never copy `manifest.sample.json` verbatim (`id: "my-app"` is a template).

#### Registration vs IDE rules

| File                         | Contains                                                                         | Tool-specific?         |
| ---------------------------- | -------------------------------------------------------------------------------- | ---------------------- |
| `kickdesk.registration.json` | `appId`, `configDir`, KD spec pointers, publish commands                         | **No** — portable      |
| `AGENTS.md` (section)        | “Read registration first”; publish/migrate triggers                              | Cursor, Claude, humans |
| `.cursor/rules/kickdesk.mdc` | Same behavior as AGENTS for matched globs                                        | **Cursor only**        |
| VS Code / other              | Equivalent: workspace rule or `copilot-instructions.md` pointing at registration | Per IDE                |

**`kickdesk.registration.json` does not grant filesystem access.** It only lists paths agents _should_ read if they can. Operators may add the Kickdesk repo to a **multi-root workspace** or set `KICKDESK_ROOT` so agents can open `MANIFEST.md` when the subscriber repo is opened alone.

#### Agent triggers (re-run publish / migrate-status)

| Event                                                                 | Run                                       |
| --------------------------------------------------------------------- | ----------------------------------------- |
| Dev port, compose host port, workflow keys, or command strings change | `publish.manifest` from registration      |
| DB up, migrate, or pull with new migrations                           | `publish.migrateStatus` from registration |

#### Cursor (recommended)

Create `.cursor/rules/kickdesk.mdc` with `globs` covering: `kickdesk.registration.json`, manifest scripts, port config (`vite.config.ts`, `docker-compose.yml`, `package.json`, etc.).

Example rule body (replace `<app-id>`):

```markdown
# Kickdesk subscriber

Read `kickdesk.registration.json` first. This app publishes to `~/.config/<app-id>/` only — do not add `.kickdesk.json` in this repo.

When changing dev ports, compose, package scripts used in workflows, or Kickdesk integration:

1. Read Kickdesk spec files (readonly) if present: `$KICKDESK_ROOT` or `../kickdesk` + paths from `kickdeskSpec.readonlyFiles`.
2. Update `scripts/kickdesk-manifest.ts` (values) — not `manifest.sample.json` verbatim.
3. Run publish command from registration after port/script/workflow changes.
4. Run migrate-status command from registration after db up / migrate / migration pull.
```

#### AGENTS.md (recommended)

Short section linking registration + optional app guide:

```markdown
## Kickdesk

Kickdesk subscriber. Read `kickdesk.registration.json`; schema: sibling `../kickdesk/docs/MANIFEST.md` (or `KICKDESK_ROOT`).

- Publish: `<publish.manifest from registration>` → `~/.config/<app-id>/manifest.json`
- Migrate status: `<publish.migrateStatus>` → `~/.config/<app-id>/migrate-status`
- Values: `scripts/kickdesk-manifest.ts`
```

#### Other IDEs

Mirror the Cursor rule: point instructions at `kickdesk.registration.json`, list the same readonly KD files, same publish/migrate triggers, same prohibition on in-repo operational `.kickdesk.json`.

### KD documentation alignment

- **NOW.md / DEV_PORTS / examples** should describe thin registry + published `~/.config/<id>/`, not “put `.kickdesk.json` in each subscriber repo.”
- Discovery **(4–5)** remain in code for backward compatibility; document them as **optional**, ranked below **(3)**.
- Legacy monolithic `apps.<id>` inline blocks in `config.json` remain supported but are not the target model.

## Validation

```bash
kickdesk config validate
kickdesk config validate --app publicweb
```

Reports manifest path per app when resolved from disk. For subscribers using the recommended path, expect `~/.config/<id>/manifest.json`.

## Kickdesk development fixtures

Internal test manifests (not for app teams to copy) live under [testdata/manifests/](../testdata/manifests/). Run from repo root:

```bash
KICKDESK_CONFIG=testdata/config.json kickdesk config validate
```

Examples for operators: [examples/config.json](../examples/config.json), [examples/profiles/](../examples/profiles/), [examples/manifest.sample.json](../examples/manifest.sample.json).
