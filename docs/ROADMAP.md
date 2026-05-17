# Kickdesk — ROADMAP (dream)

Ideas for **after** [NOW.md](NOW.md) is a daily driver. Not committed, not scheduled — a parking lot so the v0 spec stays honest.

## Experience

- **Web dashboard** — all apps on one page; start/stop/build from the browser
- **Live updates** — websockets or SSE for status changes
- **Dedicated suite port** — e.g. dashboard on `:9000`
- **`kickdesk focus`** — if not shipped in v0, polish into a persistent "working on X" mode
- **`kickdesk logs <app>`** — unified log tail per app

## Smarter status

- HTTP **health check** endpoints per app
- **Resource usage** — CPU/memory per process
- **Process ownership** — "started by kickdesk" vs orphan; PID files
- **Last health result** and timestamp in status output

## Configuration & environments

- **Config in project root** as an alternative to `~/.config/…`
- **CLI config wizard** — `kickdesk config add` interactively
- **Templates** by stack (node, docker, python, …)
- **Environment overrides** — dev / staging command sets
- **`depends_on`** — ordered `up --all` when apps rely on each other

## Extensibility

- **Plugin system** for custom app types
- **Hooks** — pre/post up, down, build
- **Custom status indicators** per app

## Operations & platform

- **`serve` split from `up`** — infra vs dev server where stacks need it
- **Automated dependency install** before up
- **CI/CD hooks** — same config drives local and pipeline steps
- **Alerts** when a process dies or port goes silent
- **Team features** — shared config, status board (only if this ever stops being a personal tool)

## Technical (if we outgrow a script)

- **Single binary**, minimal runtime deps
- **Cross-platform** — Linux, macOS, Windows
- **Sub-second cold start** everywhere

---

When promoting something from here into NOW, write *why* (pain observed) and *acceptance criteria* — same bar as v0.
