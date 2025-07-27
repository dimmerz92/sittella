package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type dbtype string

const (
	SQLITE dbtype = "sqlite"
)

var dbtypes = map[dbtype]struct{}{
	SQLITE: {},
}

// Database provides an sqlx embedded database.
type Database struct {
	*sqlx.DB
	dbtype dbtype
}

// New returns a new Database.
func New(dbtype dbtype, driver, dsn string) *Database {
	if _, ok := dbtypes[dbtype]; !ok {
		panic(fmt.Sprintf("database.New: %s is not a supported dbtype", dbtype))
	}
	if driver == "" {
		panic("database.New: driver required")
	}
	if dsn == "" {
		panic("database.New: dsn required")
	}

	db := sqlx.MustConnect(driver, dsn)

	return &Database{DB: db, dbtype: dbtype}
}

// Type returns the database type (i.e., SQLITE, POSTGRES, etc).
func (d *Database) Type() dbtype { return d.dbtype }
