package sqlite

import (
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteadminproviderroutemigrations "radishmind.local/services/platform/migrations/sqlite/admin_provider_routes"
	sqliteagentcopilotprofilemigrations "radishmind.local/services/platform/migrations/sqlite/agent_copilot_profiles"
	sqliteapikeymigrations "radishmind.local/services/platform/migrations/sqlite/api_key_records"
	sqliteapplicationcatalogmigrations "radishmind.local/services/platform/migrations/sqlite/application_catalog_records"
	sqliteapplicationdraftmigrations "radishmind.local/services/platform/migrations/sqlite/application_configuration_drafts"
	sqliteapplicationpublishmigrations "radishmind.local/services/platform/migrations/sqlite/application_publish_candidates"
	sqlitegatewaymodelpricingmigrations "radishmind.local/services/platform/migrations/sqlite/gateway_model_pricing"
	sqlitegatewayrequestquotamigrations "radishmind.local/services/platform/migrations/sqlite/gateway_request_quotas"
	sqlitegatewayrequestmigrations "radishmind.local/services/platform/migrations/sqlite/gateway_requests"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
	sqlitepromptapplicationtemplatemigrations "radishmind.local/services/platform/migrations/sqlite/prompt_application_templates"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
	sqliteworkflowdraftmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_saved_drafts"
)

func Migrations() []sqlitedev.Migration {
	migrations := make([]sqlitedev.Migration, 0, 13)
	migrations = append(migrations, sqliteadminproviderroutemigrations.Migrations()...)
	migrations = append(migrations, sqliteapplicationcatalogmigrations.Migrations()...)
	migrations = append(migrations, sqliteagentcopilotprofilemigrations.Migrations()...)
	migrations = append(migrations, sqliteapplicationdraftmigrations.Migrations()...)
	migrations = append(migrations, sqliteapplicationpublishmigrations.Migrations()...)
	migrations = append(migrations, sqliteapikeymigrations.Migrations()...)
	migrations = append(migrations, sqlitegatewayrequestmigrations.Migrations()...)
	migrations = append(migrations, sqlitegatewayrequestquotamigrations.Migrations()...)
	migrations = append(migrations, sqlitegatewaymodelpricingmigrations.Migrations()...)
	migrations = append(migrations, sqlitelocalidentitymigrations.Migrations()...)
	migrations = append(migrations, sqlitepromptapplicationtemplatemigrations.Migrations()...)
	migrations = append(migrations, sqliteworkflowdraftmigrations.Migrations()...)
	migrations = append(migrations, sqliteworkflowrunmigrations.Migrations()...)
	return migrations
}
