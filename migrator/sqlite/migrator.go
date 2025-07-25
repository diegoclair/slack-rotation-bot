package sqlite

import (
	"database/sql"
	"embed"

	"github.com/GuiaBolso/darwin"
	"github.com/diegoclair/sqlmigrator"
)

//go:embed sql/*.sql
var SqlFiles embed.FS

func Migrate(db *sql.DB) error {
	migrator := sqlmigrator.New(db, darwin.SqliteDialect{})

	return migrator.Migrate(SqlFiles, "sql")
}
