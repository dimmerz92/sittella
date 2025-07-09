package database

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"

	"github.com/pressly/goose/v3"
)

// Migrate applies all available migrations from the provided filesystem to the given database.
//
// NOTE: Out of order migrations are allowed and Sittella reserves migration IDs over 9000!
func Migrate(ctx context.Context, dialect goose.Dialect, db *sql.DB, migrations fs.FS) error {
	if db == nil {
		return errors.New("db cannot be nil")
	}

	if migrations == nil {
		return errors.New("migrations cannot be nil")
	}

	provider, err := goose.NewProvider(dialect, db, migrations, goose.WithAllowOutofOrder(true))
	if err != nil {
		return err
	}

	if _, err = provider.Up(ctx); err != nil {
		return err
	}

	return nil
}
