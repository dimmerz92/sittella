package database_test

import (
	"database/sql"
	"testing"

	"github.com/dimmerz92/sittella/database"
	"github.com/dimmerz92/sittella/database/sqlitedb"
	sqlitetestmigrations "github.com/dimmerz92/sittella/database/test_data/sqlite_migrations"
	"github.com/pressly/goose/v3"
)

func constructor(_ *sql.DB) int { return 0 }

func TestMigrate(t *testing.T) {
	t.Run("sqlite migrations", func(t *testing.T) {
		db, err := sqlitedb.NewSQLiteDatabase(sqlitedb.Config{DSN: ":memory:"}, constructor)
		if err != nil {
			t.Fatalf("failed to open database: %v", err)
		}
		defer db.DB().Close()

		err = database.Migrate(t.Context(), goose.DialectSQLite3, db.DB(), sqlitetestmigrations.Migrations)
		if err != nil {
			t.Fatalf("expected migration, got %v", err)
		}

		var out string
		if err := db.DB().QueryRow("SELECT name FROM sqlite_master WHERE name = 'test'").Scan(&out); err != nil {
			t.Fatalf("failed to execute query: %v", err)
		}

		if out != "test" {
			t.Fatalf("expected table 'test', got %s", out)
		}
	})
}
