# Agent commit log

| File | Tracked? | Role |
|------|----------|------|
| `COMMITLOG` | **no** (gitignored) | WIP commit message; agents keep it current |
| `COMMITLOG.example` | yes | Template if COMMITLOG is missing |
| `SOP.md` | yes | Stub → default SOP at `~/fleet/harness/SOP.md` |
| `SOP_VERSION` | yes | Installed pack version (`version=N`) |
| `../commit.sh` | yes | Human-gated `git add -A` + `git commit -F COMMITLOG` |

**Default SOP:** `~/fleet/harness/SOP.md` (fleet-wide). **Overrides:** project `AGENTS.md`.

**Agents:** after any meaningful change, **append** *why* to `COMMITLOG` (not a file dump). Prefix each entry: **`[c]`** Cursor, **`[oc]`** OpenCode (optional **`[h]`** human). Do not overwrite other WIP lines. Replace the file only when it is a lone `committed <timestamp>` stamp. Do **not** run `commit.sh` or git mutations unless the human grants permission in this conversation.

**Humans:** `commit` or `commit <project>` (after future `./install.sh`). Refresh harness SOP: `proj-agents harness`. Update project stubs when `sop/VERSION` bumps: `proj-agents install <dir>`.
