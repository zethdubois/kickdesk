# Agent commit log

| File | Tracked? | Role |
|------|----------|------|
| `COMMITLOG` | **no** (gitignored) | WIP commit message; agents keep it current |
| `COMMITLOG.example` | yes | Template if COMMITLOG is missing |
| `SOP.md` | yes | Portable agent SOP (now/plan, commit, closing a phase) |
| `SOP_VERSION` | yes | Installed pack version (`version=N`) |
| `../commit.sh` | yes | Human-gated `git add -A` + `git commit -F COMMITLOG` |

**Agents:** after any meaningful change, **append** *why* to `COMMITLOG` (not a file dump). Prefix each entry: **`[c]`** Cursor, **`[oc]`** OpenCode (optional **`[h]`** human). Do not overwrite other WIP lines. Replace the file only when it is a lone `committed <timestamp>` stamp. Follow `SOP.md`. Do not run `commit.sh` unless the human asks — it prompts on a real TTY.

**Humans:** `commit` or `commit <project>` (after future `./install.sh`). Update this pack with `proj-agents install <dir>` from `~/projects/future`.
