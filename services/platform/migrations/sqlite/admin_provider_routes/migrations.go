package adminproviderroutes

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "admin_provider_routes"
	MigrationID        = "0002_admin_provider_route_v2"
	StoreSchemaVersion = "admin_provider_routes_store_v2"
)

//go:embed 0001_admin_provider_routes.up.sql
var initialUpSQL string

//go:embed 0002_admin_provider_route_v2.up.sql
var routeV2UpSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{
		{
			Component:          Component,
			ID:                 "0001_admin_provider_routes",
			StoreSchemaVersion: "admin_provider_routes_store_v1",
			UpSQL:              initialUpSQL,
		},
		{
			Component:          Component,
			ID:                 MigrationID,
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              routeV2UpSQL,
		},
	}
}
