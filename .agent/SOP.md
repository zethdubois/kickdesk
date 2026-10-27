# Agent SOP (project stub)

**Pack version:** 6

**Default SOP (read this):** [`~/fleet/harness/SOP.md`](/home/golem/fleet/harness/SOP.md)

Canonical copy in git: `~/projects/future/sop/harness/SOP.md`. Refresh the harness file with `proj-agents harness` after pulling `future`.

**Local overrides:** this project’s [`AGENTS.md`](../AGENTS.md) (stack, named shared stores, exceptions). Project wins only where it explicitly overrides; otherwise the harness SOP wins.

**This pack** installs `commit.sh`, COMMITLOG helpers, and this stub — not a full duplicate of the SOP. Check `.agent/SOP_VERSION` vs `proj-agents version` / `proj-agents status .`.

**Agents:** keep `.agent/COMMITLOG` current (`[c]` / `[oc]`). Do **not** run `commit.sh` or any git mutations unless the human grants permission in-session. See harness SOP → Commit SOP.
