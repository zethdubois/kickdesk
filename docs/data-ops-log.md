# Data ops log

Append-only record of **writes to shared or production data** (migrations, backfills, direct updates, seeds against a shared DB). Newest entry first.

Local throwaway databases do not need an entry. This is not a code changelog and not a migration runbook — link the runbook; keep restore facts here.

**Rule (agents and humans):**

1. **Back up first** to `~/backups/<project>/<store>/<YYYY-MM-DD>_<slug>/` on the host that runs the write. Never `/tmp` (wiped on reboot) or the git repo (dumps hold user data). Lossless format (e.g. Mongo extended JSON, `pg_dump`) plus `SHA256SUMS`. Keep a copy of the script that performed the write. Do **not** Syncthing backup trees into `~/fleet`.
2. **Dry-run, then apply.** Agents never auto-run a shared-data write — explicit human approval first.
3. **Log it in the same session.** An operation is not finished until its entry exists below.

Record **paths and checksums**, never connection strings or secrets.

Project `AGENTS.md` should name which stores count as shared/production for this repo.

---

## Entry template

Copy below the rules; fill in; leave older entries intact.

### YYYY-MM-DD — `<slug>`

| Field | Value |
|-------|--------|
| Who / host | `[c]` / `[oc]` / `[h]` · hostname |
| Script | path in repo or under `~/backups/...` |
| Scope | DB / tenants · collections or tables · record counts |
| Backup | `~/backups/<project>/...` · checksum (see `SHA256SUMS`) |
| What changed | short factual summary |
| Verified how | e.g. re-run dry-run touches 0 records |
| Data lost or overwritten | quote identifiers / counts; or **none** |
| Restore | runnable snippet or link to steps |
| Follow-ups | open items, or **none** |
| Runbook | link to how-to doc (not duplicated here) |

---

## Log

_(No entries yet.)_
