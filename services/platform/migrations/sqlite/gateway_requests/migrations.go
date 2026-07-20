package gatewayrequests

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "gateway_requests"
	MigrationID        = "0001_gateway_requests"
	StoreSchemaVersion = "gateway_requests_store_v1"
)

//go:embed 0001_gateway_requests.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
