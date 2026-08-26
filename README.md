# Kickdesk

A personal command center for working on multiple apps at once: one place to see **status**, run **builds**, and **operate** each project without keeping several mental stacks in your head.

**Problem:** You juggle multiple repos. It's easy to forget what's running, which port is which, and how to start or build each one.

**Success (v0):** `kickdesk` shows app status, then **1–N** for each app in your profile. Each opens a startup or shutdown workflow. **Space** / **Enter** run procedure steps.

## Docs

| Doc | Purpose |
|-----|---------|
| [docs/NOW.md](docs/NOW.md) | What we're building first — scope, commands, config, non-goals |
| [docs/MANIFEST.md](docs/MANIFEST.md) | Per-app manifest contract (share with app repos) |
| [examples/manifest.sample.json](examples/manifest.sample.json) | Generic manifest template |
| [docs/DEV_PORTS.md](docs/DEV_PORTS.md) | Optional org port reference (not subscriber-required) |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Later ideas |

## Quick orientation

```bash
kickdesk                    # menu using ~/.config/kickdesk/last.cnfg
kickdesk -c                 # profile picker (0–9 hotkeys); saves last.cnfg
kickdesk status             # status table
kickdesk run <app-id> test  # run one command in manifest path
kickdesk config validate    # paths, ports, workflows, manifest paths
```

**Three config layers:**

| Layer | File | Who maintains |
|-------|------|----------------|
| Registry | `~/.config/kickdesk/config.json` | You — app ids + optional labels only |
| Profile | `~/.config/kickdesk/<name>.json` | You — menu order, tmux layout |
| Manifest | `~/.config/<app-id>/manifest.json` | **Each app repo** — `path`, ports, commands, workflows |

See [examples/manifest.sample.json](examples/manifest.sample.json). App teams publish real manifests; kickdesk does not ship per-app templates.

**tmux:** Profiles with `tmux.enabled` and `tmux.top` launch a dashboard on plain `kickdesk` (after `last.cnfg` is set, or answer **y** to the tmux prompt).

**First-time setup:**

```bash
cp examples/config.json ~/.config/kickdesk/config.json
cp examples/profiles/two-up.json ~/.config/kickdesk/two-up.json
# Edit registry app ids to match your apps; each app publishes ~/.config/<id>/manifest.json (includes path)
kickdesk config validate
kickdesk -c
```

Example profiles use `fixture-*` app ids aligned with Kickdesk test fixtures; replace with your registry keys.

**Kickdesk repo development** (generic fixtures, not for app teams):

```bash
cd ~/projects/kickdesk
KICKDESK_CONFIG=testdata/config.json kickdesk config validate
```

Legacy monolithic config: [examples/config.legacy.json](examples/config.legacy.json).

**Install:**

```bash
cd ~/projects/kickdesk && ./install.sh
```

Builds the binary, symlinks `~/.local/bin/kickdesk`, and writes a marked block to `~/.bashrc` (`KICKDESK_ROOT` + PATH) so `kickdesk` runs from any directory. Re-run after pulling to rebuild. `./install.sh --uninstall` reverses it.

If `go` 1.25+ is not on `PATH`, the installer downloads a user-local toolchain to `~/.local/share/kickdesk/go` (no sudo).

This shell (bashrc applies to new shells):

```bash
hash -r
export PATH="$HOME/.local/bin:$PATH"
kickdesk config validate
```
