package localidentityrecords

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "local_identity_records"
	MigrationID        = "0001_local_identity_records"
	StoreSchemaVersion = "local_identity_records_store_v1"
)

//go:embed 0001_local_identity_records.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
