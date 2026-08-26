# AGENTS.md

**Start here** for any agent (Cursor, OpenCode, etc.) working in this repository.

Portable procedure (now/plan, closing a phase, commit): **[.agent/SOP.md](.agent/SOP.md)**. Installed version: **[.agent/SOP_VERSION](.agent/SOP_VERSION)** (`proj-agents version` / `proj-agents status .`).

## Start here (ordered)

1. **[docs/now.md](docs/now.md)** — this week’s work; do the first unchecked item
2. **[docs/plan.md](docs/plan.md)** — roadmap + phase outcomes (only the linked section)
3. **[.agent/SOP.md](.agent/SOP.md)** — how we work (shared across projects)

## Project notes

Go CLI (`github.com/Kick-Asset-Management/kickdesk`), Linux-first. Layout: `config/`, `menu/`, `run/`, `status/`, `tmux/`. Operator config is `~/.config/kickdesk/`; per-app manifests live in `~/.config/<app-id>/`.

**Install:** `./install.sh` builds the binary, symlinks `~/.local/bin/kickdesk`, and registers a marked block in `~/.bashrc` so `kickdesk` runs from any directory. Re-run after pulling to rebuild.

## Working rules

Follow [.agent/SOP.md](.agent/SOP.md). Keep `.agent/COMMITLOG` current with *why*. Humans run `commit` / `commit <project>`. Agents do not run `commit.sh` unless asked.
