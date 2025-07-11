package sqlitedb

import (
	"context"
	"database/sql"

	"github.com/dimmerz92/sittella/database"
	"github.com/dimmerz92/sittella/database/sqlitedb/migrations"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

type SQLiteDB[Queries any] database.Database[Queries]

// Config allows configuration of the SQLite database connection and driver behaviour.
type Config struct {
	// DSN is the SQLite connection string.
	DSN string

	// Extensions is a list of shared object paths for SQLite loadable extensions.
	Extensions []string

	// ConnectHook allows custom behaviour during SQLite connection setup.
	// It is invoked after extension loading but before the connection is returned.
	// Note: the "uuid" SQL function is always registered regardless of this hook.
	ConnectHook func(*sqlite3.SQLiteConn) error

	// BeforeStart is an optional lifecycle hook that is executed just before the database is opened.
	BeforeStart func()

	// AfterStart is an optional lifecycle hook that is executed immediately after the database is successfully opened.
	AfterStart func()
}

// NewSQLiteDatabase returns an opened and ready sqlite database.
//
// Uses the mattn/go-sqlite3 driver with a uuid func registered by default.
func NewSQLiteDatabase[DB any, Queries any](config Config, constructor func(DB) Queries) (SQLiteDB[Queries], error) {
	sql.Register("sittella_sqlite3", &sqlite3.SQLiteDriver{
		ConnectHook: func(sc *sqlite3.SQLiteConn) error {
			if config.ConnectHook != nil {
				if err := config.ConnectHook(sc); err != nil {
					return err
				}
			}
			return sc.RegisterFunc("uuid", func() string { return uuid.NewString() }, true)
		},
	})

	if config.BeforeStart != nil {
		config.BeforeStart()
	}

	db, err := database.NewDatabase(config.DSN, "sittella_sqlite3", constructor)
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(context.Background(), goose.DialectSQLite3, db.DB(), migrations.Migrations); err != nil {
		return nil, err
	}

	if config.AfterStart != nil {
		config.AfterStart()
	}

	return db, err
}
