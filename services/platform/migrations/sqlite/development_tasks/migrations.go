package developmenttasks

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

//go:embed 0001_development_tasks.up.sql
var upSQL string

//go:embed 0002_development_task_resources.up.sql
var resourcesSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{
		{Component: "development_tasks", ID: "0001_development_tasks", StoreSchemaVersion: "development_task_store_v1", UpSQL: upSQL},
		{Component: "development_tasks", ID: "0002_development_task_resources", StoreSchemaVersion: "development_task_store_v2", UpSQL: resourcesSQL},
	}
}
