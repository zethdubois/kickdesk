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
kickdesk menu               # same
kickdesk status             # status table (ports, git, optional migrate pending)
kickdesk run publicweb test # run one command without the menu
kickdesk config validate    # paths, ports, workflows
```

In the app workflow: **Space** = next step, **Enter** = run full procedure, **b** = back, **c** = command catalog (single key, no typing Enter after).

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
