# Agent SOP

**SOP version:** 2

Portable procedure for humans and agents in any git project. Source of truth is the `future` repo (`~/projects/future/sop/`). The operator updates a project with:

```bash
proj-agents install <dir-or-name>
```

Do not copy these files by hand. Check the installed flag: `.agent/SOP_VERSION` (or this header). Compare with `proj-agents version` / `proj-agents status <dir>`.

Fleet hosts, SSH, and thin_host switching are **not** in this pack — see `~/projects/future` (`docs/fleet.md`, `AGENTS.md`).

If the project uses JavaScript, pin the package manager it already uses (future itself is pnpm-only).

## Start here (ordered)

Same files for humans and agents.

1. **[docs/now.md](../docs/now.md)** — this week’s work; do the first unchecked item
2. **[docs/plan.md](../docs/plan.md)** — roadmap; open **only the linked section**

Tick `docs/now.md` when a session step is done. Tick `docs/plan.md` only when that **phase outcome** is done. Do not keep the same checklist in both files.

Project-specific notes (stack, layout, local rules) live in the repo’s `AGENTS.md` if it has one.

## Closing a phase

When **This week** in `docs/now.md` is all checked (optional leftovers do not count):

1. Tick the matching **outcomes** in the linked `plan.md` section. Optional leftovers stay unchecked and do **not** block Done.
2. `plan.md` phase status table → **Done**. Section heading gets a done marker. Fold finished work into **Delivered** (or equivalent).
3. Rewrite `docs/now.md`: **Focus** = next phase; **This week** = a new command-level list (≤7) taken from that phase; finished work → **Done recently**. Never leave a fully-checked This week — that is the next agent’s “what do I do?”
4. Update any status mirror that phase named.

How we know the way: this section, not convention from an earlier phase.

## Commit SOP

`git add .; git commit -am "msg"` works at the repo root: `add .` stages new files in this directory, then `-a` restages tracked edits. It is blunt. From a subdirectory it misses the rest of the tree; `-am` without `add` misses untracked files.

Use a **WIP log + gated script** instead.

| Who | Habit |
|-----|--------|
| **Agents** | After meaningful work, **append** a short *why* to `.agent/COMMITLOG` (gitignored). Never replace existing WIP text with only this turn. If the file is only `committed <timestamp>`, replace that stamp with the new why. Not a file dump. Do not run `commit.sh` unless asked — it needs a real TTY. |
| **Humans** | `commit` (walks up to `commit.sh`) or `commit <name>` → `~/projects/<name>/commit.sh`. Dry-run, then y/n commit, then y/n push. |

What the script runs (always from the git root):

```bash
git add -A
git commit -F .agent/COMMITLOG
# then overwrites COMMITLOG with: committed <timestamp>
```

If asked to commit without a TTY: `git add -A` at the repo root, `git commit -F .agent/COMMITLOG`, then reset the log to `committed <timestamp>`. Do not push unless asked.

See [`.agent/README.md`](README.md).

## Secrets and local agent state

- Do not commit secrets (`.env`, tokens). Review scrap scripts with hardcoded tokens before reuse.
- Do not Syncthing git working trees, live Cursor DBs, or OpenCode sqlite.
- Code lives in git only.

## `commit` dispatcher

Installed from future onto `PATH` (`~/.local/bin/commit`):

- `commit` — walk up from `$PWD` until `commit.sh`
- `commit future` — `~/projects/future/commit.sh`
- `commit mr_bayes` — `~/projects/mr_bayes/commit.sh` (after that repo has the pack)
