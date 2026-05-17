package menu

import (
	"github.com/Kick-Asset-Management/kickdesk/config"
	"github.com/Kick-Asset-Management/kickdesk/status"
)

// EffectiveWorkflowKeys returns workflow command keys, omitting migrate when status says skip.
func EffectiveWorkflowKeys(cfg *config.Config, appName string, running bool, migrateStatus string) ([]string, error) {
	keys, err := cfg.WorkflowKeys(appName, running)
	if err != nil {
		return nil, err
	}
	if shouldSkipMigrate(migrateStatus) {
		keys = omitKey(keys, "migrate")
	}
	return keys, nil
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
