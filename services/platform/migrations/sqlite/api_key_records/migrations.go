package apikeyrecords

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "api_key_records"
	MigrationID        = "0001_api_key_records"
	StoreSchemaVersion = "api_key_records_store_v1"
)

//go:embed 0001_api_key_records.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
