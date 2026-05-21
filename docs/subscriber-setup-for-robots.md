# Kickdesk subscriber setup (for robots)

**Audience:** Coding agents and developers wiring a **new app repository** as a Kickdesk subscriber.

**Goal:** Publish operational config to `~/.config/<app-id>/` so Kickdesk can show status, run workflows, and read migration state. Kickdesk does **not** load manifests from inside the subscriber repo (no committed `.kickdesk.json` as the standard path).

**Normative contract (longer):** [MANIFEST.md](MANIFEST.md) — overlap with this doc is intentional; this file is the **executable** one-sheet.

---

## Minimum human input (required — ask if missing)

Some manifest fields are **organizational decisions**, not something to infer by scanning the subscriber repo. Humans often hand agents **only this guide** and omit them — **stop and ask** before implementing `kickdesk-manifest` sources or running publish.

| Input | Example | Who decides |
|-------|---------|-------------|
| **`id`** | `kickagent` | **Human** (must match kickdesk registry key). You may propose from repo/package name; human confirms. |
| **`port_family`** | `70` | **Human** (decade on this machine — see [DEV_PORTS.md](DEV_PORTS.md)). You may propose from stack type; human confirms no collision with other apps. |

Minimal handoff the human should provide (or confirm after your proposal):

```json
{
  "id": "kickagent",
  "port_family": 70
}
```

**Do not** copy `id` or `port_family` from `manifest.sample.json` (`my-app` / `50` are placeholders).

**If `id` or `port_family` is missing or unclear:** ask explicitly, e.g. “Which kickdesk app id should I register, and which port family (50xx / 70xx / 80xx or new decade) per DEV_PORTS?” Do not publish until answered.

### What you derive from the repo (after human confirms `id` + `port_family`)

| Field | Source |
|-------|--------|
| `path` | Checkout path at publish time |
| `primary_port`, `ports[]`, `url` | Repo truth (vite, compose, `.env.example`, Makefile) **within** the approved family |
| `commands`, `workflows` | Real scripts/targets in the repo (e.g. `package.json`, `Makefile`) |
| `status.migrate` | If the app has DB migrations — path under `~/.config/<id>/migrate-status` |

Use [DEV_PORTS.md](DEV_PORTS.md) to sanity-check ports against the chosen decade (+0 HTTP, +43 Postgres host, etc.). The rest of the manifest is your job from the repo; the two fields above are the human’s.

---

## Read this first (ordered)

Resolve the Kickdesk repo from the subscriber’s `kickdesk.registration.json`:

- `kickdeskSpec.rootEnv` if set (e.g. `KICKDESK_ROOT`), else `kickdeskSpec.root` (e.g. `../kickdesk`).

Then read these paths **under that root** (readonly — requirements only, not your app’s port values):

| Order | Path | Purpose |
|-------|------|---------|
| 1 | `docs/subscriber-setup-for-robots.md` | **This file** — setup checklist |
| 2 | `examples/manifest.sample.json` | JSON field shapes (template `my-app` — do not copy ports) |
| 3 | `docs/MANIFEST.md` | Discovery rules, schema, operator registry/profiles |
| 4 | `docs/DEV_PORTS.md` | Org port-family reference — **human chooses family** |

**Optional reference implementation** (copy file layout and scripts, not ports): sibling checkout `../publicweb` if present on disk.

---

## Agent access rules

### May read (Kickdesk contract only)

Only the four KD paths listed above, plus the subscriber repo’s own files.

### Must use for app values (subscriber repo)

| Source | Purpose |
|--------|---------|
| `kickdesk.registration.json` | `appId`, `configDir`, publish command names |
| `scripts/kickdesk-manifest.ts` (or `.js`) | Ports, `workflows`, `commands`, manifest `path` |
| Repo truth | `vite.config.ts`, `docker-compose.yml`, `package.json`, `Makefile`, etc. |

**Never** copy `examples/manifest.sample.json` verbatim into your manifest source (`id: "my-app"` and its ports are placeholders).

### Do not read (unless human explicitly asks)

- Kickdesk Go source, `testdata/`, internal fixtures
- Operator `~/.config/kickdesk/config.json` (other apps’ registry/layout)
- Other apps’ `~/.config/<id>/` trees

### Filesystem access

`kickdesk.registration.json` **does not grant** access to the Kickdesk repo. If the agent only has the subscriber repo open:

- Ask the human to add a **multi-root workspace** (subscriber + kickdesk), or
- Set **`KICKDESK_ROOT`** to the kickdesk checkout path.

---

## Published hooks: `~/.config/<app-id>/`

Kickdesk loads the subscriber’s **published** files (not the git repo’s operational manifest).

| File | Required | Content |
|------|----------|---------|
| `~/.config/<app-id>/manifest.json` | **Yes** | Full operational manifest including `path`, `ports`, `commands`, `workflows` |
| `~/.config/<app-id>/migrate-status` | If app has DB migrations | One line: `ok`, `pending:N`, or `unavailable` |

Set `configDir` in registration to `~/.config/<app-id>`. The manifest’s `status.migrate` field should point at `~/.config/<app-id>/migrate-status` when used.

**Discovery (summary):** Operator registry lists app **id** + optional **label** only → Kickdesk loads `~/.config/<id>/manifest.json`. Details: [MANIFEST.md](MANIFEST.md).

---

## Port family reference (after human chooses `port_family`)

Read [DEV_PORTS.md](DEV_PORTS.md) once the human has confirmed the decade. Glance rule:

- **50xx** — web / Vite-style app (`+0` HTTP, `+43` Postgres host)
- **70xx** — agent / tooling HTTP (e.g. publish **7099**)
- **80xx** — API stack (`+0` API, `+1` gateway, `+43` Postgres)

Then map **repo truth** into `primary_port`, `ports[]`, and `url` — not from `manifest.sample.json`. New product decades (e.g. **52xx**) require explicit human approval and org documentation in DEV_PORTS.

---

## App repo files to create

Mirror the reference pattern (e.g. **publicweb**). Replace `<app-id>` everywhere.

### 1. `kickdesk.registration.json` (strict JSON)

```json
{
  "appId": "<app-id>",
  "configDir": "~/.config/<app-id>",
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

Adjust `publish.*` if the stack uses `make`, `npm`, etc.

### 2. `scripts/kickdesk-manifest.ts` (values source)

Export `buildKickdeskManifest(repoPath: string)` returning an object matching [manifest.sample.json](../examples/manifest.sample.json):

- `id` — same as `appId` / registry key
- `path` — checkout directory (`~/projects/<app-id>` or absolute); publish script passes cwd
- `port_family`, `primary_port`, `ports`, `url` — from repo truth + human-approved family
- `workflows`, `commands` — see next section
- `status.migrate` — `~/.config/<app-id>/migrate-status` if DB app

### 3. `scripts/publish-kickdesk-manifest.ts`

- `mkdir -p` `~/.config/<app-id>/`
- Write `manifest.json` (pretty JSON + trailing newline)
- Set `path` from publish-time cwd (prefer `~/...` when under home)
- Optional: warn if vite/compose ports drift from manifest constants

Wire in `package.json`:

```json
"kickdesk:publish-manifest": "tsx scripts/publish-kickdesk-manifest.ts"
```

### 4. `scripts/migrate-status.ts` (if the app has a database)

**Stdout contract** (one line, exit 0 for `ok` and `pending:N`):

| Output | Meaning |
|--------|---------|
| `ok` | DB reachable, schema matches repo migrations |
| `pending:N` | N migrations not applied (N > 0) |
| `unavailable` | Cannot determine (DB down, env missing, etc.) |

Also **write the same line** to `~/.config/<app-id>/migrate-status` on every run.

Wire:

```json
"db:migrate:status": "tsx scripts/migrate-status.ts"
```

Register in manifest `commands` as `migrate-status` and in `workflows.start` after `db-up` if you use migrations.

**No DB:** omit `status`, `migrate-status` command, and migrate-status script; simplify `workflows.start` (e.g. `["up"]` only).

---

## Filling `workflows` and `commands`

`commands` values are shell strings Kickdesk runs with cwd = manifest `path`. Keys are arbitrary but must include at least **`up`**, **`down`**, **`build`**.

### `workflows.start` (ordered keys)

List command **keys** in startup order — infrastructure before app:

| Typical key | Role |
|-------------|------|
| `db-up` | Start local DB (compose, docker, make) |
| `migrate` | Apply migrations |
| `up` | Start dev server / main process |

Example: `["db-up", "migrate", "up"]`

Apps without DB: `["up"]` or `["build", "up"]` as appropriate.

### `workflows.stop` (ordered keys)

Teardown order — often app first, then infra:

| Typical key | Role |
|-------------|------|
| `stop-server` | Kill dev server port (`fuser -k PORT/tcp` or app-specific) |
| `db-down` | Stop DB containers |

Example: `["stop-server", "db-down"]`

### `commands` map

Every key referenced in `workflows` must exist in `commands`. Common entries:

| Key | Example value | Notes |
|-----|----------------|-------|
| `up` | `pnpm dev` | Required |
| `down` | `pnpm db:down` or `true` | Required |
| `build` | `pnpm build` | Required |
| `test` | `pnpm test` | Optional |
| `db-up` | `pnpm db:up` | If using local DB |
| `db-down` | `pnpm db:down` | If using local DB |
| `migrate` | `pnpm db:migrate` | If using migrations |
| `migrate-status` | `pnpm db:migrate:status` | If using migrations |
| `stop-server` | `fuser -k 5000/tcp 2>/dev/null \|\| true` | Match your HTTP port |

Use the project’s real package manager and script names.

### `status.migrate`

```json
"status": {
  "migrate": "~/.config/<app-id>/migrate-status"
}
```

Kickdesk reads this file when the **db** port is up; otherwise it may run `commands.migrate-status`.

---

## Operator steps (human)

1. Add app id to `~/.config/kickdesk/config.json` — **label only**, no `path`/ports/commands:

   ```json
   {
     "apps": {
       "<app-id>": { "label": "My App" }
     }
   }
   ```

2. In the app repo: run `publish.manifest` from registration (e.g. `pnpm kickdesk:publish-manifest`).

3. If DB app: run `publish.migrateStatus` with DB up (e.g. `pnpm db:migrate:status`).

4. Validate:

   ```bash
   kickdesk config validate --app <app-id>
   ```

   Expect manifest path: `~/.config/<app-id>/manifest.json`.

---

## Subscriber checklist (execute in order)

| Step | Who | Action |
|------|-----|--------|
| 1 | Human | Provide or confirm **`id`** and **`port_family`** ([DEV_PORTS.md](DEV_PORTS.md)); agent asks if missing |
| 2 | Agent | Add `kickdesk.registration.json` with four `readonlyFiles` |
| 3 | Agent | Add `scripts/kickdesk-manifest.*` with real ports/commands + `path` |
| 4 | Agent | Add publish script → `~/.config/<app-id>/manifest.json` |
| 5 | Agent | Add migrate-status script (if DB) → stdout + `migrate-status` file |
| 6 | Agent | Wire `package.json` (or Makefile) publish/migrate scripts |
| 7 | Agent | Optional: `AGENTS.md` section, `.cursor/rules/kickdesk.mdc`, README habit table |
| 8 | Human | Register app id in `~/.config/kickdesk/config.json` |
| 9 | Human | Run publish + migrate-status |
| 10 | Human | `kickdesk config validate --app <app-id>` |

---

## Day-to-day (after setup)

Subscriber apps often duplicate this in README; minimal reference:

| You did… | Run |
|----------|-----|
| Changed checkout `path`, dev ports, compose, workflow keys, or manifest command strings | `publish.manifest` from registration |
| DB up, migrate, or pulled new migrations | `publish.migrateStatus` from registration |

Running **`migrate`** does not require re-publish unless manifest **command strings** or workflow keys changed.

---

## Validation

```bash
kickdesk config validate --app <app-id>
```

---

## IDE hook (optional, subscriber repo)

**Cursor** — `.cursor/rules/kickdesk.mdc` with globs on `kickdesk.registration.json`, manifest scripts, port config files. Body: read registration first; read KD `readonlyFiles` for schema; update manifest source for values; run publish/migrate per table above; do not add `.kickdesk.json` in repo.

**AGENTS.md** — Short Kickdesk section: link `kickdesk.registration.json`, this robots doc (via registration), and app README for day-to-day habits.

---

## Reference implementation

**publicweb** (first full subscriber): `kickdesk.registration.json`, `scripts/kickdesk-manifest.ts`, `scripts/publish-kickdesk-manifest.ts`, `scripts/migrate-status.ts` — copy and adapt `app-id`, ports, and commands only.
