package sqlitedb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dimmerz92/sittella/database"
	_ "modernc.org/sqlite"
)

const (
	DEFAULT_DSN = "data/db.sqlite3?_fk=true&_sync=normal&_timeout=5000&_journal=wal"
	MEMORY_DSN  = "file:memorydb?mode=memory&cache=shared"
)

func New(dsn string) *database.Database {
	if dsn == DEFAULT_DSN {
		if err := os.MkdirAll(filepath.Dir(dsn), 0755); err != nil {
			panic(fmt.Sprintf("sqlite.New: %v", err))
		}
	}

	return database.New(database.SQLITE, "sqlite", dsn)
}
