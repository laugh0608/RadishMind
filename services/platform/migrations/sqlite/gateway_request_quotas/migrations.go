package gatewayrequestquotas

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "gateway_request_quotas"
	MigrationID        = "0001_gateway_request_quotas"
	StoreSchemaVersion = "gateway_request_quota_store_v1"
)

//go:embed 0001_gateway_request_quotas.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
