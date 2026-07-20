package applicationconfigurationdrafts

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "application_configuration_drafts"
	MigrationID        = "0001_application_configuration_drafts"
	StoreSchemaVersion = "application_configuration_drafts_store_v1"
)

//go:embed 0001_application_configuration_drafts.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
