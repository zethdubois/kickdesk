# Kickdesk

A personal command center for working on multiple apps at once: one place to see **status**, run **builds**, and **operate** each project without keeping three mental stacks in your head.

**Problem:** You juggle three tools/repos. It's easy to forget what's running, which port is which, and how to start or build each one.

**Success (v0):** `kickdesk` shows app status, then **1 / 2 / 3** for publicweb, kickagent, merch-api. Each opens a startup or shutdown workflow (based on whether the app is running). **Space+Enter** runs the next step; **Enter** runs all steps.

## Docs

| Doc | Purpose |
|-----|---------|
| [docs/NOW.md](docs/NOW.md) | What we're building first — scope, commands, config, non-goals |
| [docs/DEV_PORTS.md](docs/DEV_PORTS.md) | KAM port families (50xx / 70xx / 80xx) — canonical local ports |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Later ideas — dashboard, plugins, CI, team features |

Build and design decisions should satisfy **NOW** first. Anything in **ROADMAP** is explicitly out of scope until daily use proves the core.

## Quick orientation

```bash
kickdesk                    # hub: [1] publicweb [2] kickagent [3] merch-api
kickdesk -t 2               # tmux: 2 server panes on top, menu on bottom
kickdesk menu               # same as (none)
kickdesk status             # status table (ports, git, optional migrate pending)
kickdesk run publicweb test # run one command without the menu
kickdesk config validate    # paths, ports, workflows
```

**tmux dashboard (`-t N`):** requires `tmux` on PATH. Top row gets the first **N** apps from `app_order` (e.g. `-t 2` → publicweb + kickagent). Bottom pane runs kickdesk automatically. Blocking `up` commands go to the matching top pane (not a new GUI window). Outside tmux you get a **y/N** prompt to start session `kickdesk` and attach. Inside tmux, the current window is re-layouted. If you start `kickdesk menu` manually in the bottom pane (title `kickdesk-menu`), dashboard mode is detected automatically.

Single hotkey menu (no Enter): **1–3** select app, then **Space** = next step, **Enter** = all remaining, **1–N** = run one step, **b** = back, **c** = catalog, **r** = refresh, **q** = quit. After each step, **any key** continues. Without `-t`, dev servers (`up`) open in a new GUI terminal when possible (`KICKDESK_TERMINAL` or auto-detect).

Config lives at `~/.config/kickdesk/config.json` (see [docs/NOW.md](docs/NOW.md) for schema). Apps with a database can define an optional `migrate-status` command; kickdesk runs it when the db port is up and shows **MIGRATE** (`ok`, `pending:N`, or `unavailable`). See `examples/config.json` for publicweb and merch-api.

**First-time setup:**

```bash
mkdir -p ~/.config/kickdesk
cp examples/config.json ~/.config/kickdesk/config.json
kickdesk config validate
kickdesk
```

**Install on PATH** — add to `~/.bashrc`, then `go install`:

```bash
export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"
cd ~/projects/kickdesk && go install .
```

## Development (Go)

Requires Go 1.21+. If `go` is not on your PATH, install from [go.dev/dl](https://go.dev/dl/) or use `~/.local/go` after extracting the official tarball.

```bash
export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"   # add to ~/.bashrc

cd ~/projects/kickdesk
go run .                  # interactive menu
go run . status
go build -o bin/kickdesk .
./bin/kickdesk
go install .              # installs to ~/go/bin/kickdesk
```
