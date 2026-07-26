package adminproviderroutes

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "admin_provider_routes"
	MigrationID        = "0001_admin_provider_routes"
	StoreSchemaVersion = "admin_provider_routes_store_v1"
)

//go:embed 0001_admin_provider_routes.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
