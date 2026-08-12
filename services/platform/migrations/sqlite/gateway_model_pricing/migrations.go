package gatewaymodelpricing

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "gateway_model_pricing"
	MigrationID        = "0001_gateway_model_pricing"
	StoreSchemaVersion = "gateway_model_pricing_store_v1"
)

//go:embed 0001_gateway_model_pricing.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
