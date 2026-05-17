# Kam-Suite

A personal command center for working on multiple apps at once: one place to see **status**, run **builds**, and **operate** each project without keeping three mental stacks in your head.

**Problem:** You juggle three tools/repos. It's easy to forget what's running, which port is which, and how to start or build each one.

**Success (v0):** `kam-suite status` answers that in under a second. `kam-suite up <name>` and `kam-suite run <name> <task>` work reliably every time.

## Docs

| Doc | Purpose |
|-----|---------|
| [docs/NOW.md](docs/NOW.md) | What we're building first — scope, commands, config, non-goals |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Later ideas — dashboard, plugins, CI, team features |

Build and design decisions should satisfy **NOW** first. Anything in **ROADMAP** is explicitly out of scope until daily use proves the core.

## Quick orientation

```bash
kam-suite status              # what's running, ports, git branch, URLs
kam-suite up|down|build app   # lifecycle per app or --all
kam-suite run app test        # named "operate" commands (migrate, test, …)
kam-suite config validate     # paths exist, ports free
```

Config lives at `~/.config/kam-suite/config.json` (see [docs/NOW.md](docs/NOW.md) for schema).

## Development (Go)

Requires Go 1.21+. If `go` is not on your PATH, install from [go.dev/dl](https://go.dev/dl/) or use `~/.local/go` after extracting the official tarball.

```bash
export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"   # add to ~/.bashrc

cd ~/projects/kam-suite
go run .                  # banner
go run . status           # status stub
go build -o bin/kam-suite .
./bin/kam-suite status
go install .              # installs to ~/go/bin/kam-suite
```
