package sqlitedb_test

import (
	"database/sql"
	"testing"

	"github.com/dimmerz92/sittella/database/sqlitedb"
	"github.com/google/uuid"
)

type queries struct {
	db *sql.DB
}

func constructor(db *sql.DB) *queries { return &queries{db} }

func (q *queries) GetUUID() (string, error) {
	var out string
	return out, q.db.QueryRow("SELECT uuid()").Scan(&out)
}

func TestSQLiteDatabase(t *testing.T) {
	var beforeStart, afterStart string
	db, err := sqlitedb.NewSQLiteDatabase(sqlitedb.Config{
		DSN:         ":memory:",
		BeforeStart: func() { beforeStart = "I was set" },
		AfterStart:  func() { afterStart = "I was set" },
	}, constructor)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.DB().Close()

	out, err := db.Queries().GetUUID()
	if err != nil {
		t.Fatalf("expected successful query, got %v", err)
	}

	if _, err := uuid.Parse(out); err != nil {
		t.Fatalf("expected uuid, got %s", out)
	}

	if beforeStart != "I was set" {
		t.Fatal("expected before start hook to run")
	}

	if afterStart != "I was set" {
		t.Fatal("expected before start hook to run")
	}
}
