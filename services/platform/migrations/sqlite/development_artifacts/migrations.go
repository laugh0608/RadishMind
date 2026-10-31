package developmentartifacts

import (
	_ "embed"
	"radishmind.local/services/platform/internal/sqlitedev"
)

//go:embed 0001_development_artifacts.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{Component: "development_artifacts", ID: "0001_development_artifacts", StoreSchemaVersion: "development_artifact_store_v1", UpSQL: upSQL}}
}
