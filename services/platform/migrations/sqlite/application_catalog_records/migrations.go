package applicationcatalogrecords

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "application_catalog_records"
	MigrationID        = "0001_application_catalog_records"
	StoreSchemaVersion = "application_catalog_records_store_v1"
)

//go:embed 0001_application_catalog_records.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
