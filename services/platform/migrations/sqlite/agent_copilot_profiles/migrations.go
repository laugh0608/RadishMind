package agentcopilotprofiles

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "agent_copilot_profiles"
	MigrationID        = "0001_agent_copilot_profiles"
	StoreSchemaVersion = "agent_copilot_profile_store_sqlite_v1"
)

//go:embed 0001_agent_copilot_profiles.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{Component: Component, ID: MigrationID, StoreSchemaVersion: StoreSchemaVersion, UpSQL: upSQL}}
}
