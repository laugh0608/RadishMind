package gatewayrequests

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "gateway_requests"
	MigrationID        = "0003_gateway_provider_attempt_history"
	StoreSchemaVersion = "gateway_requests_store_v3"
)

//go:embed 0001_gateway_requests.up.sql
var initialUpSQL string

//go:embed 0002_gateway_request_cost_estimate.up.sql
var costEstimateUpSQL string

//go:embed 0003_gateway_provider_attempt_history.up.sql
var providerAttemptHistoryUpSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{
		{
			Component:          Component,
			ID:                 "0001_gateway_requests",
			StoreSchemaVersion: "gateway_requests_store_v1",
			UpSQL:              initialUpSQL,
		},
		{
			Component:          Component,
			ID:                 "0002_gateway_request_cost_estimate",
			StoreSchemaVersion: "gateway_requests_store_v2",
			UpSQL:              costEstimateUpSQL,
		},
		{
			Component:          Component,
			ID:                 MigrationID,
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              providerAttemptHistoryUpSQL,
		},
	}
}
