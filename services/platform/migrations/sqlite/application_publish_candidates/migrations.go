package applicationpublishcandidates

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "application_publish_candidates"
	MigrationID        = "0001_application_publish_candidates"
	StoreSchemaVersion = "application_publish_candidates_store_v1"
)

//go:embed 0001_application_publish_candidates.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
