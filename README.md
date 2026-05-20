# Kickdesk

A personal command center for working on multiple apps at once: one place to see **status**, run **builds**, and **operate** each project without keeping three mental stacks in your head.

**Problem:** You juggle three tools/repos. It's easy to forget what's running, which port is which, and how to start or build each one.

**Success (v0):** `kickdesk` shows app status, then **1–N** for each app in your profile. Each opens a startup or shutdown workflow. **Space** / **Enter** run procedure steps.

## Docs

| Doc | Purpose |
|-----|---------|
| [docs/NOW.md](docs/NOW.md) | What we're building first — scope, commands, config, non-goals |
| [docs/MANIFEST.md](docs/MANIFEST.md) | Per-app manifest contract (share with app repos) |
| [examples/manifest.sample.json](examples/manifest.sample.json) | Generic manifest template — not app-specific |
| [docs/DEV_PORTS.md](docs/DEV_PORTS.md) | KAM port families (50xx / 70xx / 80xx) |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Later ideas |

## Quick orientation

```bash
kickdesk                    # menu using ~/.config/kickdesk/last.cnfg
kickdesk -c                 # profile picker (0–9 hotkeys); saves last.cnfg
kickdesk status             # status table
kickdesk run publicweb test   # run one command
kickdesk config validate    # paths, ports, workflows, manifest paths
```

**Three config layers:**

| Layer | File | Who maintains |
|-------|------|----------------|
| Registry | `~/.config/kickdesk/config.json` | You — repo paths only |
| Profile | `~/.config/kickdesk/<name>.json` | You — menu order, tmux layout |
| Manifest | `~/.config/<app-id>/manifest.json` | **Each app repo** — ports, commands, workflows |

See [examples/manifest.sample.json](examples/manifest.sample.json) for the manifest schema. App teams publish real manifests; kickdesk does not ship per-app manifest templates.

**tmux:** Profiles with `tmux.enabled` and `tmux.top` launch a dashboard on plain `kickdesk` (not immediately after `-c` picker in the same terminal — use `kickdesk` again after `last.cnfg` is set, or answer **y** to the tmux prompt).

**First-time setup:**

```bash
cp examples/config.json ~/.config/kickdesk/config.json
cp examples/profiles/two-up.json ~/.config/kickdesk/two-up.json
# Edit paths in config.json; ensure each app publishes ~/.config/<id>/manifest.json
kickdesk config validate
kickdesk -c
```

**Kickdesk repo development** (fixtures, not for app teams):

```bash
cd ~/projects/kickdesk
KICKDESK_CONFIG=testdata/config.json kickdesk config validate
```

Legacy monolithic config: [examples/config.legacy.json](examples/config.legacy.json).

**Install:**

```bash
export PATH="$HOME/go/bin:$PATH"
cd ~/projects/kickdesk && go install .
```
