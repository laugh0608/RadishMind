package developmentattempts

import (
	_ "embed"
	"radishmind.local/services/platform/internal/sqlitedev"
)

//go:embed 0001_development_attempts.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{Component: "development_attempts", ID: "0001_development_attempts", StoreSchemaVersion: "development_attempt_store_v1", UpSQL: upSQL}}
}
