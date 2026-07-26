package promptapplicationtemplates

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "prompt_application_templates"
	MigrationID        = "0001_prompt_application_templates"
	StoreSchemaVersion = "prompt_application_template_store_sqlite_v1"
)

//go:embed 0001_prompt_application_templates.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{Component: Component, ID: MigrationID, StoreSchemaVersion: StoreSchemaVersion, UpSQL: upSQL}}
}
