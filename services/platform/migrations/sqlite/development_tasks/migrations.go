package developmenttasks

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

//go:embed 0001_development_tasks.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{Component: "development_tasks", ID: "0001_development_tasks", StoreSchemaVersion: "development_task_store_v1", UpSQL: upSQL}}
}
