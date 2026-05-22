package menu

import (
	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/status"
)

// Procedure is the resolved menu workflow for one app at a moment in time.
// Keys are the flat ordered list of command keys shown to the operator (with
// continuous 1..N numbering). When Republish is defined for a running app,
// the keys are Stop keys followed by Republish keys; RepublishFrom is the
// index in Keys where the republish section starts. When no republish
// section is shown, RepublishFrom == len(Keys).
type Procedure struct {
	Title         string
	Keys          []string
	RepublishFrom int
}

// HasRepublish reports whether the procedure includes a republish subsection.
func (p Procedure) HasRepublish() bool {
	return p.RepublishFrom < len(p.Keys)
}

// RepublishKeys returns just the republish portion of the procedure.
func (p Procedure) RepublishKeys() []string {
	if !p.HasRepublish() {
		return nil
	}
	return append([]string(nil), p.Keys[p.RepublishFrom:]...)
}

// IsRepublishIndex reports whether step i (0-based) is part of the republish
// subsection.
func (p Procedure) IsRepublishIndex(i int) bool {
	return p.HasRepublish() && i >= p.RepublishFrom
}

// EffectiveProcedure returns the procedure to render for one app.
//
// When the app is stopped: Startup procedure (workflows.start, with migrate
// optionally skipped when migrate-status says it is unnecessary).
//
// When the app is running:
//   - workflows.stop alone when workflows.republish is empty (legacy behavior).
//   - workflows.stop followed by workflows.republish (continuous numbering,
//     title "Shutdown · Republish") when workflows.republish is non-empty.
func EffectiveProcedure(cfg *config.Config, appName string, running bool, migrateStatus string) (Procedure, error) {
	keys, err := cfg.WorkflowKeys(appName, running)
	if err != nil {
		return Procedure{}, err
	}
	if !running {
		if shouldSkipMigrate(migrateStatus) {
			keys = omitKey(keys, "migrate")
		}
		return Procedure{
			Title:         "Startup",
			Keys:          keys,
			RepublishFrom: len(keys),
		}, nil
	}

	republish, err := cfg.RepublishKeys(appName)
	if err != nil {
		return Procedure{}, err
	}
	if len(republish) == 0 {
		return Procedure{
			Title:         "Shutdown",
			Keys:          keys,
			RepublishFrom: len(keys),
		}, nil
	}
	combined := make([]string, 0, len(keys)+len(republish))
	combined = append(combined, keys...)
	republishFrom := len(combined)
	combined = append(combined, republish...)
	return Procedure{
		Title:         "Shutdown · Republish",
		Keys:          combined,
		RepublishFrom: republishFrom,
	}, nil
}

// EffectiveWorkflowKeys returns workflow command keys, omitting migrate when
// status says skip. Kept for callers that only need the flat key list.
func EffectiveWorkflowKeys(cfg *config.Config, appName string, running bool, migrateStatus string) ([]string, error) {
	proc, err := EffectiveProcedure(cfg, appName, running, migrateStatus)
	if err != nil {
		return nil, err
	}
	return proc.Keys, nil
}

func shouldSkipMigrate(migrateStatus string) bool {
	return migrateStatus == "ok" || migrateStatus == "n/a"
}

func omitKey(keys []string, key string) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if k != key {
			out = append(out, k)
		}
	}
	return out
}

func migrateStatusForApp(apps []status.AppStatus, appName string) string {
	for _, a := range apps {
		if a.Name == appName {
			if a.Migrate != "" {
				return a.Migrate
			}
			return "n/a"
		}
	}
	return "n/a"
}
