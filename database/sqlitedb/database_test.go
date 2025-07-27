package sqlitedb_test

import (
	"database/sql/driver"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimmerz92/sittella/database/sqlitedb"
	"github.com/google/uuid"
	"modernc.org/sqlite"
)

func TestSQLiteDatabase(t *testing.T) {
	sqlite.RegisterScalarFunction("uuid", 0, func(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		return uuid.NewString(), nil
	})

	t.Run("in memory", func(t *testing.T) {
		db := sqlitedb.New(sqlitedb.MEMORY_DSN)
		defer db.Close()

		var out string
		if err := db.QueryRowxContext(t.Context(), "SELECT uuid()").Scan(&out); err != nil {
			t.Fatalf("failed to run uuid query: %v", err)
		}

		if _, err := uuid.Parse(out); err != nil {
			t.Fatalf("failed to parse uuid: %v", err)
		}
	})

	t.Run("default dsn", func(t *testing.T) {
		path := strings.Split(sqlitedb.DEFAULT_DSN, "?")[0]

		db := sqlitedb.New(sqlitedb.DEFAULT_DSN)
		defer db.Close()

		if _, err := os.Stat(path); err != nil {
			t.Fatalf("failed to create database: %v", err)
		}

		if err := os.RemoveAll(filepath.Dir(path)); err != nil {
			t.Fatalf("failed to remove created dir: %v", err)
		}
	})
}
