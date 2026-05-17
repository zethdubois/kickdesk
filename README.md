# Kickdesk

A personal command center for working on multiple apps at once: one place to see **status**, run **builds**, and **operate** each project without keeping three mental stacks in your head.

**Problem:** You juggle three tools/repos. It's easy to forget what's running, which port is which, and how to start or build each one.

**Success (v0):** `kickdesk status` answers that in under a second. `kickdesk up <name>` and `kickdesk run <name> <task>` work reliably every time.

## Docs

| Doc | Purpose |
|-----|---------|
| [docs/NOW.md](docs/NOW.md) | What we're building first — scope, commands, config, non-goals |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Later ideas — dashboard, plugins, CI, team features |

Build and design decisions should satisfy **NOW** first. Anything in **ROADMAP** is explicitly out of scope until daily use proves the core.

## Quick orientation

```bash
kickdesk status              # what's running, ports, git branch, URLs
kickdesk up|down|build app   # lifecycle per app or --all
kickdesk run app test        # named "operate" commands (migrate, test, …)
kickdesk config validate     # paths exist, ports free
```

Config lives at `~/.config/kickdesk/config.json` (see [docs/NOW.md](docs/NOW.md) for schema).

## Development (Go)

Requires Go 1.21+. If `go` is not on your PATH, install from [go.dev/dl](https://go.dev/dl/) or use `~/.local/go` after extracting the official tarball.

```bash
export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"   # add to ~/.bashrc

cd ~/projects/kickdesk
go run .                  # banner
go run . status           # status stub
go build -o bin/kickdesk .
./bin/kickdesk status
go install .              # installs to ~/go/bin/kickdesk
```
