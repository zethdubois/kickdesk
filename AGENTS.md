# AGENTS.md

**Start here** for any agent (Cursor, OpenCode, etc.) working in this repository.

<!-- BEGIN proj-agents -->
Portable procedure: default **[`~/fleet/harness/SOP.md`](/home/golem/fleet/harness/SOP.md)**; project stub **[.agent/SOP.md](.agent/SOP.md)** (pack **[.agent/SOP_VERSION](.agent/SOP_VERSION)** — `proj-agents version` / `status .`). Local overrides and stack notes: **this file**.

## Start here (ordered)

1. **[docs/now.md](docs/now.md)** — this week’s work; do the first unchecked item
2. **[docs/plan.md](docs/plan.md)** — roadmap + phase outcomes (only the linked section)
3. **`~/fleet/harness/SOP.md`** — default how we work (shared across projects)

## Working rules

Follow `~/fleet/harness/SOP.md`, then this file for project-specific rules. Keep `.agent/COMMITLOG` current with *why* — append each turn; prefix **`[c]`** (Cursor) or **`[oc]`** (OpenCode). **Git and `commit.sh` are for humans** — agents must ask and receive permission before any git mutation. Humans run `commit` / `commit <project>`.

**Shared / production data:** name those stores in this file’s project notes. Before any write to them: back up under `~/backups/<project>/…`, dry-run, get human approval, then apply — and append **[docs/data-ops-log.md](docs/data-ops-log.md)** in the same session (operation not done until logged). See harness SOP → Writes to shared data.
<!-- END proj-agents -->

## Project notes

_(Replace this section with stack, layout, and rules that are unique to this repo.)_
